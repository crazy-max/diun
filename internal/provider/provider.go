package provider

import (
	"github.com/crazy-max/diun/v4/internal/model"
)

// Handler is a provider interface
type Handler interface {
	ListJob() []model.Job
}

// Client represents an active provider object
type Client struct {
	Handler
}

// WalkJobs calls fn for every job returned by providers.
func WalkJobs(fn func(model.Job), providers ...*Client) {
	for _, prd := range providers {
		for _, job := range prd.ListJob() {
			fn(job)
		}
	}
}
