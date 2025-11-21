package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

const (
	noneValue    = "<none>"
	unknownValue = "<unknown>"
)

// FormatPodAge returns the age of the pod in human-readable format.
func FormatPodAge(pod *corev1.Pod) string {
	if pod.CreationTimestamp.IsZero() {
		return unknownValue
	}

	return duration.HumanDuration(time.Since(pod.CreationTimestamp.Time))
}

// FormatPodStatus returns the current status of the pod.
func FormatPodStatus(pod *corev1.Pod) string {
	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	// check for init container issues
	if initReason, initializing := checkInitContainers(pod); initializing {
		reason = initReason
	} else {
		// check main containers
		reason = checkMainContainers(pod, reason)
	}

	// handle pod deletion
	return handlePodDeletion(pod, reason)
}

func checkInitContainers(pod *corev1.Pod) (string, bool) {
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			return formatInitTerminatedReason(&container), true
		case container.State.Waiting != nil &&
			container.State.Waiting.Reason != "" &&
			container.State.Waiting.Reason != "PodInitializing":
			return "Init:" + container.State.Waiting.Reason, true
		default:
			return fmt.Sprintf("Init:%d/%d", i, len(pod.Status.InitContainerStatuses)), true
		}
	}

	return "", false
}

func formatInitTerminatedReason(container *corev1.ContainerStatus) string {
	if container.State.Terminated.Reason == "" {
		if container.State.Terminated.Signal != 0 {
			return fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
		}

		return fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
	}

	return "Init:" + container.State.Terminated.Reason
}

func checkMainContainers(pod *corev1.Pod, currentReason string) string {
	reason := currentReason
	hasRunning := false

	for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
		container := pod.Status.ContainerStatuses[i]

		newReason := getContainerReason(&container)
		if newReason != "" {
			reason = newReason
		}

		if container.Ready && container.State.Running != nil {
			hasRunning = true
		}
	}

	if reason == "Completed" && hasRunning {
		return "Running"
	}

	return reason
}

func getContainerReason(container *corev1.ContainerStatus) string {
	if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
		return container.State.Waiting.Reason
	}

	if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
		return container.State.Terminated.Reason
	}

	if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
		return formatTerminatedReason(container)
	}

	return ""
}

func formatTerminatedReason(container *corev1.ContainerStatus) string {
	if container.State.Terminated.Signal != 0 {
		return fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
	}

	return fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
}

func handlePodDeletion(pod *corev1.Pod, reason string) string {
	if pod.DeletionTimestamp != nil {
		if pod.Status.Reason == "NodeLost" {
			return "Unknown"
		}

		return "Terminating"
	}

	return reason
}

// FormatPodReady returns the number of ready containers out of total containers.
func FormatPodReady(pod *corev1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)
	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Ready {
			readyContainers++
		}
	}

	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// FormatRestarts returns the total number of container restarts in the pod.
func FormatRestarts(pod *corev1.Pod) string {
	restarts := int32(0)
	for i := range pod.Status.ContainerStatuses {
		restarts += pod.Status.ContainerStatuses[i].RestartCount
	}

	return strconv.Itoa(int(restarts))
}

// FormatLabels formats a map of labels into a comma-separated key=value string.
func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return noneValue
	}
	labelStrings := make([]string, 0, len(labels))
	for key, value := range labels {
		labelStrings = append(labelStrings, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(labelStrings, ",")
}

// GetNodeName returns the name of the node where the pod is scheduled.
func GetNodeName(pod *corev1.Pod) string {
	if pod.Spec.NodeName != "" {
		return pod.Spec.NodeName
	}

	return noneValue
}

func makeTableRow(pod *corev1.Pod, ip string, showNamespace, wide, showLabels bool) []any {
	row := []any{}

	if showNamespace {
		row = append(row, pod.Namespace)
	}

	row = append(row, pod.Name, ip, FormatPodStatus(pod))

	if wide {
		row = append(row, FormatPodReady(pod), FormatRestarts(pod), GetNodeName(pod))
	}

	row = append(row, FormatPodAge(pod))

	if showLabels {
		row = append(row, FormatLabels(pod.Labels))
	}

	return row
}

func makeTableHeaders(showNamespace, wide, showLabels bool) []metav1.TableColumnDefinition {
	columns := []metav1.TableColumnDefinition{}

	if showNamespace {
		columns = append(columns, metav1.TableColumnDefinition{
			Name: "NAMESPACE",
			Type: "string",
		})
	}

	columns = append(columns,
		metav1.TableColumnDefinition{
			Name: "NAME",
			Type: "string",
		},
		metav1.TableColumnDefinition{
			Name: "IP",
			Type: "string",
		},
		metav1.TableColumnDefinition{
			Name: "STATUS",
			Type: "string",
		},
	)

	if wide {
		columns = append(columns,
			metav1.TableColumnDefinition{
				Name:     "READY",
				Type:     "string",
				Priority: 1,
			},
			metav1.TableColumnDefinition{
				Name:     "RESTARTS",
				Type:     "string",
				Priority: 1,
			},
			metav1.TableColumnDefinition{
				Name:     "NODE",
				Type:     "string",
				Priority: 1,
			},
		)
	}

	columns = append(columns, metav1.TableColumnDefinition{
		Name: "AGE",
		Type: "string",
	})

	if showLabels {
		columns = append(columns, metav1.TableColumnDefinition{
			Name:     "LABELS",
			Type:     "string",
			Priority: 1,
		})
	}

	return columns
}
