package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	_ "time/tzdata"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var version = "dev"

type cli struct {
	Version     kong.VersionFlag `name:"version" help:"Print version information."`
	Serve       ServeCmd         `cmd:"" help:"Starts Diun server."`
	Healthcheck HealthcheckCmd   `cmd:"" help:"Check Diun health."`
	Image       ImageCmd         `cmd:"" help:"Manage image manifests."`
	Notif       NotifCmd         `cmd:"" help:"Manage notifications."`
}

type Context struct {
	Meta model.Meta
}

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errHealthcheckFailed) {
			os.Exit(1)
		}
		log.Fatal().Err(err).Send()
	}
}

func run() error {
	meta := model.Meta{
		ID:      "diun",
		Name:    "Diun",
		Desc:    "Docker image update notifier",
		URL:     "https://github.com/crazy-max/diun",
		Logo:    "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
		Author:  "CrazyMax",
		Version: version,
	}
	meta.UserAgent = fmt.Sprintf("%s/%s go/%s %s", meta.ID, meta.Version, runtime.Version()[2:], strings.Title(runtime.GOOS)) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support

	var err error
	if meta.Hostname, err = os.Hostname(); err != nil {
		return errors.Wrap(err, "cannot resolve hostname")
	}

	cmd := cli{}
	kctx := kong.Parse(&cmd,
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

	return kctx.Run(&Context{Meta: meta})
}
