package notifier

import (
	"github.com/crazy-max/diun/v4/internal/model"
)

// Handler is a notifier interface
type Handler interface {
	Name() string
	Send(entry model.NotifEntry) error
}

// BatchHandler is an optional interface for notifiers that support digest/batch mode.
// If a notifier implements this, SendBatch is called instead of looping Send in digest mode.
type BatchHandler interface {
	Handler
	SendBatch(entries *model.NotifEntries) error
}

// Notifier represents an active notifier object
type Notifier struct {
	Handler
}
