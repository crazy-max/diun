package secret

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSecret(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(secretFile, []byte("from-file"), 0o600))

	tests := []struct {
		name      string
		plainText string
		filename  string
		want      string
		wantErr   bool
	}{
		{
			name:      "plaintext",
			plainText: "from-plain",
			filename:  secretFile,
			want:      "from-plain",
		},
		{
			name:     "file",
			filename: secretFile,
			want:     "from-file",
		},
		{
			name: "empty",
		},
		{
			name:     "missing file",
			filename: filepath.Join(t.TempDir(), "missing"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSecret(tt.plainText, tt.filename)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
