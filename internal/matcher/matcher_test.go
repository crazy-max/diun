package matcher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchString(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		s    string
		want bool
	}{
		{
			name: "matches",
			exp:  `^v\d+\.\d+\.\d+$`,
			s:    "v1.2.3",
			want: true,
		},
		{
			name: "does not match",
			exp:  `^v\d+\.\d+\.\d+$`,
			s:    "latest",
		},
		{
			name: "invalid regexp",
			exp:  `[`,
			s:    "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, MatchString(tt.exp, tt.s))
		})
	}
}

func TestIsIncluded(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		includes []string
		want     bool
	}{
		{
			name: "empty includes defaults to true",
			s:    "latest",
			want: true,
		},
		{
			name:     "matches one include pattern",
			s:        "v1.2.3",
			includes: []string{`^latest$`, `^v\d+\.\d+\.\d+$`},
			want:     true,
		},
		{
			name:     "matches no include patterns",
			s:        "dev",
			includes: []string{`^latest$`, `^v\d+\.\d+\.\d+$`},
		},
		{
			name:     "invalid pattern does not include",
			s:        "v1.2.3",
			includes: []string{`[`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsIncluded(tt.s, tt.includes))
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		excludes []string
		want     bool
	}{
		{
			name: "empty excludes defaults to false",
			s:    "latest",
		},
		{
			name:     "matches one exclude pattern",
			s:        "dev",
			excludes: []string{`^dev$`, `^test$`},
			want:     true,
		},
		{
			name:     "matches no exclude patterns",
			s:        "v1.2.3",
			excludes: []string{`^dev$`, `^test$`},
		},
		{
			name:     "invalid pattern does not exclude",
			s:        "dev",
			excludes: []string{`[`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsExcluded(tt.s, tt.excludes))
		})
	}
}
