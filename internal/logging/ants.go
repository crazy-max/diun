package logging

import (
	"github.com/rs/zerolog/log"
)

// AntsLogger is a logger for ants module
type AntsLogger struct{}

// Printf must have the same semantics as log.Printf
func (w *AntsLogger) Printf(format string, args ...interface{}) {
	log.Debug().Msgf(format, args...)
}
