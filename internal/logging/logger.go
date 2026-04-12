package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	if !opts.LogJSON {
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
}
