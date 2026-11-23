package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStatic(t *testing.T) {
	file := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(file, []byte("file-value\n"), 0o600))
	missing := file + ".missing"
	for _, tc := range []struct {
		name      string
		opts      StaticOptions
		want      Static
		wantError bool
	}{
		{
			name: "anonymous",
		},
		{
			name: "files preserve content",
			opts: StaticOptions{UsernameFile: file, PasswordFile: file},
			want: Static{Username: "file-value\n", Password: "file-value\n"},
		},
		{
			name: "inline takes precedence",
			opts: StaticOptions{Username: "user", Password: "pass", UsernameFile: missing, PasswordFile: missing},
			want: Static{Username: "user", Password: "pass"},
		},
		{
			name:      "username failure retains password",
			opts:      StaticOptions{UsernameFile: missing, Password: "pass"},
			want:      Static{Password: "pass"},
			wantError: true,
		},
		{
			name:      "password failure retains username",
			opts:      StaticOptions{Username: "user", PasswordFile: missing},
			want:      Static{Username: "user"},
			wantError: true,
		},
		{
			name:      "both files missing",
			opts:      StaticOptions{UsernameFile: missing, PasswordFile: missing},
			wantError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewStatic(tc.opts)
			if tc.wantError {
				require.ErrorIs(t, err, os.ErrNotExist)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}
