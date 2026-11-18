package docker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClosesConnectionOnVersionError(t *testing.T) {
	closed := make(chan struct{}, 1)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ApiVersion":`))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateClosed {
			select {
			case closed <- struct{}{}:
			default:
			}
		}
	}
	server.Start()
	defer server.Close()

	cli, err := New(Options{Endpoint: server.URL, APIVersion: "1.52"})
	require.Error(t, err)
	require.Nil(t, cli)
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("failed initialization left its HTTP connection open")
	}
}
