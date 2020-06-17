package k8s

import (
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodList returns Kubernetes pods
func (c *Client) PodList(opts metav1.ListOptions) ([]v1.Pod, error) {
	var podList []v1.Pod

	for _, ns := range c.namespaces {
		pods, err := c.API.CoreV1().Pods(ns).List(c.ctx, opts)
		if err != nil {
			return nil, err
		}
		podList = append(podList, pods.Items...)
	}

	sort.Slice(podList, func(i, j int) bool {
		return podList[i].Name < podList[j].Name
	})

	return podList, nil
}
