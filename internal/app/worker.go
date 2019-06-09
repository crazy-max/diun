package app

import (
	"sync"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/rs/zerolog/log"
)

type Job struct {
	ImageStr string
	Item     model.Item
	Reg      *docker.RegistryClient
	Wg       *sync.WaitGroup
}

type worker struct {
	id         int
	diun       *Diun
	workerPool chan chan Job
	jobChannel chan Job
	end        chan bool
}

// Start method starts the run loop for the worker
func (w *worker) Start() {
	go func() {
		for {
			w.workerPool <- w.jobChannel
			select {
			case job := <-w.jobChannel:
				if err := w.diun.analyze(job, w.id); err != nil {
					log.Error().Err(err).
						Str("image", job.ImageStr).
						Int("worker_id", w.id).
						Msg("Error analyzing image")
				}
			case <-w.end:
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w *worker) Stop() {
	w.end <- true
}
