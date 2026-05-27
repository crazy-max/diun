package metrics

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/secret"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// Server exposes Prometheus metrics over HTTP.
type Server struct {
	httpServer *http.Server
	path       string
	tokenHash  [sha256.Size]byte
	auth       bool
}

// NewServer creates a Prometheus metrics HTTP server.
func NewServer(cfg *model.Metrics, registry *prometheus.Registry) (*Server, error) {
	token, err := secret.GetSecret(cfg.Token, cfg.TokenFile)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		path: cfg.Path,
		auth: token != "",
	}
	if srv.auth {
		srv.tokenHash = sha256.Sum256([]byte("Bearer " + token))
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, srv.authHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog: promLogger{},
	})))

	srv.httpServer = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return srv, nil
}

// Listen opens the Prometheus metrics HTTP listener.
func (s *Server) Listen() (net.Listener, error) {
	return (&net.ListenConfig{}).Listen(context.Background(), "tcp", s.httpServer.Addr)
}

// Serve starts the Prometheus metrics HTTP server.
func (s *Server) Serve(lis net.Listener) error {
	log.Info().Str("addr", lis.Addr().String()).Str("path", s.path).Msg("Prometheus metrics server listening")

	if err := s.httpServer.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Start starts the Prometheus metrics HTTP server.
func (s *Server) Start() error {
	lis, err := s.Listen()
	if err != nil {
		return err
	}
	return s.Serve(lis)
}

// Shutdown gracefully stops the Prometheus metrics HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) authHandler(next http.Handler) http.Handler {
	if !s.auth {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHash := sha256.Sum256([]byte(r.Header.Get("Authorization")))
		if subtle.ConstantTimeCompare(gotHash[:], s.tokenHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="diun metrics"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type promLogger struct{}

func (promLogger) Println(v ...any) {
	event := log.Error()
	for _, arg := range v {
		if err, ok := arg.(error); ok {
			event = event.Err(err)
			break
		}
	}
	event.Msg(strings.TrimSpace(fmt.Sprintln(v...)))
}
