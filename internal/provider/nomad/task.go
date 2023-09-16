package nomad

import (
	"reflect"
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/imdario/mergo"
)

func parseServiceTags(tags []string) map[string]string {
	labels := map[string]string{}

	for _, tag := range tags {
		tagParts := strings.SplitN(tag, "=", 2)
		if len(tagParts) < 2 {
			continue
		}

		labels[tagParts[0]] = tagParts[1]
	}

	return labels
}

func (c *Client) listTaskImages() []model.Image {
	config := &nomad.Config{
		Address:   c.config.Address,
		Region:    c.config.Region,
		SecretID:  c.config.SecretID,
		Namespace: c.config.Namespace,
	}

	if *c.config.TLSInsecure {
		config.TLSConfig.Insecure = true
	}

	client, err := nomad.NewClient(config)
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot create Nomad client")
		return []model.Image{}
	}

	jobs, _, err := client.Jobs().List(nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot list Nomad jobs")
	}

	var list []model.Image

	for _, job := range jobs {
		jobInfo, _, err := client.Jobs().Info(job.ID, nil)
		if err != nil {
			c.logger.Error().Err(err).Msg("Cannot get info for job")
		}

		for _, taskGroup := range jobInfo.TaskGroups {
			// Get task group service labels
			groupLabels := map[string]string{}
			if err := mergo.Merge(&groupLabels, taskGroup.Meta, mergo.WithOverride); err != nil {
				c.logger.Error().Err(err).Msg("Cannot merge group labels and meta")
			}

			for _, service := range taskGroup.Services {
				if err := mergo.Merge(&groupLabels, parseServiceTags(service.Tags), mergo.WithOverride); err != nil {
					c.logger.Error().Err(err).Msg("Cannot merge group service labels")
				}
			}

			for _, task := range taskGroup.Tasks {
				if task.Driver != "docker" {
					continue
				}

				if taskImage, ok := task.Config["image"]; ok {
					imageName := taskImage.(string)
					if imageName == "${meta.connect.sidecar_image}" {
						c.logger.Debug().
							Str("job_id", job.ID).
							Str("task_group", *taskGroup.Name).
							Str("task_name", task.Name).
							Msg("Skipping connect sidecar")
						continue
					}

					// Get task service labels
					labels := map[string]string{}
					if err := mergo.Merge(&labels, groupLabels, mergo.WithOverride); err != nil {
						c.logger.Error().Err(err).Msg("Cannot merge task service with group service labels")
					}
					for _, service := range task.Services {
						if err := mergo.Merge(&labels, parseServiceTags(service.Tags), mergo.WithOverride); err != nil {
							c.logger.Error().Err(err).Msg("Cannot merge task service labels")
						}
					}

					// Finally, merge task meta values
					if err := mergo.Merge(&labels, task.Meta, mergo.WithOverride); err != nil {
						c.logger.Error().Err(err).Msg("Cannot merge task metadata")
					}

					image, err := provider.ValidateImage(imageName, metadata(job, taskGroup, task), labels, *c.config.WatchByDefault, *c.imageDefaults)
					if err != nil {
						c.logger.Error().
							Err(err).
							Str("job_id", job.ID).
							Str("task_group", *taskGroup.Name).
							Str("task_name", task.Name).
							Str("image_name", imageName).
							Msg("Error validating image")
						continue
					} else if reflect.DeepEqual(image, model.Image{}) {
						c.logger.Debug().
							Str("job_id", job.ID).
							Str("task_group", *taskGroup.Name).
							Str("task_name", task.Name).
							Str("image_name", imageName).
							Msg("Watch disabled")
						continue
					}

					list = append(list, image)
				}
			}
		}
	}

	return list
}

func metadata(job *nomad.JobListStub, taskGroup *nomad.TaskGroup, task *nomad.Task) map[string]string {
	return map[string]string{
		"job_id":         job.ID,
		"job_name":       job.Name,
		"job_status":     job.Status,
		"job_namespace":  job.Namespace,
		"taskgroup_name": *taskGroup.Name,
		"task_name":      task.Name,
		"task_driver":    task.Driver,
		"task_user":      task.User,
	}
}
