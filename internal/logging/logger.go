package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
)

// Configure configures logger
func Configure(cli *model.Cli, location *time.Location) {
	var err error
	var w io.Writer

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(location)
	}

	if !cli.LogJSON {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC1123,
		}
	} else {
		w = os.Stdout
	}

	ctx := zerolog.New(w).With().Timestamp()
	if cli.LogCaller {
		ctx = ctx.Caller()
	}

	log.Logger = ctx.Logger()

	logLevel, err := zerolog.ParseLevel(cli.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}

	logrusLevel, err := logrus.ParseLevel(cli.LogLevel)
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
