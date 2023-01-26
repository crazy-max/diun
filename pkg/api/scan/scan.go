package metrics

import (
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

// New is a factory function creating a new  Handler instance
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
