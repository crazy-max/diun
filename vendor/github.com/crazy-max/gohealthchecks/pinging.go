package gohealthchecks

import (
	"context"
	"fmt"
	"net/http"
)

// PingingOptions holds parameters for the Pinging API.
type PingingOptions struct {
	UUID string
	Logs string
}

// Success sends a success request to Healthchecks to indicate that a job has completed.
func (c *Client) Success(ctx context.Context, po PingingOptions) (err error) {
	return c.request(ctx, http.MethodPost, po.UUID, []byte(po.Logs))
}

// Fail sends a fail request to Healthchecks to indicate that an error has occurred.
func (c *Client) Fail(ctx context.Context, po PingingOptions) (err error) {
	return c.request(ctx, http.MethodPost, fmt.Sprintf("%s/fail", po.UUID), []byte(po.Logs))
}

// Start sends a start request to Healthchecks to indicate that a job has started.
func (c *Client) Start(ctx context.Context, po PingingOptions) (err error) {
	return c.request(ctx, http.MethodPost, fmt.Sprintf("%s/start", po.UUID), []byte(po.Logs))
}
