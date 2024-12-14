package gotgbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	RequestWithContext(ctx context.Context, token string, method string, params map[string]string, data map[string]FileReader, opts *RequestOpts) (json.RawMessage, error)
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
	Params map[string]string
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

	return context.WithTimeout(parentCtx, DefaultTimeout)
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
		return context.WithTimeout(parentCtx, opts.Timeout)

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
func (bot *BaseBotClient) RequestWithContext(parentCtx context.Context, token string, method string, params map[string]string, data map[string]FileReader, opts *RequestOpts) (json.RawMessage, error) {
	ctx, cancel := bot.getTimeoutContext(parentCtx, opts)
	defer cancel()

	var requestBody io.Reader

	var contentType string
	// Check if there are any files to upload. If yes, use multipart; else, use JSON.
	if len(data) > 0 {
		pr, pw := io.Pipe()
		defer pr.Close() // avoid writer goroutine leak
		mw := multipart.NewWriter(pw)
		contentType = mw.FormDataContentType()
		requestBody = pr
		// Write the request data asynchronously from another goroutine
		// to the multipart.Writer which will be piped into the pipe reader
		// which is tied to the request to be sent
		go func() {
			writerError := fillBuffer(mw, params, data)
			// Close the writer with error of multipart writer.
			// If the error is nil, this will act just like pw.Close()
			_ = pw.CloseWithError(writerError)
		}()
	} else {
		contentType = "application/json"
		bodyBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode parameters as JSON: %w", err)
		}
		requestBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bot.methodEndpoint(token, method, opts), requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build POST request to %s: %w", method, err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute POST request to %s: %w", method, err)
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

// Fill the buffer of multipart.Writer with data which is going to be sent.
func fillBuffer(w *multipart.Writer, params map[string]string, data map[string]FileReader) error {
	for k, v := range params {
		err := w.WriteField(k, v)
		if err != nil {
			return fmt.Errorf("failed to write multipart field %s with value %s: %w", k, v, err)
		}
	}

	for field, file := range data {
		fileName := file.Name
		if fileName == "" {
			fileName = field
		}

		part, err := w.CreateFormFile(field, fileName)
		if err != nil {
			return fmt.Errorf("failed to create form file for field %s and fileName %s: %w", field, fileName, err)
		}

		_, err = io.Copy(part, file.Data)
		if err != nil {
			return fmt.Errorf("failed to copy file contents of field %s to form: %w", field, err)
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart form writer: %w", err)
	}

	return nil
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
