package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestAppendPodSkipsDuplicateOwnerReference(t *testing.T) {
	pods := []v1.Pod{{
		ObjectMeta: objectMeta("api-1", "replicaset-uid"),
	}}

	got := appendPod(pods, v1.Pod{
		ObjectMeta: objectMeta("api-2", "replicaset-uid"),
	})

	assert.Equal(t, pods, got)
}

func TestAppendPodKeepsPodsWithoutMatchingOwnerReference(t *testing.T) {
	pods := []v1.Pod{{
		ObjectMeta: objectMeta("api-1", "replicaset-uid"),
	}}
	standalone := v1.Pod{}
	standalone.Name = "standalone"

	got := appendPod(pods, standalone)

	assert.Equal(t, []v1.Pod{pods[0], standalone}, got)
}

func objectMeta(name string, uid types.UID) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
		OwnerReferences: []metav1.OwnerReference{
			{UID: uid},
		},
	}
}
