package app

import (
	"bytes"
	"context"
	"net/url"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/secret"
	"github.com/crazy-max/gohealthchecks"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type healthchecksClient struct {
	*gohealthchecks.Client
	UUID string
}

func newHealthchecksClient(cfg *model.Healthchecks) (*healthchecksClient, error) {
	var baseURL *url.URL
	if len(cfg.BaseURL) > 0 {
		var err error
		baseURL, err = url.Parse(cfg.BaseURL)
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse Healthchecks base URL")
		}
	}
	uuid, err := secret.GetSecret(cfg.UUID, cfg.UUIDFile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot retrieve Healthchecks UUID")
	}
	return &healthchecksClient{
		Client: gohealthchecks.NewClient(&gohealthchecks.ClientOptions{
			BaseURL: baseURL,
		}),
		UUID: uuid,
	}, nil
}

func (c *healthchecksClient) Start(ctx context.Context) error {
	return c.Client.Start(ctx, gohealthchecks.PingingOptions{
		UUID: c.UUID,
	})
}

func (c *healthchecksClient) Success(ctx context.Context, logs string) error {
	return c.Client.Success(ctx, gohealthchecks.PingingOptions{
		UUID: c.UUID,
		Logs: logs,
	})
}

func (c *healthchecksClient) Fail(ctx context.Context, logs string) error {
	return c.Client.Fail(ctx, gohealthchecks.PingingOptions{
		UUID: c.UUID,
		Logs: logs,
	})
}

func (di *Diun) HealthchecksStart() {
	if di.hc == nil {
		return
	}
	if err := di.hc.Start(context.Background()); err != nil {
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
		return
	}
	if err := di.hc.Success(context.Background(), logsBuf.String()); err != nil {
		log.Error().Err(err).Msgf("Cannot send Healthchecks success event")
	}
}

func (di *Diun) HealthchecksFail(logs string) {
	if di.hc == nil {
		return
	}
	if err := di.hc.Fail(context.Background(), logs); err != nil {
		log.Error().Err(err).Msgf("Cannot send Healthchecks fail event")
	}
}
