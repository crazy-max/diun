package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/diun/internal/app"
	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/logging"
	"github.com/crazy-max/diun/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	diun    *app.Diun
	cli     model.Cli
	version = "dev"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse command line
	_ = kong.Parse(&cli,
		kong.Name("diun"),
		kong.Description(`Docker image update notifier. More info: https://github.com/crazy-max/diun`),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s", version),
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	// Load timezone location
	location, err := time.LoadLocation(cli.Timezone)
	if err != nil {
		log.Panic().Err(err).Msgf("Cannot load timezone %s", cli.Timezone)
	}

	// Init
	logging.Configure(&cli, location)
	log.Info().Msgf("Starting Diun %s", version)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		diun.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := config.Load(cli, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	cfg.Display()

	// Init
	if diun, err = app.New(cfg, location); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize Diun")
	}

	// Start
	if err = diun.Start(); err != nil {
		log.Fatal().Err(err).Msg("Cannot start Diun")
	}
}
