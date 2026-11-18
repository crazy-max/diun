package k8s

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestPodListAllPreservesReplicas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/pods", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v1.PodList{Items: []v1.Pod{
			{
				ObjectMeta: objectMeta("api-1", "replicaset"),
			},
			{
				ObjectMeta: objectMeta("api-2", "replicaset"),
			},
			{
				Name: "excluded", Namespace: "excluded",
			},
		}})
	}))
	defer server.Close()
	t.Setenv("KUBECONFIG", "")
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	cli, err := NewWithContext(context.Background(), Options{Endpoint: server.URL, Namespaces: []string{"!excluded"}})
	require.NoError(t, err)
	pods, err := cli.PodListAll(metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, pods, 2)
	pods, err = cli.PodList(metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, pods, 1)
}

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
	standalone := v1.Pod{
		Name: "standalone",
	}

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
