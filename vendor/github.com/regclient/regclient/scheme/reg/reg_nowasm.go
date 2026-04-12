//go:build !wasm

package reg

import (
	"log/slog"

	"github.com/sirupsen/logrus"

	"github.com/regclient/regclient/internal/reghttp"
	"github.com/regclient/regclient/internal/sloghandle"
)

// WithLog injects a logrus Logger configuration
func WithLog(log *logrus.Logger) Opts {
	return func(r *Reg) {
		r.slog = slog.New(sloghandle.Logrus(log))
		r.reghttpOpts = append(r.reghttpOpts, reghttp.WithLog(r.slog))
	}
}
