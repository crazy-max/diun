package gotgbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultAPIURL is the default telegram API URL.
	DefaultAPIURL = "https://api.telegram.org"
	// DefaultTimeout is the default timeout to be set for all requests.
	DefaultTimeout = time.Second * 5
)

type BotClient interface {
	// RequestWithContext submits a POST HTTP request a bot API instance.
	RequestWithContext(ctx context.Context, token string, method string, params map[string]any, opts *RequestOpts) (json.RawMessage, error)
	// GetAPIURL gets the URL of the API either in use by the bot or defined in the request opts.
	GetAPIURL(opts *RequestOpts) string
	// FileURL gets the URL of a file at the API address that the bot is interacting with.
	FileURL(token string, tgFilePath string, opts *RequestOpts) string
}

var _ BotClient = &BaseBotClient{}

type BaseBotClient struct {
	// Client is the HTTP Client used for all HTTP requests made for this bot.
	Client http.Client
	// UseTestEnvironment defines whether this bot was created to run on telegram's test environment.
	// Enabling this uses a slightly different API path.
	// See https://core.telegram.org/bots/webapps#using-bots-in-the-test-environment for more details.
	UseTestEnvironment bool
	// Default opts to use for all requests, when no other request opts are specified.
	DefaultRequestOpts *RequestOpts
}

type Response struct {
	// Ok: if true, request was successful, and result can be found in the Result field.
	// If false, error can be explained in the Description.
	Ok bool `json:"ok"`
	// Result: result of requests (if Ok)
	Result json.RawMessage `json:"result"`
	// ErrorCode: Integer error code of request. Subject to change in the future.
	ErrorCode int `json:"error_code"`
	// Description: contains a human readable description of the error result.
	Description string `json:"description"`
	// Parameters: Optional extra data which can help automatically handle the error.
	Parameters *ResponseParameters `json:"parameters"`
}

type TelegramError struct {
	// The telegram method which raised the error.
	Method string
	// The HTTP parameters which raised the error.
	Params map[string]any
	// The error code returned by telegram.
	Code int
	// The error description returned by telegram.
	Description string
	// The additional parameters returned by telegram
	ResponseParams *ResponseParameters
}

func (t *TelegramError) Error() string {
	return fmt.Sprintf("unable to %s: %s", t.Method, t.Description)
}

// RequestOpts defines any request-specific options used to interact with the telegram API.
type RequestOpts struct {
	// Timeout for the HTTP request to the telegram API.
	Timeout time.Duration
	// Custom API URL to use for requests.
	APIURL string
	// OverrideParams can be used to override existing parameters, or override existing ones.
	OverrideParams map[string]any
}

// getTimeoutContext returns the appropriate context for the current settings.
func (bot *BaseBotClient) getTimeoutContext(parentCtx context.Context, opts *RequestOpts) (context.Context, context.CancelFunc) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}

	if opts != nil {
		ctx, cancelFunc := timeoutFromOpts(parentCtx, opts)
		if ctx != nil {
			return ctx, cancelFunc
		}
	}

	if bot.DefaultRequestOpts != nil {
		ctx, cancelFunc := timeoutFromOpts(parentCtx, bot.DefaultRequestOpts)
		if ctx != nil {
			return ctx, cancelFunc
		}
	}

	return context.WithTimeout(parentCtx, DefaultTimeout) //nolint:gosec
}

func timeoutFromOpts(parentCtx context.Context, opts *RequestOpts) (context.Context, context.CancelFunc) {
	// nothing? no timeout.
	if opts == nil {
		return nil, nil
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}

	if opts.Timeout > 0 {
		return context.WithTimeout(parentCtx, opts.Timeout) //nolint:gosec

	} else if opts.Timeout < 0 {
		// < 0  no timeout; infinite.
		return parentCtx, func() {}
	}
	// 0 == nothing defined, use defaults.
	return nil, nil
}

// RequestWithContext allows sending a POST request to the telegram bot API with an existing context.
//   - ctx: the timeout contexts to be used.
//   - method: the telegram API method to call.
//   - params: map of parameters to be sending to the telegram API. eg: chat_id, user_id, etc.
//   - data: map of any files to be sending to the telegram API.
//   - opts: request opts to use.
func (bot *BaseBotClient) RequestWithContext(parentCtx context.Context, token string, method string, params map[string]any, opts *RequestOpts) (json.RawMessage, error) {
	ctx, cancel := bot.getTimeoutContext(parentCtx, opts)
	defer cancel()

	if opts != nil {
		maps.Copy(params, opts.OverrideParams)
	}

	req, err := bot.buildRequest(ctx, params, token, method, opts)
	if err != nil {
		return nil, err
	}

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute POST request to %s: %w", method, sanitizeError(token, err))
	}
	defer resp.Body.Close()

	var r Response
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to decode POST request to %s: %w", method, err)
	}

	if !r.Ok {
		return nil, &TelegramError{
			Method:         method,
			Params:         params,
			Code:           r.ErrorCode,
			Description:    r.Description,
			ResponseParams: r.Parameters,
		}
	}

	return r.Result, nil
}

func allFilesSeekable(params map[string]any) bool {
	for _, v := range params {
		if f, ok := v.(*FileReader); ok && f.Data != nil {
			if _, ok := f.Data.(io.Seeker); !ok {
				return false
			}
		}
	}
	return true
}

func (bot *BaseBotClient) buildRequest(ctx context.Context, params map[string]any, token string, method string, opts *RequestOpts) (*http.Request, error) {
	body, contentType := buildMultipart(ctx, params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bot.methodEndpoint(token, method, opts), body)
	if err != nil {
		return nil, fmt.Errorf("failed to build POST request to %s: %w", method, err)
	}

	req.Header.Set("Content-Type", contentType)

	if allFilesSeekable(params) {
		req.GetBody = func() (io.ReadCloser, error) {
			retryBody, contentType := buildMultipart(ctx, params)
			req.Header.Set("Content-Type", contentType)
			return io.NopCloser(retryBody), nil
		}
	}
	return req, nil
}

// buildMultipart creates a lazy multipart reader/writer which only writes while it gets read.
// If the ctx gets closed, so does this.
func buildMultipart(ctx context.Context, params map[string]any) (io.Reader, string) {
	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)

	go func() {
		done := make(chan struct{})

		go func() {
			select {
			case <-ctx.Done():
				// If the context is done before we finish writing, abort the pipe.
				// This unblocks any pw.Write() call and lets the goroutine exit.
				pw.CloseWithError(ctx.Err())
			case <-done:
				// writer finished cleanly, nothing to do
			}
		}()
		defer close(done)

		if len(params) == 0 {
			if err := w.WriteField("_empty", ""); err != nil {
				pw.CloseWithError(fmt.Errorf("failed to write empty multipart field: %w", err))
				return
			}
		}

		for k, v := range params {
			contents, err := getFieldContents(v, k, w)
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			if err := w.WriteField(k, contents); err != nil {
				pw.CloseWithError(fmt.Errorf("failed to write multipart field %s with value %v: %w", k, v, err))
				return
			}
		}

		if err := w.Close(); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to close multipart form writer: %w", err))
			return
		}
		pw.Close()
	}()

	return pr, w.FormDataContentType()
}

// Sanitize the error to avoid token leak.
func sanitizeError(token string, err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		urlErr.URL = strings.ReplaceAll(urlErr.URL, token, "<TOKEN>")
		return urlErr
	}

	return err
}

func getFieldContents(v any, k string, w *multipart.Writer) (string, error) {
	// Check if the value is a simple type that can be converted directly
	switch val := v.(type) {
	case string:
		return val, nil

	case *string:
		if val == nil {
			return "", nil
		}
		return *val, nil

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return fmt.Sprint(val), nil

	case Attach:
		err := val.Attach(k, w)
		if err != nil {
			return "", fmt.Errorf("failed to read input multipart field: %w", err)
		}

		// In case of a simple inputfile attachment, rely on files
		if inputFile, ok := val.(InputFile); ok {
			return inputFile.getValue(), nil
		}

		// For complex types (structs, maps, slices, etc.), marshal as JSON
		bs, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("failed to marshal field %s to JSON: %w", k, err)
		}
		return string(bs), nil

	default:
		// For complex types (structs, maps, slices, etc.), marshal as JSON
		bs, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("failed to marshal field %s to JSON: %w", k, err)
		}
		return string(bs), nil
	}
}

// GetAPIURL returns the currently used API endpoint.
func (bot *BaseBotClient) GetAPIURL(opts *RequestOpts) string {
	if opts != nil && opts.APIURL != "" {
		return strings.TrimSuffix(opts.APIURL, "/")
	}
	if bot.DefaultRequestOpts != nil && bot.DefaultRequestOpts.APIURL != "" {
		return strings.TrimSuffix(bot.DefaultRequestOpts.APIURL, "/")
	}
	return DefaultAPIURL
}

func (bot *BaseBotClient) FileURL(token string, tgFilePath string, opts *RequestOpts) string {
	return fmt.Sprintf("%s/file/%s/%s", bot.GetAPIURL(opts), bot.getEnvAuth(token), tgFilePath)
}

func (bot *BaseBotClient) getEnvAuth(token string) string {
	if bot.UseTestEnvironment {
		return "bot" + token + "/test"
	}
	return "bot" + token
}

func (bot *BaseBotClient) methodEndpoint(token string, method string, opts *RequestOpts) string {
	return fmt.Sprintf("%s/%s/%s", bot.GetAPIURL(opts), bot.getEnvAuth(token), method)
}
