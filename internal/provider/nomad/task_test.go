package nomad

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	nomadapi "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServiceTags(t *testing.T) {
	testCases := []struct {
		input    []string
		expected map[string]string
	}{
		{
			input: []string{
				"noequal",
			},
			expected: map[string]string{},
		},
		{
			input: []string{
				"emptyequal=",
			},
			expected: map[string]string{
				"emptyequal": "",
			},
		},
		{
			input: []string{
				"key=value",
			},
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			input: []string{
				"withequal=a=b",
			},
			expected: map[string]string{
				"withequal": "a=b",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.input[0], func(t *testing.T) {
			result := parseServiceTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNamespaces(t *testing.T) {
	testCases := []struct {
		name     string
		config   *model.PrdNomad
		expected []string
	}{
		{
			name:     "all namespaces by default",
			config:   &model.PrdNomad{},
			expected: []string{nomadapi.AllNamespacesNamespace},
		},
		{
			name: "legacy namespace",
			config: &model.PrdNomad{
				Namespace: "legacy",
			},
			expected: []string{"legacy"},
		},
		{
			name: "namespaces",
			config: &model.PrdNomad{
				Namespace:  "legacy",
				Namespaces: []string{" dev ", "", "prod"},
			},
			expected: []string{"dev", "prod"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{config: tt.config}
			assert.Equal(t, tt.expected, client.namespaces())
		})
	}
}

func TestDeprecatedNamespace(t *testing.T) {
	namespace, ok := (&Client{config: &model.PrdNomad{Namespace: "legacy"}}).deprecatedNamespace()
	assert.True(t, ok)
	assert.Equal(t, "legacy", namespace)

	namespace, ok = (&Client{config: &model.PrdNomad{Namespaces: []string{"dev"}}}).deprecatedNamespace()
	assert.False(t, ok)
	assert.Empty(t, namespace)

	namespace, ok = (&Client{config: &model.PrdNomad{Namespace: "legacy", Namespaces: []string{"dev"}}}).deprecatedNamespace()
	assert.False(t, ok)
	assert.Empty(t, namespace)
}

func TestJobNamespace(t *testing.T) {
	assert.Equal(t, "prod", jobNamespace(&nomadapi.JobListStub{Namespace: "prod"}, nomadapi.AllNamespacesNamespace))
	assert.Equal(t, nomadapi.DefaultNamespace, jobNamespace(&nomadapi.JobListStub{}, nomadapi.AllNamespacesNamespace))
	assert.Equal(t, "dev", jobNamespace(&nomadapi.JobListStub{}, "dev"))
}

func TestListTaskImagesQueriesAllNamespacesByDefault(t *testing.T) {
	var listNamespaces []string
	var infoNamespaces []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/jobs":
			listNamespaces = append(listNamespaces, r.URL.Query().Get("namespace"))
			writeJSON(t, w, []*nomadapi.JobListStub{
				{ID: "api", Name: "api", Namespace: "dev", Status: "running"},
				{ID: "api", Name: "api", Namespace: "prod", Status: "running"},
			})
		case "/v1/job/api":
			namespace := r.URL.Query().Get("namespace")
			infoNamespaces = append(infoNamespaces, namespace)
			writeJSON(t, w, nomadTestJob("registry.example.com/api:"+namespace))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &Client{
		config: &model.PrdNomad{
			Address:        server.URL,
			TLSInsecure:    ptr(false),
			WatchByDefault: ptr(true),
		},
		logger: zerolog.Nop(),
	}

	images := client.listTaskImages()

	require.Len(t, images, 2)
	assert.Equal(t, []string{nomadapi.AllNamespacesNamespace}, listNamespaces)
	assert.ElementsMatch(t, []string{"dev", "prod"}, infoNamespaces)
	assert.ElementsMatch(t, []string{"registry.example.com/api:dev", "registry.example.com/api:prod"}, imageNames(images))
	assert.ElementsMatch(t, []string{"dev", "prod"}, imageMetadataValues(images, "job_namespace"))
}

func TestListTaskImagesQueriesConfiguredNamespaces(t *testing.T) {
	var listNamespaces []string
	var infoNamespaces []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.URL.Query().Get("namespace")

		switch r.URL.Path {
		case "/v1/jobs":
			listNamespaces = append(listNamespaces, namespace)
			writeJSON(t, w, []*nomadapi.JobListStub{
				{ID: "api-" + namespace, Name: "api-" + namespace, Status: "running"},
			})
		case "/v1/job/api-dev", "/v1/job/api-prod":
			infoNamespaces = append(infoNamespaces, namespace)
			writeJSON(t, w, nomadTestJob("registry.example.com/api:"+namespace))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &Client{
		config: &model.PrdNomad{
			Address:        server.URL,
			Namespace:      "legacy",
			Namespaces:     []string{" dev ", "", "prod"},
			TLSInsecure:    ptr(false),
			WatchByDefault: ptr(true),
		},
		logger: zerolog.Nop(),
	}

	images := client.listTaskImages()

	require.Len(t, images, 2)
	assert.Equal(t, []string{"dev", "prod"}, listNamespaces)
	assert.Equal(t, []string{"dev", "prod"}, infoNamespaces)
	assert.ElementsMatch(t, []string{"registry.example.com/api:dev", "registry.example.com/api:prod"}, imageNames(images))
	assert.ElementsMatch(t, []string{"dev", "prod"}, imageMetadataValues(images, "job_namespace"))
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("failed to write JSON response: %v", err)
	}
}

func nomadTestJob(image string) *nomadapi.Job {
	return &nomadapi.Job{
		TaskGroups: []*nomadapi.TaskGroup{
			{
				Name: ptr("group"),
				Tasks: []*nomadapi.Task{
					{
						Name:   "task",
						Driver: "docker",
						Config: map[string]interface{}{
							"image": image,
						},
					},
				},
			},
		},
	}
}

func imageNames(images []model.Image) []string {
	names := make([]string, 0, len(images))
	for _, image := range images {
		names = append(names, image.Name)
	}
	return names
}

func imageMetadataValues(images []model.Image, key string) []string {
	values := make([]string, 0, len(images))
	for _, image := range images {
		values = append(values, image.Metadata[key])
	}
	return values
}

func ptr[T any](v T) *T {
	return &v
}
