package k8s

import (
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodList returns Kubernetes pods
func (c *Client) PodList(opts metav1.ListOptions) ([]v1.Pod, error) {
	pods, err := c.API.CoreV1().Pods("").List(c.ctx, opts)
	if err != nil {
		return nil, err
	}

	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})

	return pods.Items, nil
}
