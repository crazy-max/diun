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
	assert.Equal(t, "1.2.3", images["docker.io/crazymax/diun"][0].Tag)
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

func TestNotifTestWithoutNotifier(t *testing.T) {
	client, _ := newTestClient(t)

	resp, err := client.NotifTest(context.Background(), &pb.NotifTestRequest{})
	require.NoError(t, err)
	assert.Equal(t, "No notifier available", resp.Message)
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
		Created:  &created,
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
