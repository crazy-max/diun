package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/crazy-max/diun/v4/internal/httputil"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/internal/secret"
	"github.com/pkg/errors"
)

const (
	teamsMaxRateLimitAttempts = 3
	teamsRateLimitMessage     = "Microsoft Teams endpoint returned HTTP error 429" // https://learn.microsoft.com/en-gb/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using?tabs=cURL%2Ctext1#rate-limiting-for-connectors
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifTeams
	meta model.Meta
}

// New creates a new webhook notification instance
func New(config *model.NotifTeams, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "teams"
}

// Sections is grouping data together containing title, subtitle and facts and creating a nested json element
type Sections struct {
	ActivityTitle    string `json:"activityTitle"`
	ActivitySubtitle string `json:"activitySubtitle"`
	Facts            []Fact `json:"facts"`
}

// Fact is grouping data together to create a nested json element containing a name and an associated value
type Fact struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type messageCardPayload struct {
	Type       string     `json:"@type"`
	Context    string     `json:"@context"`
	ThemeColor string     `json:"themeColor"`
	Summary    string     `json:"summary"`
	Sections   []Sections `json:"sections"`
}

type adaptiveCardPayload struct {
	Type        string                   `json:"type"`
	Attachments []adaptiveCardAttachment `json:"attachments"`
}

type adaptiveCardAttachment struct {
	ContentType string              `json:"contentType"`
	Content     adaptiveCardContent `json:"content"`
}

type adaptiveCardContent struct {
	Schema  string                `json:"$schema"`
	Type    string                `json:"type"`
	Version string                `json:"version"`
	Body    []adaptiveCardElement `json:"body"`
}

type adaptiveCardElement struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	Wrap     bool               `json:"wrap,omitempty"`
	Size     string             `json:"size,omitempty"`
	Weight   string             `json:"weight,omitempty"`
	Color    string             `json:"color,omitempty"`
	IsSubtle bool               `json:"isSubtle,omitempty"`
	Spacing  string             `json:"spacing,omitempty"`
	Facts    []adaptiveCardFact `json:"facts,omitempty"`
}

type adaptiveCardFact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	webhookURL, err := secret.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve webhook URL for Teams notifier")
	}

	message, err := msg.New(msg.Options{
		Meta:         c.meta,
		Entry:        entry,
		TemplateBody: c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	_, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	jsonBody, err := c.payload(entry, string(body))
	if err != nil {
		return err
	}

	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Teams notifier")
	}

	for attempt := 1; attempt <= teamsMaxRateLimitAttempts; attempt++ {
		cancelCtx, cancel := context.WithCancelCause(context.Background())
		timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent

		req, err := http.NewRequestWithContext(timeoutCtx, "POST", webhookURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			cancel(errors.WithStack(context.Canceled))
			return err
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.meta.UserAgent)

		resp, err := hc.Do(req)
		if err != nil {
			cancel(errors.WithStack(context.Canceled))
			return err
		}

		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); err == nil {
			err = closeErr
		}
		cancel(errors.WithStack(context.Canceled))
		if err != nil {
			return errors.Wrap(err, "cannot read Teams response")
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices && !teamsRateLimited(resp, body) {
			return nil
		}

		err = teamsResponseError(resp, body)
		if !teamsRateLimited(resp, body) || attempt == teamsMaxRateLimitAttempts {
			return err
		}

		time.Sleep(teamsRetryAfter(resp, attempt))
	}

	return nil
}

func (c *Client) payload(entry model.NotifEntry, body string) ([]byte, error) {
	facts := c.facts(entry)
	if c.cfg.CardType == model.NotifTeamsCardTypeAdaptiveCard {
		return json.Marshal(c.adaptiveCardPayload(entry, body, facts))
	}
	return json.Marshal(c.messageCardPayload(entry, body, facts))
}

func (c *Client) facts(entry model.NotifEntry) []Fact {
	if !*c.cfg.RenderFacts {
		return nil
	}

	return []Fact{
		{"Hostname", c.meta.Hostname},
		{"Provider", entry.Provider},
		{"Created", entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC")},
		{"Digest", entry.Manifest.Digest.String()},
		{"Platform", entry.Manifest.Platform},
	}
}

func (c *Client) messageCardPayload(entry model.NotifEntry, body string, facts []Fact) messageCardPayload {
	themeColor := "68CA00"
	if entry.Status == model.ImageStatusUpdate {
		themeColor = "0076D7"
	}

	return messageCardPayload{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		ThemeColor: themeColor,
		Summary:    body,
		Sections: []Sections{
			{
				ActivityTitle:    body,
				ActivitySubtitle: fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
				Facts:            facts,
			},
		},
	}
}

func (c *Client) adaptiveCardPayload(entry model.NotifEntry, body string, facts []Fact) adaptiveCardPayload {
	color := "Good"
	if entry.Status == model.ImageStatusUpdate {
		color = "Accent"
	}

	elements := []adaptiveCardElement{
		{
			Type:   "TextBlock",
			Text:   body,
			Wrap:   true,
			Weight: "Bolder",
			Color:  color,
		},
	}
	if len(facts) > 0 {
		adaptiveFacts := make([]adaptiveCardFact, 0, len(facts))
		for _, fact := range facts {
			adaptiveFacts = append(adaptiveFacts, adaptiveCardFact{
				Title: fact.Name,
				Value: fact.Value,
			})
		}
		elements = append(elements, adaptiveCardElement{
			Type:  "FactSet",
			Facts: adaptiveFacts,
		})
	}
	elements = append(elements, adaptiveCardElement{
		Type:     "TextBlock",
		Text:     fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
		Wrap:     true,
		Size:     "Small",
		IsSubtle: true,
		Spacing:  "Small",
	})

	return adaptiveCardPayload{
		Type: "message",
		Attachments: []adaptiveCardAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: adaptiveCardContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.2",
					Body:    elements,
				},
			},
		},
	}
}

func teamsRateLimited(resp *http.Response, body []byte) bool {
	return resp.StatusCode == http.StatusTooManyRequests || strings.Contains(string(body), teamsRateLimitMessage)
}

func teamsResponseError(resp *http.Response, body []byte) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return errors.Errorf("unexpected Teams response: %s", string(body))
	}
	return errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
}

// https://learn.microsoft.com/en-us/microsoftteams/platform/bots/build-conversational-capability?tabs=dotnet%2Capp-manifest-v112-or-later%2Cdotnet2%2Cdotnet3%2Cdotnet4%2Cdotnet5%2Ccsharp2%2Cdotnet6%2Ccsharp1#status-codes-from-bot-conversational-apis
func teamsRetryAfter(resp *http.Response, attempt int) time.Duration {
	if value := resp.Header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second
		}
		if retryAt, err := http.ParseTime(value); err == nil {
			delay := time.Until(retryAt)
			if delay > 0 {
				return delay
			}
			return 0
		}
	}

	return time.Duration(1<<uint(attempt-1)) * time.Second
}
