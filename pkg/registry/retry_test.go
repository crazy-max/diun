package registry

import (
	"context"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.podman.io/image/v5/docker"
)

func TestRetry(t *testing.T) {
	for _, tt := range []struct {
		name     string
		failures int
		attempts int
	}{
		{
			name:     "immediate success",
			attempts: 1,
		},
		{
			name:     "recovery",
			failures: 2,
			attempts: 3,
		},
		{
			name:     "retry limit",
			failures: 4,
			attempts: 4,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				attempts := 0
				previous := time.Now()
				failure := errors.Wrap(io.EOF, "registry disconnected")
				err := retry(t.Context(), func() error {
					if attempts > 0 {
						backoff := time.Second << (attempts - 1)
						assert.GreaterOrEqual(t, time.Since(previous), backoff/2)
						assert.Less(t, time.Since(previous), backoff)
					}
					previous = time.Now()
					attempts++
					if attempts <= tt.failures {
						return failure
					}
					return nil
				})
				assert.Equal(t, tt.attempts, attempts)
				if tt.failures == tt.attempts {
					assert.Same(t, failure, err)
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

func TestRetryCancellation(t *testing.T) {
	for _, cause := range []error{context.Canceled, stderrors.New("registry watcher stopped")} {
		t.Run(cause.Error(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancelCause(t.Context())
				defer cancel(context.Canceled)
				attempts := 0
				done := make(chan error, 1)
				go func() {
					done <- retry(ctx, func() error {
						attempts++
						return io.EOF
					})
				}()
				synctest.Wait()
				start := time.Now()
				cancel(cause)
				err := <-done
				require.ErrorIs(t, err, context.Canceled)
				require.ErrorIs(t, err, cause)
				assert.Zero(t, time.Since(start))
				assert.Equal(t, 1, attempts)
			})
		})
	}
}

func TestRetryCanceledBeforeAttempt(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	cause := stderrors.New("registry watcher stopped")
	cancel(cause)
	err := retry(ctx, func() error {
		t.Fatal("operation called with a canceled context")
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, cause)
}

func TestRetryDeadline(t *testing.T) {
	for _, cause := range []error{context.DeadlineExceeded, stderrors.New("registry timeout")} {
		t.Run(cause.Error(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithTimeoutCause(t.Context(), 250*time.Millisecond, cause)
				defer cancel()
				attempts := 0
				start := time.Now()
				err := retry(ctx, func() error {
					attempts++
					return io.EOF
				})
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.ErrorIs(t, err, cause)
				assert.Equal(t, 250*time.Millisecond, time.Since(start))
				assert.Equal(t, 1, attempts)
			})
		})
	}
}

func TestRetryPermanentError(t *testing.T) {
	attempts := 0
	failure := stderrors.New("invalid manifest")
	err := retry(t.Context(), func() error {
		attempts++
		return failure
	})
	assert.Same(t, failure, err)
	assert.Equal(t, 1, attempts)
}

func TestRetryable(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "EOF",
			err:  io.EOF,
			want: true,
		},
		{
			name: "unexpected EOF",
			err:  io.ErrUnexpectedEOF,
			want: true,
		},
		{
			name: "connection reset",
			err:  syscall.ECONNRESET,
			want: true,
		},
		{
			name: "broken pipe",
			err:  syscall.EPIPE,
			want: true,
		},
		{
			name: "closed connection",
			err:  net.ErrClosed,
			want: true,
		},
		{
			name: "network timeout",
			err:  &net.DNSError{IsTimeout: true},
			want: true,
		},
		{
			name: "temporary DNS failure",
			err:  &net.DNSError{IsTemporary: true},
			want: true,
		},
		{
			name: "unknown host",
			err:  &net.DNSError{IsNotFound: true},
		},
		{
			name: "rate limit handled upstream",
			err:  docker.ErrTooManyRequests,
		},
		{
			name: "HTTP rate limit handled upstream",
			err:  docker.UnexpectedHTTPStatusError{StatusCode: http.StatusTooManyRequests},
		},
		{
			name: "canceled",
			err:  context.Canceled,
		},
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, retryable(errors.Wrap(tt.err, "registry request failed")))
		})
	}
}
