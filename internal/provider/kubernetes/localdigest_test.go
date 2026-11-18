package kubernetes

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestPodDigests(t *testing.T) {
	old, current := digest.FromString("old"), digest.FromString("current")
	cfg := (&model.PrdKubernetes{}).GetDefaults()
	pod := v1.Pod{
		Annotations: map[string]string{"diun.enable": "true"},
		Spec:        v1.PodSpec{Containers: []v1.Container{{Name: "app", Image: "alpine:latest"}}},
		Status: v1.PodStatus{
			Phase:             v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{{Name: "app", ImageID: "docker-pullable://alpine@" + old.String(), State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}}},
		},
	}
	replica := pod.DeepCopy()
	replica.Status.ContainerStatuses[0].ImageID = "docker.io/library/alpine@" + current.String()
	key := "docker.io/library/alpine:latest"
	require.Equal(t, [][]digest.Digest{{old}, {current}}, podDigests([]v1.Pod{pod, *replica}, cfg, nil)[key])
	for _, tc := range []struct {
		name string
		id   string
	}{
		{
			name: "config digest",
			id:   "docker://" + current.String(),
		},
		{
			name: "opaque digest",
			id:   current.String(),
		},
		{
			name: "missing status identity",
			id:   "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			replica.Status.ContainerStatuses[0].ImageID = tc.id
			require.Equal(t, [][]digest.Digest{nil}, podDigests([]v1.Pod{*replica}, cfg, nil)[key])
		})
	}
	replica.Status.ContainerStatuses = nil
	require.Equal(t, [][]digest.Digest{nil}, podDigests([]v1.Pod{*replica}, cfg, nil)[key])
	replica.Status.Phase = v1.PodSucceeded
	require.Empty(t, podDigests([]v1.Pod{*replica}, cfg, nil))
	pod.Annotations["diun.enable"] = "false"
	require.Empty(t, podDigests([]v1.Pod{pod}, cfg, nil))
}
