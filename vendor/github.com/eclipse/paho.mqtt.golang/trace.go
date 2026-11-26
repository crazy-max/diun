/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"context"
	"io"
	"log/slog"
)

type (
	// Logger interface allows implementations to provide to this package any
	// object that implements the methods defined in it.
	Logger interface {
		Println(v ...interface{})
		Printf(format string, v ...interface{})
	}

	// NOOPLogger implements the logger that does not perform any operation
	// by default. This allows us to efficiently discard the unwanted messages.
	NOOPLogger struct{}
)

func (NOOPLogger) Println(v ...interface{})               {}
func (NOOPLogger) Printf(format string, v ...interface{}) {}

// Internal levels of library output that are initialised to not print
// anything but can be overridden by programmer
var (
	ERROR    Logger = NOOPLogger{}
	CRITICAL Logger = NOOPLogger{}
	WARN     Logger = NOOPLogger{}
	DEBUG    Logger = NOOPLogger{}
)

var noopSLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

type logWriter struct {
	l Logger
}

func (w logWriter) Write(p []byte) (n int, err error) {
	w.l.Printf("%s", string(p))
	return len(p), nil
}

// LogWrapper implements slog.Handler to bridge slog logging to legacy Logger interfaces
type LogWrapper struct {
	handler  slog.Handler
	ERROR    slog.Handler
	CRITICAL slog.Handler
	WARN     slog.Handler
	DEBUG    slog.Handler
}

// NewLogWrapper creates a new LogWrapper that bridges slog to legacy loggers
func NewLogWrapper(handler slog.Handler) *LogWrapper {
	return &LogWrapper{
		handler:  handler,
		ERROR:    slog.NewTextHandler(logWriter{ERROR}, &slog.HandlerOptions{}),
		CRITICAL: slog.NewTextHandler(logWriter{CRITICAL}, &slog.HandlerOptions{}),
		WARN:     slog.NewTextHandler(logWriter{WARN}, &slog.HandlerOptions{}),
		DEBUG:    slog.NewTextHandler(logWriter{DEBUG}, &slog.HandlerOptions{}),
	}
}

func (w *LogWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return w.handler.Enabled(ctx, level)
}

func (w *LogWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogWrapper{
		handler:  w.handler.WithAttrs(attrs),
		ERROR:    w.ERROR.WithAttrs(attrs),
		CRITICAL: w.CRITICAL.WithAttrs(attrs),
		WARN:     w.WARN.WithAttrs(attrs),
		DEBUG:    w.DEBUG.WithAttrs(attrs),
	}
}

func (w *LogWrapper) WithGroup(name string) slog.Handler {
	return &LogWrapper{
		handler:  w.handler.WithGroup(name),
		ERROR:    w.ERROR.WithGroup(name),
		CRITICAL: w.CRITICAL.WithGroup(name),
		WARN:     w.WARN.WithGroup(name),
		DEBUG:    w.DEBUG.WithGroup(name),
	}
}

func (w *LogWrapper) Handle(ctx context.Context, record slog.Record) error {
	switch {
	case record.Level == slog.LevelError:
		w.ERROR.Handle(ctx, record)
	case record.Level == slog.LevelWarn:
		w.CRITICAL.Handle(ctx, record)
	case record.Level == slog.LevelInfo:
		w.WARN.Handle(ctx, record)
	case record.Level == slog.LevelDebug:
		w.DEBUG.Handle(ctx, record)
	}

	return w.handler.Handle(ctx, record)
}
