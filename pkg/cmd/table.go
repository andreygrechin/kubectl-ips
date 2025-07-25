package cmd

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type podIPWithPod struct {
	pod *corev1.Pod
	ip  string
}

func generateTable(pods *corev1.PodList, showNamespace, wide, showLabels bool) *metav1.Table {
	podIPList := extractPodIPsWithPods(pods)
	sortPodIPsWithPods(podIPList)

	table := &metav1.Table{
		ColumnDefinitions: makeTableHeaders(showNamespace, wide, showLabels),
	}

	for _, item := range podIPList {
		row := metav1.TableRow{
			Cells: makeTableRow(item.pod, item.ip, showNamespace, wide, showLabels),
			Object: runtime.RawExtension{
				Object: item.pod,
			},
		}
		table.Rows = append(table.Rows, row)
	}

	return table
}

func extractPodIPsWithPods(pods *corev1.PodList) []podIPWithPod {
	var podIPs []podIPWithPod
	uniqueIPs := make(map[string]bool)

	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.PodIP != "" {
			podIPs = append(podIPs, podIPWithPod{
				pod: pod,
				ip:  pod.Status.PodIP,
			})
			uniqueIPs[pod.Status.PodIP] = true
		}

		for _, ip := range pod.Status.PodIPs {
			if ip.IP != "" && !uniqueIPs[ip.IP] {
				podIPs = append(podIPs, podIPWithPod{
					pod: pod,
					ip:  ip.IP,
				})
				uniqueIPs[ip.IP] = true
			}
		}
	}

	return podIPs
}

func sortPodIPsWithPods(podIPs []podIPWithPod) {
	sort.Slice(podIPs, func(i, j int) bool {
		if podIPs[i].pod.Namespace != podIPs[j].pod.Namespace {
			return podIPs[i].pod.Namespace < podIPs[j].pod.Namespace
		}
		if podIPs[i].pod.Name != podIPs[j].pod.Name {
			return podIPs[i].pod.Name < podIPs[j].pod.Name
		}

		return podIPs[i].ip < podIPs[j].ip
	})
}
