package docker

import (
	"context"
	"fmt"
	"github.com/crazy-max/diun/internal/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"time"
)

// New creates and starts a new docker watcher
func New(cfg *config.Config) error {
	if !cfg.Watch.Docker {
		// The feature is disabled
		stopped = true
		return nil
	}

	// Try to connect to the docker daemon
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		stopped = true
		return fmt.Errorf("couldn't connect to docker daemon: %v", err)
	}

	log.Info().Msg("Connection to docker container established")

	update(cfg, cli)
	go watch(cfg, cli)
	return nil
}

var shouldStop = make(chan bool)
var hasStopped = make(chan bool)
var stopped = false

func Stop() {
	// Stop the update loop, notify watch() that it should stop, and wait until it has stopped
	if stopped {
		return
	}
	stopped = true
	select {
	case shouldStop <- true:
		<-hasStopped
	case <-hasStopped:
	}
}

func watch(cfg *config.Config, cli *client.Client) {
	// Filter events so we only get relevant ones
	args := filters.NewArgs()
	args.Add("type", "container")
	if cfg.Watch.StoppedContainers {
		args.Add("event", "create")
		args.Add("event", "destroy")
	} else {
		args.Add("event", "start")
		args.Add("event", "stop")
	}
	if !cfg.Watch.UnlabeledContainers {
		args.Add("label", "diun")
		args.Add("label", "diun.enable")
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	for !stopped {
		// Wait for the relevant events from the docker daemon
		events, dockerErrors := cli.Events(ctx, types.EventsOptions{Filters: args})
		log.Debug().Msg("Now listening to docker events")
		for !stopped {
			select {

			case <-events:
				// An update will be done during the next iteration of the inner for loop
				err := update(cfg, cli)
				for ; err != nil; err = update(cfg, cli) {
					// The update function threw an error
					log.Error().Err(err).Msg("Error when updating watched docker containers, waiting 30 seconds before trying again")
					time.Sleep(30 * time.Second)
				}

			case err := <-dockerErrors:
				// The docker events channel threw an error
				log.Error().Err(err).Msg("Error in docker event channel, waiting 30 seconds before trying again")
				time.Sleep(30 * time.Second)
				break // Recreate the events channel in the outer for loop

			case <-shouldStop:
				// `stopped` is now true, so we'll quit both for loops
				break

			}
		}
	}
	// Cancel the docker event context and notify the Close() function that everything has been stopped
	ctxCancel()
	hasStopped <- true
}
