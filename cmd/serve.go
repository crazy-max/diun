package main

import (
	"os"
	"os/signal"
	"path"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"
)

// ServeCmd holds serve command args and flags
type ServeCmd struct {
	Cfgfile       string `kong:"name='config',env='CONFIG',help='Diun configuration file.'"`
	ProfilerPath  string `kong:"name='profiler-path',env='PROFILER_PATH',help='Base path where profiling files are written.'"`
	Profiler      string `kong:"name='profiler',env='PROFILER',help='Profiler to use.'"`
	LogLevel      string `kong:"name='log-level',env='LOG_LEVEL',default='info',help='Set log level.'"`
	LogJSON       bool   `kong:"name='log-json',env='LOG_JSON',default='false',help='Enable JSON logging output.'"`
	LogCaller     bool   `kong:"name='log-caller',env='LOG_CALLER',default='false',help='Add file:line of the caller to log output.'"`
	LogNoColor    bool   `kong:"name='log-nocolor',env='LOG_NOCOLOR',default='false',help='Disables the colorized output.'"`
	GRPCAuthority string `kong:"name='grpc-authority',env='GRPC_AUTHORITY',default=':42286',help='Address used to expose the gRPC server.'"`
}

func (s *ServeCmd) Run(ctx *Context) error {
	var diun *app.Diun

	// Logging
	logging.Configure(logging.Options{
		LogLevel:   s.LogLevel,
		LogJSON:    s.LogJSON,
		LogCaller:  s.LogCaller,
		LogNoColor: s.LogNoColor,
	})
	log.Info().Str("version", version).Msgf("Starting %s", ctx.Meta.Name)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, utl.SIGTERM)
	go func() {
		sig := <-channel
		diun.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := config.Load(s.Cfgfile)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	log.Debug().Msg(cfg.String())

	// Profiler
	if len(s.Profiler) > 0 && len(s.ProfilerPath) > 0 {
		profilerPath := path.Clean(s.ProfilerPath)
		if err = os.MkdirAll(profilerPath, os.ModePerm); err != nil {
			log.Fatal().Err(err).Msg("Cannot create profiler folder")
		}
		profilePath := profile.ProfilePath(profilerPath)
		switch s.Profiler {
		case "cpu":
			defer profile.Start(profile.CPUProfile, profilePath).Stop()
		case "mem":
			defer profile.Start(profile.MemProfile, profilePath).Stop()
		case "alloc":
			defer profile.Start(profile.MemProfileAllocs, profilePath).Stop()
		case "heap":
			defer profile.Start(profile.MemProfileHeap, profilePath).Stop()
		case "routines":
			defer profile.Start(profile.GoroutineProfile, profilePath).Stop()
		case "mutex":
			defer profile.Start(profile.MutexProfile, profilePath).Stop()
		case "threads":
			defer profile.Start(profile.ThreadcreationProfile, profilePath).Stop()
		case "block":
			defer profile.Start(profile.BlockProfile, profilePath).Stop()
		default:
			log.Fatal().Msgf("Unknown profiler: %s", s.Profiler)
		}
	}

	// Init
	diun, err = app.New(ctx.Meta, cfg, s.GRPCAuthority)
	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot initialize %s", ctx.Meta.Name)
	}

	// Start
	err = diun.Start()
	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot start %s", ctx.Meta.Name)
	}

	return nil
}
