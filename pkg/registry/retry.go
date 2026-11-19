package registry

import (
	"context"
	stderrors "errors"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.podman.io/image/v5/docker"
)

const maxRetryBackoff = 8 * time.Second

// retry retries transient failures with jittered exponential backoff until the
// backoff reaches maxRetryBackoff. All attempts and waits share the caller's
// timeout. f must be safe to repeat.
func retry(ctx context.Context, f func() error) error {
	for backoff := time.Second; ; backoff *= 2 {
		if context.Cause(ctx) != nil {
			return contextFailure(ctx)
		}
		err := f()
		if err == nil {
			return nil
		}
		if context.Cause(ctx) != nil {
			return contextFailure(ctx)
		}
		if backoff >= maxRetryBackoff || !retryable(err) {
			return err
		}

		delay := backoff/2 + rand.N(backoff/2) //nolint:gosec // Jitter spreads retries; cryptographic randomness is unnecessary.
		log.Debug().Err(err).Dur("delay", delay).Msg("Retrying registry request")
		select {
		case <-ctx.Done():
			return contextFailure(ctx)
		case <-time.After(delay):
		}
	}
}

func contextFailure(ctx context.Context) error {
	err := ctx.Err() //nolint:forbidigo // Preserve cancellation status alongside a custom cause.
	cause := context.Cause(ctx)
	if stderrors.Is(cause, err) {
		return errors.WithStack(cause)
	}
	return errors.WithStack(stderrors.Join(err, cause))
}

func retryable(err error) bool {
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if httpErr, ok := stderrors.AsType[docker.UnexpectedHTTPStatusError](err); ok {
		switch httpErr.StatusCode {
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		return false
	}
	if stderrors.Is(err, io.EOF) || stderrors.Is(err, io.ErrUnexpectedEOF) || stderrors.Is(err, syscall.ECONNRESET) || stderrors.Is(err, syscall.EPIPE) || stderrors.Is(err, net.ErrClosed) {
		return true
	}
	if netErr, ok := stderrors.AsType[net.Error](err); ok {
		return netErr.Timeout() || netErr.Temporary() //nolint:staticcheck // Some network errors report transient failures only through Temporary.
	}
	return false
}
