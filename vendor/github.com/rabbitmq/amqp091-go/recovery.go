// Copyright (c) 2026 Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries. All rights reserved.

package amqp091

import (
	"net/url"
	"time"
)

const (
	// DefaultMaxRetryCount is the default maximum number of retries for recovery.
	DefaultMaxRetryCount = 5

	// DefaultRetryInterval is the default interval between retries for recovery.
	DefaultRetryInterval = 5 * time.Second
)

var (
	// defaultRecoverableErrorCodes contains the default exception codes that trigger recovery.
	defaultRecoverableErrorCodes = []int{ConnectionForced, InternalError}

	// DefaultReconnectionConfig is the default reconnection config settings.
	DefaultReconnectionConfig = &ReconnectionConfig{
		MaxRetryCount:         DefaultMaxRetryCount,
		RetryInterval:         DefaultRetryInterval,
		RecoverableErrorCodes: cloneRecoverableErrorCodes(defaultRecoverableErrorCodes),
	}
)

// cloneRecoverableErrorCodes returns a clone of given RecoverableErrorCodes slice.
// It is used to avoid modifying the original slice.
func cloneRecoverableErrorCodes(inRecoverableErrorCodes []int) []int {
	if inRecoverableErrorCodes == nil {
		return nil
	}
	codes := make([]int, len(inRecoverableErrorCodes))
	copy(codes, inRecoverableErrorCodes)
	return codes
}

// ReconnectionConfig is the configuration for the reconnection.
type ReconnectionConfig struct {
	MaxRetryCount         int           // The maximum number of retries.
	RetryInterval         time.Duration // The interval between retries.
	RecoverableErrorCodes []int         // The error codes that trigger recovery.
}

// Clone returns a deep copy of the ReconnectionConfig.
func (rc *ReconnectionConfig) Clone() *ReconnectionConfig {
	if rc == nil {
		return nil
	}
	return &ReconnectionConfig{
		MaxRetryCount:         rc.MaxRetryCount,
		RetryInterval:         rc.RetryInterval,
		RecoverableErrorCodes: cloneRecoverableErrorCodes(rc.RecoverableErrorCodes),
	}
}

// ConnectionRecovery is the interface for the connection recovery.
//
// The err parameter in OnConnectionClose and OnChannelClose provides the reason
// why the connection or channel was closed. Under the hood, DefaultConnectionRecovery
// performs conditional recovery based on RecoverableErrorCodes. You can also customize
// the list of recoverable errors dynamically using Connection.SetRecoverableErrorCodes and
// Connection.AddRecoverableErrorCodes, or use custom implementations of this interface to
// perform more advanced logic, log errors to external monitoring systems (e.g., Prometheus),
// or trigger alerts.
type ConnectionRecovery interface {
	OnConnectionClose(conn *Connection, err *Error) // Called when the connection is closed.
	OnChannelClose(ch *Channel, err *Error)         // Called when the channel is closed.
}

// Recovery is the configuration for the recovery.
type Recovery struct {
	ReconnectionConfig *ReconnectionConfig // The configuration for the reconnection.
	ConnectionRecovery ConnectionRecovery  // The implementation of the connection recovery.
}

// DefaultConnectionRecovery is the default implementation of the connection recovery.
type DefaultConnectionRecovery struct{}

func (d *DefaultConnectionRecovery) OnConnectionClose(conn *Connection, err *Error) {
	Logger.Printf("Connection closed with error: %v", err)

	parsedURL, err1 := url.Parse(conn.url)
	if err1 != nil {
		Logger.Printf("Error parsing connection URL: %v", err1)
		return
	}

	if !conn.IsRecoveryEnabled() {
		Logger.Printf("Connection %s recovery is not enabled, skipping reconnect. ", parsedURL.Redacted())
		return
	}

	if !conn.isRecoverable(err) {
		code := 0
		if err != nil {
			code = err.Code
		}
		Logger.Printf("Connection %s closed with non-recoverable error code %d, skipping reconnect.", parsedURL.Redacted(), code)
		return
	}

	Logger.Printf("Initiating connection recovery for %s.", parsedURL.Redacted())
	// Reconnect connection
	if err := conn.Reconnect(); err != nil {
		Logger.Printf("Connection %s recovery failed: %v.", parsedURL.Redacted(), err)
		conn.cleanup()
	}
}

func (d *DefaultConnectionRecovery) OnChannelClose(ch *Channel, err *Error) {
	Logger.Printf("Channel %d closed with error: %v", ch.id, err)
	if !ch.connection.IsRecoveryEnabled() {
		Logger.Printf("Channel %d recovery is not enabled, skipping reconnect.", ch.id)
		return
	}

	if ch.connection.IsClosed() {
		Logger.Printf("Connection is closed, letting connection recovery handle channel %d.", ch.id)
		return
	}

	if !ch.connection.isRecoverable(err) {
		code := 0
		if err != nil {
			code = err.Code
		}
		Logger.Printf("Channel %d closed with non-recoverable error code %d, skipping reconnect.", ch.id, code)
		return
	}

	Logger.Printf("Initiating channel %d recovery", ch.id)
	// Reconnect channel
	if err := ch.Reconnect(); err != nil {
		Logger.Printf("Channel %d recovery failed: %v.", ch.id, err)
		ch.cleanup()
	}
}
