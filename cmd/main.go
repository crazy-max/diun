package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	_ "time/tzdata"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	version = "dev"
	cli     struct {
		Version kong.VersionFlag
		Serve   ServeCmd `kong:"cmd,help='Starts Diun server.'"`
		Image   ImageCmd `kong:"cmd,help='Manage image manifests.'"`
		Notif   NotifCmd `kong:"cmd,help='Manage notifications.'"`
	}
)

type Context struct {
	Meta model.Meta
}

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	meta := model.Meta{
		ID:      "diun",
		Name:    "Diun",
		Desc:    "Docker image update notifier",
		URL:     "https://github.com/crazy-max/diun",
		Logo:    "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
		Author:  "CrazyMax",
		Version: version,
	}
	meta.UserAgent = fmt.Sprintf("%s/%s go/%s %s", meta.ID, meta.Version, runtime.Version()[2:], strings.Title(runtime.GOOS))
	if meta.Hostname, err = os.Hostname(); err != nil {
		log.Fatal().Err(err).Msg("Cannot resolve hostname")
	}

	ctx := kong.Parse(&cli,
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

	ctx.FatalIfErrorf(ctx.Run(&Context{Meta: meta}))
}
