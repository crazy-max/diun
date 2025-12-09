package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Client represents an active elasticsearch notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifElasticsearch
	meta model.Meta
}

// New creates a new elasticsearch notification instance
func New(config *model.NotifElasticsearch, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "elasticsearch"
}

// Send creates and sends an elasticsearch notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	username, err := utl.GetValueOrFileContents(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		return err
	}

	password, err := utl.GetValueOrFileContents(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		return err
	}

	// Use the same JSON structure as webhook notifier
	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	body, err := message.RenderJSON()
	if err != nil {
		return err
	}

	// Parse the JSON to add the client field
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}

	// Add the current time
	doc["@timestamp"] = time.Now().Format(time.RFC3339Nano)

	// Add the client field from the configuration
	doc["client"] = c.cfg.Client

	// Re-marshal the JSON with the client field
	body, err = json.Marshal(doc)
	if err != nil {
		return err
	}

	// Build the Elasticsearch indexing URL
	// This uses the Index API (POST /{index}/_doc) to create a document with an auto-generated _id:
	// https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-create
	u, err := url.Parse(c.cfg.Address)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, c.cfg.Index, "_doc")

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	tlsConfig, err := utl.LoadTLSConfig(c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot load TLS configuration for Elasticsearch notifier")
	}
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequestWithContext(timeoutCtx, "POST", u.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	// Add authentication if provided
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errBody struct {
			Status int `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return errors.Wrapf(err, "cannot decode JSON error response for HTTP %d %s status",
				resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return errors.Errorf("%d %s: %s", errBody.Status, errBody.Error.Type, errBody.Error.Reason)
	}

	return nil
}
