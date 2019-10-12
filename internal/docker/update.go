package docker

import (
	"context"
	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/model"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var queued = false
var working = false
var workingMutex = sync.Mutex{}

func update(cfg *config.Config, cli *client.Client) error {
	if working {
		queued = true
	}
	working = true
	workingMutex.Lock()

	// Filter containers so we only get relevant ones
	args := filters.NewArgs()
	if !cfg.Watch.StoppedContainers {
		args.Add("status", "running")
		args.Add("status", "restarting")
		args.Add("status", "paused")
	}
	if !cfg.Watch.UnlabeledContainers {
		args.Add("label", "diun")
		args.Add("label", "diun.enable=true")
	}

	time.Sleep(2 * time.Second)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: args})
	if err != nil {
		return err
	}

	// watched contains true for all existing containers that are being watched, for easier lookup when removing old images
	watched := map[string]bool{}

	countAdd := 0
	countUpd := 0
	countDel := 0
	countErr := 0
	log.Debug().Msgf("Old Config: %+v", cfg.Image)

	for _, container := range containers {
		if container.Labels["diun.enable"] == "false" {
			continue
		}

		watched[container.ID] = true

		found := false
		for i := range cfg.Image {
			if cfg.Image[i].SourceContainer != "" && cfg.Image[i].SourceContainer == container.ID {
				// Image exists, update the entry
				found = true
				if err := cfg.SetImage(i, container2image(container)); err != nil {
					countErr++
					log.Error().Str("container", name(container.Names, container.ID)).Err(err).Msg("Invalid configuration")
				} else {
					countUpd++
				}
				break
			}
		}
		if !found {
			// Image doesn't exist, create the entry
			if err := cfg.AddImage(container2image(container)); err != nil {
				countErr++
				log.Error().Str("container", name(container.Names, container.ID)).Err(err).Msg("Invalid configuration")
			} else {
				countAdd++
			}
		}
	}

	for i := 0; i < len(cfg.Image); i++ {
		if cfg.Image[i].SourceContainer != "" && !watched[cfg.Image[i].SourceContainer] {
			// Container doesn't exist, remove the entry
			cfg.RemoveImage(i)
			countDel++
			i--
		}
	}

	if countErr > 0 || countAdd > 0 || countDel > 0 || countUpd > 0 {
		log.Info().Msgf("Containers updated: %d added, %d unmodified/updated, %d removed, %d failed", countAdd, countUpd, countDel, countErr)
	}

	working = false
	if queued {
		return update(cfg, cli)
	}
	workingMutex.Unlock()
	return nil
}

func container2image(container types.Container) model.Image {
	img := model.Image{
		SourceContainer: container.ID,
		Name:            container.Image,
		Os:              container.Labels["diun.os"],
		Arch:            container.Labels["diun.arch"],
		RegOptsID:       container.Labels["diun.regopts_id"],
		WatchRepo:       container.Labels["diun.watch_repo"] == "true",
		MaxTags:         0,
		IncludeTags:     []string{},
		ExcludeTags:     []string{},
	}
	if maxTags, ok := container.Labels["diun.max_tags"]; ok {
		var err error
		img.MaxTags, err = strconv.Atoi(maxTags)
		if err != nil {
			log.Warn().Str("container", name(container.Names, container.ID)).Err(err).Msg("diun.max_tags is not an integer, using default value 0")
		}
	}
	for l, v := range container.Labels {
		if l == "diun.include_tags" || strings.HasPrefix(l, "diun.include_tags.") || (l == "diun" && v != "true" && v != "false" && v != "") {
			if _, ok := container.Labels["diun.watch_repo"]; !ok {
				img.WatchRepo = true
			}
			img.IncludeTags = append(img.IncludeTags, v)
		}
		if l == "diun.exclude_tags" || strings.HasPrefix(l, "diun.exclude_tags.") {
			img.ExcludeTags = append(img.ExcludeTags, v)
		}
	}

	return img
}

func name(n []string, id string) string {
	if len(n) > 0 {
		return n[0]
	} else {
		return id[:11]
	}
}
