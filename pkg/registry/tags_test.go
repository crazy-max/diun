package registry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	regconfig "github.com/regclient/regclient/config"
	regplatform "github.com/regclient/regclient/types/platform"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTagsWithBearerPagination(t *testing.T) {
	var issuedToken bool
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/token":
			issuedToken = true
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"token":"test-token","expires_in":300}`))
		case "/v2/test/app/tags/list":
			if req.URL.RawQuery == "n=2&last=v2" {
				if req.Header.Get("Authorization") != "Bearer test-token" {
					rw.WriteHeader(http.StatusUnauthorized)
					return
				}
				rw.Header().Set("Content-Type", "application/json")
				_, _ = rw.Write([]byte(`{"name":"test/app","tags":["v3"]}`))
				return
			}
			if req.Header.Get("Authorization") != "Bearer test-token" {
				rw.Header().Set("Www-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry.test"`, server.URL))
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Add("Link", `</v2/test/app/tags/list?n=2&last=v2>; rel="next"`)
			_, _ = rw.Write([]byte(`{"name":"test/app","tags":["v1","v2"]}`))
		case "/v2/":
			rw.WriteHeader(http.StatusOK)
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	imageName := strings.TrimPrefix(server.URL, "https://") + "/test/app:latest"
	img, err := ParseImage(ParseImageOptions{Name: imageName})
	require.NoError(t, err)

	rc := New(Options{
		Host: &regconfig.Host{
			Name: img.Domain,
			TLS:  regconfig.TLSInsecure,
		},
		Platform: regplatform.Local(),
	})

	tags, err := rc.Tags(TagsOptions{Image: img})
	require.NoError(t, err)
	require.True(t, issuedToken)
	require.Equal(t, []string{"v1", "v2", "v3"}, tags.List)
	require.Equal(t, 3, tags.Total)
}

func TestTagsAppliesIncludeExcludeAndMax(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/v2/test/app/tags/list":
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"name":"test/app","tags":["old","v1","v2","keep","v3"]}`))
		case "/v2/":
			rw.WriteHeader(http.StatusOK)
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	imageName := strings.TrimPrefix(server.URL, "https://") + "/test/app:latest"
	img, err := ParseImage(ParseImageOptions{Name: imageName})
	require.NoError(t, err)

	rc := New(Options{
		Host: &regconfig.Host{
			Name: img.Domain,
			TLS:  regconfig.TLSInsecure,
		},
		Platform: regplatform.Local(),
	})

	tags, err := rc.Tags(TagsOptions{
		Image:   img,
		Max:     2,
		Include: []string{`^v`, `^keep$`},
		Exclude: []string{`^v2$`},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"v1", "keep"}, tags.List)
	require.Equal(t, 5, tags.Total)
	require.Equal(t, 1, tags.NotIncluded)
	require.Equal(t, 1, tags.Excluded)
}

func TestTagsRetries429WithLegacyBackoffDelay(t *testing.T) {
	var tagListRequests int

	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/v2/test/app/tags/list":
			tagListRequests++
			if tagListRequests == 1 {
				rw.WriteHeader(http.StatusTooManyRequests)
				_, _ = rw.Write([]byte(`{"errors":[{"code":"TOOMANYREQUESTS","message":"rate limited"}]}`))
				return
			}
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"name":"test/app","tags":["v1"]}`))
		case "/v2/":
			rw.WriteHeader(http.StatusOK)
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	imageName := strings.TrimPrefix(server.URL, "https://") + "/test/app:latest"
	img, err := ParseImage(ParseImageOptions{Name: imageName})
	require.NoError(t, err)

	rc := New(Options{
		Host: &regconfig.Host{
			Name: img.Domain,
			TLS:  regconfig.TLSInsecure,
		},
		Platform: regplatform.Local(),
	})

	start := time.Now()
	tags, err := rc.Tags(TagsOptions{Image: img})
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Equal(t, []string{"v1"}, tags.List)
	require.Equal(t, 2, tagListRequests)
	require.GreaterOrEqual(t, elapsed, 1500*time.Millisecond)
}

func TestTags(t *testing.T) {
	client := New(Options{
		Host: regconfig.HostNewName("docker.io"),
		Platform: regplatform.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	})

	image, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:3.0.0",
	})
	require.NoError(t, err)

	tags, err := client.Tags(TagsOptions{
		Image: image,
	})
	require.NoError(t, err)

	require.Greater(t, tags.Total, 0)
	require.Greater(t, len(tags.List), 0)
}

func TestTagsWithDigest(t *testing.T) {
	client := New(Options{
		Host: regconfig.HostNewName("docker.io"),
		Platform: regplatform.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	})

	image, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:latest@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	})
	require.NoError(t, err)

	tags, err := client.Tags(TagsOptions{
		Image: image,
	})
	require.NoError(t, err)

	require.Greater(t, tags.Total, 0)
	require.Greater(t, len(tags.List), 0)
}

func TestTagsSort(t *testing.T) {
	testCases := []struct {
		name     string
		sortTag  SortTag
		expected []string
	}{
		{
			name:    "sort default",
			sortTag: SortTagDefault,
			expected: []string{
				"0.1.0",
				"0.4.0",
				"3.0.0-beta.1",
				"3.0.0-beta.3",
				"3.0.0-beta.4",
				"4",
				"4.0.0",
				"4.0.0-beta.1",
				"4.1.0",
				"4.1.1",
				"4.10.0",
				"4.11.0",
				"4.12.0",
				"4.13.0",
				"4.14.0",
				"4.19.0",
				"4.2.0",
				"4.20",
				"4.20.0",
				"4.20.1",
				"4.21",
				"4.21.0",
				"4.3.0",
				"4.3.1",
				"4.4.0",
				"4.6.1",
				"4.7.0",
				"4.8.0",
				"4.8.1",
				"4.9.0",
				"ubuntu-5.0",
				"alpine-5.0",
				"edge",
				"latest",
			},
		},
		{
			name:    "sort lexicographical",
			sortTag: SortTagLexicographical,
			expected: []string{
				"0.1.0",
				"0.4.0",
				"3.0.0-beta.1",
				"3.0.0-beta.3",
				"3.0.0-beta.4",
				"4",
				"4.0.0",
				"4.0.0-beta.1",
				"4.1.0",
				"4.1.1",
				"4.10.0",
				"4.11.0",
				"4.12.0",
				"4.13.0",
				"4.14.0",
				"4.19.0",
				"4.2.0",
				"4.20",
				"4.20.0",
				"4.20.1",
				"4.21",
				"4.21.0",
				"4.3.0",
				"4.3.1",
				"4.4.0",
				"4.6.1",
				"4.7.0",
				"4.8.0",
				"4.8.1",
				"4.9.0",
				"alpine-5.0",
				"edge",
				"latest",
				"ubuntu-5.0",
			},
		},
		{
			name:    "sort reverse",
			sortTag: SortTagReverse,
			expected: []string{
				"latest",
				"edge",
				"alpine-5.0",
				"ubuntu-5.0",
				"4.9.0",
				"4.8.1",
				"4.8.0",
				"4.7.0",
				"4.6.1",
				"4.4.0",
				"4.3.1",
				"4.3.0",
				"4.21.0",
				"4.21",
				"4.20.1",
				"4.20.0",
				"4.20",
				"4.2.0",
				"4.19.0",
				"4.14.0",
				"4.13.0",
				"4.12.0",
				"4.11.0",
				"4.10.0",
				"4.1.1",
				"4.1.0",
				"4.0.0-beta.1",
				"4.0.0",
				"4",
				"3.0.0-beta.4",
				"3.0.0-beta.3",
				"3.0.0-beta.1",
				"0.4.0",
				"0.1.0",
			},
		},
		{
			name:    "sort semver",
			sortTag: SortTagSemver,
			expected: []string{
				"alpine-5.0",
				"ubuntu-5.0",
				"4.21.0",
				"4.21",
				"4.20.1",
				"4.20.0",
				"4.20",
				"4.19.0",
				"4.14.0",
				"4.13.0",
				"4.12.0",
				"4.11.0",
				"4.10.0",
				"4.9.0",
				"4.8.1",
				"4.8.0",
				"4.7.0",
				"4.6.1",
				"4.4.0",
				"4.3.1",
				"4.3.0",
				"4.2.0",
				"4.1.1",
				"4.1.0",
				"4.0.0",
				"4",
				"4.0.0-beta.1",
				"3.0.0-beta.4",
				"3.0.0-beta.3",
				"3.0.0-beta.1",
				"0.4.0",
				"0.1.0",
				"edge",
				"latest",
			},
		},
	}
	for _, tt := range testCases {
		repotags := []string{
			"0.1.0",
			"0.4.0",
			"3.0.0-beta.1",
			"3.0.0-beta.3",
			"3.0.0-beta.4",
			"4",
			"4.0.0",
			"4.0.0-beta.1",
			"4.1.0",
			"4.1.1",
			"4.10.0",
			"4.11.0",
			"4.12.0",
			"4.13.0",
			"4.14.0",
			"4.19.0",
			"4.2.0",
			"4.20",
			"4.20.0",
			"4.20.1",
			"4.21",
			"4.21.0",
			"4.3.0",
			"4.3.1",
			"4.4.0",
			"4.6.1",
			"4.7.0",
			"4.8.0",
			"4.8.1",
			"4.9.0",
			"ubuntu-5.0",
			"alpine-5.0",
			"edge",
			"latest",
		}
		t.Run(tt.name, func(t *testing.T) {
			tags := SortTags(repotags, tt.sortTag)
			require.Equal(t, tt.expected, tags)
		})
	}
}
