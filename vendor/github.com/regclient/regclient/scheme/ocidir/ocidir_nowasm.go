//go:build !wasm

package ocidir

import (
	"log/slog"

	"github.com/sirupsen/logrus"

	"github.com/regclient/regclient/internal/sloghandle"
)

// WithLog provides a logrus logger.
// By default logging is disabled.
func WithLog(log *logrus.Logger) Opts {
	return func(c *ociConf) {
		c.slog = slog.New(sloghandle.Logrus(log))
	}
}
