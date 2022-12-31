package kubernetes

import (
	"reflect"
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) listPodImage() []model.Image {
	cli, err := k8s.New(k8s.Options{
		Endpoint:         c.config.Endpoint,
		Token:            c.config.Token,
		TokenFile:        c.config.TokenFile,
		CertAuthFilePath: c.config.CertAuthFilePath,
		TLSInsecure:      c.config.TLSInsecure,
		Namespaces:       c.config.Namespaces,
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
		// for _, ctn := range pod.Spec.Containers {
		for _, ctn := range pod.Status.ContainerStatuses {
			c.logger.Debug().
				Str("pod_name", pod.Name).
				Interface("pod_annot", pod.Annotations).
				Str("ctn_name", ctn.Name).
				Str("ctn_image", ctn.Image).
				Msg("Validate image")

			image, err := provider.ValidateImageWithDigest(ctn.Image, metadata(pod, ctn), pod.Annotations, *c.config.WatchByDefault, ctn.ImageID)

			if err != nil {
				c.logger.Error().Err(err).
					Str("pod_name", pod.Name).
					Interface("pod_annot", pod.Annotations).
					Str("ctn_name", ctn.Name).
					Str("ctn_image", ctn.Image).
					Msg("Invalid image")
				continue
			} else if reflect.DeepEqual(image, model.Image{}) {
				c.logger.Debug().
					Str("pod_name", pod.Name).
					Interface("pod_annot", pod.Annotations).
					Str("ctn_name", ctn.Name).
					Str("ctn_image", ctn.Image).
					Msg("Watch disabled")
				continue
			}

			list = append(list, image)
		}
	}

	return list
}

func metadata(pod v1.Pod, ctn v1.Container) map[string]string {
	return map[string]string{
		"pod_name":      pod.Name,
		"pod_status":    pod.Status.String(),
		"pod_namespace": pod.Namespace,
		"pod_createdat": pod.CreationTimestamp.String(),
		"ctn_name":      ctn.Name,
		"ctn_command":   strings.Join(ctn.Command, " "),
	}
}
