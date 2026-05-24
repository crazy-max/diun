package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	digest "github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	podmanmanifest "go.podman.io/image/v5/manifest"
)

type testRegistry struct {
	t *testing.T

	repo   string
	server *httptest.Server

	mu          sync.Mutex
	requests    map[string]int
	tags        map[string][]string
	tagLinks    map[string]string
	manifests   map[string]testRegistryBlob
	configBlobs map[string][]byte
}

type testRegistryBlob struct {
	mediaType string
	body      []byte
	digest    digest.Digest
}

type testRegistryImage struct {
	manifest testRegistryBlob
	config   testRegistryBlob
	layer    digest.Digest
}

func newTestRegistry(t *testing.T, repo string) *testRegistry {
	t.Helper()

	r := &testRegistry{
		t:           t,
		repo:        repo,
		requests:    map[string]int{},
		tags:        map[string][]string{},
		tagLinks:    map[string]string{},
		manifests:   map[string]testRegistryBlob{},
		configBlobs: map[string][]byte{},
	}
	r.server = httptest.NewTLSServer(http.HandlerFunc(r.handle))
	t.Cleanup(r.server.Close)

	return r
}

func (r *testRegistry) host() string {
	return strings.TrimPrefix(r.server.URL, "https://")
}

func (r *testRegistry) imageName(ref string) string {
	return fmt.Sprintf("%s/%s:%s", r.host(), r.repo, ref)
}

func (r *testRegistry) addTagsPage(page string, tags []string, link string) {
	r.tags[page] = tags
	r.tagLinks[page] = link
}

func (r *testRegistry) addImage(ref string, image testRegistryImage) {
	r.addManifest(ref, image.manifest)
	r.configBlobs[image.config.digest.String()] = image.config.body
}

func (r *testRegistry) addManifest(ref string, manifest testRegistryBlob) {
	r.manifests[ref] = manifest
	r.manifests[manifest.digest.String()] = manifest
}

func (r *testRegistry) requestCount(method, path string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.requests[method+" "+path]
}

func (r *testRegistry) handle(w http.ResponseWriter, req *http.Request) {
	r.record(req)

	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/v2/":
		w.WriteHeader(http.StatusOK)
	case req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/v2/%s/tags/list", r.repo):
		r.handleTags(w, req)
	case strings.HasPrefix(req.URL.Path, fmt.Sprintf("/v2/%s/manifests/", r.repo)):
		r.handleManifest(w, req)
	case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, fmt.Sprintf("/v2/%s/blobs/", r.repo)):
		r.handleBlob(w, req)
	default:
		r.t.Errorf("unexpected registry request: %s %s", req.Method, req.URL.RequestURI())
		http.NotFound(w, req)
	}
}

func (r *testRegistry) record(req *http.Request) {
	key := req.Method + " " + req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[key]++
}

func (r *testRegistry) handleTags(w http.ResponseWriter, req *http.Request) {
	page := req.URL.Query().Get("page")
	tags, ok := r.tags[page]
	if !ok {
		http.NotFound(w, req)
		return
	}
	if link := r.tagLinks[page]; link != "" {
		w.Header().Set("Link", link)
	}
	w.Header().Set("Content-Type", "application/json")
	require.NoError(r.t, json.NewEncoder(w).Encode(map[string]any{
		"name": r.repo,
		"tags": tags,
	}))
}

func (r *testRegistry) handleManifest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodHead && req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ref := strings.TrimPrefix(req.URL.Path, fmt.Sprintf("/v2/%s/manifests/", r.repo))
	manifest, ok := r.manifests[ref]
	if !ok {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Docker-Content-Digest", manifest.digest.String())
	w.Header().Set("Content-Type", manifest.mediaType)
	if req.Method == http.MethodHead {
		return
	}
	_, err := w.Write(manifest.body)
	require.NoError(r.t, err)
}

func (r *testRegistry) handleBlob(w http.ResponseWriter, req *http.Request) {
	ref := strings.TrimPrefix(req.URL.Path, fmt.Sprintf("/v2/%s/blobs/", r.repo))
	blob, ok := r.configBlobs[ref]
	if !ok {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err := w.Write(blob)
	require.NoError(r.t, err)
}

func newTestRegistryClient(t *testing.T, opts Options) *Client {
	t.Helper()

	opts.InsecureTLS = true
	if opts.ImageOs == "" {
		opts.ImageOs = "linux"
	}
	if opts.ImageArch == "" {
		opts.ImageArch = "amd64"
	}

	client, err := New(opts)
	require.NoError(t, err)
	return client
}

func newTestRegistryImage(t *testing.T, platform imgspecv1.Platform, dockerVersion string, labels map[string]string) testRegistryImage {
	t.Helper()

	configBody, err := json.Marshal(struct {
		Created       string `json:"created"`
		DockerVersion string `json:"docker_version,omitempty"`
		Architecture  string `json:"architecture"`
		Variant       string `json:"variant,omitempty"`
		OS            string `json:"os"`
		Config        *struct {
			Labels map[string]string `json:"Labels,omitempty"`
		} `json:"config,omitempty"`
	}{
		Created:       "2026-05-24T00:00:00Z",
		DockerVersion: dockerVersion,
		Architecture:  platform.Architecture,
		Variant:       platform.Variant,
		OS:            platform.OS,
		Config: &struct {
			Labels map[string]string `json:"Labels,omitempty"`
		}{
			Labels: labels,
		},
	})
	require.NoError(t, err)

	config := testRegistryBlob{
		mediaType: podmanmanifest.DockerV2Schema2ConfigMediaType,
		body:      configBody,
		digest:    digest.FromBytes(configBody),
	}
	layer := digest.FromString(fmt.Sprintf("%s/%s/%s-layer", platform.OS, platform.Architecture, platform.Variant))

	manifestBody, err := json.Marshal(struct {
		SchemaVersion int                      `json:"schemaVersion"`
		MediaType     string                   `json:"mediaType"`
		Config        testRegistryDescriptor   `json:"config"`
		Layers        []testRegistryDescriptor `json:"layers"`
	}{
		SchemaVersion: 2,
		MediaType:     podmanmanifest.DockerV2Schema2MediaType,
		Config: testRegistryDescriptor{
			MediaType: podmanmanifest.DockerV2Schema2ConfigMediaType,
			Size:      int64(len(configBody)),
			Digest:    config.digest,
		},
		Layers: []testRegistryDescriptor{
			{
				MediaType: podmanmanifest.DockerV2Schema2LayerMediaType,
				Size:      42,
				Digest:    layer,
			},
		},
	})
	require.NoError(t, err)

	return testRegistryImage{
		manifest: testRegistryBlob{
			mediaType: podmanmanifest.DockerV2Schema2MediaType,
			body:      manifestBody,
			digest:    digest.FromBytes(manifestBody),
		},
		config: config,
		layer:  layer,
	}
}

type testRegistryDescriptor struct {
	MediaType string        `json:"mediaType"`
	Size      int64         `json:"size"`
	Digest    digest.Digest `json:"digest"`
}

func newTestManifestList(t *testing.T, instances ...testManifestListInstance) testRegistryBlob {
	t.Helper()

	descriptors := make([]testManifestListDescriptor, 0, len(instances))
	for _, instance := range instances {
		descriptors = append(descriptors, testManifestListDescriptor{
			MediaType: instance.manifest.mediaType,
			Size:      int64(len(instance.manifest.body)),
			Digest:    instance.manifest.digest,
			Platform:  instance.platform,
		})
	}

	body, err := json.Marshal(struct {
		SchemaVersion int                          `json:"schemaVersion"`
		MediaType     string                       `json:"mediaType"`
		Manifests     []testManifestListDescriptor `json:"manifests"`
	}{
		SchemaVersion: 2,
		MediaType:     podmanmanifest.DockerV2ListMediaType,
		Manifests:     descriptors,
	})
	require.NoError(t, err)

	return testRegistryBlob{
		mediaType: podmanmanifest.DockerV2ListMediaType,
		body:      body,
		digest:    digest.FromBytes(body),
	}
}

type testManifestListInstance struct {
	manifest testRegistryBlob
	platform imgspecv1.Platform
}

type testManifestListDescriptor struct {
	MediaType string             `json:"mediaType"`
	Size      int64              `json:"size"`
	Digest    digest.Digest      `json:"digest"`
	Platform  imgspecv1.Platform `json:"platform"`
}
