package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/crazy-max/cron"
	"github.com/crazy-max/diun/internal/app"
	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/logging"
	"github.com/crazy-max/diun/internal/model"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	diun    *app.Diun
	flags   model.Flags
	c       *cron.Cron
	version = "dev"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse command line
	kingpin.Flag("config", "Diun configuration file.").Envar("CONFIG").Required().StringVar(&flags.Cfgfile)
	kingpin.Flag("timezone", "Timezone assigned to Diun.").Envar("TZ").Default("UTC").StringVar(&flags.Timezone)
	kingpin.Flag("log-level", "Set log level.").Envar("LOG_LEVEL").Default("info").StringVar(&flags.LogLevel)
	kingpin.Flag("log-json", "Enable JSON logging output.").Envar("LOG_JSON").Default("false").BoolVar(&flags.LogJson)
	kingpin.Flag("log-caller", "Enable to add file:line of the caller.").Envar("LOG_CALLER").Default("false").BoolVar(&flags.LogCaller)
	kingpin.Flag("run-startup", "Run on startup.").Envar("RUN_STARTUP").Default("false").BoolVar(&flags.RunStartup)
	kingpin.Flag("docker", "Enable Docker mode.").Envar("DOCKER").Default("false").BoolVar(&flags.Docker)
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("CrazyMax")
	kingpin.CommandLine.Name = "diun"
	kingpin.CommandLine.Help = `Docker image update notifier. More info on https://github.com/crazy-max/diun`
	kingpin.Parse()

	// Load timezone location
	location, err := time.LoadLocation(flags.Timezone)
	if err != nil {
		log.Panic().Err(err).Msgf("Cannot load timezone %s", flags.Timezone)
	}

	// Init
	logging.Configure(&flags, location)
	log.Info().Msgf("Starting Diun %s", version)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		if c != nil {
			c.Stop()
		}
		diun.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load and check configuration
	cfg, err := config.Load(flags, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	if err := cfg.Check(); err != nil {
		cfg.Display()
		log.Fatal().Err(err).Msg("Improper configuration")
	}
	cfg.Display()

	// Init
	if diun, err = app.New(cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize Diun")
	}

	// Run on startup
	if flags.RunStartup {
		diun.Run()
	}

	// Start scheduler
	c = cron.NewWithLocation(location)
	log.Info().Msgf("Start watcher with schedule %s", cfg.Watch.Schedule)
	if err := c.AddJob(cfg.Watch.Schedule, diun); err != nil {
		log.Fatal().Err(err).Msg("Cannot create cron task")
	}
	c.Start()

	select {}
}
