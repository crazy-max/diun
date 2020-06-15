package kubernetes

import (
	"reflect"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) listPodImage() []model.Image {
	cli, err := k8s.New(k8s.Options{
		Endpoint:    c.config.Endpoint,
		Token:       c.config.Token,
		TokenFile:   c.config.TokenFile,
		TLSCAFile:   c.config.TLSCAFile,
		TLSInsecure: c.config.TLSInsecure,
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot create Kubernetes client")
		return []model.Image{}
	}

	pods, err := cli.PodList(metav1.ListOptions{})
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot list Kubernetes pods")
		return []model.Image{}
	}

	var list []model.Image
	for _, pod := range pods {
		for _, ctn := range pod.Spec.Containers {
			image, err := provider.ValidateContainerImage(ctn.Image, pod.Labels, *c.config.WatchByDefault)
			if err != nil {
				c.logger.Error().Err(err).Msgf("Cannot get image from container %s (pod %s)", ctn.Name, pod.Name)
				continue
			} else if reflect.DeepEqual(image, model.Image{}) {
				c.logger.Debug().Msgf("Watch disabled for container %s (pod %s)", ctn.Name, pod.Name)
				continue
			}
			list = append(list, image)
		}
	}

	return list
}
