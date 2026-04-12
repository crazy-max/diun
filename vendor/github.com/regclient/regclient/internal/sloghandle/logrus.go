//go:build !wasm

// Package sloghandle provides a transition handler for migrating from logrus to slog.
package sloghandle

import (
	"context"
	"log/slog"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/regclient/regclient/types"
)

func Logrus(logger *logrus.Logger) *logrusHandler {
	return &logrusHandler{
		logger: logger,
	}
}

type logrusHandler struct {
	logger *logrus.Logger
	attrs  []slog.Attr
	groups []string
}

func (h *logrusHandler) Enabled(_ context.Context, level slog.Level) bool {
	ll := h.logger.GetLevel()
	if curLevel, ok := logrusToSlog[ll]; ok {
		return level >= curLevel
	}
	return true
}

func (h *logrusHandler) Handle(ctx context.Context, r slog.Record) error {
	log := logrus.NewEntry(h.logger).WithContext(ctx)
	if !r.Time.IsZero() {
		log = log.WithTime(r.Time)
	}
	fields := logrus.Fields{}
	for _, a := range h.attrs {
		if a.Key != "" {
			fields[a.Key] = a.Value
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "" {
			fields[a.Key] = a.Value
		}
		return true
	})
	if len(fields) > 0 {
		log = log.WithFields(fields)
	}
	log.Log(slogToLogrus(r.Level), r.Message)
	return nil
}

func (h *logrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ret := h.clone()
	prefix := ""
	if len(h.groups) > 0 {
		prefix = strings.Join(h.groups, ":") + ":"
	}
	for _, a := range attrs {
		if a.Key == "" {
			continue
		}
		ret.attrs = append(ret.attrs, slog.Attr{
			Key:   prefix + a.Key,
			Value: a.Value,
		})
	}
	return ret
}

func (h *logrusHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	ret := h.clone()
	ret.groups = append(ret.groups, name)
	return ret
}

func (h *logrusHandler) clone() *logrusHandler {
	attrs := make([]slog.Attr, len(h.attrs))
	copy(attrs, h.attrs)
	groups := make([]string, len(h.groups))
	copy(groups, h.groups)
	return &logrusHandler{
		logger: h.logger,
		attrs:  attrs,
		groups: groups,
	}
}

var logrusToSlog = map[logrus.Level]slog.Level{
	logrus.TraceLevel: types.LevelTrace,
	logrus.DebugLevel: slog.LevelDebug,
	logrus.InfoLevel:  slog.LevelInfo,
	logrus.WarnLevel:  slog.LevelWarn,
	logrus.ErrorLevel: slog.LevelError,
	logrus.FatalLevel: slog.LevelError + 4,
	logrus.PanicLevel: slog.LevelError + 8,
}

func slogToLogrus(level slog.Level) logrus.Level {
	if level <= types.LevelTrace {
		return logrus.TraceLevel
	} else if level <= slog.LevelDebug {
		return logrus.DebugLevel
	} else if level <= slog.LevelInfo {
		return logrus.InfoLevel
	} else if level <= slog.LevelWarn {
		return logrus.WarnLevel
	} else if level <= slog.LevelError {
		return logrus.ErrorLevel
	} else if level <= slog.LevelError+4 {
		return logrus.FatalLevel
	} else {
		return logrus.PanicLevel
	}
}
