package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDocker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, tc := range []struct {
		name      string
		config    string
		want      Credentials
		wantError bool
	}{
		{
			name:   "username and password",
			config: `{"auths":{"diun-auth-test.invalid":{"auth":"dXNlcjpwYXNz"}}}`,
			want: Credentials{
				Username: "user",
				Password: "pass",
			},
		},
		{
			name:   "identity token",
			config: `{"auths":{"diun-auth-test.invalid":{"auth":"Og==","identitytoken":"token"}}}`,
			want: Credentials{
				IdentityToken: "token",
			},
		},
		{
			name:   "no matching credentials",
			config: `{"auths":{}}`,
		},
		{
			name:      "invalid config",
			config:    `{`,
			wantError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(tc.config), 0o600))
			provider, err := NewDocker("diun-auth-test.invalid")
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, Docker{credentials: tc.want}, provider)
			credentials, err := provider.Credentials(t.Context(), "diun-auth-test.invalid")
			require.NoError(t, err)
			require.Equal(t, tc.want, credentials)
			providers := map[Provider]bool{provider: true}
			require.True(t, providers[Docker{credentials: tc.want}])
		})
	}
}
