package logging

import (
	"context"
	"fmt"
	"log/slog"

	regtypes "github.com/regclient/regclient/types"
	"github.com/rs/zerolog"
)

func NewRegclientLogger(logger zerolog.Logger) *slog.Logger {
	return slog.New(&regclientHandler{logger: logger})
}

type regclientHandler struct {
	logger zerolog.Logger
	attrs  []slog.Attr
	group  string
}

func (h *regclientHandler) Enabled(_ context.Context, level slog.Level) bool {
	current := h.effectiveLevel()
	return regclientLevel(current, level) >= current
}

func (h *regclientHandler) Handle(_ context.Context, record slog.Record) error {
	if h.skipMessage(record.Message) {
		return nil
	}
	eventLogger, current := h.eventLogger()
	level := regclientLevel(current, record.Level)
	event := eventLogger.WithLevel(level)
	if event == nil {
		return nil
	}
	if !record.Time.IsZero() {
		event.Time(zerolog.TimestampFieldName, record.Time)
	}

	fields := h.fields(record)
	message := "[regclient] " + record.Message
	if record.Message == "reg http request" {
		message = h.httpRequestMessage(fields)
	}

	includeHeaders := current <= zerolog.TraceLevel
	for key, value := range fields {
		if !includeHeaders && (key == "req-headers" || key == "resp-headers") {
			continue
		}
		event.Interface(key, value)
	}
	event.Msg(message)
	return nil
}

func (h *regclientHandler) skipMessage(message string) bool {
	switch message {
	case "regclient initialized", "Auth request parsed", "Sleeping for backoff":
		return true
	default:
		return false
	}
}

func (h *regclientHandler) effectiveLevel() zerolog.Level {
	level := h.logger.GetLevel()
	if level == zerolog.NoLevel || zerolog.GlobalLevel() > level {
		return zerolog.GlobalLevel()
	}
	return level
}

func (h *regclientHandler) eventLogger() (zerolog.Logger, zerolog.Level) {
	level := h.effectiveLevel()
	if h.logger.GetLevel() != level {
		return h.logger.Level(level), level
	}
	return h.logger, level
}

func (h *regclientHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ret := h.clone()
	prefix := ret.group
	if prefix != "" {
		prefix += ":"
	}
	for _, attr := range attrs {
		if attr.Key == "" {
			continue
		}
		ret.attrs = append(ret.attrs, slog.Attr{
			Key:   prefix + attr.Key,
			Value: attr.Value,
		})
	}
	return ret
}

func (h *regclientHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	ret := h.clone()
	if ret.group == "" {
		ret.group = name
	} else {
		ret.group += ":" + name
	}
	return ret
}

func (h *regclientHandler) clone() *regclientHandler {
	attrs := make([]slog.Attr, len(h.attrs))
	copy(attrs, h.attrs)
	return &regclientHandler{
		logger: h.logger,
		attrs:  attrs,
		group:  h.group,
	}
}

func (h *regclientHandler) fields(record slog.Record) map[string]any {
	fields := make(map[string]any, len(h.attrs))
	for _, attr := range h.attrs {
		appendAttr(fields, attr, "")
	}
	record.Attrs(func(attr slog.Attr) bool {
		appendAttr(fields, attr, "")
		return true
	})
	return fields
}

func appendAttr(fields map[string]any, attr slog.Attr, prefix string) {
	if attr.Key == "" {
		return
	}
	key := prefix + attr.Key
	if attr.Value.Kind() == slog.KindGroup {
		nextPrefix := key + ":"
		for _, groupAttr := range attr.Value.Group() {
			appendAttr(fields, groupAttr, nextPrefix)
		}
		return
	}
	fields[key] = attrValue(attr.Value)
}

func attrValue(value slog.Value) any {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindUint64:
		return value.Uint64()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindBool:
		return value.Bool()
	case slog.KindDuration:
		return value.Duration()
	case slog.KindTime:
		return value.Time()
	case slog.KindAny:
		return value.Any()
	default:
		return value.Any()
	}
}

func (h *regclientHandler) httpRequestMessage(fields map[string]any) string {
	method, _ := fields["req-method"].(string)
	url, _ := fields["req-url"].(string)
	status, hasStatus := fields["resp-status"]
	errValue, hasErr := fields["err"]

	delete(fields, "req-method")
	delete(fields, "req-url")
	delete(fields, "resp-status")
	delete(fields, "err")

	parts := []string{"[regclient]"}
	if method != "" {
		parts = append(parts, method)
	}
	if url != "" {
		parts = append(parts, url)
	}
	if hasStatus {
		parts = append(parts, fmt.Sprintf("status=%v", status))
	}
	if hasErr {
		parts = append(parts, fmt.Sprintf("err=%v", errValue))
	}
	message := parts[0]
	for _, part := range parts[1:] {
		message += " " + part
	}
	return message
}

func regclientLevel(current zerolog.Level, level slog.Level) zerolog.Level {
	switch {
	case level <= regtypes.LevelTrace:
		if current <= zerolog.TraceLevel {
			return zerolog.TraceLevel
		}
		return zerolog.DebugLevel
	case level <= slog.LevelDebug:
		return zerolog.DebugLevel
	case level <= slog.LevelInfo:
		return zerolog.InfoLevel
	case level <= slog.LevelWarn:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}

var _ slog.Handler = (*regclientHandler)(nil)
