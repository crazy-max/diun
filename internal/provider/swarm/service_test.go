package swarm

import (
	"testing"
	"time"

	mobyswarm "github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestMetadataFormatsService(t *testing.T) {
	created := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)
	updated := time.Date(2026, 5, 25, 13, 35, 57, 0, time.UTC)

	got := metadata(mobyswarm.Service{
		ID: "service-id",
		Meta: mobyswarm.Meta{
			CreatedAt: created,
			UpdatedAt: updated,
		},
		Spec: mobyswarm.ServiceSpec{
			Annotations: mobyswarm.Annotations{
				Name: "api",
			},
		},
	})

	assert.Equal(t, map[string]string{
		"svc_id":        "service-id",
		"svc_createdat": created.String(),
		"svc_updatedat": updated.String(),
		"ctn_name":      "api",
	}, got)
}
