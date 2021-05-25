package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
)

type Options struct {
	LogLevel   string
	LogJSON    bool
	LogCaller  bool
	LogNoColor bool
}

// Configure configures logger
func Configure(opts Options) {
	var err error
	var w io.Writer

	// Adds support for NO_COLOR. More info https://no-color.org/
	_, noColor := os.LookupEnv("NO_COLOR")

	if !cli.LogJSON {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    noColor || opts.LogNoColor,
			TimeFormat: time.RFC1123,
		}
	} else {
		w = os.Stdout
	}

	ctx := zerolog.New(w).With().Timestamp()
	if opts.LogCaller {
		ctx = ctx.Caller()
	}

	log.Logger = ctx.Logger()

	logLevel, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}

	logrusLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		logrus.SetLevel(logrusLevel)
	}
	logrus.SetFormatter(new(LogrusFormatter))
}

// LogrusFormatter is a logrus formatter
type LogrusFormatter struct{}

// Format renders a single log entry from logrus entry to zerolog
func (f *LogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := fmt.Sprintf("[containers/image] %s", entry.Message)
	switch entry.Level {
	case logrus.ErrorLevel:
		log.Error().Fields(entry.Data).Msg(message)
	case logrus.WarnLevel:
		log.Warn().Fields(entry.Data).Msg(message)
	case logrus.DebugLevel:
		log.Debug().Fields(entry.Data).Msg(message)
	default:
		log.Info().Fields(entry.Data).Msg(message)
	}
	return nil, nil
}

// AntsLogger is a logger for ants module
type AntsLogger struct{}

// Printf must have the same semantics as log.Printf
func (w *AntsLogger) Printf(format string, args ...interface{}) {
	log.Debug().Msgf(format, args...)
}
