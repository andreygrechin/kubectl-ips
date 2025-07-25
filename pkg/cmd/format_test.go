package cmd_test

import (
	"testing"
	"time"

	"github.com/andreygrechin/kubectl-ips/pkg/cmd"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFormatPodAge(t *testing.T) {
	tests := map[string]struct {
		pod      *corev1.Pod
		expected string
	}{
		"pod with creation timestamp": {
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * time.Hour)),
				},
			},
			expected: "120m",
		},
		"pod without creation timestamp": {
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "<unknown>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.FormatPodAge(tc.pod)
			if tc.expected == "<unknown>" {
				assert.Equal(t, tc.expected, result)
			} else {
				assert.Contains(t, result, "m")
			}
		})
	}
}

func TestFormatPodStatus(t *testing.T) {
	tests := map[string]struct {
		pod      *corev1.Pod
		expected string
	}{
		"running pod": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			expected: "Running",
		},
		"pending pod": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			expected: "Pending",
		},
		"succeeded pod": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
			expected: "Succeeded",
		},
		"failed pod": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
			expected: "Failed",
		},
		"terminating pod": {
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			expected: "Terminating",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.FormatPodStatus(tc.pod)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatPodReady(t *testing.T) {
	tests := map[string]struct {
		pod      *corev1.Pod
		expected string
	}{
		"all containers ready": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: true},
					},
				},
			},
			expected: "2/2",
		},
		"some containers ready": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: false},
						{Ready: true},
					},
				},
			},
			expected: "2/3",
		},
		"no containers ready": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: false},
						{Ready: false},
					},
				},
			},
			expected: "0/2",
		},
		"no containers": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{},
			},
			expected: "0/0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.FormatPodReady(tc.pod)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatRestarts(t *testing.T) {
	tests := map[string]struct {
		pod      *corev1.Pod
		expected string
	}{
		"no restarts": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{RestartCount: 0},
						{RestartCount: 0},
					},
				},
			},
			expected: "0",
		},
		"multiple restarts": {
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{RestartCount: 2},
						{RestartCount: 3},
						{RestartCount: 1},
					},
				},
			},
			expected: "6",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.FormatRestarts(tc.pod)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatLabels(t *testing.T) {
	tests := map[string]struct {
		labels   map[string]string
		expected string
	}{
		"no labels": {
			labels:   map[string]string{},
			expected: "<none>",
		},
		"single label": {
			labels:   map[string]string{"app": "nginx"},
			expected: "app=nginx",
		},
		"multiple labels": {
			labels: map[string]string{
				"app":     "nginx",
				"version": "1.0",
			},
			expected: "app=nginx,version=1.0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.FormatLabels(tc.labels)
			if len(tc.labels) > 1 {
				assert.Contains(t, result, "=")
				assert.Contains(t, result, ",")
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetNodeName(t *testing.T) {
	tests := map[string]struct {
		pod      *corev1.Pod
		expected string
	}{
		"pod with node": {
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					NodeName: "worker-1",
				},
			},
			expected: "worker-1",
		},
		"pod without node": {
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			expected: "<none>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := cmd.GetNodeName(tc.pod)
			assert.Equal(t, tc.expected, result)
		})
	}
}
