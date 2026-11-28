package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCredentialsPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(home, "run"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("DOCKER_CONFIG", filepath.Join(home, "docker"))
	containersFile := filepath.Join(home, ".config", "containers", "auth.json")
	dockerFile := filepath.Join(home, "docker", "config.json")
	for _, path := range []string{containersFile, dockerFile} {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	}
	require.NoError(t, os.WriteFile(containersFile, []byte(`{"auths":{"example.com":{"identitytoken":"containers"}}}`), 0o600))
	require.NoError(t, os.WriteFile(dockerFile, []byte(`{"auths":{"example.com":{"identitytoken":"docker"}}}`), 0o600))
	auth, err := GetCredentials("example.com")
	require.NoError(t, err)
	assert.Equal(t, "docker", auth.IdentityToken)
	require.NoError(t, os.WriteFile(dockerFile, []byte(`{"auths":{"other.example":{"identitytoken":"docker"}}}`), 0o600))
	auth, err = GetCredentials("example.com")
	require.NoError(t, err)
	assert.Equal(t, "containers", auth.IdentityToken)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".dockercfg"), []byte(`{"example.com":{"identitytoken":"legacy"}}`), 0o600))
	auth, err = GetCredentials("example.com")
	require.NoError(t, err)
	assert.Equal(t, "legacy", auth.IdentityToken)
}

func TestCredentialHelper(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixture runs in the Linux test workflow")
	}
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	helper := "#!/bin/sh\nread host\n[ \"$1\" = get ] && [ \"$host\" = \"$DIUN_TEST_HELPER_HOST\" ] || exit 1\nprintf '%s' '{\"Username\":\"<token>\",\"Secret\":\"helper-token\"}'\n"
	// #nosec G306 -- the temporary credential-helper fixture must be executable.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-credential-diun-test"), []byte(helper), 0o700))
	for _, tt := range []struct {
		key    string
		domain string
	}{
		{
			key:    "example.com",
			domain: "example.com",
		},
		{
			key:    "https://index.docker.io/v1/",
			domain: "docker.io",
		},
		{
			key:    "index.docker.io",
			domain: "docker.io",
		},
		{
			key:    "registry-1.docker.io",
			domain: "docker.io",
		},
	} {
		t.Run(tt.key, func(t *testing.T) {
			t.Setenv("DIUN_TEST_HELPER_HOST", tt.key)
			path := filepath.Join(dir, "config.json")
			body := fmt.Sprintf(`{"credHelpers":{%q:"diun-test"},"credsStore":"nonexistent","auths":{%q:{"auth":"dXNlcjpzZWNyZXQ="}}}`, tt.key, tt.domain)
			require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
			auth, err := credentialsFromFile(path, tt.domain)
			require.NoError(t, err)
			assert.Equal(t, AuthConfig{IdentityToken: "helper-token"}, auth)
		})
	}
}

func TestCredentialsFromFile(t *testing.T) {
	for _, tt := range []struct {
		name      string
		file      string
		body      string
		domain    string
		want      AuthConfig
		wantError bool
	}{
		{
			name:   "Docker Hub alias",
			body:   `{"auths":{"https://index.docker.io/v1/":{"auth":"dXNlcjpzZWNyZXQ="}}}`,
			domain: "docker.io",
			want:   AuthConfig{Username: "user", Password: "secret"},
		},
		{
			name:   "containers identity token",
			body:   `{"auths":{"example.com":{"identitytoken":"token"}}}`,
			domain: "example.com",
			want:   AuthConfig{IdentityToken: "token"},
		},
		{
			name:   "legacy Docker config",
			file:   ".dockercfg",
			body:   `{"https://index.docker.io/v1/":{"auth":"dXNlcjpzZWNyZXQ="}}`,
			domain: "docker.io",
			want:   AuthConfig{Username: "user", Password: "secret"},
		},
		{
			name:   "repository credentials do not become host credentials",
			body:   `{"auths":{"example.com/private":{"auth":"dXNlcjpzZWNyZXQ="}}}`,
			domain: "example.com",
		},
		{
			name:   "other registry",
			body:   `{"auths":{"other.example":{"auth":"dXNlcjpzZWNyZXQ="}}}`,
			domain: "example.com",
		},
		{
			name:      "malformed auth",
			body:      `{"auths":{"example.com":{"auth":"!"}}}`,
			domain:    "example.com",
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.file
			if name == "" {
				name = "auth.json"
			}
			path := filepath.Join(t.TempDir(), name)
			require.NoError(t, os.WriteFile(path, []byte(tt.body), 0o600))
			got, err := credentialsFromFile(path, tt.domain)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
