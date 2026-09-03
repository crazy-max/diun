// Copyright (c) 2026 Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries. All rights reserved.

package amqp091

import (
	"fmt"
	"net/url"
	"time"
)

// TopologyRecoveryEntityType identifies the topology recovery entity kind.
type TopologyRecoveryEntityType byte

const (
	TopologyEntityExchange TopologyRecoveryEntityType = iota
	TopologyEntityQueue
	TopologyEntityQueueBinding
	TopologyEntityExchangeBinding
	TopologyEntityConsumer
)

func (t TopologyRecoveryEntityType) String() string {
	switch t {
	case TopologyEntityExchange:
		return "exchange"
	case TopologyEntityQueue:
		return "queue"
	case TopologyEntityQueueBinding:
		return "queue-binding"
	case TopologyEntityExchangeBinding:
		return "exchange-binding"
	case TopologyEntityConsumer:
		return "consumer"
	default:
		return "unknown"
	}
}

// TopologyRecoveryEntity describes a single topology entity that failed during recovery.
// It is passed to Recovery.OnTopologyEntityError and collected in StateChanged.SkippedTopologyEntities.
//
// For non-binding entity types, EntityName uniquely identifies the entity.
// For binding types (QueueBinding, ExchangeBinding), use EntityName + SecondaryName + RoutingKey
// together — the same queue or exchange may appear in multiple bindings with different
// routing keys or destination exchanges.
type TopologyRecoveryEntity struct {
	EntityType TopologyRecoveryEntityType
	// EntityName is the primary name of the entity:
	//   - Exchange: the exchange name
	//   - Queue: the queue name
	//   - QueueBinding: the queue name
	//   - ExchangeBinding: the source exchange name
	//   - Consumer: the consumer tag
	EntityName string
	// SecondaryName disambiguates binding entities that share the same EntityName:
	//   - QueueBinding: the exchange the queue is bound to
	//   - ExchangeBinding: the destination exchange
	//   - Empty for all other entity types.
	SecondaryName string
	// RoutingKey is set for binding entities (QueueBinding, ExchangeBinding).
	// Empty for all other entity types.
	RoutingKey string
	ChannelID  uint16
	Err        error // the underlying broker or network error during the recovery.
}

func (e TopologyRecoveryEntity) Error() string {
	switch e.EntityType {
	case TopologyEntityQueueBinding:
		return fmt.Sprintf("topology recovery: queue-binding %q → %q [%s] on channel %d: %v",
			e.EntityName, e.SecondaryName, e.RoutingKey, e.ChannelID, e.Err)
	case TopologyEntityExchangeBinding:
		return fmt.Sprintf("topology recovery: exchange-binding %q → %q [%s] on channel %d: %v",
			e.EntityName, e.SecondaryName, e.RoutingKey, e.ChannelID, e.Err)
	default:
		return fmt.Sprintf("topology recovery: %s %q on channel %d: %v",
			e.EntityType, e.EntityName, e.ChannelID, e.Err)
	}
}

// Unwrap allows errors.As and errors.Is to inspect the underlying broker error,
// enabling callers to match on *amqp091.Error reply codes inside OnTopologyEntityError.
func (e TopologyRecoveryEntity) Unwrap() error { return e.Err }

const (
	// DefaultMaxRetryCount is the default maximum number of retries for recovery.
	DefaultMaxRetryCount = 5

	// DefaultRetryInterval is the default interval between retries for recovery.
	DefaultRetryInterval = 5 * time.Second
)

var (
	// DefaultReconnectionConfig is the default reconnection config settings.
	DefaultReconnectionConfig = &ReconnectionConfig{
		MaxRetryCount: DefaultMaxRetryCount,
		RetryInterval: DefaultRetryInterval,
	}
)

// ReconnectionConfig is the configuration for the reconnection.
type ReconnectionConfig struct {
	MaxRetryCount int           // The maximum number of retries.
	RetryInterval time.Duration // The interval between retries.
}

// Clone returns a deep copy of the ReconnectionConfig.
func (rc *ReconnectionConfig) Clone() *ReconnectionConfig {
	if rc == nil {
		return nil
	}
	return &ReconnectionConfig{
		MaxRetryCount: rc.MaxRetryCount,
		RetryInterval: rc.RetryInterval,
	}
}

// ConnectionRecovery is the interface for the connection recovery.
//
// The err parameter in OnConnectionClose and OnChannelClose provides the reason
// why the connection or channel was closed. Custom implementations of this interface
// can perform advanced logic, log errors to external monitoring systems (e.g., Prometheus),
// or trigger alerts.
type ConnectionRecovery interface {
	OnConnectionClose(conn *Connection, err *Error) // Called when the connection is closed.
	OnChannelClose(ch *Channel, err *Error)         // Called when the channel is closed.
}

// TopologyRecovery is the interface for the topology recovery.
//
// The default implementation is DefaultTopologyRecovery, which redeclares tracked
// exchanges, queues, and bindings, as well as re-subscribing active consumers.
//
// Custom implementations of this interface can be provided to:
//   - Instrument recovery with logging, tracing, or external metrics.
//   - Load and declare topology dynamically from an external config or registry.
//   - Rate-limit or stagger declarations to avoid overloading the broker after a reconnect.
//   - Perform pre-recovery checks or customized failover/fallback routines.
//
// RecoverTopology returns any entity-level errors that were skipped (via the
// Recovery.OnTopologyEntityError callback) as the first return value. A non-nil
// second return value means recovery failed fatally and the retry cycle continues.
type TopologyRecovery interface {
	RecoverTopology(conn *Connection, channels []*Channel) ([]TopologyRecoveryEntity, error)
}

// TopologyRecoveryMode controls which topology entities are recovered after a
// connection or channel is re-established. The modes are mutually exclusive.
type TopologyRecoveryMode byte

const (
	// TopologyRecoveryAllEnabled recovers all tracked topology: exchanges, queues,
	// bindings, exchange-to-exchange bindings, and active consumers.
	//
	// This is the default (the zero value of TopologyRecoveryMode).
	TopologyRecoveryAllEnabled TopologyRecoveryMode = iota

	// TopologyRecoveryOnlyTransient recovers only connection-scoped (transient)
	// entities: queues declared as exclusive and/or auto-delete (which includes
	// server-named queues), auto-delete exchanges, and any bindings that reference
	// one of those transient entities. Active consumers are still re-subscribed,
	// because consumer subscriptions are always lost on reconnect regardless of
	// queue durability.
	//
	// Durable, non-auto-delete exchanges and queues (and bindings between them) are
	// skipped because the broker retains them across a network interruption. Use
	// this mode when durable topology is managed declaratively or out-of-band and
	// only the connection-scoped entities need to be restored by the client.
	TopologyRecoveryOnlyTransient

	// TopologyRecoveryDisabled disables topology recovery completely. Neither
	// entities nor consumers are recovered. Connection and channel recovery still
	// occur if otherwise enabled.
	TopologyRecoveryDisabled
)

// Recovery is the configuration for the recovery.
type Recovery struct {
	ReconnectionConfig *ReconnectionConfig // The configuration for the reconnection.
	ConnectionRecovery ConnectionRecovery  // The implementation of the connection recovery.
	TopologyRecovery   TopologyRecovery    // The implementation of the topology recovery.

	// TopologyRecoveryMode controls which topology entities are recovered. The zero
	// value (TopologyRecoveryAllEnabled) recovers all tracked topology.
	// Setting it to TopologyRecoveryDisabled disables topology and consumer recovery entirely.
	TopologyRecoveryMode TopologyRecoveryMode

	// OnTopologyEntityError is called each time a single topology entity fails to
	// recover. Return true to skip the entity and continue recovering the remaining
	// entities. Return false to abort topology recovery and trigger the normal retry
	// cycle.
	//
	// The entity e carries enough information to identify the failed entity without
	// any additional lookup:
	//
	//	e.EntityType   — what kind of entity failed (exchange, queue, binding, consumer)
	//	e.EntityName   — primary name: exchange/queue name, source exchange, consumer tag
	//	e.SecondaryName — for bindings only: exchange (queue-binding) or destination exchange (exchange-binding)
	//	e.RoutingKey   — for bindings only: the routing key
	//	e.ChannelID    — the channel on which the failure occurred
	//	e.Err          — the underlying broker or network error
	//
	// For bindings, EntityName alone is not sufficient to identify the specific
	// binding — the same queue or exchange may be bound multiple times with different
	// routing keys. Use EntityName + SecondaryName + RoutingKey together.
	//
	// The conn parameter is the connection on which recovery is running. Use it to
	// retrieve the full declaration arguments of the failed entity when needed:
	//
	//	cfg := conn.Channel(e.ChannelID).TopologyConfiguration(true)
	//
	// The global flag merges topology from all channels into one connection-scoped
	// view — appropriate because AMQP exchanges, queues, and bindings are scoped to
	// the TCP connection, not individual channels. Pass false to limit the view to
	// entities declared on that channel only.
	//
	//	switch e.EntityType {
	//	case amqp091.TopologyEntityQueue:
	//	    qc, ok := cfg.Queues[e.EntityName]    // QueueConfig with Durable, AutoDelete, Args, etc.
	//	case amqp091.TopologyEntityExchange:
	//	    ec, ok := cfg.Exchanges[e.EntityName]  // ExchangeConfig with Kind, Durable, Args, etc.
	//	case amqp091.TopologyEntityQueueBinding:
	//	    // Match by Queue + Exchange + Key (args may further distinguish headers bindings).
	//	    for _, b := range cfg.Bindings {
	//	        if b.Queue == e.EntityName && b.Exchange == e.SecondaryName && b.Key == e.RoutingKey { ... }
	//	    }
	//	case amqp091.TopologyEntityExchangeBinding:
	//	    // Match by Source + Destination + Key.
	//	    for _, eb := range cfg.ExchangeBindings {
	//	        if eb.Source == e.EntityName && eb.Destination == e.SecondaryName && eb.Key == e.RoutingKey { ... }
	//	    }
	//	case amqp091.TopologyEntityConsumer:
	//	    // Consumer internals are not exposed via TopologyConfiguration.
	//	    // The consumer tag is available in e.EntityName.
	//	}
	//
	// Skipped-entity errors are collected and delivered to NotifyStateChange listeners
	// in the SkippedTopologyEntities field of the StateReconnecting→StateOpen transition, so the
	// application can observe what was not restored even on an otherwise successful reconnect.
	//
	// Default: nil — entity is skipped on failure and recovery continues (same as returning true).
	OnTopologyEntityError func(conn *Connection, e TopologyRecoveryEntity) bool
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

	Logger.Printf("Initiating connection recovery for %s.", parsedURL.Redacted())
	// Reconnect connection
	if err := conn.Reconnect(); err != nil {
		Logger.Printf("Connection %s recovery failed: %v.", parsedURL.Redacted(), err)
		conn.cleanup(err)
	}
}

func (d *DefaultConnectionRecovery) OnChannelClose(ch *Channel, err *Error) {
	Logger.Printf("Channel %d closed with error: %v", ch.id, err)
	if !ch.connection.IsRecoveryEnabled() {
		Logger.Printf("Channel %d recovery is not enabled, skipping reconnect.", ch.id)
		return
	}

	// Guard against concurrent recovery loops if the connection is already reconnecting
	if ch.connection.IsClosed() || ch.connection.lifeCycle.State() == StateReconnecting {
		Logger.Printf("Connection is closed or reconnecting, letting connection recovery handle channel %d.", ch.id)
		return
	}

	// Guard against a redundant, competing recovery pass: an in-flight
	// recoverConnectionTopology call already owns this channel's
	// reopen/redeclare sequence and will reopen it itself (see
	// Channel.reopenIfClosed) if a broker soft error closes it here.
	if ch.recoveringTopology.Load() {
		Logger.Printf("Channel %d topology recovery already in progress, skipping redundant reconnect.", ch.id)
		return
	}

	Logger.Printf("Initiating channel %d recovery", ch.id)
	// Reconnect channel
	if err := ch.Reconnect(); err != nil {
		Logger.Printf("Channel %d recovery failed: %v.", ch.id, err)
		ch.cleanup(err)
		ch.connection.releaseChannel(ch)
	}
}

// DefaultTopologyRecovery is the default implementation of the topology recovery.
type DefaultTopologyRecovery struct{}

func (d *DefaultTopologyRecovery) RecoverTopology(conn *Connection, channels []*Channel) ([]TopologyRecoveryEntity, error) {
	return conn.recoverConnectionTopology(channels)
}
