package script

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRunsCommandWithNotificationEnv(t *testing.T) {
	captureFile := filepath.Join(t.TempDir(), "env.json")
	workDir := t.TempDir()
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("DIUN_SCRIPT_CAPTURE", captureFile)

	cmd, err := os.Executable()
	require.NoError(t, err)

	client := newTestClient(cmd, workDir)

	err = client.Send(testEntry(t))
	require.NoError(t, err)

	file, err := os.Open(captureFile)
	require.NoError(t, err)
	defer file.Close()

	var got map[string]string
	require.NoError(t, json.NewDecoder(file).Decode(&got))
	assert.Equal(t, workDir, got["PWD"])
	assert.Equal(t, "4.0.0", got["DIUN_VERSION"])
	assert.Equal(t, "node-1", got["DIUN_HOSTNAME"])
	assert.Equal(t, "update", got["DIUN_ENTRY_STATUS"])
	assert.Equal(t, "file", got["DIUN_ENTRY_PROVIDER"])
	assert.Equal(t, "docker.io/library/alpine:latest", got["DIUN_ENTRY_IMAGE"])
	assert.Equal(t, "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", got["DIUN_ENTRY_DIGEST"])
	assert.Equal(t, "linux/amd64", got["DIUN_ENTRY_PLATFORM"])
	assert.Equal(t, "ops", got["DIUN_ENTRY_METADATA_OWNER"])
}

func TestSendReturnsCommandStderr(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("DIUN_SCRIPT_MODE", "fail")

	cmd, err := os.Executable()
	require.NoError(t, err)

	err = newTestClient(cmd, "").Send(testEntry(t))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "script failed")
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	os.Exit(runHelperProcess())
}

func runHelperProcess() int {
	if os.Getenv("DIUN_SCRIPT_MODE") == "fail" {
		fmt.Fprintln(os.Stderr, "script failed")
		return 42
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	//nolint:gosec // test helper receives a t.TempDir capture path from the parent test.
	file, err := os.Create(os.Getenv("DIUN_SCRIPT_CAPTURE"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	err = json.NewEncoder(file).Encode(map[string]string{
		"PWD":                       workDir,
		"DIUN_VERSION":              os.Getenv("DIUN_VERSION"),
		"DIUN_HOSTNAME":             os.Getenv("DIUN_HOSTNAME"),
		"DIUN_ENTRY_STATUS":         os.Getenv("DIUN_ENTRY_STATUS"),
		"DIUN_ENTRY_PROVIDER":       os.Getenv("DIUN_ENTRY_PROVIDER"),
		"DIUN_ENTRY_IMAGE":          os.Getenv("DIUN_ENTRY_IMAGE"),
		"DIUN_ENTRY_DIGEST":         os.Getenv("DIUN_ENTRY_DIGEST"),
		"DIUN_ENTRY_PLATFORM":       os.Getenv("DIUN_ENTRY_PLATFORM"),
		"DIUN_ENTRY_METADATA_OWNER": os.Getenv("DIUN_ENTRY_METADATA_OWNER"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = file.Close()
		return 2
	}

	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	return 0
}

func newTestClient(cmd, dir string) *Client {
	return &Client{
		cfg: &model.NotifScript{
			Cmd:  cmd,
			Args: []string{"-test.run=TestHelperProcess"},
			Dir:  dir,
		},
		meta: model.Meta{
			Version:  "4.0.0",
			Hostname: "node-1",
		},
	}
}

func testEntry(t *testing.T) model.NotifEntry {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "docker.io/library/alpine:latest",
	})
	require.NoError(t, err)

	return model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "file",
		Image:    image,
		Manifest: registry.Manifest{
			Digest:   digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
			Platform: "linux/amd64",
		},
		Metadata: map[string]string{
			"owner": "ops",
		},
	}
}
