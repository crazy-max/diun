package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	_ "time/tzdata"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"
)

var (
	diun    *app.Diun
	cli     model.Cli
	version = "dev"
	meta    = model.Meta{
		ID:     "diun",
		Name:   "Diun",
		Desc:   "Docker image update notifier",
		URL:    "https://github.com/crazy-max/diun",
		Logo:   "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
		Author: "CrazyMax",
	}
)

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	meta.Version = version
	meta.UserAgent = fmt.Sprintf("%s/%s go/%s %s", meta.ID, meta.Version, runtime.Version()[2:], strings.Title(runtime.GOOS))
	if meta.Hostname, err = os.Hostname(); err != nil {
		log.Fatal().Err(err).Msg("Cannot resolve hostname")
	}

	// Parse command line
	_ = kong.Parse(&cli,
		kong.Name(meta.ID),
		kong.Description(fmt.Sprintf("%s. More info: %s", meta.Desc, meta.URL)),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	// Init
	logging.Configure(&cli)
	log.Info().Str("version", version).Msgf("Starting %s", meta.Name)

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
	cfg, err := config.Load(cli)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	log.Debug().Msg(cfg.String())

	// Profiler
	if len(cli.Profiler) > 0 && len(cli.ProfilerPath) > 0 {
		profilerPath := path.Clean(cli.ProfilerPath)
		if err = os.MkdirAll(profilerPath, os.ModePerm); err != nil {
			log.Fatal().Err(err).Msg("Cannot create profiler folder")
		}
		profilePath := profile.ProfilePath(profilerPath)
		switch cli.Profiler {
		case "cpu":
			defer profile.Start(profile.CPUProfile, profilePath).Stop()
		case "mem":
			defer profile.Start(profile.MemProfile, profilePath).Stop()
		case "alloc":
			defer profile.Start(profile.MemProfile, profile.MemProfileAllocs(), profilePath).Stop()
		case "heap":
			defer profile.Start(profile.MemProfile, profile.MemProfileHeap(), profilePath).Stop()
		case "routines":
			defer profile.Start(profile.GoroutineProfile, profilePath).Stop()
		case "mutex":
			defer profile.Start(profile.MutexProfile, profilePath).Stop()
		case "threads":
			defer profile.Start(profile.ThreadcreationProfile, profilePath).Stop()
		case "block":
			defer profile.Start(profile.BlockProfile, profilePath).Stop()
		default:
			log.Fatal().Msgf("Unknown profiler: %s", cli.Profiler)
		}
	}

	// Init
	if diun, err = app.New(meta, cli, cfg); err != nil {
		log.Fatal().Err(err).Msgf("Cannot initialize %s", meta.Name)
	}

	// Test notif
	if cli.TestNotif {
		diun.TestNotif()
		return
	}

	// Start
	if err = diun.Start(); err != nil {
		log.Fatal().Err(err).Msgf("Cannot start %s", meta.Name)
	}
}
