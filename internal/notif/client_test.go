package notif

import (
	"errors"
	"testing"
)

func TestSanitizeUrlTokens(t *testing.T) {
	tests := []struct {
		input    error
		expected string
	}{
		{
			input:    errors.New(`Post "http://gotify:9265/message?token=supersecret": dial tcp ...`),
			expected: `Post "http://gotify:9265/message?token=[REDACTED]": dial tcp ...`,
		},
		{
			input:    errors.New(`GET /api?apikey=12345&auth=abcdef`),
			expected: `GET /api?apikey=[REDACTED]&auth=[REDACTED]`,
		},
		{
			input:    errors.New(`https://foo.com?token=abc&apikey=def&password=ghi`),
			expected: `https://foo.com?token=[REDACTED]&apikey=[REDACTED]&password=[REDACTED]`,
		},
		{
			input:    errors.New(`https://bar.com?sessionid=xyz&key=123`),
			expected: `https://bar.com?sessionid=[REDACTED]&key=[REDACTED]`,
		},
		{
			input:    errors.New(`No sensitive params here`),
			expected: `No sensitive params here`,
		},
		{
			input:    errors.New(`Post "http://gotify:9265/message?otherparam=asdf": dial tcp ...`),
			expected: `Post "http://gotify:9265/message?otherparam=asdf": dial tcp ...`,
		},
	}

	for _, tt := range tests {
		result := SanitizeUrlTokens(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeUrlTokens(%q) = %q; want %q", tt.input.Error(), result, tt.expected)
		}
	}
}
