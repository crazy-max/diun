package logging

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"slices"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logURLPattern = regexp.MustCompile(`(?i)https?://[^\s"<>]+`)

func redactLogURLs(value string) string {
	return logURLPattern.ReplaceAllStringFunc(value, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			return "[redacted URL]"
		}
		u.RawQuery, u.Fragment, u.RawFragment, u.User = "", "", "", nil
		u.ForceQuery = false
		return u.String()
	})
}

// SlogHandler routes registry diagnostics through DIUN's configured logger.
type SlogHandler struct {
	attrs  []slog.Attr
	groups []string
}

func (h SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return log.Logger.WithLevel(slogLevel(level)).Enabled()
}

func (h SlogHandler) Handle(_ context.Context, record slog.Record) error {
	event := log.Logger.WithLevel(slogLevel(record.Level))
	fields := make(map[string]any)
	for _, attr := range h.attrs {
		addSlogAttr(fields, attr)
	}
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool { attrs = append(attrs, attr); return true })
	for _, attr := range h.groupAttrs(attrs) {
		addSlogAttr(fields, attr)
	}
	event.Fields(fields).Msg("[regclient] " + redactLogURLs(record.Message))
	return nil
}

func (h SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(slices.Clone(h.attrs), h.groupAttrs(attrs)...)
	return h
}

func (h SlogHandler) WithGroup(name string) slog.Handler {
	if name != "" {
		h.groups = append(slices.Clone(h.groups), name)
	}
	return h
}

func (h SlogHandler) groupAttrs(attrs []slog.Attr) []slog.Attr {
	for _, group := range slices.Backward(h.groups) {
		attrs = []slog.Attr{slog.GroupAttrs(group, attrs...)}
	}
	return attrs
}

func addSlogAttr(fields map[string]any, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	// Headers and URL queries can contain credentials or signed redirect URLs.
	if attr.Key == "req-headers" || attr.Key == "resp-headers" {
		return
	}
	if attr.Value.Kind() == slog.KindString {
		attr.Value = slog.StringValue(redactLogURLs(attr.Value.String()))
	} else if err, ok := attr.Value.Any().(error); ok {
		attr.Value = slog.StringValue(redactLogURLs(err.Error()))
	}
	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key == "" {
			for _, child := range attr.Value.Group() {
				addSlogAttr(fields, child)
			}
			return
		}
		nested, ok := fields[attr.Key].(map[string]any)
		if !ok {
			nested = make(map[string]any)
		}
		for _, child := range attr.Value.Group() {
			addSlogAttr(nested, child)
		}
		if len(nested) > 0 {
			fields[attr.Key] = nested
		}
		return
	}
	fields[attr.Key] = attr.Value.Any()
}

func slogLevel(level slog.Level) zerolog.Level {
	switch {
	case level < slog.LevelDebug:
		return zerolog.TraceLevel
	case level < slog.LevelInfo:
		return zerolog.DebugLevel
	case level < slog.LevelWarn:
		return zerolog.InfoLevel
	case level < slog.LevelError:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}
