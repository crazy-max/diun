// Package reghttp is used for HTTP requests to a registry
package reghttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/internal/auth"
	"github.com/regclient/regclient/internal/pqueue"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/warning"
)

var (
	defaultDelayInit, _ = time.ParseDuration("0.1s")
	defaultDelayMax, _  = time.ParseDuration("30s")
	warnRegexp          = regexp.MustCompile(`^299\s+-\s+"([^"]+)"`)
)

const (
	DefaultRetryLimit = 5 // number of times a request will be retried
)

// Client is an HTTP client wrapper.
// It handles features like authentication, retries, backoff delays, TLS settings.
type Client struct {
	httpClient    *http.Client              // upstream [http.Client], this is wrapped per repository for an auth handler on redirects
	getConfigHost func(string) *config.Host // call-back to get the [config.Host] for a specific registry
	host          map[string]*clientHost    // host specific settings, wrap access with a mutex lock
	rootCAPool    [][]byte                  // list of root CAs for configuring the http.Client transport
	rootCADirs    []string                  // list of directories for additional root CAs
	retryLimit    int                       // number of retries before failing a request, this applies to each host, and each request
	delayInit     time.Duration             // how long to initially delay requests on a failure
	delayMax      time.Duration             // maximum time to delay a request
	slog          *slog.Logger              // logging for tracing and failures
	userAgent     string                    // user agent to specify in http request headers
	mu            sync.Mutex                // mutex to prevent data races
}

type clientHost struct {
	config      *config.Host                // config entry
	httpClient  *http.Client                // modified http client for registry specific settings
	userAgent   string                      // user agent to specify in http request headers
	slog        *slog.Logger                // logging for tracing and failures
	auth        map[string]*auth.Auth       // map of auth handlers by repository
	backoffLast time.Time                   // time a backoff was last seen, used to deprioritize mirrors for later requests
	reqFreq     time.Duration               // how long between submitting requests for this host
	reqNext     time.Time                   // time to release the next request
	throttle    *pqueue.Queue[reqmeta.Data] // limit concurrent requests to the host
	mu          sync.Mutex                  // mutex to prevent data races
}

// Req is a request to send to a registry.
type Req struct {
	MetaKind    reqmeta.Kind                  // kind of request for the priority queue
	Host        string                        // registry name, hostname and mirrors will be looked up from host configuration
	Method      string                        // http method to call
	DirectURL   *url.URL                      // url to query, overrides repository, path, and query
	Repository  string                        // repository to scope the request
	Path        string                        // path of the request within a repository
	Query       url.Values                    // url query parameters
	BodyLen     int64                         // length of body to send
	BodyBytes   []byte                        // bytes of the body, overridden by BodyFunc
	BodyFunc    func() (io.ReadCloser, error) // function to return a new body
	Headers     http.Header                   // headers to send in the request
	NoPrefix    bool                          // do not include the repository prefix
	NoMirrors   bool                          // do not send request to a mirror
	ExpectLen   int64                         // expected size of the returned body
	TransactLen int64                         // size of an overall transaction for the priority queue
	IgnoreErr   bool                          // ignore http errors and do not trigger backoffs
}

// Resp is used to handle the result of a request.
type Resp struct {
	ctx              context.Context
	client           *Client
	req              *Req
	resp             *http.Response
	mirror           string
	done             bool
	reader           io.Reader
	readCur, readMax int64
	retryCount       int
	backoffCur       int
	backoffLast      time.Time
	throttleDone     func()
}

// Opts is used to configure client options.
type Opts func(*Client)

// NewClient returns a client for handling requests.
func NewClient(opts ...Opts) *Client {
	c := Client{
		httpClient: &http.Client{},
		host:       map[string]*clientHost{},
		retryLimit: DefaultRetryLimit,
		delayInit:  defaultDelayInit,
		delayMax:   defaultDelayMax,
		slog:       slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		rootCAPool: [][]byte{},
		rootCADirs: []string{},
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

// WithCerts adds certificates.
func WithCerts(certs [][]byte) Opts {
	return func(c *Client) {
		c.rootCAPool = append(c.rootCAPool, certs...)
	}
}

// WithCertDirs adds directories to check for host specific certs.
func WithCertDirs(dirs []string) Opts {
	return func(c *Client) {
		c.rootCADirs = append(c.rootCADirs, dirs...)
	}
}

// WithCertFiles adds certificates by filename.
func WithCertFiles(files []string) Opts {
	return func(c *Client) {
		for _, f := range files {
			//#nosec G304 command is run by a user accessing their own files
			cert, err := os.ReadFile(f)
			if err != nil {
				c.slog.Warn("Failed to read certificate",
					slog.String("err", err.Error()),
					slog.String("file", f))
			} else {
				c.rootCAPool = append(c.rootCAPool, cert)
			}
		}
	}
}

// WithConfigHostFn adds the callback to request a [config.Host] struct.
// The function must normalize the hostname for Docker Hub support.
func WithConfigHostFn(gch func(string) *config.Host) Opts {
	return func(c *Client) {
		c.getConfigHost = gch
	}
}

// WithDelay initial time to wait between retries (increased with exponential backoff).
func WithDelay(delayInit time.Duration, delayMax time.Duration) Opts {
	return func(c *Client) {
		if delayInit > 0 {
			c.delayInit = delayInit
		}
		// delayMax must be at least delayInit, if 0 initialize to 30x delayInit
		if delayMax > c.delayInit {
			c.delayMax = delayMax
		} else if delayMax > 0 {
			c.delayMax = c.delayInit
		} else {
			c.delayMax = c.delayInit * 30
		}
	}
}

// WithHTTPClient uses a specific http client with retryable requests.
func WithHTTPClient(hc *http.Client) Opts {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithRetryLimit restricts the number of retries (defaults to 5).
func WithRetryLimit(rl int) Opts {
	return func(c *Client) {
		if rl > 0 {
			c.retryLimit = rl
		}
	}
}

// WithLog injects a slog Logger configuration.
func WithLog(slog *slog.Logger) Opts {
	return func(c *Client) {
		c.slog = slog
	}
}

// WithTransport uses a specific http transport with retryable requests.
func WithTransport(t *http.Transport) Opts {
	return func(c *Client) {
		c.httpClient = &http.Client{Transport: t}
	}
}

// WithUserAgent sets a user agent header.
func WithUserAgent(ua string) Opts {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// Do runs a request, returning the response result.
func (c *Client) Do(ctx context.Context, req *Req) (*Resp, error) {
	resp := &Resp{
		ctx:     ctx,
		client:  c,
		req:     req,
		readCur: 0,
		readMax: req.ExpectLen,
	}
	err := resp.next()
	return resp, err
}

// next sends requests until a mirror responds or all requests fail.
func (resp *Resp) next() error {
	var err error
	c := resp.client
	req := resp.req
	// lookup reqHost entry
	reqHost := c.getHost(req.Host)
	// create sorted list of mirrors, based on backoffs, upstream, and priority
	hosts := make([]*clientHost, 0, 1+len(reqHost.config.Mirrors))
	if !req.NoMirrors {
		for _, m := range reqHost.config.Mirrors {
			hosts = append(hosts, c.getHost(m))
		}
	}
	hosts = append(hosts, reqHost)
	sort.Slice(hosts, sortHostsCmp(hosts, reqHost.config.Name))
	// loop over requests to mirrors and retries
	curHost := 0
	for {
		backoff := false
		dropHost := false
		retryHost := false
		if len(hosts) == 0 {
			if err != nil {
				return err
			}
			return errs.ErrAllRequestsFailed
		}
		if curHost >= len(hosts) {
			curHost = 0
		}
		h := hosts[curHost]
		resp.mirror = h.config.Name
		// there is an intentional extra retry in this check to allow for auth requests
		if resp.retryCount > c.retryLimit {
			return errs.ErrRetryLimitExceeded
		}
		resp.retryCount++

		// check that context isn't canceled/done
		ctxErr := resp.ctx.Err()
		if ctxErr != nil {
			return ctxErr
		}
		// wait for other concurrent requests to this host
		throttleDone, throttleErr := h.throttle.Acquire(resp.ctx, reqmeta.Data{
			Kind: req.MetaKind,
			Size: req.BodyLen + req.ExpectLen + req.TransactLen,
		})
		if throttleErr != nil {
			return throttleErr
		}

		// try each host in a closure to handle all the backoff/dropHost from one place
		loopErr := func() error {
			var err error
			if req.Method == "HEAD" && h.config.APIOpts != nil {
				var disableHead bool
				disableHead, err = strconv.ParseBool(h.config.APIOpts["disableHead"])
				if err == nil && disableHead {
					dropHost = true
					return fmt.Errorf("head requests disabled for host \"%s\": %w", h.config.Name, errs.ErrUnsupportedAPI)
				}
			}

			// build the url
			var u url.URL
			if req.DirectURL != nil {
				u = *req.DirectURL
			} else {
				u = url.URL{
					Host:   h.config.Hostname,
					Scheme: "https",
				}
				path := strings.Builder{}
				path.WriteString("/v2")
				if h.config.PathPrefix != "" && !req.NoPrefix {
					path.WriteString("/" + h.config.PathPrefix)
				}
				if req.Repository != "" {
					path.WriteString("/" + req.Repository)
				}
				path.WriteString("/" + req.Path)
				u.Path = path.String()
				if h.config.TLS == config.TLSDisabled {
					u.Scheme = "http"
				}
				query := url.Values{}
				if req.Query != nil {
					query = req.Query
				}
				if h.config.Hostname != reqHost.config.Hostname {
					query.Set("ns", reqHost.config.Hostname)
				}
				u.RawQuery = query.Encode()
			}
			// close previous response
			if resp.resp != nil && resp.resp.Body != nil {
				_ = resp.resp.Body.Close()
			}
			// delay for backoff if needed
			bu := resp.backoffGet()
			if !bu.IsZero() && bu.After(time.Now()) {
				sleepTime := time.Until(bu)
				c.slog.Debug("Sleeping for backoff",
					slog.String("Host", h.config.Name),
					slog.Duration("Duration", sleepTime))
				select {
				case <-resp.ctx.Done():
					return errs.ErrCanceled
				case <-time.After(sleepTime):
				}
			}
			var httpReq *http.Request
			httpReq, err = http.NewRequestWithContext(resp.ctx, req.Method, u.String(), nil)
			if err != nil {
				dropHost = true
				return err
			}
			if req.BodyFunc != nil {
				body, err := req.BodyFunc()
				if err != nil {
					dropHost = true
					return err
				}
				httpReq.Body = body
				httpReq.GetBody = req.BodyFunc
				httpReq.ContentLength = req.BodyLen
			} else if len(req.BodyBytes) > 0 {
				body := io.NopCloser(bytes.NewReader(req.BodyBytes))
				httpReq.Body = body
				httpReq.GetBody = func() (io.ReadCloser, error) { return body, nil }
				httpReq.ContentLength = req.BodyLen
			}
			if len(req.Headers) > 0 {
				httpReq.Header = req.Headers.Clone()
			}
			if c.userAgent != "" && httpReq.Header.Get("User-Agent") == "" {
				httpReq.Header.Add("User-Agent", c.userAgent)
			}
			if resp.readCur > 0 && resp.readMax > 0 {
				if req.Headers.Get("Range") == "" {
					httpReq.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", resp.readCur, resp.readMax))
				} else {
					// TODO: support Seek within a range request
					dropHost = true
					return fmt.Errorf("unable to resume a connection within a range request")
				}
			}

			hAuth := h.getAuth(req.Repository)
			if hAuth != nil {
				// include docker generated scope to emulate docker clients
				if req.Repository != "" {
					scope := "repository:" + req.Repository + ":pull"
					if req.Method != "HEAD" && req.Method != "GET" {
						scope = scope + ",push"
					}
					_ = hAuth.AddScope(h.config.Hostname, scope)
				}
				// add auth headers
				err = hAuth.UpdateRequest(httpReq)
				if err != nil {
					if errors.Is(err, errs.ErrHTTPUnauthorized) {
						dropHost = true
					} else {
						backoff = true
					}
					return err
				}
			}

			// delay for the rate limit
			if h.reqFreq > 0 {
				sleep := time.Duration(0)
				h.mu.Lock()
				if time.Now().Before(h.reqNext) {
					sleep = time.Until(h.reqNext)
					h.reqNext = h.reqNext.Add(h.reqFreq)
				} else {
					h.reqNext = time.Now().Add(h.reqFreq)
				}
				h.mu.Unlock()
				if sleep > 0 {
					time.Sleep(sleep)
				}
			}

			// send request
			hc := h.getHTTPClient(req.Repository)
			//#nosec G704 inputs are user controlled and sanitized
			resp.resp, err = hc.Do(httpReq)
			if err != nil {
				c.slog.Debug("Request failed",
					slog.String("URL", u.String()),
					slog.String("err", err.Error()))
				backoff = true
				return err
			}

			statusCode := resp.resp.StatusCode
			if statusCode < 200 || statusCode >= 300 {
				switch statusCode {
				case http.StatusUnauthorized:
					// if auth can be done, retry same host without delay, otherwise drop/backoff
					if hAuth != nil {
						err = hAuth.HandleResponse(resp.resp)
					} else {
						err = fmt.Errorf("authentication handler unavailable")
					}
					if err != nil {
						if errors.Is(err, errs.ErrEmptyChallenge) || errors.Is(err, errs.ErrNoNewChallenge) || errors.Is(err, errs.ErrHTTPUnauthorized) {
							c.slog.Debug("Failed to handle auth request",
								slog.String("URL", u.String()),
								slog.String("Err", err.Error()))
						} else {
							c.slog.Warn("Failed to handle auth request",
								slog.String("URL", u.String()),
								slog.String("Err", err.Error()))
						}
						dropHost = true
					} else {
						err = fmt.Errorf("authentication required")
						retryHost = true
					}
					return err
				case http.StatusNotFound:
					// if not found, drop mirror for this req, but other requests don't need backoff
					dropHost = true
				case http.StatusRequestedRangeNotSatisfiable:
					// if range request error (blob push), drop mirror for this req, but other requests don't need backoff
					dropHost = true
				case http.StatusTooManyRequests, http.StatusRequestTimeout, http.StatusGatewayTimeout, http.StatusBadGateway, http.StatusInternalServerError:
					// server is likely overloaded, backoff but still retry
					backoff = true
				default:
					// all other errors indicate a bigger issue, don't retry and set backoff
					backoff = true
					dropHost = true
				}
				errHTTP := HTTPError(resp.resp.StatusCode)
				errBody, _ := io.ReadAll(resp.resp.Body)
				_ = resp.resp.Body.Close()
				return fmt.Errorf("request failed: %w: %s", errHTTP, errBody)
			}

			resp.reader = resp.resp.Body
			resp.done = false
			// set variables from headers if found
			clHeader := resp.resp.Header.Get("Content-Length")
			if resp.readCur == 0 && clHeader != "" {
				cl, parseErr := strconv.ParseInt(clHeader, 10, 64)
				if parseErr != nil {
					c.slog.Debug("failed to parse content-length header",
						slog.String("err", parseErr.Error()),
						slog.String("header", clHeader))
				} else if resp.readMax > 0 {
					if resp.readMax != cl {
						return fmt.Errorf("unexpected content-length, expected %d, received %d", resp.readMax, cl)
					}
				} else {
					resp.readMax = cl
				}
			}
			// verify Content-Range header when range request used, fail if missing
			if httpReq.Header.Get("Range") != "" && resp.resp.Header.Get("Content-Range") == "" {
				dropHost = true
				_ = resp.resp.Body.Close()
				return fmt.Errorf("range request not supported by server")
			}
			return nil
		}()
		// return on success
		if loopErr == nil {
			resp.throttleDone = throttleDone
			return nil
		}
		// backoff, dropHost, and/or go to next host in the list
		if backoff {
			if req.IgnoreErr {
				// don't set a backoff, immediately drop the host when errors ignored
				dropHost = true
			} else {
				boErr := resp.backoffSet()
				if boErr != nil {
					// reached backoff limit
					dropHost = true
				}
			}
		}
		throttleDone()
		// when error does not allow retries, abort with the last known err value
		if err != nil && errors.Is(loopErr, errs.ErrNotRetryable) {
			return err
		}
		err = loopErr
		if dropHost {
			hosts = slices.Delete(hosts, curHost, curHost+1)
		} else if !retryHost {
			curHost++
		}
	}
}

// GetThrottle returns the current [pqueue.Queue] for a host used to throttle connections.
// This can be used to acquire multiple throttles before performing a request across multiple hosts.
func (c *Client) GetThrottle(host string) *pqueue.Queue[reqmeta.Data] {
	ch := c.getHost(host)
	return ch.throttle
}

// HTTPResponse returns the [http.Response] from the last request.
func (resp *Resp) HTTPResponse() *http.Response {
	return resp.resp
}

// Read provides a retryable read from the body of the response.
func (resp *Resp) Read(b []byte) (int, error) {
	if resp.done {
		return 0, io.EOF
	}
	if resp.resp == nil {
		return 0, errs.ErrNotFound
	}
	// perform the read
	i, err := resp.reader.Read(b)
	resp.readCur += int64(i)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		if resp.resp.Request.Method == "HEAD" || resp.readCur >= resp.readMax {
			resp.backoffReset()
			resp.done = true
		} else {
			// short read, retry?
			resp.client.slog.Debug("EOF before reading all content, retrying",
				slog.Int64("curRead", resp.readCur),
				slog.Int64("contentLen", resp.readMax))
			// retry
			respErr := resp.backoffSet()
			if respErr == nil {
				respErr = resp.next()
			}
			// unrecoverable EOF
			if respErr != nil {
				resp.client.slog.Warn("Failed to recover from short read",
					slog.String("err", respErr.Error()))
				resp.done = true
				return i, err
			}
			// retry successful, no EOF
			return i, nil
		}
	}

	if err == nil {
		return i, nil
	}
	return i, err
}

// Close frees up resources from the request.
func (resp *Resp) Close() error {
	if resp.throttleDone != nil {
		resp.throttleDone()
		resp.throttleDone = nil
	}
	if resp.resp == nil {
		return errs.ErrNotFound
	}
	if !resp.done {
		resp.backoffReset()
	}
	resp.done = true
	return resp.resp.Body.Close()
}

// Seek provides a limited ability seek within the request response.
func (resp *Resp) Seek(offset int64, whence int) (int64, error) {
	newOffset := resp.readCur
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset += offset
	case io.SeekEnd:
		if resp.readMax <= 0 {
			return resp.readCur, fmt.Errorf("seek from end is not supported")
		} else if resp.readMax+offset < 0 {
			return resp.readCur, fmt.Errorf("seek past beginning of the file is not supported")
		}
		newOffset = resp.readMax + offset
	default:
		return resp.readCur, fmt.Errorf("unknown value of whence: %d", whence)
	}
	if newOffset != resp.readCur {
		resp.readCur = newOffset
		// rerun the request to restart
		resp.retryCount-- // do not count a seek as a retry
		err := resp.next()
		if err != nil {
			return resp.readCur, err
		}
	}
	return resp.readCur, nil
}

func (resp *Resp) backoffGet() time.Time {
	if resp.backoffCur > 0 {
		delay := resp.client.delayInit << resp.backoffCur
		delay = min(delay, resp.client.delayMax)
		next := resp.backoffLast.Add(delay)
		now := time.Now()
		if now.After(next) {
			next = now
		}
		resp.backoffLast = next
		return next
	}
	// reset a stale "retry-after" time
	if !resp.backoffLast.IsZero() && resp.backoffLast.Before(time.Now()) {
		resp.backoffLast = time.Time{}
	}
	return resp.backoffLast
}

func (resp *Resp) backoffSet() error {
	c := resp.client
	now := time.Now()
	// check rate limit header and use that directly if possible
	if resp.resp != nil && resp.resp.Header.Get("Retry-After") != "" {
		ras := resp.resp.Header.Get("Retry-After")
		ra, _ := time.ParseDuration(ras + "s")
		if ra > 0 {
			next := now.Add(ra)
			if resp.backoffLast.Before(next) {
				resp.backoffLast = next
			}
			resp.backoffHostSet(next)
			return nil
		}
	}
	// Track backoffs for this request only. Shared host backoff state caused later
	// requests to fail after a previous request exhausted its own retry budget.
	resp.backoffCur++
	if resp.backoffLast.IsZero() {
		resp.backoffLast = now
	}
	resp.backoffHostSet(resp.backoffLast)
	if resp.backoffCur >= c.retryLimit {
		return fmt.Errorf("%w: backoffs %d", errs.ErrBackoffLimit, resp.backoffCur)
	}

	return nil
}

func (resp *Resp) backoffReset() {
	resp.backoffCur = 0
	resp.backoffLast = time.Time{}
}

func (resp *Resp) backoffHostSet(next time.Time) {
	ch := resp.client.getHost(resp.mirror)
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if ch.backoffLast.Before(next) {
		ch.backoffLast = next
	}
}

// getHost looks up or creates a clientHost for a given registry.
func (c *Client) getHost(host string) *clientHost {
	c.mu.Lock()
	defer c.mu.Unlock()
	if h, ok := c.host[host]; ok {
		return h
	}
	var conf *config.Host
	if c.getConfigHost != nil {
		conf = c.getConfigHost(host)
	} else {
		conf = config.HostNewName(host)
	}
	if conf.Name != host {
		if h, ok := c.host[conf.Name]; ok {
			return h
		}
	}
	h := &clientHost{
		config:    conf,
		userAgent: c.userAgent,
		slog:      c.slog,
		auth:      map[string]*auth.Auth{},
	}
	if h.config.ReqPerSec > 0 {
		h.reqFreq = time.Duration(float64(time.Second) / h.config.ReqPerSec)
	}
	if h.config.ReqConcurrent > 0 {
		h.throttle = pqueue.New(pqueue.Opts[reqmeta.Data]{Max: int(h.config.ReqConcurrent), Next: reqmeta.DataNext})
	}
	// copy the http client and configure registry specific settings
	hc := *c.httpClient
	h.httpClient = &hc
	if h.httpClient.Transport == nil {
		h.httpClient.Transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	// configure transport for insecure requests and root certs
	if h.config.TLS == config.TLSInsecure || len(c.rootCAPool) > 0 || len(c.rootCADirs) > 0 || h.config.RegCert != "" || (h.config.ClientCert != "" && h.config.ClientKey != "") {
		t, ok := h.httpClient.Transport.(*http.Transport)
		if ok {
			var tlsc *tls.Config
			if t.TLSClientConfig != nil {
				tlsc = t.TLSClientConfig.Clone()
			} else {
				//#nosec G402 the default TLS 1.2 minimum version is allowed to support older registries
				tlsc = &tls.Config{}
			}
			if h.config.TLS == config.TLSInsecure {
				tlsc.InsecureSkipVerify = true
			} else {
				rootPool, err := makeRootPool(c.rootCAPool, c.rootCADirs, h.config.Hostname, h.config.RegCert)
				if err != nil {
					c.slog.Warn("failed to setup CA pool",
						slog.String("err", err.Error()))
				} else {
					tlsc.RootCAs = rootPool
				}
			}
			if h.config.ClientCert != "" && h.config.ClientKey != "" {
				cert, err := tls.X509KeyPair([]byte(h.config.ClientCert), []byte(h.config.ClientKey))
				if err != nil {
					c.slog.Warn("failed to configure client certs",
						slog.String("err", err.Error()))
				} else {
					tlsc.Certificates = []tls.Certificate{cert}
				}
			}
			t.TLSClientConfig = tlsc
			h.httpClient.Transport = t
		}
	}
	// wrap the transport for logging and to handle warning headers
	h.httpClient.Transport = &wrapTransport{c: c, orig: h.httpClient.Transport}

	c.host[conf.Name] = h
	if conf.Name != host {
		// save another reference for faster lookups
		c.host[host] = h
	}
	return h
}

// getHTTPClient returns a client specific to the repo being queried.
// Repository specific authentication needs a dedicated CheckRedirect handler.
func (ch *clientHost) getHTTPClient(repo string) *http.Client {
	hc := *ch.httpClient
	hc.CheckRedirect = ch.checkRedirect(repo, hc.CheckRedirect)
	return &hc
}

// checkRedirect wraps http.CheckRedirect to inject auth headers to specific hosts in the redirect chain
func (ch *clientHost) checkRedirect(repo string, orig func(req *http.Request, via []*http.Request) error) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		// fail on too many redirects
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		// add auth headers if appropriate for the target host
		hAuth := ch.getAuth(repo)
		err := hAuth.UpdateRequest(req)
		if err != nil {
			return err
		}
		// wrap original redirect check
		if orig != nil {
			return orig(req, via)
		}
		return nil
	}
}

// getAuth returns an auth, which may be repository specific.
func (ch *clientHost) getAuth(repo string) *auth.Auth {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if !ch.config.RepoAuth {
		repo = "" // without RepoAuth, unset the provided repo
	}
	if _, ok := ch.auth[repo]; !ok {
		ch.auth[repo] = auth.NewAuth(
			auth.WithLog(ch.slog),
			auth.WithHTTPClient(ch.httpClient),
			auth.WithCreds(ch.AuthCreds()),
			auth.WithClientID(ch.userAgent),
		)
	}
	return ch.auth[repo]
}

func (ch *clientHost) AuthCreds() func(h string) auth.Cred {
	if ch == nil || ch.config == nil {
		return auth.DefaultCredsFn
	}
	return func(h string) auth.Cred {
		hCred := ch.config.GetCred()
		return auth.Cred{User: hCred.User, Password: hCred.Password, Token: hCred.Token}
	}
}

type wrapTransport struct {
	c    *Client
	orig http.RoundTripper
}

func (wt *wrapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := wt.orig.RoundTrip(req)
	// copy headers to censor auth field
	reqHead := req.Header.Clone()
	if reqHead.Get("Authorization") != "" {
		reqHead.Set("Authorization", "[censored]")
	}
	if err != nil {
		wt.c.slog.Debug("reg http request",
			slog.String("req-method", req.Method),
			slog.String("req-url", req.URL.String()),
			slog.Any("req-headers", reqHead),
			slog.String("err", err.Error()))
	} else {
		// extract any warnings
		for _, wh := range resp.Header.Values("Warning") {
			if match := warnRegexp.FindStringSubmatch(wh); len(match) == 2 {
				// TODO(bmitch): pass other fields (registry hostname) with structured logging
				warning.Handle(req.Context(), wt.c.slog, match[1])
			}
		}
		wt.c.slog.Log(req.Context(), types.LevelTrace, "reg http request",
			slog.String("req-method", req.Method),
			slog.String("req-url", req.URL.String()),
			slog.Any("req-headers", reqHead),
			slog.String("resp-status", resp.Status),
			slog.Any("resp-headers", resp.Header))
	}
	return resp, err
}

// HTTPError returns an error based on the status code.
func HTTPError(statusCode int) error {
	switch statusCode {
	case 401:
		return fmt.Errorf("%w [http %d]", errs.ErrHTTPUnauthorized, statusCode)
	case 403:
		return fmt.Errorf("%w [http %d]", errs.ErrHTTPUnauthorized, statusCode)
	case 404:
		return fmt.Errorf("%w [http %d]", errs.ErrNotFound, statusCode)
	case 429:
		return fmt.Errorf("%w [http %d]", errs.ErrHTTPRateLimit, statusCode)
	default:
		return fmt.Errorf("%w: %s [http %d]", errs.ErrHTTPStatus, http.StatusText(statusCode), statusCode)
	}
}

func makeRootPool(rootCAPool [][]byte, rootCADirs []string, hostname string, hostcert string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	for _, ca := range rootCAPool {
		if ok := pool.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("failed to load ca: %s", ca)
		}
	}
	for _, dir := range rootCADirs {
		hostDir := filepath.Join(dir, hostname)
		files, err := os.ReadDir(hostDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read directory %s: %w", hostDir, err)
			}
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if strings.HasSuffix(f.Name(), ".crt") {
				f := filepath.Join(hostDir, f.Name())
				//#nosec G304 file from a known directory and extension read by the user running the command on their own host
				cert, err := os.ReadFile(f)
				if err != nil {
					return nil, fmt.Errorf("failed to read %s: %w", f, err)
				}
				if ok := pool.AppendCertsFromPEM(cert); !ok {
					return nil, fmt.Errorf("failed to import cert from %s", f)
				}
			}
		}
	}
	if hostcert != "" {
		if ok := pool.AppendCertsFromPEM([]byte(hostcert)); !ok {
			// try to parse the certificate and generate a useful error
			block, _ := pem.Decode([]byte(hostcert))
			if block == nil {
				err = fmt.Errorf("pem.Decode is nil")
			} else {
				_, err = x509.ParseCertificate(block.Bytes)
			}
			return nil, fmt.Errorf("failed to load host specific ca (registry: %s): %w: %s", hostname, err, hostcert)
		}
	}
	return pool, nil
}

// sortHostCmp to sort host list of mirrors.
func sortHostsCmp(hosts []*clientHost, upstream string) func(i, j int) bool {
	now := time.Now()
	// sort by backoff first, then priority decending, then upstream name last
	return func(i, j int) bool {
		if now.Before(hosts[i].backoffLast) || now.Before(hosts[j].backoffLast) {
			return hosts[i].backoffLast.Before(hosts[j].backoffLast)
		}
		if hosts[i].config.Priority != hosts[j].config.Priority {
			return hosts[i].config.Priority < hosts[j].config.Priority
		}
		return hosts[i].config.Name != upstream
	}
}
