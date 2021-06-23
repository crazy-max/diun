package matrix

import (
	"fmt"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/matrix-org/gomatrix"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	defer func() {
		if _, err := m.Logout(); err != nil {
			log.Error().Err(err).Msg("Cannot logout")
		}
	}()

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

	message, err := msg.New(msg.Options{
		Meta:         c.meta,
		Entry:        entry,
		TemplateBody: c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	_, msgText, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	_, msgHTML, err := message.RenderHTML()
	if err != nil {
		return err
	}

	if _, err := m.SendMessageEvent(joined.RoomID, "m.room.message", gomatrix.HTMLMessage{
		Body:          string(msgText),
		MsgType:       fmt.Sprintf("m.%s", c.cfg.MsgType),
		Format:        "org.matrix.custom.html",
		FormattedBody: string(msgHTML),
	}); err != nil {
		return errors.Wrap(err, "failed to submit message to Matrix")
	}

	return nil
}
