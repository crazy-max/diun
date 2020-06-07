package script

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client represents an active script notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifScript
	meta model.Meta
}

// New creates a new script notification instance
func New(config *model.NotifScript, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "script"
}

// Send creates and sends a script notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	cmd := exec.Command(c.cfg.Cmd, c.cfg.Args...)
	setSysProcAttr(cmd)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set working dir
	if c.cfg.Dir != "" {
		cmd.Dir = c.cfg.Dir
	}

	// Set env vars
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("DIUN_VERSION=%s", c.meta.Version),
		fmt.Sprintf("DIUN_ENTRY_STATUS=%s", string(entry.Status)),
		fmt.Sprintf("DIUN_ENTRY_PROVIDER=%s", entry.Provider),
		fmt.Sprintf("DIUN_ENTRY_IMAGE=%s", entry.Image.String()),
		fmt.Sprintf("DIUN_ENTRY_MIMETYPE=%s", entry.Manifest.MIMEType),
		fmt.Sprintf("DIUN_ENTRY_DIGEST=%s", entry.Manifest.Digest),
		fmt.Sprintf("DIUN_ENTRY_CREATED=%s", entry.Manifest.Created),
		fmt.Sprintf("DIUN_ENTRY_PLATFORM=%s", entry.Manifest.Platform),
	}...)

	// Run
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, strings.TrimSpace(stderr.String()))
	}

	log.Debug().Msgf(strings.TrimSpace(stdout.String()))
	return nil
}
