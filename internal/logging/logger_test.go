package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureSetsGlobalLevels(t *testing.T) {
	oldLogger := log.Logger
	oldGlobalLevel := zerolog.GlobalLevel()
	oldLogrusLevel := logrus.GetLevel()
	oldFormatter := logrus.StandardLogger().Formatter
	defer func() {
		log.Logger = oldLogger
		zerolog.SetGlobalLevel(oldGlobalLevel)
		logrus.SetLevel(oldLogrusLevel)
		logrus.SetFormatter(oldFormatter)
	}()

	Configure(Options{
		LogLevel: "debug",
		LogJSON:  true,
	})

	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	_, ok := logrus.StandardLogger().Formatter.(*LogrusFormatter)
	assert.True(t, ok)
}

func TestLogrusFormatterMapsLevels(t *testing.T) {
	testCases := []struct {
		name     string
		level    logrus.Level
		expected string
	}{
		{name: "error", level: logrus.ErrorLevel, expected: "error"},
		{name: "warn", level: logrus.WarnLevel, expected: "warn"},
		{name: "debug", level: logrus.DebugLevel, expected: "debug"},
		{name: "trace", level: logrus.TraceLevel, expected: "trace"},
		{name: "info", level: logrus.InfoLevel, expected: "info"},
		{name: "panic falls back to info", level: logrus.PanicLevel, expected: "info"},
		{name: "fatal falls back to info", level: logrus.FatalLevel, expected: "info"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := formatLogrusEntry(t, &logrus.Entry{
				Logger: logrus.New(),
				Data: logrus.Fields{
					"blob": "sha256:deadbeef",
				},
				Level:   tc.level,
				Message: "extracting",
			})

			assert.Equal(t, tc.expected, event["level"])
			assert.Equal(t, "[containers/image] extracting", event["message"])
			assert.Equal(t, "sha256:deadbeef", event["blob"])
		})
	}
}

func TestAntsLoggerPrintf(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := log.Logger
	oldLevel := zerolog.GlobalLevel()
	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer func() {
		log.Logger = oldLogger
		zerolog.SetGlobalLevel(oldLevel)
	}()

	new(AntsLogger).Printf("worker %d %s", 7, "ready")

	var event map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &event))
	assert.Equal(t, "debug", event["level"])
	assert.Equal(t, "worker 7 ready", event["message"])
}

func formatLogrusEntry(t *testing.T, entry *logrus.Entry) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	oldLogger := log.Logger
	oldLevel := zerolog.GlobalLevel()
	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	defer func() {
		log.Logger = oldLogger
		zerolog.SetGlobalLevel(oldLevel)
	}()

	payload, err := new(LogrusFormatter).Format(entry)
	require.NoError(t, err)
	assert.Nil(t, payload)

	var event map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &event))
	return event
}
