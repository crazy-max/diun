package registry

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/regclient/regclient/types/errs"
)

// registryTransport validates pagination before regclient can follow a Link
// as a direct request, which bypasses its HTTP redirect checks.
type registryTransport struct {
	base *http.Transport
}

func (t registryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	for _, value := range resp.Header.Values("Link") {
		if err := validatePaginationLinks(req.URL, value); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
	}
	return resp, nil
}

func validatePaginationLinks(base *url.URL, value string) error {
	for value = strings.TrimSpace(value); value != ""; value = strings.TrimSpace(value) {
		// Require RFC 8288 angle brackets, including for relative references.
		if !strings.HasPrefix(value, "<") {
			return errors.Wrap(errs.ErrNotRetryable, "invalid registry pagination link")
		}
		raw, rest, ok := strings.Cut(value[1:], ">")
		if !ok {
			return errors.Wrap(errs.ErrNotRetryable, "invalid registry pagination link")
		}
		next, err := base.Parse(raw)
		if err != nil || next.User != nil || !strings.EqualFold(next.Scheme, base.Scheme) || !strings.EqualFold(next.Host, base.Host) {
			return errors.Wrap(errs.ErrNotRetryable, "unsafe registry pagination link")
		}
		// Commas inside quoted parameters do not start another link.
		if strings.Contains(rest, `\`) {
			return errors.Wrap(errs.ErrNotRetryable, "invalid registry pagination link")
		}
		quoted := false
		value = ""
		for i, char := range rest {
			if char == '"' {
				quoted = !quoted
			} else if !quoted && char == ',' {
				value = rest[i+1:]
				break
			}
		}
		if quoted {
			return errors.Wrap(errs.ErrNotRetryable, "invalid registry pagination link")
		}
	}
	return nil
}
