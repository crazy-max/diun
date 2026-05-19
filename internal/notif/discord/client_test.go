package discord

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
)

// mockServerConfig holds configuration for the mock Discord server
type mockServerConfig struct {
	t       *testing.T
	handler func(w http.ResponseWriter, r *http.Request, count *atomic.Int32)
}

// createMockServer creates a mock Discord server with the given handler
func createMockServer(cfg mockServerConfig) (server *httptest.Server, requestCount *atomic.Int32) {
	cfg.t.Helper()
	var counter atomic.Int32
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		io.ReadAll(r.Body)
		cfg.handler(w, r, &counter)
	}))
	cfg.t.Cleanup(func() { server.Close() })
	return server, &counter
}

// createTestClient creates a test client with the given webhook URL
func createTestClient(t *testing.T, webhookURL string) *Client {
	t.Helper()
	timeout := 10 * time.Second
	cfg := &model.NotifDiscord{
		WebhookURL:   webhookURL,
		RenderEmbeds: ptr(false),
		Timeout:      &timeout,
	}
	return &Client{
		cfg: cfg,
		meta: model.Meta{
			Name:      "Test",
			Hostname:  "test",
			UserAgent: "test",
		},
	}
}

// createTestEntry creates a standard test notification entry
func createTestEntry() model.NotifEntry {
	img, _ := registry.ParseImage(registry.ParseImageOptions{
		Name: "test/image:latest",
	})
	return model.NotifEntry{
		Provider: "test",
		Image:    img,
		Manifest: registry.Manifest{
			Created:  ptr(time.Now()),
			Digest:   digest.Digest("sha256:test"),
			Platform: "linux/amd64",
		},
	}
}

func TestSendWith429Retry(t *testing.T) {
	// Create mock server that returns 429 once, then success
	server, requestCount := createMockServer(mockServerConfig{
		t: t,
		handler: func(w http.ResponseWriter, r *http.Request, count *atomic.Int32) {
			if count.Load() == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"message": "You are being rate limited.", "retry_after": 0.1, "global": false}`)
				t.Logf("Request 1: returning 429 (retry_after: 1.0s)")
			} else {
				w.WriteHeader(http.StatusNoContent)
				t.Logf("Request %d: returning 204 (success)", count.Load())
			}
		},
	})

	client := createTestClient(t, server.URL)
	entry := createTestEntry()

	err := client.Send(entry)

	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	// Should have made 2 requests (1 failure + 1 success)
	if requestCount.Load() != 2 {
		t.Errorf("Expected 2 requests, got %d", requestCount.Load())
	}
}

func TestSendWith429MaxRetries(t *testing.T) {
	server, requestCount := createMockServer(mockServerConfig{
		t: t,
		handler: func(w http.ResponseWriter, r *http.Request, count *atomic.Int32) {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"message": "You are being rate limited.", "retry_after": 0.1, "global": false}`)
		},
	})

	client := createTestClient(t, server.URL)
	entry := createTestEntry()

	err := client.Send(entry)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	if requestCount.Load() != 3 {
		t.Errorf("Expected 3 requests (max retries), got %d", requestCount.Load())
	}

	t.Logf("Correctly failed after max retries: %v", err)
}

func ptr[T any](v T) *T { return &v }
