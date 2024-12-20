package telegram

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParseChatIDs(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string
		expected []chatID
		err      error
	}{
		{
			name:    "valid chat IDS",
			entries: []string{"8547439", "1234567"},
			expected: []chatID{
				{id: 8547439},
				{id: 1234567},
			},
			err: nil,
		},
		{
			name:    "valid strings with topics",
			entries: []string{"567891234:25", "891256734:25;12"},
			expected: []chatID{
				{id: 567891234, topics: []int64{25}},
				{id: 891256734, topics: []int64{25, 12}},
			},
			err: nil,
		},
		{
			name:     "invalid format",
			entries:  []string{"invalid_format"},
			expected: nil,
			err:      errors.New(`invalid chat ID: strconv.ParseInt: parsing "invalid_format": invalid syntax`),
		},
		{
			name:     "empty string",
			entries:  []string{""},
			expected: nil,
			err:      errors.New(`invalid chat ID: strconv.ParseInt: parsing "": invalid syntax`),
		},
		{
			name:     "string with invalid topic",
			entries:  []string{"567891234:invalid"},
			expected: nil,
			err:      errors.New(`invalid topic "invalid" for chat ID 567891234: strconv.ParseInt: parsing "invalid": invalid syntax`),
		},
		{
			name:     "invalid format with too many parts",
			entries:  []string{"567891234:25:extra"},
			expected: nil,
			err:      errors.New(`invalid chat ID "567891234:25:extra"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseChatIDs(tt.entries)
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, res)
		})
	}
}
