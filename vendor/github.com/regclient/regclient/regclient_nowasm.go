//go:build !wasm

package regclient

import (
	"log/slog"

	"github.com/sirupsen/logrus"

	"github.com/regclient/regclient/internal/sloghandle"
)

// WithLog configuring logging with a logrus Logger.
// Note that regclient has switched to log/slog for logging and my eventually deprecate logrus support.
func WithLog(log *logrus.Logger) Opt {
	return func(rc *RegClient) {
		rc.slog = slog.New(sloghandle.Logrus(log))
	}
}
