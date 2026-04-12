// Package warning is used to handle HTTP warning headers
package warning

import (
	"context"
	"log/slog"
	"slices"
	"sync"
)

type contextKey string

var key contextKey = "key"

type Warning struct {
	List []string
	Hook *func(context.Context, *slog.Logger, string)
	mu   sync.Mutex
}

func (w *Warning) Handle(ctx context.Context, slog *slog.Logger, msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if slices.Contains(w.List, msg) {
		return
	}
	w.List = append(w.List, msg)
	// handle new warning if hook defined
	if w.Hook != nil {
		(*w.Hook)(ctx, slog, msg)
	}
}

func NewContext(ctx context.Context, w *Warning) context.Context {
	return context.WithValue(ctx, key, w)
}

func FromContext(ctx context.Context) *Warning {
	wAny := ctx.Value(key)
	if wAny == nil {
		return nil
	}
	w, ok := wAny.(*Warning)
	if !ok {
		return nil
	}
	return w
}

func NewHook(log *slog.Logger) *func(context.Context, *slog.Logger, string) {
	hook := func(_ context.Context, _ *slog.Logger, msg string) {
		logMsg(log, msg)
	}
	return &hook
}

func DefaultHook() *func(context.Context, *slog.Logger, string) {
	hook := func(_ context.Context, slog *slog.Logger, msg string) {
		logMsg(slog, msg)
	}
	return &hook
}

func Handle(ctx context.Context, slog *slog.Logger, msg string) {
	// check for context
	if w := FromContext(ctx); w != nil {
		w.Handle(ctx, slog, msg)
		return
	}

	// fallback to log
	logMsg(slog, msg)
}

func logMsg(log *slog.Logger, msg string) {
	log.Warn("Registry warning message", slog.String("warning", msg))
}
