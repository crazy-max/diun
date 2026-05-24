package msg

import (
	"encoding/json"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdown(t *testing.T) {
	client := newTestClient(t, Options{
		TemplateTitle: "\n{{ upper .Entry.Image.Tag }}\n",
		TemplateBody:  "\n{{ .Meta.Hostname }} saw {{ .Entry.Image }} from {{ .Entry.Provider }}\n",
		TemplateFuncs: template.FuncMap{
			"upper": strings.ToUpper,
		},
	})

	title, body, err := client.RenderMarkdown()
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", string(title))
	assert.Equal(t, "node-1 saw docker.io/crazymax/diun:1.2.3 from file", string(body))
}

func TestRenderMarkdownReturnsTemplateErrors(t *testing.T) {
	client := newTestClient(t, Options{
		TemplateTitle: "{{ unknown .Entry.Image }}",
		TemplateBody:  "body",
	})

	_, _, err := client.RenderMarkdown()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse title template")
}

func TestRenderHTMLPreservesLeadingPAndSanitizesBody(t *testing.T) {
	client := newTestClient(t, Options{
		TemplateTitle: "title",
		TemplateBody:  "patched <script>alert(1)</script>",
	})

	title, body, err := client.RenderHTML()
	require.NoError(t, err)

	assert.Equal(t, "title", string(title))
	assert.Contains(t, string(body), "patched")
	assert.NotContains(t, string(body), "<script>")
}

func TestRenderJSON(t *testing.T) {
	client := newTestClient(t, Options{})

	body, err := client.RenderJSON()
	require.NoError(t, err)

	var payload struct {
		Version  string            `json:"diun_version"`
		Hostname string            `json:"hostname"`
		Status   string            `json:"status"`
		Provider string            `json:"provider"`
		Image    string            `json:"image"`
		HubLink  string            `json:"hub_link"`
		MIMEType string            `json:"mime_type"`
		Digest   digest.Digest     `json:"digest"`
		Created  *time.Time        `json:"created"`
		Platform string            `json:"platform"`
		Metadata map[string]string `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	assert.Equal(t, "4.0.0", payload.Version)
	assert.Equal(t, "node-1", payload.Hostname)
	assert.Equal(t, "update", payload.Status)
	assert.Equal(t, "file", payload.Provider)
	assert.Equal(t, "docker.io/crazymax/diun:1.2.3", payload.Image)
	assert.Equal(t, "https://hub.docker.com/r/crazymax/diun", payload.HubLink)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", payload.MIMEType)
	assert.Equal(t, digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"), payload.Digest)
	assert.Equal(t, "linux/amd64", payload.Platform)
	assert.Equal(t, map[string]string{"owner": "ops", "ticket": "DIUN-123"}, payload.Metadata)
	require.NotNil(t, payload.Created)
	assert.Equal(t, time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC), *payload.Created)
}

func TestRenderEnv(t *testing.T) {
	client := newTestClient(t, Options{})

	assert.ElementsMatch(t, []string{
		"DIUN_VERSION=4.0.0",
		"DIUN_HOSTNAME=node-1",
		"DIUN_ENTRY_STATUS=update",
		"DIUN_ENTRY_PROVIDER=file",
		"DIUN_ENTRY_IMAGE=docker.io/crazymax/diun:1.2.3",
		"DIUN_ENTRY_HUBLINK=https://hub.docker.com/r/crazymax/diun",
		"DIUN_ENTRY_MIMETYPE=application/vnd.docker.distribution.manifest.v2+json",
		"DIUN_ENTRY_DIGEST=sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"DIUN_ENTRY_CREATED=2026-05-24 12:34:56 +0000 UTC",
		"DIUN_ENTRY_PLATFORM=linux/amd64",
		"DIUN_ENTRY_METADATA_OWNER=ops",
		"DIUN_ENTRY_METADATA_TICKET=DIUN-123",
	}, client.RenderEnv())
}

func newTestClient(t *testing.T, overrides Options) *Client {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:1.2.3",
	})
	require.NoError(t, err)
	image.HubLink = "https://hub.docker.com/r/crazymax/diun"

	opts := Options{
		Meta: model.Meta{
			Version:  "4.0.0",
			Hostname: "node-1",
		},
		Entry: model.NotifEntry{
			Status:   model.ImageStatusUpdate,
			Provider: "file",
			Image:    image,
			Manifest: registry.Manifest{
				MIMEType: "application/vnd.docker.distribution.manifest.v2+json",
				Digest:   "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Created:  new(time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)),
				Platform: "linux/amd64",
			},
			Metadata: map[string]string{
				"owner":  "ops",
				"ticket": "DIUN-123",
			},
		},
		TemplateTitle: "{{ .Entry.Image }}",
		TemplateBody:  "{{ .Entry.Provider }}",
	}

	if overrides.TemplateTitle != "" {
		opts.TemplateTitle = overrides.TemplateTitle
	}
	if overrides.TemplateBody != "" {
		opts.TemplateBody = overrides.TemplateBody
	}
	if overrides.TemplateFuncs != nil {
		opts.TemplateFuncs = overrides.TemplateFuncs
	}

	client, err := New(opts)
	require.NoError(t, err)
	return client
}
