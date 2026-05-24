package kubernetes

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestMetadataFormatsPodContainer(t *testing.T) {
	created := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "api-pod",
			Namespace:         "prod",
			CreationTimestamp: metav1.NewTime(created),
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	ctn := v1.Container{
		Name:    "api",
		Command: []string{"diun", "serve"},
	}

	got := metadata(pod, ctn)

	assert.Equal(t, map[string]string{
		"pod_name":      "api-pod",
		"pod_status":    pod.Status.String(),
		"pod_namespace": "prod",
		"pod_createdat": metav1.NewTime(created).String(),
		"ctn_name":      "api",
		"ctn_command":   "diun serve",
	}, got)
}
