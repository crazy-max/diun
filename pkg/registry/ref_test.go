package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sha256digestHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	sha256digest    = "@sha256:" + sha256digestHex
)

func TestParseReference(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{
			input:    "busybox",
			expected: "docker.io/library/busybox:latest",
		},
		{
			input:    "//busybox:notlatest",
			expected: "docker.io/library/busybox:notlatest",
		},
		{
			input:    "//busybox" + sha256digest,
			expected: "docker.io/library/busybox" + sha256digest,
		},
		{
			input:    "//busybox",
			expected: "docker.io/library/busybox:latest",
		},
		{
			input:    "//busybox:latest" + sha256digest,
			expected: "",
			wantErr:  true,
		},
		{
			input:    "//docker.io/library/busybox:latest",
			expected: "docker.io/library/busybox:latest",
		},
		{
			input:    "//UPPERCASEISINVALID",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			ref, err := ParseReference(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ref.DockerReference().String(), tt.input)
		})
	}
}
