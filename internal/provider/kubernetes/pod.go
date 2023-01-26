package kubernetes

import (
	"reflect"

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
		for _, ctn := range pod.Status.ContainerStatuses {
			c.logger.Debug().
				Str("pod_name", pod.Name).
				Interface("pod_annot", pod.Annotations).
				Str("ctn_name", ctn.Name).
				Str("ctn_image", ctn.Image).
				Msg("Validate image")

			digests := make([]string, 1, 4)
			digests[0] = ctn.ImageID
			image, err := provider.ValidateImageWithDigest(ctn.Image, metadata(pod, ctn), pod.Annotations, *c.config.WatchByDefault, digests)
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

func metadata(pod v1.Pod, ctn v1.ContainerStatus) map[string]string {
	return map[string]string{
		"pod_name":      pod.Name,
		"pod_status":    pod.Status.String(),
		"pod_namespace": pod.Namespace,
		"pod_createdat": pod.CreationTimestamp.String(),
		"ctn_name":      ctn.Name,
		"ctn_command":   "",
		"ctn_names":     pod.Name,
	}
}
