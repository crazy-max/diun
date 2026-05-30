package containerd

import (
	"testing"
	"time"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMetadataFormatsContainer(t *testing.T) {
	created := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)
	updated := time.Date(2026, 5, 25, 13, 35, 57, 0, time.UTC)

	got := metadata("default", &containersapi.Container{
		ID: "container-id",
		Labels: map[string]string{
			"nerdctl/name": "redis",
		},
		Image:       "docker.io/library/redis:6.2.3-alpine",
		Runtime:     &containersapi.Container_Runtime{Name: "io.containerd.runc.v2"},
		Snapshotter: "overlayfs",
		SnapshotKey: "container-id",
		CreatedAt:   timestamppb.New(created),
		UpdatedAt:   timestamppb.New(updated),
	}, tasktypes.Status_RUNNING)

	assert.Equal(t, map[string]string{
		"ctn_id":           "container-id",
		"ctn_name":         "redis",
		"ctn_image":        "docker.io/library/redis:6.2.3-alpine",
		"ctn_namespace":    "default",
		"ctn_createdat":    created.String(),
		"ctn_updatedat":    updated.String(),
		"ctn_runtime":      "io.containerd.runc.v2",
		"ctn_snapshotter":  "overlayfs",
		"ctn_snapshot_key": "container-id",
		"ctn_status":       "running",
	}, got)
}

func TestMetadataFallsBackToContainerIDForName(t *testing.T) {
	got := metadata("default", &containersapi.Container{
		ID: "container-id",
	}, tasktypes.Status_UNKNOWN)

	assert.Equal(t, "container-id", got["ctn_name"])
	assert.Equal(t, "", got["ctn_status"])
}

func TestTaskStatusesPrefersRunningStatus(t *testing.T) {
	got := taskStatuses([]*tasktypes.Process{
		{ContainerID: "container-id", Status: tasktypes.Status_STOPPED},
		{ContainerID: "container-id", Status: tasktypes.Status_RUNNING},
		{ID: "fallback-id", Status: tasktypes.Status_RUNNING},
		{ContainerID: "", Status: tasktypes.Status_RUNNING},
	})

	assert.Equal(t, map[string]tasktypes.Status{
		"container-id": tasktypes.Status_RUNNING,
		"fallback-id":  tasktypes.Status_RUNNING,
	}, got)
}

func TestTaskContainerIDFallsBackToProcessID(t *testing.T) {
	assert.Equal(t, "container-id", taskContainerID(&tasktypes.Process{
		ContainerID: "container-id",
		ID:          "process-id",
	}))
	assert.Equal(t, "process-id", taskContainerID(&tasktypes.Process{
		ID: "process-id",
	}))
}
