package k8s

import (
	"slices"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodList returns Kubernetes pods
func (c *Client) PodList(opts metav1.ListOptions) ([]v1.Pod, error) {
	return c.podList(opts, true)
}

// PodListAll preserves replicas so callers can inspect mixed deployments.
func (c *Client) PodListAll(opts metav1.ListOptions) ([]v1.Pod, error) {
	return c.podList(opts, false)
}

func (c *Client) podList(opts metav1.ListOptions, deduplicate bool) ([]v1.Pod, error) {
	var podList []v1.Pod

	for _, ns := range c.namespaces {
		pods, err := c.API.CoreV1().Pods(ns).List(c.ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			if slices.Contains(c.namespacesExcludes, pod.Namespace) {
				continue
			}
			if deduplicate {
				podList = appendPod(podList, pod)
			} else {
				podList = append(podList, pod)
			}
		}
	}

	sort.Slice(podList, func(i, j int) bool {
		return podList[i].Name < podList[j].Name
	})

	return podList, nil
}

func appendPod(pods []v1.Pod, i v1.Pod) []v1.Pod {
	for _, pod := range pods {
		if len(pod.OwnerReferences) > 0 && len(i.OwnerReferences) > 0 && pod.OwnerReferences[0].UID == i.OwnerReferences[0].UID {
			return pods
		}
	}
	return append(pods, i)
}
