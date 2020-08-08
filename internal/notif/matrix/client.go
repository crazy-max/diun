package matrix

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/matrix-org/gomatrix"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday/v2"
)

// Client represents an active rocketchat notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifMatrix
	meta model.Meta
}

// New creates a new rocketchat notification instance
func New(config *model.NotifMatrix, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "matrix"
}

// Send creates and sends a matrix notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	m, err := gomatrix.NewClient(c.cfg.HomeserverURL, "", "")
	if err != nil {
		return errors.Wrap(err, "failed to initialize Matrix client")
	}
	defer m.Logout()

	user, err := utl.GetSecret(c.cfg.User, c.cfg.UserFile)
	if err != nil {
		return errors.New("Cannot retrieve username secret for Matrix notifier")
	}
	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		return errors.New("Cannot retrieve password secret for Matrix notifier")
	}

	r, err := m.Login(&gomatrix.ReqLogin{
		Type:                     "m.login.password",
		User:                     user,
		Password:                 password,
		InitialDeviceDisplayName: c.meta.Name,
	})
	if err != nil {
		return errors.Wrap(err, "failed to authenticate Matrix user")
	}
	m.SetCredentials(r.UserID, r.AccessToken)

	joined, err := m.JoinRoom(c.cfg.RoomID, "", nil)
	if err != nil {
		return errors.Wrap(err, "failed to join room")
	}

	tagTpl := "**{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}**"
	if len(entry.Image.HubLink) > 0 {
		tagTpl = "[**{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}**]({{ .Entry.Image.HubLink }})"
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("text").Parse(fmt.Sprintf("Docker tag %s which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status \"new\") }}newly added{{ else }}updated{{ end }} on {{ .Meta.Hostname }}.", tagTpl)))
	if err := msgTpl.Execute(&msgBuf, struct {
		Meta  model.Meta
		Entry model.NotifEntry
	}{
		Meta:  c.meta,
		Entry: entry,
	}); err != nil {
		return err
	}

	msgHTML := bluemonday.UGCPolicy().SanitizeBytes(
		blackfriday.Run(msgBuf.Bytes()),
	)

	if _, err := m.SendMessageEvent(joined.RoomID, "m.room.message", gomatrix.HTMLMessage{
		Body:          msgBuf.String(),
		MsgType:       fmt.Sprintf("m.%s", c.cfg.MsgType),
		Format:        "org.matrix.custom.html",
		FormattedBody: string(msgHTML),
	}); err != nil {
		return errors.Wrap(err, "failed to submit message to Matrix")
	}

	return nil
}
