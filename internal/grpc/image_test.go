package grpc

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	"github.com/crazy-max/diun/v4/pb"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestImageListAndInspect(t *testing.T) {
	client, _ := newTestClient(t)
	older := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	seedManifest(t, client, "crazymax/diun:1.2.0", older)
	seedManifest(t, client, "crazymax/diun:1.2.3", newer)
	seedManifest(t, client, "alpine:3.20", older)

	list, err := client.ImageList(context.Background(), &pb.ImageListRequest{})
	require.NoError(t, err)
	require.Len(t, list.Images, 2)

	diun := findImageListEntry(t, list.Images, "docker.io/crazymax/diun")
	assert.Equal(t, int64(2), diun.ManifestsCount)
	require.NotNil(t, diun.Latest)
	assert.Equal(t, "1.2.3", diun.Latest.Tag)
	assert.Equal(t, newer, diun.Latest.Created.AsTime())

	inspect, err := client.ImageInspect(context.Background(), &pb.ImageInspectRequest{
		Name: "crazymax/diun",
	})
	require.NoError(t, err)
	require.NotNil(t, inspect.Image)
	assert.Equal(t, "docker.io/crazymax/diun", inspect.Image.Name)
	assert.ElementsMatch(t, []string{"1.2.0", "1.2.3"}, manifestTags(inspect.Image.Manifests))
}

func TestImageInspectReturnsNotFound(t *testing.T) {
	client, _ := newTestClient(t)

	_, err := client.ImageInspect(context.Background(), &pb.ImageInspectRequest{
		Name: "missing/image",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker.io/missing/image not found in database")
}

func TestImageRemoveByTag(t *testing.T) {
	client, dbClient := newTestClient(t)
	created := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	seedManifest(t, client, "crazymax/diun:1.2.0", created)
	seedManifest(t, client, "crazymax/diun:1.2.3", created)

	removed, err := client.ImageRemove(context.Background(), &pb.ImageRemoveRequest{
		Name: "crazymax/diun:1.2.0",
	})
	require.NoError(t, err)
	require.Len(t, removed.Manifests, 1)
	assert.Equal(t, "1.2.0", removed.Manifests[0].Tag)
	assert.Positive(t, removed.Manifests[0].Size)

	images, err := dbClient.ListImage()
	require.NoError(t, err)
	require.Len(t, images["docker.io/crazymax/diun"], 1)
	assert.Equal(t, "1.2.3", images["docker.io/crazymax/diun"][0].Manifest.Tag)
}

func TestImageRemoveWithoutTagRemovesAllImageManifests(t *testing.T) {
	client, dbClient := newTestClient(t)
	created := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	seedManifest(t, client, "crazymax/diun:1.2.0", created)
	seedManifest(t, client, "crazymax/diun:1.2.3", created)
	seedManifest(t, client, "alpine:3.20", created)

	removed, err := client.ImageRemove(context.Background(), &pb.ImageRemoveRequest{
		Name: "crazymax/diun",
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.2.0", "1.2.3"}, manifestTags(removed.Manifests))

	images, err := dbClient.ListImage()
	require.NoError(t, err)
	assert.NotContains(t, images, "docker.io/crazymax/diun")
	assert.Contains(t, images, "docker.io/library/alpine")
}

func TestImagePrune(t *testing.T) {
	client, dbClient := newTestClient(t)
	created := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	seedManifest(t, client, "crazymax/diun:1.2.3", created)
	seedManifest(t, client, "alpine:3.20", created)

	removed, err := client.ImagePrune(context.Background(), &pb.ImagePruneRequest{})
	require.NoError(t, err)
	require.Len(t, removed.Images, 2)

	manifests, err := dbClient.ListManifest()
	require.NoError(t, err)
	assert.Empty(t, manifests)
}

func TestImageRemovalUsesStoredReferences(t *testing.T) {
	old := digest.FromString("deployed").String()
	other := digest.FromString("another deployment").String()
	refs := []string{
		"alpine:latest",
		"alpine:latest@" + old,
		"alpine:latest@" + other,
		"alpine@" + old,
		"alpine:latest-other",
		"alpine2:latest",
	}
	for _, tt := range []struct {
		name      string
		request   string
		prune     bool
		remaining []string
	}{
		{
			name:      "repository",
			request:   "alpine",
			remaining: refs[5:],
		},
		{
			name:      "tag and its aliases",
			request:   "alpine:latest",
			remaining: refs[3:],
		},
		{
			name:      "exact tag plus digest",
			request:   "alpine:latest@" + old,
			remaining: []string{refs[0], refs[2], refs[3], refs[4], refs[5]},
		},
		{
			name:      "exact digest only",
			request:   "alpine@" + old,
			remaining: []string{refs[0], refs[1], refs[2], refs[4], refs[5]},
		},
		{
			name:      "missing digest does not remove siblings",
			request:   "alpine@" + digest.FromString("missing").String(),
			remaining: refs,
		},
		{
			name:  "prune",
			prune: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, store := newTestClient(t)
			created := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
			for _, ref := range refs {
				manifest := seedManifest(t, client, ref, created)
				manifest.Tag = "latest"
				manifest.Digest = digest.FromString("registry now")
				image, err := registry.ParseImage(registry.ParseImageOptions{Name: ref})
				require.NoError(t, err)
				require.NoError(t, store.PutManifest(image, manifest))
			}
			var removed []*pb.Manifest
			if tt.prune {
				response, err := client.ImagePrune(t.Context(), &pb.ImagePruneRequest{})
				require.NoError(t, err)
				for _, image := range response.Images {
					removed = append(removed, image.Manifests...)
				}
			} else {
				response, err := client.ImageRemove(t.Context(), &pb.ImageRemoveRequest{Name: tt.request})
				require.NoError(t, err)
				removed = response.Manifests
			}
			require.Len(t, removed, len(refs)-len(tt.remaining))
			for _, manifest := range removed {
				require.Equal(t, digest.FromString("registry now").String(), manifest.Digest)
			}
			images, err := store.ListImage()
			require.NoError(t, err)
			var remaining []string
			for _, entries := range images {
				for _, entry := range entries {
					remaining = append(remaining, entry.Reference)
				}
			}
			var want []string
			for _, ref := range tt.remaining {
				image, err := registry.ParseImage(registry.ParseImageOptions{Name: ref})
				require.NoError(t, err)
				want = append(want, image.String())
			}
			require.ElementsMatch(t, want, remaining)
		})
	}
}

func TestNotifTestWithoutNotifier(t *testing.T) {
	client, _ := newTestClient(t)

	resp, err := client.NotifTest(context.Background(), &pb.NotifTestRequest{})
	require.NoError(t, err)
	assert.Equal(t, "No notifier available", resp.Message)
}

func TestHealthServiceDefaults(t *testing.T) {
	client, _ := newTestClient(t)

	overall, err := client.health.Check(context.Background(), &healthpb.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, healthpb.HealthCheckResponse_NOT_SERVING, overall.Status)

	scheduler, err := client.health.Check(context.Background(), &healthpb.HealthCheckRequest{
		Service: HealthServiceScheduler,
	})
	require.NoError(t, err)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVICE_UNKNOWN, scheduler.Status)
}

func newTestClient(t *testing.T) (*Client, *db.Client) {
	t.Helper()

	dbClient, err := db.New(model.Db{Path: filepath.Join(t.TempDir(), "diun.db")})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dbClient.Close())
	})

	notifClient, err := notif.New(nil, model.Meta{})
	require.NoError(t, err)

	client, err := New("127.0.0.1:0", dbClient, notifClient)
	require.NoError(t, err)
	return client, dbClient
}

func seedManifest(t *testing.T, client *Client, imageName string, created time.Time) registry.Manifest {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{Name: imageName})
	require.NoError(t, err)

	manifest := registry.Manifest{
		Name:     image.Name(),
		Tag:      image.Tag,
		MIMEType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:   digest.FromString(image.String()),
		Created:  new(created),
		Labels: map[string]string{
			"org.opencontainers.image.title": image.Name(),
		},
		Platform: "linux/amd64",
	}
	require.NoError(t, client.db.PutManifest(image, manifest))
	return manifest
}

func findImageListEntry(t *testing.T, images []*pb.ImageListResponse_Image, name string) *pb.ImageListResponse_Image {
	t.Helper()

	for _, image := range images {
		if image.Name == name {
			return image
		}
	}
	t.Fatalf("image %q not found", name)
	return nil
}

func manifestTags(manifests []*pb.Manifest) []string {
	tags := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		tags = append(tags, manifest.Tag)
	}
	return tags
}
