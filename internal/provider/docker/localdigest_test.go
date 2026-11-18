package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestLocalDigests(t *testing.T) {
	dgst := digest.FromString("current")
	var repoLess atomic.Bool
	var tagMissing atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/version"):
			_, _ = w.Write([]byte(`{"ApiVersion":"1.52"}`))
		case strings.HasSuffix(r.URL.Path, "/containers/json"):
			_, _ = w.Write([]byte(`[
{"Id":"watched","Image":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ImageID":"actual-image","Labels":{"diun.enable":"true"}},
{"Id":"excluded","Image":"alpine:latest","ImageID":"excluded-image","Labels":{"diun.enable":"false"}},
{"Id":"unknown","Image":"redis:latest","ImageID":"missing","Labels":{"diun.enable":"true"}}
]`))
		case strings.HasSuffix(r.URL.Path, "/images/actual-image/json"), strings.HasSuffix(r.URL.Path, "/images/excluded-image/json"):
			if repoLess.Load() {
				_ = json.NewEncoder(w).Encode(map[string]any{"Descriptor": map[string]string{"digest": dgst.String()}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"RepoDigests": []string{"other@" + string(digest.FromString("other")), "alpine@" + string(dgst)}})
		case strings.HasSuffix(r.URL.Path, "/images/alpine:latest/json"):
			if tagMissing.Load() {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"tag missing"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"RepoDigests": []string{"other@" + string(digest.FromString("other")), "alpine@" + string(dgst)}})
		case strings.HasSuffix(r.URL.Path, "/containers/watched/json"):
			_, _ = w.Write([]byte(`{"Config":{"Image":"alpine:latest"}}`))
		case strings.HasSuffix(r.URL.Path, "/images/missing/json"), strings.HasSuffix(r.URL.Path, "/images/redis:latest/json"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"missing"}`))
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	cfg := (&model.PrdDocker{}).GetDefaults()
	cfg.Endpoint = server.URL
	cfg.APIVersion = "1.52"
	got, err := LocalDigests(context.Background(), cfg, nil)
	require.NoError(t, err)
	require.Equal(t, [][]digest.Digest{{dgst}}, got["docker.io/library/alpine:latest"])
	require.Equal(t, [][]digest.Digest{nil}, got["docker.io/library/redis:latest"])
	repoLess.Store(true)
	got, err = LocalDigests(context.Background(), cfg, nil)
	require.NoError(t, err)
	require.Equal(t, [][]digest.Digest{{dgst}}, got["docker.io/library/alpine:latest"])
	// The next watch run must keep monitoring the tag, rather than replace
	// its registry target with the old running image's immutable digest.
	jobs := New(cfg, nil).ListJob()
	require.Len(t, jobs, 1)
	require.Equal(t, "alpine:latest", jobs[0].Image.Name)
	// A removed tag must still be watched when its running image is known.
	repoLess.Store(false)
	tagMissing.Store(true)
	jobs = New(cfg, nil).ListJob()
	require.Len(t, jobs, 1)
	require.Equal(t, "alpine:latest", jobs[0].Image.Name)
}

func TestLocalDigestsDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	cfg := (&model.PrdDocker{}).GetDefaults()
	cfg.Endpoint = server.URL
	cfg.APIVersion = "1.52"
	ctx, cancel := context.WithTimeoutCause(context.Background(), 50*time.Millisecond, context.DeadlineExceeded)
	defer cancel()
	_, err := LocalDigests(ctx, cfg, nil)
	require.Error(t, err)
	require.ErrorIs(t, context.Cause(ctx), context.DeadlineExceeded)
}
