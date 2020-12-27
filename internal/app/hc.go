package app

import (
	"bytes"
	"context"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/gohealthchecks"
	"github.com/rs/zerolog/log"
)

func (di *Diun) HealthchecksStart() {
	if di.hc == nil {
		return
	}

	if err := di.hc.Start(context.Background(), gohealthchecks.PingingOptions{
		UUID: di.cfg.Watch.Healthchecks.UUID,
	}); err != nil {
		log.Error().Err(err).Msgf("Cannot send Healthchecks start event")
	}
}

func (di *Diun) HealthchecksSuccess(entries *model.NotifEntries) {
	if di.hc == nil {
		return
	}

	var logsBuf bytes.Buffer
	logsTpl := template.Must(template.New("").Parse(`{{ .CountTotal }} tag(s) have been scanned:
* {{ .CountNew }} new tag(s) found
* {{ .CountUpdate }} tag(s) updated
* {{ .CountUnchange }} tag(s) unchanged
* {{ .CountSkip }} tag(s) skipped
* {{ .CountError }} tag(s) with error`))
	if err := logsTpl.Execute(&logsBuf, entries); err != nil {
		log.Error().Err(err).Msgf("Cannot create logs for Healthchecks success event")
	}

	if err := di.hc.Success(context.Background(), gohealthchecks.PingingOptions{
		UUID: di.cfg.Watch.Healthchecks.UUID,
		Logs: logsBuf.String(),
	}); err != nil {
		log.Error().Err(err).Msgf("Cannot send Healthchecks success event")
	}
}

func (di *Diun) HealthchecksFail(logs string) {
	if di.hc == nil {
		return
	}

	if err := di.hc.Fail(context.Background(), gohealthchecks.PingingOptions{
		UUID: di.cfg.Watch.Healthchecks.UUID,
		Logs: logs,
	}); err != nil {
		log.Error().Err(err).Msgf("Cannot send Healthchecks fail event")
	}
}
