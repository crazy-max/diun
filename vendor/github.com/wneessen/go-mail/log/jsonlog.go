// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

package log

import (
	"fmt"
	"io"
	"log/slog"
)

// JSONlog is the default structured JSON logger that satisfies the Logger interface
type JSONlog struct {
	level Level
	log   *slog.Logger
}

// NewJSON returns a new JSONlog type that satisfies the Logger interface
func NewJSON(output io.Writer, level Level) *JSONlog {
	logOpts := slog.HandlerOptions{}
	switch level {
	case LevelDebug:
		logOpts.Level = slog.LevelDebug
	case LevelInfo:
		logOpts.Level = slog.LevelInfo
	case LevelWarn:
		logOpts.Level = slog.LevelWarn
	case LevelError:
		logOpts.Level = slog.LevelError
	default:
		logOpts.Level = slog.LevelDebug
	}
	logHandler := slog.NewJSONHandler(output, &logOpts)
	return &JSONlog{
		level: level,
		log:   slog.New(logHandler),
	}
}

// logMessage is a helper function to handle different log levels and formats.
func logMessage(level Level, log *slog.Logger, logData Log, formatFunc func(string, ...interface{}) string) {
	lGroup := log.WithGroup(DirString).With(
		slog.String(DirFromString, logData.directionFrom()),
		slog.String(DirToString, logData.directionTo()),
	)
	switch level {
	case LevelDebug:
		lGroup.Debug(formatFunc(logData.Format, logData.Messages...))
	case LevelInfo:
		lGroup.Info(formatFunc(logData.Format, logData.Messages...))
	case LevelWarn:
		lGroup.Warn(formatFunc(logData.Format, logData.Messages...))
	case LevelError:
		lGroup.Error(formatFunc(logData.Format, logData.Messages...))
	}
}

// Debugf logs a debug message via the structured JSON logger
func (l *JSONlog) Debugf(log Log) {
	if l.level >= LevelDebug {
		logMessage(LevelDebug, l.log, log, fmt.Sprintf)
	}
}

// Infof logs a info message via the structured JSON logger
func (l *JSONlog) Infof(log Log) {
	if l.level >= LevelInfo {
		logMessage(LevelInfo, l.log, log, fmt.Sprintf)
	}
}

// Warnf logs a warn message via the structured JSON logger
func (l *JSONlog) Warnf(log Log) {
	if l.level >= LevelWarn {
		logMessage(LevelWarn, l.log, log, fmt.Sprintf)
	}
}

// Errorf logs a warn message via the structured JSON logger
func (l *JSONlog) Errorf(log Log) {
	if l.level >= LevelError {
		logMessage(LevelError, l.log, log, fmt.Sprintf)
	}
}
