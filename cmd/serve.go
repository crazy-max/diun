package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"
)

// ServeCmd holds serve command args and flags
type ServeCmd struct {
	Cfgfile       string `name:"config" env:"CONFIG" help:"Diun configuration file."`
	ProfilerPath  string `name:"profiler-path" env:"PROFILER_PATH" help:"Base path where profiling files are written."`
	Profiler      string `name:"profiler" env:"PROFILER" help:"Profiler to use."`
	LogLevel      string `name:"log-level" env:"LOG_LEVEL" default:"info" help:"Set log level."`
	LogJSON       bool   `name:"log-json" env:"LOG_JSON" default:"false" help:"Enable JSON logging output.'"`
	LogCaller     bool   `name:"log-caller" env:"LOG_CALLER" default:"false" help:"Add file:line of the caller to log output."`
	LogNoColor    bool   `name:"log-nocolor" env:"LOG_NOCOLOR" default:"false" help:"Disables the colorized output."`
	GRPCAuthority string `name:"grpc-authority" env:"GRPC_AUTHORITY" default:":42286" help:"Address used to expose the gRPC server."`
}

func (s *ServeCmd) Run(ctx *Context) error {
	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logging.Configure(logging.Options{
		LogLevel:   s.LogLevel,
		LogJSON:    s.LogJSON,
		LogCaller:  s.LogCaller,
		LogNoColor: s.LogNoColor,
	})
	log.Info().Str("version", version).Msgf("Starting %s", ctx.Meta.Name)

	cfg, err := config.Load(s.Cfgfile)
	if err != nil {
		return errors.Wrap(err, "cannot load configuration")
	}
	log.Debug().Interface("config", cfg).Msg("Configuration")

	if len(s.Profiler) > 0 && len(s.ProfilerPath) > 0 {
		profilerPath := path.Clean(s.ProfilerPath)
		if err = os.MkdirAll(profilerPath, os.ModePerm); err != nil {
			return errors.Wrap(err, "cannot create profiler folder")
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
			return errors.Errorf("unknown profiler: %s", s.Profiler)
		}
	}

	diun, err := app.New(ctx.Meta, cfg, s.GRPCAuthority)
	if err != nil {
		return errors.Wrapf(err, "cannot initialize %s", ctx.Meta.Name)
	}

	if err = diun.Start(runCtx); err != nil {
		return errors.Wrapf(err, "cannot start %s", ctx.Meta.Name)
	}

	if cause := context.Cause(runCtx); cause != nil {
		log.Warn().Msg(strings.Title(cause.Error())) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support
	}

	return nil
}
