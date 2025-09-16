package mqtt

import "net/url"

type ConnectionNotificationType int64

const (
	ConnectionNotificationTypeConnected ConnectionNotificationType = iota
	ConnectionNotificationTypeConnecting
	ConnectionNotificationTypeFailed
	ConnectionNotificationTypeLost
	ConnectionNotificationTypeBroker
	ConnectionNotificationTypeBrokerFailed
)

type ConnectionNotification interface {
	Type() ConnectionNotificationType
}

// Connected

type ConnectionNotificationConnected struct {
}

func (n ConnectionNotificationConnected) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnected
}

// Connecting

type ConnectionNotificationConnecting struct {
	IsReconnect bool
	Attempt     int
}

func (n ConnectionNotificationConnecting) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnecting
}

// Connection Failed

type ConnectionNotificationFailed struct {
	Reason error
}

func (n ConnectionNotificationFailed) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeFailed
}

// Connection Lost

type ConnectionNotificationLost struct {
	Reason error // may be nil
}

func (n ConnectionNotificationLost) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeLost
}

// Broker Connection

type ConnectionNotificationBroker struct {
	Broker *url.URL
}

func (n ConnectionNotificationBroker) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeBroker
}

// Broker Connection Failed

type ConnectionNotificationBrokerFailed struct {
	Broker *url.URL
	Reason error
}

func (n ConnectionNotificationBrokerFailed) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeBrokerFailed
}
