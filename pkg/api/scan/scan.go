package metrics

import (
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

var (
	lock chan bool
)

// New is a factory function creating a new  Handler instance
func New2(updateFn func(), updateLock chan bool) *Handler {
	if updateLock != nil {
		lock = updateLock
	} else {
		lock = make(chan bool, 1)
		lock <- true
	}

	return &Handler{
		fn:   updateFn,
		Path: "/v1/scan",
	}
}

func New(scanFn func()) *Handler {
	return &Handler{
		fn:   scanFn,
		Path: "/v1/scan",
	}
}

// Handler is an API handler used for triggering container update scans
type Handler struct {
	fn   func()
	Path string
}

// Handle is the actual http.Handle function doing all the heavy lifting
func (handle *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Updates triggered by HTTP API request.")

	_, err := io.Copy(os.Stdout, r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error")
		return
	}

	handle.fn()
}
