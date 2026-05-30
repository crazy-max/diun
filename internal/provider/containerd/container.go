package containerd

import (
	"reflect"
	"strings"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	ctd "github.com/crazy-max/diun/v4/pkg/containerd"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) listContainerImage() []model.Image {
	cli, err := ctd.New(ctd.Options{
		Endpoint: c.config.Endpoint,
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot create containerd client")
		return []model.Image{}
	}
	defer func() {
		if err := cli.Close(); err != nil {
			c.logger.Warn().Err(err).Msg("Cannot close containerd client")
		}
	}()

	var list []model.Image
	for _, namespace := range c.config.Namespaces {
		ctns, err := cli.ContainerList(namespace)
		if err != nil {
			c.logger.Error().Err(err).Str("namespace", namespace).Msg("Cannot list containerd containers")
			continue
		}

		statuses, ok := c.listTaskStatuses(cli, namespace)
		if !ok && !*c.config.WatchStopped {
			continue
		}

		for _, ctn := range ctns {
			imageName := ctn.Image
			if imageName == "" {
				c.logger.Debug().
					Str("namespace", namespace).
					Str("ctn_id", ctn.ID).
					Interface("ctn_labels", ctn.Labels).
					Msg("Skip container without image")
				continue
			}

			status := statuses[ctn.ID]
			if !*c.config.WatchStopped && !isRunningStatus(status) {
				c.logger.Debug().
					Str("namespace", namespace).
					Str("ctn_id", ctn.ID).
					Str("ctn_image", imageName).
					Str("ctn_status", statusString(status)).
					Msg("Skip stopped container")
				continue
			}

			c.logger.Debug().
				Str("namespace", namespace).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Interface("ctn_labels", ctn.Labels).
				Msg("Validate image")
			image, err := provider.ValidateImage(imageName, metadata(namespace, ctn, status), ctn.Labels, *c.config.WatchByDefault, c.defaults)

			if err != nil {
				c.logger.Error().Err(err).
					Str("namespace", namespace).
					Str("ctn_id", ctn.ID).
					Str("ctn_image", imageName).
					Interface("ctn_labels", ctn.Labels).
					Msg("Invalid image")
				continue
			} else if reflect.DeepEqual(image, model.Image{}) {
				c.logger.Debug().
					Str("namespace", namespace).
					Str("ctn_id", ctn.ID).
					Str("ctn_image", imageName).
					Interface("ctn_labels", ctn.Labels).
					Msg("Watch disabled")
				continue
			}

			list = append(list, image)
		}
	}

	return list
}

func (c *Client) listTaskStatuses(cli *ctd.Client, namespace string) (map[string]tasktypes.Status, bool) {
	tasks, err := cli.TaskList(namespace)
	if err != nil {
		c.logger.Error().Err(err).Str("namespace", namespace).Msg("Cannot list containerd tasks")
		return nil, false
	}
	return taskStatuses(tasks), true
}

func taskStatuses(tasks []*tasktypes.Process) map[string]tasktypes.Status {
	statuses := map[string]tasktypes.Status{}
	for _, task := range tasks {
		if task.ContainerID == "" {
			continue
		}
		current := statuses[task.ContainerID]
		if !isRunningStatus(current) || isRunningStatus(task.Status) {
			statuses[task.ContainerID] = task.Status
		}
	}
	return statuses
}

func isRunningStatus(status tasktypes.Status) bool {
	return status == tasktypes.Status_RUNNING
}

func metadata(namespace string, ctn *containersapi.Container, status tasktypes.Status) map[string]string {
	return map[string]string{
		"ctn_id":           ctn.ID,
		"ctn_name":         containerName(ctn),
		"ctn_image":        ctn.Image,
		"ctn_namespace":    namespace,
		"ctn_createdat":    timestampString(ctn.CreatedAt),
		"ctn_updatedat":    timestampString(ctn.UpdatedAt),
		"ctn_runtime":      runtimeName(ctn),
		"ctn_snapshotter":  ctn.Snapshotter,
		"ctn_snapshot_key": ctn.SnapshotKey,
		"ctn_status":       statusString(status),
	}
}

func containerName(ctn *containersapi.Container) string {
	if name := ctn.Labels["nerdctl/name"]; name != "" {
		return name
	}
	return ctn.ID
}

func runtimeName(ctn *containersapi.Container) string {
	if ctn.Runtime == nil {
		return ""
	}
	return ctn.Runtime.Name
}

func timestampString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().String()
}

func statusString(status tasktypes.Status) string {
	if status == tasktypes.Status_UNKNOWN {
		return ""
	}
	return strings.ToLower(status.String())
}
