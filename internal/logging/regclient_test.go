package logging

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	regtypes "github.com/regclient/regclient/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestRegclientLoggerFormatsHTTPRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)
	reglogger := NewRegclientLogger(logger)

	reglogger.LogAttrs(context.Background(), regtypes.LevelTrace, "reg http request",
		slog.String("req-method", "GET"),
		slog.String("req-url", "https://registry-1.docker.io/v2/"),
		slog.String("resp-status", "401 Unauthorized"),
		slog.Any("req-headers", http.Header{"Authorization": []string{"[censored]"}}),
	)

	output := buf.String()
	require.Contains(t, output, "[regclient] GET https://registry-1.docker.io/v2/ status=401 Unauthorized")
	require.NotContains(t, output, "Authorization")
}

func TestRegclientLoggerIncludesHeadersAtTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.TraceLevel)
	reglogger := NewRegclientLogger(logger)

	reglogger.LogAttrs(context.Background(), regtypes.LevelTrace, "reg http request",
		slog.String("req-method", "GET"),
		slog.String("req-url", "https://registry-1.docker.io/v2/"),
		slog.String("resp-status", "401 Unauthorized"),
		slog.Any("req-headers", http.Header{"Authorization": []string{"[censored]"}}),
	)

	output := buf.String()
	require.Contains(t, output, "[regclient] GET https://registry-1.docker.io/v2/ status=401 Unauthorized")
	require.True(t, strings.Contains(output, "Authorization") || strings.Contains(output, "req-headers"))
}

func TestRegclientLoggerSkipsLowValueNoise(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)
	reglogger := NewRegclientLogger(logger)

	reglogger.Debug("regclient initialized")
	reglogger.Debug("Auth request parsed")
	reglogger.Debug("Sleeping for backoff")

	require.Empty(t, buf.String())
}

func TestRegclientLoggerUsesInheritedGlobalLevel(t *testing.T) {
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer zerolog.SetGlobalLevel(prev)

	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	reglogger := NewRegclientLogger(logger)

	reglogger.LogAttrs(context.Background(), regtypes.LevelTrace, "reg http request",
		slog.String("req-method", "HEAD"),
		slog.String("req-url", "https://registry-1.docker.io/v2/test/app/manifests/latest"),
		slog.String("resp-status", "200 OK"),
		slog.Any("req-headers", http.Header{"Authorization": []string{"[censored]"}}),
	)

	output := buf.String()
	require.Contains(t, output, "[regclient] HEAD https://registry-1.docker.io/v2/test/app/manifests/latest status=200 OK")
	require.NotContains(t, output, "Authorization")
}
