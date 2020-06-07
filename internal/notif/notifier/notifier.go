package notifier

import (
	"github.com/crazy-max/diun/v4/internal/model"
)

// Handler is a notifier interface
type Handler interface {
	Name() string
	Send(entry model.NotifEntry) error
}

// Notifier represents an active notifier object
type Notifier struct {
	Handler
}
