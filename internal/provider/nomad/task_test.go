package nomad_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/provider/nomad"
	"github.com/stretchr/testify/assert"
)

func TestParseServiceTags(t *testing.T) {
	testCases := []struct {
		input    []string
		expected map[string]string
	}{
		{
			input: []string{
				"noequal",
			},
			expected: map[string]string{},
		},
		{
			input: []string{
				"emptyequal=",
			},
			expected: map[string]string{
				"emptyequal": "",
			},
		},
		{
			input: []string{
				"key=value",
			},
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			input: []string{
				"withequal=a=b",
			},
			expected: map[string]string{
				"withequal": "a=b",
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.input[0], func(t *testing.T) {
			result := nomad.ParseServiceTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
