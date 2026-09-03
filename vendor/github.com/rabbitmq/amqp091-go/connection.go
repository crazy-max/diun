// Copyright (c) 2021 VMware, Inc. or its affiliates. All Rights Reserved.
// Copyright (c) 2012-2021, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package amqp091

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxChannelMax = (2 << 15) - 1

	defaultHeartbeat         = 10 * time.Second
	defaultConnectionTimeout = 30 * time.Second
	defaultProduct           = "AMQP 0.9.1 Client"
	buildVersion             = "1.13.0"
	platform                 = "golang"
	// Safer default that makes channel leaks a lot easier to spot
	// before they create operational headaches. See https://github.com/rabbitmq/rabbitmq-server/issues/1593.
	defaultChannelMax = uint16((2 << 10) - 1)
	defaultLocale     = "en_US"
)

// Config is used in DialConfig and Open to specify the desired tuning
// parameters used during a connection open handshake.  The negotiated tuning
// will be stored in the returned connection's Config field.
type Config struct {
	// The SASL mechanisms to try in the client request, and the successful
	// mechanism used on the Connection object.
	// If SASL is nil, PlainAuth from the URL is used.
	SASL []Authentication

	// Vhost specifies the namespace of permissions, exchanges, queues and
	// bindings on the server.  Dial sets this to the path parsed from the URL.
	Vhost string

	ChannelMax uint16        // 0 max channels means 2^16 - 1
	FrameSize  int           // 0 max bytes means unlimited
	Heartbeat  time.Duration // less than 1s uses the server's interval

	// TLSClientConfig specifies the client configuration of the TLS connection
	// when establishing a tls transport.
	// If the URL uses an amqps scheme, then an empty tls.Config with the
	// ServerName from the URL is used.
	TLSClientConfig *tls.Config

	// Properties is table of properties that the client advertises to the server.
	// This is an optional setting - if the application does not set this,
	// the underlying library will use a generic set of client properties.
	Properties Table

	// Connection locale that we expect to always be en_US
	// Even though servers must return it as per the AMQP 0-9-1 spec,
	// we are not aware of it being used other than to satisfy the spec requirements
	Locale string

	// Dial returns a net.Conn prepared for a TLS handshake with TSLClientConfig,
	// then an AMQP connection handshake.
	// If Dial is nil, net.DialTimeout with a 30s connection and 30s deadline is
	// used during TLS and AMQP handshaking.
	Dial func(network, addr string) (net.Conn, error)

	// Recovery configuration for automatic reconnection and topology recovery.
	//
	// Experimental: This is an experimental feature and may be subject to API or
	// behavioral changes in future releases.
	//
	// When a network failure occurs, the connection and all its channels will automatically
	// attempt to reconnect, and their topology (including queues, exchanges, bindings, and active consumers)
	// will be recovered based on the parameters specified in the Recovery configuration.
	//
	// If Recovery is nil, automatic reconnection and topology recovery are disabled.
	// If Recovery.ReconnectionConfig is nil, a default reconnection configuration (DefaultReconnectionConfig) is used.
	// If Recovery.ConnectionRecovery is nil, a default connection recovery implementation (DefaultConnectionRecovery) is used.
	// If Recovery.TopologyRecovery is nil, a default topology recovery implementation (DefaultTopologyRecovery) is used.
	//
	// Topology recovery scope is controlled by Recovery.TopologyRecoveryMode:
	//   - TopologyRecoveryAllEnabled (default): recovers all tracked topology (exchanges, queues,
	//     bindings, exchange-to-exchange bindings, and consumers).
	//   - TopologyRecoveryOnlyTransient: recovers only transient entities
	//     (exclusive/auto-delete queues, auto-delete exchanges, and bindings to them).
	//   - TopologyRecoveryDisabled: skips topology and consumer recovery entirely.
	//
	// During the recovery process, applications can monitor state changes (such as reconnecting
	// or closed) by registering a listener using `Connection.NotifyStateChange` and
	// `Channel.NotifyStateChange`.
	Recovery *Recovery
}

// NewConnectionProperties creates an amqp.Table to be used as amqp.Config.Properties.
//
// Defaults to library-defined values. For empty properties, use make(amqp.Table) instead.
func NewConnectionProperties() Table {
	return Table{
		"product":  defaultProduct,
		"version":  buildVersion,
		"platform": platform,
	}
}

// setSASL populates the SASL configuration from URI if it's not already set.
func (config *Config) setSASL(uri URI) error {
	if config.SASL == nil {
		if uri.AuthMechanism != nil {
			for _, identifier := range uri.AuthMechanism {
				switch strings.ToUpper(identifier) {
				case "PLAIN":
					config.SASL = append(config.SASL, uri.PlainAuth())
				case "AMQPLAIN":
					config.SASL = append(config.SASL, uri.AMQPlainAuth())
				case "EXTERNAL":
					config.SASL = append(config.SASL, &ExternalAuth{})
				default:
					return fmt.Errorf("unsupported auth_mechanism: %v", identifier)
				}
			}
		} else {
			config.SASL = []Authentication{uri.PlainAuth()}
		}
	}
	return nil
}

// Connection manages the serialization and deserialization of frames from IO
// and dispatches the frames to the appropriate channel.  All RPC methods and
// asynchronous Publishing, Delivery, Ack, Nack and Return messages are
// multiplexed on this channel.  There must always be active receivers for
// every asynchronous message on this connection.
type Connection struct {
	destructorM sync.Mutex   // Mutex for connection teardown: notifying close/block listeners, closing channels, and closing the underlying socket
	destructed  bool         // true when the connection has been destructed (teardown is initiated or completed)
	closeM      sync.Mutex   // Mutex for connection close handshake: sending a single connection.close frame to the broker
	closeInit   bool         // true when a connection close has been initiated (connection.close frame has been or is being sent)
	sendM       sync.Mutex   // conn writer mutex
	m           sync.Mutex   // struct field mutex
	notifyM     sync.RWMutex // Mutex for notifying close/block listeners; mirrors Channel.notifyM

	conn io.ReadWriteCloser

	rpc       chan message
	writer    *writer
	sends     chan time.Time     // timestamps of each frame sent
	deadlines chan readDeadliner // heartbeater updates read deadlines

	allocator *allocator // id generator valid after openTune
	channels  map[uint16]*Channel

	topologyM             sync.Mutex                        // Mutex for protecting connection-level topology configuration
	topologyConfiguration map[uint16]*TopologyConfiguration // connection-level topology indexed by channel ID

	noNotify bool // true when we will never notify again
	closes   []chan *Error
	blocks   []chan Blocking

	errors chan *Error
	// if connection is closed should close this chan
	close chan struct{}

	Config Config // The negotiated Config after connection.open

	url string // Connection URL stored for recovery

	Major      int      // Server's major version
	Minor      int      // Server's minor version
	Properties Table    // Server properties
	Locales    []string // Server locales

	closed atomic.Bool // Will be true if the connection is closed, false otherwise.

	// maxFrameSize mirrors Config.FrameSize once negotiated via connection.tune,
	// letting the reader goroutine reject over-sized frames before allocating
	// memory for them. 0 means no limit is enforced yet (or none was negotiated).
	maxFrameSize atomic.Uint32

	reconnecting sync.Mutex // Mutex for protecting reconnect/recovery operations to ensure serialization and prevent race conditions.
	lifeCycle    *lifeCycle // The current state of the connection.

	recoveryCancels []chan struct{} // listeners for connection recovery cancellation
}

type readDeadliner interface {
	SetReadDeadline(time.Time) error
}

// DefaultDial establishes a connection when config.Dial is not provided
func DefaultDial(connectionTimeout time.Duration) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, connectionTimeout)
		if err != nil {
			return nil, err
		}

		// Heartbeating hasn't started yet, don't stall forever on a dead server.
		// A deadline is set for TLS and AMQP handshaking. After AMQP is established,
		// the deadline is cleared in openComplete.
		if err := conn.SetDeadline(time.Now().Add(connectionTimeout)); err != nil {
			return nil, err
		}

		return conn, nil
	}
}

// Dial accepts a string in the AMQP URI format and returns a new Connection
// over TCP using PlainAuth.  Defaults to a server heartbeat interval of 10
// seconds and sets the handshake deadline to 30 seconds. After handshake,
// deadlines are cleared.
//
// Dial uses the zero value of tls.Config when it encounters an amqps://
// scheme.  It is equivalent to calling DialTLS(amqp, nil).
func Dial(url string) (*Connection, error) {
	return DialConfig(url, Config{
		Locale: defaultLocale,
	})
}

// DialTLS accepts a string in the AMQP URI format and returns a new Connection
// over TCP using PlainAuth.  Defaults to a server heartbeat interval of 10
// seconds and sets the initial read deadline to 30 seconds.
//
// DialTLS uses the provided tls.Config when encountering an amqps:// scheme.
// Note: If you provide a custom tls.Config, you should explicitly set a secure
// MinVersion (such as tls.VersionTLS12 or tls.VersionTLS13) as the library
// does not override it.
func DialTLS(url string, amqps *tls.Config) (*Connection, error) {
	return DialConfig(url, Config{
		TLSClientConfig: amqps,
		Locale:          defaultLocale,
	})
}

// DialTLS_ExternalAuth accepts a string in the AMQP URI format and returns a
// new Connection over TCP using EXTERNAL auth. Defaults to a server heartbeat
// interval of 10 seconds and sets the initial read deadline to 30 seconds.
//
// This mechanism is used, when RabbitMQ is configured for EXTERNAL auth with
// ssl_cert_login plugin for userless/passwordless logons
//
// DialTLS_ExternalAuth uses the provided tls.Config when encountering an
// amqps:// scheme.
// Note: If you provide a custom tls.Config, you should explicitly set a secure
// MinVersion (such as tls.VersionTLS12 or tls.VersionTLS13) as the library
// does not override it.
func DialTLS_ExternalAuth(url string, amqps *tls.Config) (*Connection, error) {
	return DialConfig(url, Config{
		TLSClientConfig: amqps,
		SASL:            []Authentication{&ExternalAuth{}},
	})
}

// DialConfig accepts a string in the AMQP URI format and a configuration for
// the transport and connection setup, returning a new Connection.  Defaults to
// a server heartbeat interval of 10 seconds and sets the initial read deadline
// to 30 seconds. The heartbeat interval specified in the AMQP URI takes precedence
// over the value specified in the config. To disable heartbeats, you must use
// the AMQP URI and set heartbeat=0 there.
func DialConfig(url string, config Config) (*Connection, error) {
	var err error
	var conn net.Conn

	uri, err := ParseURI(url)
	if err != nil {
		return nil, err
	}

	if config.Locale == "" {
		config.Locale = defaultLocale
	}

	if err := config.setSASL(uri); err != nil {
		return nil, err
	}

	if config.Vhost == "" {
		config.Vhost = uri.Vhost
	}

	if uri.Heartbeat.hasValue {
		config.Heartbeat = uri.Heartbeat.value
	} else {
		if config.Heartbeat == 0 {
			config.Heartbeat = defaultHeartbeat
		}
	}

	if config.ChannelMax == 0 {
		config.ChannelMax = uri.ChannelMax
	}

	connectionTimeout := defaultConnectionTimeout
	if uri.ConnectionTimeout != 0 {
		connectionTimeout = time.Duration(uri.ConnectionTimeout) * time.Millisecond
	}

	addr := net.JoinHostPort(uri.Host, strconv.FormatInt(int64(uri.Port), 10))

	dialer := config.Dial
	if dialer == nil {
		dialer = DefaultDial(connectionTimeout)
	}

	conn, err = dialer("tcp", addr)
	if err != nil {
		return nil, err
	}

	if uri.Scheme == "amqps" {
		if config.TLSClientConfig == nil {
			tlsConfig, err := tlsConfigFromURI(uri)
			if err != nil {
				return nil, fmt.Errorf("create TLS config from URI: %w", err)
			}
			config.TLSClientConfig = tlsConfig
		}

		// If ServerName has not been specified in TLSClientConfig,
		// set it to the URI host used for this connection.
		if config.TLSClientConfig.ServerName == "" {
			config.TLSClientConfig.ServerName = uri.Host
		}

		client := tls.Client(conn, config.TLSClientConfig)
		if err := client.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}

		conn = client
	}

	if config.Recovery != nil {
		if config.Recovery.ReconnectionConfig == nil {
			config.Recovery.ReconnectionConfig = DefaultReconnectionConfig.Clone()
		}
		if config.Recovery.ConnectionRecovery == nil {
			config.Recovery.ConnectionRecovery = &DefaultConnectionRecovery{}
		}
		if config.Recovery.TopologyRecovery == nil && config.Recovery.TopologyRecoveryMode != TopologyRecoveryDisabled {
			config.Recovery.TopologyRecovery = &DefaultTopologyRecovery{}
		}
	}

	c, err := Open(conn, config)
	if c != nil && c.IsRecoveryEnabled() {
		c.url = url
		c.watchConnection()
	}

	return c, err
}

/*
Open accepts an already established connection, or other io.ReadWriteCloser as
a transport.  Use this method if you have established a TLS connection or wish
to use your own custom transport.
*/
func Open(conn io.ReadWriteCloser, config Config) (*Connection, error) {
	c := &Connection{
		conn:                  conn,
		writer:                &writer{bufio.NewWriter(conn)},
		channels:              make(map[uint16]*Channel),
		topologyConfiguration: make(map[uint16]*TopologyConfiguration),
		rpc:                   make(chan message),
		sends:                 make(chan time.Time),
		errors:                make(chan *Error, 1),
		close:                 make(chan struct{}),
		deadlines:             make(chan readDeadliner, 1),
		Config:                config,
		lifeCycle:             newLifeCycle(),
	}
	go c.reader(conn)
	err := c.open(config)
	if err == nil {
		c.lifeCycle.SetState(StateOpen, nil)
	}
	return c, err
}

/*
UpdateSecret updates the secret used to authenticate this connection. It is used when
secrets have an expiration date and need to be renewed, like OAuth 2 tokens.

It returns an error if the operation is not successful, or if the connection is closed.
*/
func (c *Connection) UpdateSecret(newSecret, reason string) error {
	if c.IsClosed() {
		return ErrClosed
	}
	return c.call(&connectionUpdateSecret{
		NewSecret: newSecret,
		Reason:    reason,
	}, &connectionUpdateSecretOk{})
}

/*
LocalAddr returns the local TCP peer address, or ":0" (the zero value of net.TCPAddr)
as a fallback default value if the underlying transport does not support LocalAddr().
*/
func (c *Connection) LocalAddr() net.Addr {
	if conn, ok := c.conn.(interface {
		LocalAddr() net.Addr
	}); ok {
		return conn.LocalAddr()
	}
	return &net.TCPAddr{}
}

/*
RemoteAddr returns the remote TCP peer address, if known.
*/
func (c *Connection) RemoteAddr() net.Addr {
	if conn, ok := c.conn.(interface {
		RemoteAddr() net.Addr
	}); ok {
		return conn.RemoteAddr()
	}
	return &net.TCPAddr{}
}

// ConnectionState returns basic TLS details of the underlying transport.
// Returns a zero value when the underlying connection does not implement
// ConnectionState() tls.ConnectionState.
func (c *Connection) ConnectionState() tls.ConnectionState {
	if conn, ok := c.conn.(interface {
		ConnectionState() tls.ConnectionState
	}); ok {
		return conn.ConnectionState()
	}
	return tls.ConnectionState{}
}

// NotifyStateChange registers a listener for state changes.
//
// It is necessary to continuously consume from the channel passed to NotifyStateChange
// to avoid blocking internal state dispatch routines and leaking goroutines.
func (c *Connection) NotifyStateChange(ch chan *StateChanged) {
	c.lifeCycle.notifyStateChange(ch)
}

/*
NotifyClose registers a listener for close events either initiated by an error
accompanying a connection.close method or by a normal shutdown.

The chan provided will be closed when the Connection is closed and on a
graceful close, no error will be sent.

In case of a non graceful close the error will be notified synchronously by the library
so that it will be necessary to consume the Channel from the caller in order to avoid deadlocks

To reconnect after a transport or protocol error, register a listener here and
re-run your setup process.
*/
func (c *Connection) NotifyClose(receiver chan *Error) chan *Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.noNotify {
		close(receiver)
	} else {
		c.closes = append(c.closes, receiver)
	}

	return receiver
}

/*
NotifyRecoveryCancel registers a listener that is notified (via a channel close)
when connection recovery has been canceled or aborted (for example, when Close()
or CloseDeadline() is called during an active reconnect process).

The returned channel will be closed immediately if the connection is already closing
or closed, or when Close() is called.
*/
func (c *Connection) NotifyRecoveryCancel(receiver chan struct{}) chan struct{} {
	c.m.Lock()
	defer c.m.Unlock()

	state := c.lifeCycle.State()
	if state == StateClosing || state == StateClosed {
		close(receiver)
	} else {
		c.recoveryCancels = append(c.recoveryCancels, receiver)
	}

	return receiver
}

// closeRecovery stops any active connection recovery process by notifying
// and closing all recovery cancellation listeners.
func (c *Connection) closeRecovery() {
	c.m.Lock()
	defer c.m.Unlock()

	for _, listener := range c.recoveryCancels {
		close(listener)
	}
	c.recoveryCancels = nil
}

/*
NotifyBlocked registers a listener for RabbitMQ specific TCP flow control
method extensions connection.blocked and connection.unblocked.  Flow control is
active with a reason when Blocking.Blocked is true.  When a Connection is
blocked, all methods will block across all connections until server resources
become free again.

This optional extension is supported by the server when the
"connection.blocked" server capability key is true.
*/
func (c *Connection) NotifyBlocked(receiver chan Blocking) chan Blocking {
	c.m.Lock()
	defer c.m.Unlock()

	if c.noNotify {
		close(receiver)
	} else {
		c.blocks = append(c.blocks, receiver)
	}

	return receiver
}

/*
Close requests and waits for the response to close the AMQP connection.

It's advisable to use this message when publishing to ensure all kernel buffers
have been flushed on the server and client before exiting.

An error indicates that server may not have received this request to close but
the connection should be treated as closed regardless.

After returning from this call, all resources associated with this connection,
including the underlying io, Channels, Notify listeners and Channel consumers
will also be closed.
*/
func (c *Connection) Close() error {
	c.closeRecovery() // Stop any active recovery process

	if c.IsClosed() {
		return ErrClosed
	}

	c.lifeCycle.SetState(StateClosing, nil)

	var handshakeErr error
	var initiated bool

	c.closeM.Lock()
	if !c.closeInit {
		c.closeInit = true
		initiated = true
	}
	c.closeM.Unlock()

	if initiated {
		defer c.shutdown(nil)
		handshakeErr = c.call(
			&connectionClose{
				ReplyCode: replySuccess,
				ReplyText: "kthxbai",
			},
			&connectionCloseOk{},
		)
	}
	if !initiated {
		return ErrClosed
	}
	return handshakeErr
}

// CloseDeadline requests and waits for the response to close this AMQP connection.
//
// Accepts a deadline for waiting the server response. The deadline is passed
// to the low-level connection i.e. network socket.
//
// Regardless of the error returned, the connection is considered closed, and it
// should not be used after calling this function.
//
// In the event of an I/O timeout, connection-closed listeners are NOT informed.
//
// After returning from this call, all resources associated with this connection,
// including the underlying io, Channels, Notify listeners and Channel consumers
// will also be closed.
func (c *Connection) CloseDeadline(deadline time.Time) error {
	c.closeRecovery() // Stop any active recovery process

	if c.IsClosed() {
		return ErrClosed
	}

	var handshakeErr error
	var initiated bool

	c.closeM.Lock()
	if !c.closeInit {
		c.closeInit = true
		initiated = true
	}
	c.closeM.Unlock()

	if initiated {
		defer c.shutdown(nil)
		if err := c.setDeadline(deadline); err != nil {
			return err
		}
		handshakeErr = c.call(
			&connectionClose{
				ReplyCode: replySuccess,
				ReplyText: "kthxbai",
			},
			&connectionCloseOk{},
		)
	}
	if !initiated {
		return ErrClosed
	}
	return handshakeErr
}

func (c *Connection) closeWith(err *Error) error {
	if c.IsClosed() {
		return ErrClosed
	}

	var handshakeErr error
	var initiated bool

	c.closeM.Lock()
	if !c.closeInit {
		c.closeInit = true
		initiated = true
	}
	c.closeM.Unlock()

	if initiated {
		defer c.shutdown(err)
		handshakeErr = c.call(
			&connectionClose{
				ReplyCode: uint16(err.Code),
				ReplyText: err.Reason,
			},
			&connectionCloseOk{},
		)
	}
	if !initiated {
		return ErrClosed
	}
	return handshakeErr
}

// IsClosed returns true if the connection is marked as closed, otherwise false
// is returned.
func (c *Connection) IsClosed() bool {
	return c.closed.Load()
}

// setDeadline is a wrapper to type assert Connection.conn and set an I/O
// deadline in the underlying TCP connection socket, by calling
// net.Conn.SetDeadline(). It returns an error, in case the type assertion fails,
// although this should never happen.
func (c *Connection) setDeadline(t time.Time) error {
	con, ok := c.conn.(net.Conn)
	if !ok {
		return errInvalidTypeAssertion
	}
	return con.SetDeadline(t)
}

func (c *Connection) send(f frame) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.sendM.Lock()
	err := c.writer.WriteFrame(f)
	c.sendM.Unlock()

	if err != nil {
		// shutdown could be re-entrant from signaling notify chans
		go c.shutdown(&Error{
			Code:   FrameError,
			Reason: err.Error(),
		})
	} else {
		// Broadcast we sent a frame, reducing heartbeats, only
		// if there is something that can receive - like a non-reentrant
		// call or if the heartbeater isn't running
		select {
		case c.sends <- time.Now():
		default:
		}
	}

	return err
}

// This method is intended to be used with sendUnflushed() to end a sequence
// of sendUnflushed() calls and flush the connection
func (c *Connection) endSendUnflushed() error {
	c.sendM.Lock()
	defer c.sendM.Unlock()
	return c.flush()
}

// sendUnflushed performs an *Unflushed* write. It is otherwise equivalent to
// send(), and we provide a separate flush() function to explicitly flush the
// buffer after all Frames are written.
//
// Why is this a thing?
//
// send() method uses writer.WriteFrame(), which will write the Frame then
// flush the buffer. For cases like the sendOpen() method on Channel, which
// sends multiple Frames (methodFrame, headerFrame, N x bodyFrame), flushing
// after each Frame is inefficient as it negates much of the benefit of using a
// buffered writer, and results in more syscalls than necessary. Flushing buffers
// after every frame can have a significant performance impact when sending
// (basicPublish) small messages, so this method performs an *Unflushed* write
// but is otherwise equivalent to send() method, and we provide a separate
// flush method to explicitly flush the buffer after all Frames are written.
func (c *Connection) sendUnflushed(f frame) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.sendM.Lock()
	err := c.writer.WriteFrameNoFlush(f)
	c.sendM.Unlock()

	if err != nil {
		// shutdown could be re-entrant from signaling notify chans
		go c.shutdown(&Error{
			Code:   FrameError,
			Reason: err.Error(),
		})
	}

	return err
}

// This method is intended to be used with sendUnflushed() to explicitly flush
// the buffer after all required Frames have been written to the buffer.
func (c *Connection) flush() (err error) {
	if buf, ok := c.writer.w.(*bufio.Writer); ok {
		err = buf.Flush()

		// Moving send notifier to flush increases basicPublish for the small message
		// case. As sendUnflushed + flush is used for the case of sending semantically
		// related Frames (e.g. a Message like basicPublish) there is no real advantage
		// to sending per Frame vice per "group of related Frames" and for the case of
		// small messages time.Now() is (relatively) expensive.
		if err == nil {
			// Broadcast we sent a frame, reducing heartbeats, only
			// if there is something that can receive - like a non-reentrant
			// call or if the heartbeater isn't running
			select {
			case c.sends <- time.Now():
			default:
			}
		}
	}

	return
}

func (c *Connection) shutdown(err *Error) {
	c.closed.Store(true)

	c.destructorM.Lock()
	if c.destructed {
		c.destructorM.Unlock()
		return
	}
	c.destructed = true
	defer c.destructorM.Unlock()

	c.m.Lock()
	defer c.m.Unlock()

	c.notifyM.Lock()
	defer c.notifyM.Unlock()

	if err != nil {
		for _, listener := range c.closes {
			select {
			case listener <- err:
			default:
				// Channel is full; deliver in a background goroutine so we never deadlock
				// the shutdown sequence. The goroutine holds notifyM.RLock() for the
				// duration of the send, which is mutually exclusive with cleanup()'s
				// notifyM.Lock(), preventing a concurrent send+close data race.
				go func(listener chan *Error, err *Error) {
					defer func() { _ = recover() }()
					c.notifyM.RLock()
					defer c.notifyM.RUnlock()
					select {
					case listener <- err:
					case <-time.After(5 * time.Second):
					}
				}(listener, err)
			}
		}

		select {
		case c.errors <- err:
		default:
		}
	}

	// Shutdown handler goroutine can still receive the result.
	close(c.errors)

	if err == nil || !c.IsRecoveryEnabled() {
		for _, listener := range c.closes {
			close(listener)
		}
		for _, block := range c.blocks {
			close(block)
		}
		c.closes, c.blocks = nil, nil // nil to prevent double-close
	}

	// Shutdown the channel, but do not use closeChannel() as it calls
	// releaseChannel() which requires the connection lock.
	//
	// Ranging over c.channels and calling releaseChannel() that mutates
	// c.channels is racy - see commit 6063341 for an example.
	for _, ch := range c.channels {
		ch.shutdown(err)
	}

	c.conn.Close()
	// reader exit
	close(c.close)

	if err == nil || !c.IsRecoveryEnabled() {
		c.channels = nil
		c.allocator = nil
		c.noNotify = true

		var e error
		if err != nil {
			e = fmt.Errorf("connection shutdown error: %w", err) // errors.As(e, &target) still unwraps to *Error
		}
		c.lifeCycle.SetState(StateClosed, e)
	}
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (c *Connection) demux(f frame) {
	if f.channel() == 0 {
		c.dispatch0(f)
	} else {
		c.dispatchN(f)
	}
}

func (c *Connection) dispatch0(f frame) {
	switch mf := f.(type) {
	case *methodFrame:
		switch m := mf.Method.(type) {
		case *connectionClose:
			// Send immediately as shutdown will close our side of the writer.
			f := &methodFrame{ChannelId: 0, Method: &connectionCloseOk{}}
			if err := c.send(f); err != nil {
				Logger.Printf("error sending connectionCloseOk, error: %+v", err)
			}
			c.shutdown(newError(m.ReplyCode, m.ReplyText))
		case *connectionBlocked:
			c.m.Lock()
			blocks := c.blocks
			c.m.Unlock()
			notifyAll(blocks, Blocking{Active: true, Reason: m.Reason})
		case *connectionUnblocked:
			c.m.Lock()
			blocks := c.blocks
			c.m.Unlock()
			notifyAll(blocks, Blocking{Active: false})
		default:
			select {
			case <-c.close:
				return
			case c.rpc <- m:
			}
		}
	case *heartbeatFrame:
		// kthx - all reads reset our deadline.  so we can drop this
	default:
		// lolwat - channel0 only responds to methods and heartbeats
		// closeWith use call don't block reader
		go func() {
			if err := c.closeWith(ErrUnexpectedFrame); err != nil {
				Logger.Printf("error sending connectionCloseOk with ErrUnexpectedFrame, error: %+v", err)
			}
		}()
	}
}

func (c *Connection) dispatchN(f frame) {
	c.m.Lock()
	channel, ok := c.channels[f.channel()]
	if ok {
		updateChannel(f, channel)
	} else {
		Logger.Printf("[debug] dropping frame, channel %d does not exist", f.channel())
	}
	c.m.Unlock()

	// Note: this could result in concurrent dispatch depending on
	// how channels are managed in an application
	if ok {
		channel.recv(channel, f)
	} else {
		c.dispatchClosed(f)
	}
}

// section 2.3.7: "When a peer decides to close a channel or connection, it
// sends a Close method.  The receiving peer MUST respond to a Close with a
// Close-Ok, and then both parties can close their channel or connection.  Note
// that if peers ignore Close, deadlock can happen when both peers send Close
// at the same time."
//
// When we don't have a channel, so we must respond with close-ok on a close
// method.  This can happen between a channel exception on an asynchronous
// method like basic.publish and a synchronous close with channel.close.
// In that case, we'll get both a channel.close and channel.close-ok in any
// order.
func (c *Connection) dispatchClosed(f frame) {
	// Only consider method frames, drop content/header frames
	if mf, ok := f.(*methodFrame); ok {
		switch mf.Method.(type) {
		case *channelClose:
			f := &methodFrame{ChannelId: f.channel(), Method: &channelCloseOk{}}
			if err := c.send(f); err != nil {
				Logger.Printf("error sending channelCloseOk, channel id: %d error: %+v", f.channel(), err)
			}
		case *channelCloseOk:
			// we are already closed, so do nothing
		default:
			// unexpected method on closed channel
			// closeWith use call don't block reader
			go func() {
				if err := c.closeWith(ErrClosed); err != nil {
					Logger.Printf("error sending connectionCloseOk with ErrClosed, error: %+v", err)
				}
			}()
		}
	}
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (c *Connection) reader(r io.Reader) {
	buf := bufio.NewReader(r)
	frames := &reader{r: buf, maxFrameSize: &c.maxFrameSize}
	conn, haveDeadliner := r.(readDeadliner)

	defer close(c.rpc)

	for {
		frame, err := frames.ReadFrame()
		if err != nil {
			c.shutdown(&Error{Code: FrameError, Reason: err.Error()})
			return
		}

		c.demux(frame)

		if haveDeadliner {
			select {
			case c.deadlines <- conn:
			default:
				// On c.Close() c.heartbeater() might exit just before c.deadlines <- conn is called.
				// Which results in this goroutine being stuck forever.
			}
		}
	}
}

// Ensures that at least one frame is being sent at the tuned interval with a
// jitter tolerance of 1s
func (c *Connection) heartbeater(interval time.Duration, done chan struct{}) {
	const maxServerHeartbeatsInFlight = 3

	var sendTicks <-chan time.Time
	if interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		sendTicks = ticker.C
	}

	lastSent := time.Now()

	for {
		select {
		case at, stillSending := <-c.sends:
			// When actively sending, depend on sent frames to reset server timer
			if stillSending {
				lastSent = at
			} else {
				return
			}

		case at := <-sendTicks:
			// When idle, fill the space with a heartbeat frame
			if at.Sub(lastSent) > interval-time.Second {
				if err := c.send(&heartbeatFrame{}); err != nil {
					// send heartbeats even after close/closeOk so we
					// tick until the connection starts erroring
					return
				}
			}

		case conn := <-c.deadlines:
			// When reading, reset our side of the deadline, if we've negotiated one with
			// a deadline that covers at least 2 server heartbeats
			if interval > 0 {
				if err := conn.SetReadDeadline(time.Now().Add(maxServerHeartbeatsInFlight * interval)); err != nil {
					var opErr *net.OpError
					if !errors.As(err, &opErr) {
						Logger.Printf("error setting read deadline in heartbeater: %+v", err)
						return
					}
				}
			}

		case <-done:
			return
		}
	}
}

// Convenience method to inspect the Connection.Properties["capabilities"]
// Table for server identified capabilities like "basic.ack" or
// "confirm.select".
func (c *Connection) isCapable(featureName string) bool {
	capabilities, _ := c.Properties["capabilities"].(Table)
	hasFeature, _ := capabilities[featureName].(bool)
	return hasFeature
}

// allocateChannel records but does not open a new channel with a unique id.
// This method is the initial part of the channel lifecycle and paired with
// releaseChannel
func (c *Connection) allocateChannel() (*Channel, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.IsClosed() {
		return nil, ErrClosed
	}

	id, ok := c.allocator.next()
	if !ok {
		return nil, ErrChannelMax
	}

	ch := newChannel(c, uint16(id))
	c.channels[uint16(id)] = ch

	return ch, nil
}

// releaseChannel removes a channel from the registry as the final part of the
// channel lifecycle
func (c *Connection) releaseChannel(ch *Channel) {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.IsClosed() {
		got, ok := c.channels[ch.id]
		if ok && got == ch {
			delete(c.channels, ch.id)
			c.allocator.release(int(ch.id))
			c.removeChannelTopology(ch.id)
		}
	}
}

// reregisterChannel re-adds a channel to the registry if it is missing, without
// allocating a new id. Used by Channel.reopenIfClosed to restore a channel that
// topology recovery kept around (see closeChannel) after a broker soft error.
func (c *Connection) reregisterChannel(ch *Channel) {
	c.m.Lock()
	defer c.m.Unlock()

	if _, present := c.channels[ch.id]; !present && c.allocator != nil {
		c.channels[ch.id] = ch
		c.allocator.reserve(int(ch.id))
	}
}

// openChannel allocates and opens a channel, must be paired with closeChannel
func (c *Connection) openChannel() (*Channel, error) {
	ch, err := c.allocateChannel()
	if err != nil {
		return nil, err
	}

	if err := ch.open(); err != nil {
		c.releaseChannel(ch)
		return nil, err
	}

	ch.lifeCycle.SetState(StateOpen, nil)

	if c.IsRecoveryEnabled() {
		ch.watchChannel()
	}

	return ch, nil
}

// closeChannel releases and initiates a shutdown of the channel.  All channel
// closures should be initiated here for proper channel lifecycle management on
// this connection.
func (c *Connection) closeChannel(ch *Channel, e *Error) {
	// If we are actively recovering topology, do NOT purge the channel from memory yet.
	// We keep it in c.channels so that:
	//   1. Channel.reopenIfClosed can find and reuse it if we skip-and-continue.
	//   2. connection.cleanup can access it to force a full shutdown if recovery fails completely.
	if e == nil || !c.IsRecoveryEnabled() {
		c.releaseChannel(ch)
	}
	ch.shutdown(e)
}

/*
Channel opens a unique, concurrent server channel to process the bulk of AMQP
messages.  Any error from methods on this receiver will render the receiver
invalid and a new Channel should be opened.

Channels are not thread-safe. To avoid unexpected behavior, do not share
a single Channel instance between multiple goroutines. Concurrent calls
to Channel methods may result in race conditions or unpredictable outcomes.
*/
func (c *Connection) Channel() (*Channel, error) {
	return c.openChannel()
}

func (c *Connection) call(req message, res ...message) error {
	// Special case for when the protocol header frame is sent instead of a
	// request method
	if req != nil {
		if err := c.send(&methodFrame{ChannelId: 0, Method: req}); err != nil {
			return err
		}
	}

	c.m.Lock()
	errors := c.errors
	rpc := c.rpc
	c.m.Unlock()

	var msg message
	select {
	case e, ok := <-errors:
		if ok {
			return e
		}
		return ErrClosed
	case msg = <-rpc:
	}

	// Try to match one of the result types
	for _, try := range res {
		if reflect.TypeOf(msg) == reflect.TypeOf(try) {
			// *res = *msg
			vres := reflect.ValueOf(try).Elem()
			vmsg := reflect.ValueOf(msg).Elem()
			vres.Set(vmsg)
			return nil
		}
	}
	return ErrCommandInvalid
}

// Communication flow to open, use and close a connection. 'C:' are
// frames sent by the Client. 'S:' are frames sent by the Server.
//
//	Connection          = open-Connection *use-Connection close-Connection
//
//	open-Connection     = C:protocol-header
//	                      S:START C:START-OK
//	                      *challenge
//	                      S:TUNE C:TUNE-OK
//	                      C:OPEN S:OPEN-OK
//
//	challenge           = S:SECURE C:SECURE-OK
//
//	use-Connection      = *channel
//
//	close-Connection    = C:CLOSE S:CLOSE-OK
//	                      S:CLOSE C:CLOSE-OK
func (c *Connection) open(config Config) error {
	if err := c.send(&protocolHeader{}); err != nil {
		return err
	}

	return c.openStart(config)
}

func (c *Connection) openStart(config Config) error {
	start := &connectionStart{}

	if err := c.call(nil, start); err != nil {
		return err
	}

	c.Major = int(start.VersionMajor)
	c.Minor = int(start.VersionMinor)
	c.Properties = start.ServerProperties
	c.Locales = strings.Split(start.Locales, " ")

	// eventually support challenge/response here by also responding to
	// connectionSecure.
	auth, ok := pickSASLMechanism(config.SASL, strings.Split(start.Mechanisms, " "))
	if !ok {
		return ErrSASL
	}

	// Save this mechanism off as the one we chose
	c.Config.SASL = []Authentication{auth}

	// Set the connection locale to client locale
	c.Config.Locale = config.Locale

	return c.openTune(config, auth)
}

func (c *Connection) openTune(config Config, auth Authentication) error {
	if len(config.Properties) == 0 {
		config.Properties = NewConnectionProperties()
	}

	config.Properties["capabilities"] = Table{
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"basic.nack":             true,
		"publisher_confirms":     true,
	}

	ok := &connectionStartOk{
		ClientProperties: config.Properties,
		Mechanism:        auth.Mechanism(),
		Response:         auth.Response(),
		Locale:           config.Locale,
	}
	tune := &connectionTune{}

	if err := c.call(ok, tune); err != nil {
		// per spec, a connection can only be closed when it has been opened
		// so at this point, we know it's an auth error, but the socket
		// was closed instead.  Return a meaningful error.
		return ErrCredentials
	}

	// Edge case that may race with c.shutdown()
	// https://github.com/rabbitmq/amqp091-go/issues/170
	c.m.Lock()

	// When the server and client both use default 0, then the max channel is
	// only limited by uint16.
	c.Config.ChannelMax = pickUInt16(config.ChannelMax, tune.ChannelMax)
	if c.Config.ChannelMax == 0 {
		c.Config.ChannelMax = defaultChannelMax
	}
	c.Config.ChannelMax = minUInt16(c.Config.ChannelMax, maxChannelMax)

	c.allocator = newAllocator(1, int(c.Config.ChannelMax))
	// Reserve all the channels that are already open
	// This is to avoid allocating the same channel ID after a reconnection
	for id := range c.channels {
		c.allocator.reserve(int(id))
	}

	c.m.Unlock()

	// Frame size includes headers and end byte (len(payload)+8). Enforce the spec
	// minimum floor of frameMinSize (4096 bytes) to prevent malicious servers
	// from forcing extreme fragmentation and CPU overhead.
	c.Config.FrameSize = negotiateFrameSize(config.FrameSize, int(tune.FrameMax))
	// This is the only place maxFrameSize is written. reader.ReadFrame relies on
	// any nonzero value here being >= frameMinSize (negotiateFrameSize's floor)
	// to safely subtract frameHeaderSize without underflow — keep it that way if
	// another write path is ever added.
	c.maxFrameSize.Store(uint32(c.Config.FrameSize))

	// Save this off for resetDeadline()
	c.Config.Heartbeat = time.Second * time.Duration(pick(
		int(config.Heartbeat/time.Second),
		int(tune.Heartbeat)))

	// "The client should start sending heartbeats after receiving a
	// Connection.Tune method"
	go c.heartbeater(c.Config.Heartbeat/2, c.close)

	if err := c.send(&methodFrame{
		ChannelId: 0,
		Method: &connectionTuneOk{
			ChannelMax: uint16(c.Config.ChannelMax),
			FrameMax:   uint32(c.Config.FrameSize),
			Heartbeat:  uint16(c.Config.Heartbeat / time.Second),
		},
	}); err != nil {
		return err
	}

	return c.openVhost(config)
}

func (c *Connection) openVhost(config Config) error {
	req := &connectionOpen{VirtualHost: config.Vhost}
	res := &connectionOpenOk{}

	if err := c.call(req, res); err != nil {
		// Cannot be closed yet, but we know it's a vhost problem
		return ErrVhost
	}

	c.Config.Vhost = config.Vhost

	return c.openComplete()
}

// openComplete performs any final Connection initialization dependent on the
// connection handshake and clears any state needed for TLS and AMQP handshaking.
func (c *Connection) openComplete() error {
	// We clear the deadlines and let the heartbeater reset the read deadline if requested.
	// RabbitMQ uses TCP flow control at this point for pushback so Writes can
	// intentionally block.
	if deadliner, ok := c.conn.(interface {
		SetDeadline(time.Time) error
	}); ok {
		_ = deadliner.SetDeadline(time.Time{})
	}

	// Zero out credentials after successful open-ok
	for i := range c.Config.SASL {
		if pa, ok := c.Config.SASL[i].(*PlainAuth); ok {
			pa.Password = ""
		} else if apa, ok := c.Config.SASL[i].(*AMQPlainAuth); ok {
			apa.Password = ""
		}
	}

	return nil
}

// tlsConfigFromURI tries to create TLS configuration based on query parameters.
// Returns default (empty) config in case no suitable client cert and/or client key not provided.
// Returns error in case certificates can not be parsed.
func tlsConfigFromURI(uri URI) (*tls.Config, error) {
	var certPool *x509.CertPool
	if uri.CACertFile != "" {
		data, err := os.ReadFile(uri.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("read CA certificate: %w", err)
		}

		certPool = x509.NewCertPool()
		certPool.AppendCertsFromPEM(data)
	} else if sysPool, err := x509.SystemCertPool(); err != nil {
		return nil, fmt.Errorf("load system certificates: %w", err)
	} else {
		certPool = sysPool
	}

	if uri.CertFile == "" || uri.KeyFile == "" {
		// no client auth (mTLS), just server auth
		return &tls.Config{
			RootCAs:    certPool,
			ServerName: uri.ServerName,
			MinVersion: tls.VersionTLS12,
		}, nil
	}

	certificate, err := tls.LoadX509KeyPair(uri.CertFile, uri.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ServerName:   uri.ServerName,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxUInt16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minUInt16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

func pickUInt16(client, server uint16) uint16 {
	if client == 0 || server == 0 {
		return maxUInt16(client, server)
	} else {
		return minUInt16(client, server)
	}
}

func pick(client, server int) int {
	if client == 0 || server == 0 {
		return max(client, server)
	}
	return min(client, server)
}

func negotiateFrameSize(client, server int) int {
	size := pick(client, server)
	if size > 0 && size < frameMinSize {
		return frameMinSize
	}
	return size
}

// cleanup releases registered resources and performs final teardown of the connection.
func (c *Connection) cleanup(err error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.notifyM.Lock()
	defer c.notifyM.Unlock()

	for _, listener := range c.closes {
		close(listener)
	}
	for _, block := range c.blocks {
		close(block)
	}
	c.closes, c.blocks = nil, nil // nil to prevent double-close

	for _, ch := range c.channels {
		ch.cleanup(err)
	}

	for _, listener := range c.recoveryCancels {
		close(listener)
	}

	c.recoveryCancels = nil
	c.channels = nil
	c.allocator = nil
	c.noNotify = true

	c.topologyM.Lock()
	c.topologyConfiguration = nil
	c.topologyM.Unlock()

	var e error
	if err != nil {
		e = fmt.Errorf("connection cleanup error: %w", err)
	}

	c.lifeCycle.SetState(StateClosed, e)
}

// watchConnection watches the connection for close events and triggers recovery if needed.
func (c *Connection) watchConnection() {
	errCh := c.NotifyClose(make(chan *Error, 1))
	go func() {
		for err := range errCh {
			if err != nil {
				Logger.Printf("Connection closed unexpectedly: %v", err)
				if c.IsConnectionRecoveryEnabled() {
					c.Config.Recovery.ConnectionRecovery.OnConnectionClose(c, err)
				}
			}
		}
	}()
}

// Reconnect establishes a new socket connection, replaces the existing closed/closing connection,
// and recovers both connection and channel states.
// It performs a retry loop to dial the broker, negotiate the AMQP handshake, and recover all
// active channels and their registered consumers sequentially.
func (c *Connection) Reconnect() error {
	if !c.IsRecoveryEnabled() {
		return ErrClosed
	}

	c.reconnecting.Lock()
	defer c.reconnecting.Unlock()

	if !c.IsClosed() {
		return nil
	}

	c.lifeCycle.SetState(StateReconnecting, nil)

	cancelCh := c.NotifyRecoveryCancel(make(chan struct{}))

	var err error
	for i := 0; i < c.MaxRetryCount(); i++ {
		Logger.Printf("Connection recovery attempt %d of %d", i+1, c.MaxRetryCount())
		jitter := time.Duration(rand.Intn(500)) * time.Millisecond // Random 500ms jitter to avoid thundering herd

		// Wait with select to allow immediate interruption of sleep
		select {
		case <-cancelCh:
			Logger.Printf("Connection recovery aborted: connection closed during backoff.")
			return ErrClosed
		case <-time.After(c.RetryInterval() + jitter):
		}

		// Re-dial
		var conn net.Conn
		dialer := c.Config.Dial
		if dialer == nil {
			dialer = DefaultDial(defaultConnectionTimeout)
		}

		// We need to parse URL to get addr
		var uri URI
		uri, err = ParseURI(c.url)
		if err != nil {
			Logger.Printf("Connection recovery failed to parse URI: %v", err)
			return err
		}

		// Reset SASL to recover from zeroed out credentials
		c.Config.SASL = nil
		if err = c.Config.setSASL(uri); err != nil {
			Logger.Printf("Connection recovery failed to set SASL: %v", err)
			return err
		}

		addr := net.JoinHostPort(uri.Host, strconv.FormatInt(int64(uri.Port), 10))

		conn, err = dialer("tcp", addr)
		if err != nil {
			Logger.Printf("Connection recovery dial error: %v", err)
			continue
		}

		// TLS handshake if needed
		if uri.Scheme == "amqps" {
			client := tls.Client(conn, c.Config.TLSClientConfig)
			if err = client.Handshake(); err != nil {
				Logger.Printf("Connection recovery TLS handshake error: %v", err)
				conn.Close()
				continue
			}
			conn = client
		}

		// Reset state and swap connection under locks to ensure atomicity
		// and prevent data races with concurrent readers/writers.
		// Acquiring destructorM -> closeM -> m avoids any lock ordering deadlocks.
		c.destructorM.Lock()
		c.closeM.Lock()
		c.m.Lock()

		// Swap the connection
		c.conn = conn
		c.writer = &writer{bufio.NewWriter(conn)}

		c.resetState()

		c.m.Unlock()
		c.closeM.Unlock()
		c.destructorM.Unlock()

		go c.reader(conn)

		if err = c.open(c.Config); err != nil {
			Logger.Printf("Connection recovery open error: %v", err)
			conn.Close()
			continue
		}

		// Reconnect channels
		c.m.Lock()
		channels := make([]*Channel, 0, len(c.channels))
		for _, ch := range c.channels {
			channels = append(channels, ch)
		}
		c.m.Unlock()

		// Phase 1: Reconnect and open all channel sessions, apply QoS and Confirms
		for _, ch := range channels {
			if err = ch.reconnectChannel(); err != nil {
				Logger.Printf("Connection recovery failed to reconnect channel %d: %v", ch.id, err)
				conn.Close()
				break
			}
		}

		var skippedTopologyEntities []TopologyRecoveryEntity
		if err == nil && c.IsTopologyRecoveryEnabled() {
			// Phase 2: Recover topology across all channels via the configured implementation
			skippedTopologyEntities, err = c.Config.Recovery.TopologyRecovery.RecoverTopology(c, channels)
			if err != nil {
				Logger.Printf("Connection recovery failed to recover topology: %v", err)
				conn.Close()
			}
		}

		// This error check is for retrying when the channels are not able to reconnect
		if err != nil {
			continue
		}

		if len(skippedTopologyEntities) > 0 {
			Logger.Printf("Connection recovery succeeded with %d skipped topology entities", len(skippedTopologyEntities))
		} else {
			Logger.Printf("Connection recovery successful")
		}
		c.lifeCycle.setStateOpen(skippedTopologyEntities)
		return nil
	}

	Logger.Printf("Connection recovery exhausted all %d retries", c.MaxRetryCount())
	if c.conn != nil {
		c.conn.Close()
	}
	c.closed.Store(true)

	return err
}

// resetState clears the shutdown and close flags and re-initializes the internal
// channels so the connection can be reused after a successful reconnection.
// The caller must hold c.destructorM, c.closeM and c.m.
func (c *Connection) resetState() {
	c.closed.Store(false)
	c.destructed = false
	c.closeInit = false

	c.errors = make(chan *Error, 1)
	c.close = make(chan struct{})

	// Re-create the rpc channel so we don't read stale messages from the previous connection
	c.rpc = make(chan message)
}

// IsRecoveryEnabled checks if the recovery is enabled.
func (c *Connection) IsRecoveryEnabled() bool {
	if c == nil {
		return false
	}
	c.closeM.Lock()
	closedOrClosing := c.closeInit
	c.closeM.Unlock()
	if closedOrClosing {
		return false
	}
	return c.Config.Recovery != nil &&
		c.Config.Recovery.ReconnectionConfig != nil &&
		c.Config.Recovery.ReconnectionConfig.MaxRetryCount > 0
}

// IsTopologyRecoveryEnabled checks if the topology recovery is enabled.
//
// Topology recovery is disabled when recovery is off, when no TopologyRecovery
// implementation is configured, or when the configured TopologyRecoveryMode is
// TopologyRecoveryDisabled. This single check gates both entity recovery and
// consumer re-subscription.
func (c *Connection) IsTopologyRecoveryEnabled() bool {
	if c == nil {
		return false
	}
	return c.IsRecoveryEnabled() &&
		c.Config.Recovery.TopologyRecovery != nil &&
		c.Config.Recovery.TopologyRecoveryMode != TopologyRecoveryDisabled
}

// topologyRecoveryMode returns the configured topology recovery mode, defaulting
// to TopologyRecoveryAllEnabled when recovery is not configured.
func (c *Connection) topologyRecoveryMode() TopologyRecoveryMode {
	if c == nil || c.Config.Recovery == nil {
		return TopologyRecoveryAllEnabled
	}
	return c.Config.Recovery.TopologyRecoveryMode
}

// IsConnectionRecoveryEnabled checks if the connection recovery is enabled.
func (c *Connection) IsConnectionRecoveryEnabled() bool {
	if c == nil {
		return false
	}
	return c.IsRecoveryEnabled() && c.Config.Recovery.ConnectionRecovery != nil
}

// MaxRetryCount returns the maximum number of retries if recovery is enabled, otherwise returns 0.
func (c *Connection) MaxRetryCount() int {
	if c.IsRecoveryEnabled() {
		return c.Config.Recovery.ReconnectionConfig.MaxRetryCount
	}
	return 0
}

// RetryInterval returns the interval between retries if recovery is enabled, otherwise returns 0.
func (c *Connection) RetryInterval() time.Duration {
	if c.IsRecoveryEnabled() {
		return c.Config.Recovery.ReconnectionConfig.RetryInterval
	}
	return 0
}

// channelTopology returns the TopologyConfiguration for the given channel,
// creating one if it does not exist. Caller must hold c.topologyM.
func (c *Connection) channelTopology(channelID uint16) *TopologyConfiguration {
	if c.topologyConfiguration == nil {
		c.topologyConfiguration = make(map[uint16]*TopologyConfiguration)
	}
	if _, ok := c.topologyConfiguration[channelID]; !ok {
		c.topologyConfiguration[channelID] = newTopologyConfiguration()
	}
	return c.topologyConfiguration[channelID]
}

func (c *Connection) recordExchange(channelID uint16, ec ExchangeConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()
	c.channelTopology(channelID).Exchanges[ec.Name] = ec
}

// removeExchange removes an exchange and its bindings from the topology store.
// It returns the exchange names that were sources in exchange-to-exchange bindings
// pointing TO the deleted exchange — these are candidates for auto-delete cascade.
func (c *Connection) removeExchange(name string) []string {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration == nil {
		return nil
	}
	var cascadeSources []string
	for _, config := range c.topologyConfiguration {
		delete(config.Exchanges, name)

		// Clean up related queue bindings in-place (0 allocations)
		if config.Bindings != nil {
			oldBindings := config.Bindings
			active := config.Bindings[:0]
			for _, b := range oldBindings {
				if b.Exchange != name {
					active = append(active, b)
				}
			}
			for i := len(active); i < len(oldBindings); i++ {
				oldBindings[i] = BindingConfig{} // zero-value tail slots to avoid memory leak
			}
			config.Bindings = active
		}
		if config.ExchangeBindings != nil {
			oldExchangeBindings := config.ExchangeBindings
			active := config.ExchangeBindings[:0]
			for _, eb := range oldExchangeBindings {
				if eb.Destination != name && eb.Source != name {
					active = append(active, eb)
				} else if eb.Destination == name {
					// This exchange was the destination; its source may now be auto-deletable.
					cascadeSources = append(cascadeSources, eb.Source)
				}
			}
			for i := len(active); i < len(oldExchangeBindings); i++ {
				oldExchangeBindings[i] = ExchangeBindingConfig{} // zero-value tail slots to avoid memory leak
			}
			config.ExchangeBindings = active
		}
	}
	return cascadeSources
}

// deleteRecordedExchange removes an exchange from tracking and cascades to any
// auto-delete exchange that sourced an exchange-to-exchange binding pointing to it.
func (c *Connection) deleteRecordedExchange(name string) {
	for _, src := range c.removeExchange(name) {
		c.maybeDeleteRecordedAutoDeleteExchange(src)
	}
}

// maybeDeleteRecordedAutoDeleteExchange forgets an exchange from the topology store
// when it is auto-delete and all bindings sourced from it have been removed.
func (c *Connection) maybeDeleteRecordedAutoDeleteExchange(exchangeName string) {
	if !c.IsTopologyRecoveryEnabled() || exchangeName == "" {
		return
	}

	c.topologyM.Lock()
	isAutoDelete := false
	hasBindings := false
	if c.topologyConfiguration != nil {
		for _, cfg := range c.topologyConfiguration {
			if ec, ok := cfg.Exchanges[exchangeName]; ok && ec.AutoDelete {
				isAutoDelete = true
			}
			for _, b := range cfg.Bindings {
				if b.Exchange == exchangeName {
					hasBindings = true
					break
				}
			}
			if !hasBindings {
				for _, eb := range cfg.ExchangeBindings {
					if eb.Source == exchangeName {
						hasBindings = true
						break
					}
				}
			}
			if hasBindings {
				break
			}
		}
	}
	c.topologyM.Unlock()

	if !isAutoDelete || hasBindings {
		return
	}

	c.deleteRecordedExchange(exchangeName)
}

func (c *Connection) recordQueue(channelID uint16, qc QueueConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()
	c.channelTopology(channelID).Queues[qc.ActualName] = qc
}

// removeQueue removes a queue and its bindings from the topology store.
// It returns the exchange names that sourced bindings pointing to this queue —
// these are candidates for auto-delete cascade.
func (c *Connection) removeQueue(name string) []string {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration == nil {
		return nil
	}
	var sources []string
	for _, config := range c.topologyConfiguration {
		delete(config.Queues, name)

		// Clean up related bindings in-place (0 allocations)
		if config.Bindings != nil {
			oldBindings := config.Bindings
			active := config.Bindings[:0]
			for _, b := range oldBindings {
				if b.Queue != name {
					active = append(active, b)
				} else {
					sources = append(sources, b.Exchange)
				}
			}
			for i := len(active); i < len(oldBindings); i++ {
				oldBindings[i] = BindingConfig{} // zero-value tail slots to avoid memory leak
			}
			config.Bindings = active
		}
	}
	return sources
}

// deleteRecordedQueue removes a queue from tracking and cascades to any auto-delete
// exchange that sourced a binding pointing to it.
func (c *Connection) deleteRecordedQueue(name string) {
	for _, src := range c.removeQueue(name) {
		c.maybeDeleteRecordedAutoDeleteExchange(src)
	}
}

// maybeDeleteRecordedAutoDeleteQueue forgets a queue from the topology store when
// it is auto-delete and all its consumers on this connection have been cancelled.
func (c *Connection) maybeDeleteRecordedAutoDeleteQueue(queueName string) {
	if !c.IsTopologyRecoveryEnabled() {
		return
	}

	// Fast path: bail early if not tracked as auto-delete.
	c.topologyM.Lock()
	isAutoDelete := false
	if c.topologyConfiguration != nil {
		for _, cfg := range c.topologyConfiguration {
			if qc, ok := cfg.Queues[queueName]; ok && qc.AutoDelete {
				isAutoDelete = true
				break
			}
		}
	}
	c.topologyM.Unlock()
	if !isAutoDelete {
		return
	}

	// Check whether any channel still has a consumer on this queue.
	c.m.Lock()
	channels := make([]*Channel, 0, len(c.channels))
	for _, ch := range c.channels {
		channels = append(channels, ch)
	}
	c.m.Unlock()

	for _, ch := range channels {
		if ch.consumers.hasConsumerForQueue(queueName) {
			return
		}
	}

	// No consumers remain; remove the queue and cascade.
	c.deleteRecordedQueue(queueName)
}

func (c *Connection) recordBinding(channelID uint16, bc BindingConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	config := c.channelTopology(channelID)

	for _, b := range config.Bindings {
		if b.Queue == bc.Queue && b.Exchange == bc.Exchange && b.Key == bc.Key && reflect.DeepEqual(b.Args, bc.Args) {
			return
		}
	}

	config.Bindings = append(config.Bindings, bc)
}

func (c *Connection) removeBinding(bc BindingConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration == nil {
		return
	}

	for _, config := range c.topologyConfiguration {
		oldBindings := config.Bindings
		active := config.Bindings[:0]
		for _, b := range oldBindings {
			if b.Queue != bc.Queue || b.Key != bc.Key || b.Exchange != bc.Exchange || !reflect.DeepEqual(b.Args, bc.Args) {
				active = append(active, b)
			}
		}
		for i := len(active); i < len(oldBindings); i++ {
			oldBindings[i] = BindingConfig{} // zero-value tail slots to avoid memory leak
		}
		config.Bindings = active
	}
}

func (c *Connection) recordExchangeBinding(channelID uint16, ebc ExchangeBindingConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	config := c.channelTopology(channelID)

	for _, eb := range config.ExchangeBindings {
		if eb.Source == ebc.Source && eb.Destination == ebc.Destination && eb.Key == ebc.Key && reflect.DeepEqual(eb.Args, ebc.Args) {
			return
		}
	}

	config.ExchangeBindings = append(config.ExchangeBindings, ebc)
}

func (c *Connection) removeExchangeBinding(ebc ExchangeBindingConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration == nil {
		return
	}

	for _, config := range c.topologyConfiguration {
		oldExchangeBindings := config.ExchangeBindings
		active := config.ExchangeBindings[:0]
		for _, eb := range oldExchangeBindings {
			if eb.Destination != ebc.Destination || eb.Key != ebc.Key || eb.Source != ebc.Source || !reflect.DeepEqual(eb.Args, ebc.Args) {
				active = append(active, eb)
			}
		}
		for i := len(active); i < len(oldExchangeBindings); i++ {
			oldExchangeBindings[i] = ExchangeBindingConfig{} // zero-value tail slots to avoid memory leak
		}
		config.ExchangeBindings = active
	}
}

func (c *Connection) recordQos(channelID uint16, qos QosConfig) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()
	c.channelTopology(channelID).Qos = &qos
}

func (c *Connection) removeChannelTopology(channelID uint16) {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration != nil {
		delete(c.topologyConfiguration, channelID)
	}
}

// getTopologyConfiguration returns a snapshot of topology for the given channel.
// When global is false, only entities tracked on channelID are returned (channel-local view).
// When global is true, entities from all channels are merged into a single connection-level
// view — because AMQP topology (exchanges, queues, bindings) is scoped to the TCP connection,
// not individual channels, the merged view is what must be recovered after a reconnect.
// QoS is always channel-scoped and reflects only channelID's configuration in both modes.
func (c *Connection) getTopologyConfiguration(channelID uint16, global bool) TopologyConfiguration {
	c.topologyM.Lock()
	defer c.topologyM.Unlock()

	if c.topologyConfiguration == nil {
		return *newTopologyConfiguration()
	}

	if !global {
		config, ok := c.topologyConfiguration[channelID]
		if !ok {
			return *newTopologyConfiguration()
		}
		return *config.Clone() // Clone already deep-copies Qos for this channel.
	}

	// QoS is per-channel; pin channelID's QoS onto the merged result.
	var qos *QosConfig
	if config, ok := c.topologyConfiguration[channelID]; ok && config.Qos != nil {
		qos = &QosConfig{
			PrefetchCount: config.Qos.PrefetchCount,
			PrefetchSize:  config.Qos.PrefetchSize,
			Global:        config.Qos.Global,
		}
	}

	// Merge entities from ALL channels into a single connection-level view.
	mergedExchanges := make(map[string]ExchangeConfig)
	mergedQueues := make(map[string]QueueConfig)

	// Bindings are stored as slices, so we need explicit dedup during merge.
	// The identity is (Queue, Exchange, Key, Args) — the same tuple used by
	// recordBinding and removeBinding. Args must participate: bindings that
	// differ only by arguments (e.g. headers-exchange bindings) are distinct and
	// must both survive the merge. Args is a Table (a map, not comparable), so a
	// map key cannot be used; the binding lists are small, so a linear scan that
	// mirrors recordBinding's dedup is used instead.
	var mergedBindings []BindingConfig
	var mergedExchangeBindings []ExchangeBindingConfig

	for _, config := range c.topologyConfiguration {
		for name, ec := range config.Exchanges {
			mergedExchanges[name] = ec
		}
		for name, qc := range config.Queues {
			mergedQueues[name] = qc
		}
		for _, b := range config.Bindings {
			seen := false
			for _, existing := range mergedBindings {
				if existing.Queue == b.Queue && existing.Exchange == b.Exchange && existing.Key == b.Key && reflect.DeepEqual(existing.Args, b.Args) {
					seen = true
					break
				}
			}
			if !seen {
				mergedBindings = append(mergedBindings, b)
			}
		}
		for _, eb := range config.ExchangeBindings {
			seen := false
			for _, existing := range mergedExchangeBindings {
				if existing.Source == eb.Source && existing.Destination == eb.Destination && existing.Key == eb.Key && reflect.DeepEqual(existing.Args, eb.Args) {
					seen = true
					break
				}
			}
			if !seen {
				mergedExchangeBindings = append(mergedExchangeBindings, eb)
			}
		}
	}

	return TopologyConfiguration{
		Qos:              qos,
		Exchanges:        mergedExchanges,
		Queues:           mergedQueues,
		Bindings:         mergedBindings,
		ExchangeBindings: mergedExchangeBindings,
	}
}

// filterTransientTopology returns only the connection-scoped (transient) subset of
// the given topology, used by TopologyRecoveryOnlyTransient.
//
// An exchange is transient when it is auto-delete; a queue is transient when it is
// exclusive and/or auto-delete (server-named queues are exclusive+auto-delete and
// are therefore included). A queue-to-exchange binding is recovered when either its
// queue or its exchange is transient: re-declaring a transient queue drops all of
// its bindings, including those to durable exchanges, so such bindings must be
// recreated. An exchange-to-exchange binding is recovered when its source or
// destination exchange is transient. Durable, non-auto-delete entities (and bindings
// purely between them) are dropped because the broker retains them across a network
// interruption.
func filterTransientTopology(
	config *TopologyConfiguration,
	globalTransientQueues map[string]bool,
	globalTransientExchanges map[string]bool,
) {
	filteredExchanges := make(map[string]ExchangeConfig)
	for name, ec := range config.Exchanges {
		if ec.AutoDelete {
			filteredExchanges[name] = ec
		}
	}

	filteredQueues := make(map[string]QueueConfig)
	for name, qc := range config.Queues {
		if qc.Exclusive || qc.AutoDelete {
			filteredQueues[name] = qc
		}
	}

	filteredBindings := make([]BindingConfig, 0, len(config.Bindings))
	for _, b := range config.Bindings {
		if globalTransientQueues[b.Queue] || globalTransientExchanges[b.Exchange] {
			filteredBindings = append(filteredBindings, b)
		}
	}

	filteredExchangeBindings := make([]ExchangeBindingConfig, 0, len(config.ExchangeBindings))
	for _, eb := range config.ExchangeBindings {
		if globalTransientExchanges[eb.Source] || globalTransientExchanges[eb.Destination] {
			filteredExchangeBindings = append(filteredExchangeBindings, eb)
		}
	}

	config.Exchanges = filteredExchanges
	config.Queues = filteredQueues
	config.Bindings = filteredBindings
	config.ExchangeBindings = filteredExchangeBindings
}

// recoverConnectionTopology re-establishes all tracked AMQP entities on the
// reconnected channels in dependency order: exchanges → queues → bindings →
// exchange-to-exchange bindings → consumers.
//
// The topology snapshot is cloned under topologyM before any network I/O so the
// lock is not held during broker round-trips. When TopologyRecoveryOnlyTransient
// is active the snapshot is filtered down to transient entities (exclusive /
// auto-delete queues, auto-delete exchanges, and any bindings that reference
// them) before the recovery loop begins; the global transient sets are built by
// scanning all channels so that cross-channel bindings are retained correctly.
//
// Server-generated queue names (DeclaredName == "") may differ on each
// reconnect. When the broker assigns a new name the old name is removed from the
// persistent topology store, the new name is inserted, and a nameReplacements
// map is populated so that any bindings and consumer configs that still reference
// the old name are patched in a single pass before bindings are re-created.
//
// Consumer re-subscription sends basic.consume directly via ch.call rather than
// calling ch.Consume, which intentionally reuses the existing delivery-channel
// pipeline (buffer goroutine + outer chan Delivery) that was wired up before the
// connection dropped. This means the application's delivery channel remains valid
// across reconnects without any intervention from the caller.
func (c *Connection) recoverConnectionTopology(channels []*Channel) ([]TopologyRecoveryEntity, error) {
	if !c.IsTopologyRecoveryEnabled() {
		return nil, nil
	}

	// Clone the topology snapshot under lock so network I/O below does not
	// hold topologyM and deadlock with concurrent record* / remove* calls.
	c.topologyM.Lock()
	topologyMap := make(map[uint16]*TopologyConfiguration, len(c.topologyConfiguration))
	for chID, config := range c.topologyConfiguration {
		topologyMap[chID] = config.Clone()
	}
	c.topologyM.Unlock()

	// Build a channel-ID → Channel map for O(1) lookup during the recovery loops.
	// Mark each channel as owned by this pass so watchChannel's
	// NotifyClose listener (registered once, for the channel's lifetime) won't
	// start a redundant, competing Reconnect+RecoverTopology pass if a broker
	// soft error closes the channel while this pass is already
	// reopening/redeclaring it below.
	channelMap := make(map[uint16]*Channel, len(channels))
	for _, ch := range channels {
		channelMap[ch.id] = ch
		ch.recoveringTopology.Store(true)
		defer ch.recoveringTopology.Store(false)
	}

	// When only transient topology should be recovered, filter the snapshot down
	// to auto-delete exchanges and exclusive/auto-delete queues (and their
	// bindings). The global transient sets are built by scanning every channel
	// first so that a binding on ch2 referencing a transient queue declared on
	// ch1 is correctly retained even though that queue does not appear in ch2's
	// own topology entry.
	if c.topologyRecoveryMode() == TopologyRecoveryOnlyTransient {
		globalTransientQueues := make(map[string]bool)
		globalTransientExchanges := make(map[string]bool)
		for _, config := range topologyMap {
			for name, qc := range config.Queues {
				if qc.Exclusive || qc.AutoDelete {
					globalTransientQueues[name] = true
				}
			}
			for name, ec := range config.Exchanges {
				if ec.AutoDelete {
					globalTransientExchanges[name] = true
				}
			}
		}
		// Filter transient topology for all channels.
		for _, config := range topologyMap {
			filterTransientTopology(config, globalTransientQueues, globalTransientExchanges)
		}
	}

	var skipped []TopologyRecoveryEntity
	cb := c.Config.Recovery.OnTopologyEntityError

	// skipOrAbort calls cb with the connection and entity error. If cb returns true
	// (or is nil, matching DefaultOnTopologyEntityError behaviour) the entity is
	// appended to skipped and the caller continues the loop. If cb returns false
	// the entity error is returned as a fatal error, aborting topology recovery.
	skipOrAbort := func(e TopologyRecoveryEntity) (skip bool, fatal error) {
		if cb == nil || cb(c, e) {
			skipped = append(skipped, e)
			return true, nil
		}
		return false, fmt.Errorf("%w", e)
	}

	// 1. Recover exchanges across all channels.
	for chID, config := range topologyMap {
		ch, ok := channelMap[chID]
		if !ok {
			continue
		}
		for _, ec := range config.Exchanges {
			if err := ch.ExchangeDeclare(ec.Name, ec.Kind, ec.Durable, ec.AutoDelete, ec.Internal, ec.NoWait, ec.Args); err != nil {
				Logger.Printf("failed to recover exchange %s on channel %d: %v", ec.Name, chID, err)
				e := TopologyRecoveryEntity{EntityType: TopologyEntityExchange, EntityName: ec.Name, ChannelID: chID, Err: err}
				if cont, fatal := skipOrAbort(e); !cont {
					return skipped, fatal
				}
				ch.reopenIfClosed()
			}
		}
	}

	// 2. Recover queues across all channels.
	//
	// nameReplacements collects old→new name mappings for server-generated queues
	// whose broker-assigned name changed on this reconnect; these are applied to
	// bindings and consumer configs after all queues have been declared.
	//
	// declaredQueues deduplicates named queues: the same queue may appear in
	// multiple channels' topology when a binding or consumer on ch2 references a
	// queue that was declared on ch1 — declaring it twice would be a no-op on
	// the broker but wastes a round-trip.
	nameReplacements := make(map[string]string)
	declaredQueues := make(map[string]bool)

	for chID, config := range topologyMap {
		ch, ok := channelMap[chID]
		if !ok {
			continue
		}
		for _, qc := range config.Queues {
			if qc.DeclaredName != "" {
				if declaredQueues[qc.ActualName] {
					continue
				}
				declaredQueues[qc.ActualName] = true
			}

			q, err := ch.QueueDeclare(qc.DeclaredName, qc.Durable, qc.AutoDelete, qc.Exclusive, qc.NoWait, qc.Args)
			if err != nil {
				Logger.Printf("failed to recover queue %s on channel %d: %v", qc.ActualName, chID, err)
				e := TopologyRecoveryEntity{EntityType: TopologyEntityQueue, EntityName: qc.ActualName, ChannelID: chID, Err: err}
				if cont, fatal := skipOrAbort(e); !cont {
					return skipped, fatal
				}
				ch.reopenIfClosed()
				continue
			}

			// Server-generated queues (DeclaredName == "") receive a fresh broker-assigned
			// name on every reconnect. Record the rename so bindings and consumers can be
			// updated before they are re-created/re-subscribed below.
			if qc.DeclaredName == "" && q.Name != qc.ActualName {
				nameReplacements[qc.ActualName] = q.Name

				// Patch the persistent topology store so future recoveries use the new name.
				c.topologyM.Lock()
				if c.topologyConfiguration != nil {
					if connConfig, found := c.topologyConfiguration[chID]; found && connConfig.Queues != nil {
						delete(connConfig.Queues, qc.ActualName)
						newQc := qc
						newQc.ActualName = q.Name
						connConfig.Queues[q.Name] = newQc
					}
				}
				c.topologyM.Unlock()
			}
		}
	}

	// Propagate server-generated name changes to bindings and consumer configs
	// before re-creating bindings (step 3) so they reference the live queue name.
	if len(nameReplacements) > 0 {
		// Patch the persistent binding records so subsequent recoveries are consistent.
		c.topologyM.Lock()
		if c.topologyConfiguration != nil {
			for _, connConfig := range c.topologyConfiguration {
				for i := range connConfig.Bindings {
					if newName, found := nameReplacements[connConfig.Bindings[i].Queue]; found {
						connConfig.Bindings[i].Queue = newName
					}
				}
			}
		}
		c.topologyM.Unlock()

		// Patch the local snapshot used for the binding recovery loop below.
		for _, config := range topologyMap {
			for i := range config.Bindings {
				if newName, found := nameReplacements[config.Bindings[i].Queue]; found {
					config.Bindings[i].Queue = newName
				}
			}
		}

		// Patch consumer configs so re-subscription (step 5) targets the new queue name.
		for _, ch := range channels {
			ch.consumers.Lock()
			for tag, config := range ch.consumers.configs {
				if newName, found := nameReplacements[config.Queue]; found {
					config.Queue = newName
					ch.consumers.configs[tag] = config
				}
			}
			ch.consumers.Unlock()
		}
	}

	// 3. Recover queue-to-exchange bindings across all channels.
	for chID, config := range topologyMap {
		ch, ok := channelMap[chID]
		if !ok {
			continue
		}
		for _, b := range config.Bindings {
			if err := ch.QueueBind(b.Queue, b.Key, b.Exchange, b.NoWait, b.Args); err != nil {
				Logger.Printf("failed to recover binding of queue %s to exchange %s on channel %d: %v", b.Queue, b.Exchange, chID, err)
				e := TopologyRecoveryEntity{EntityType: TopologyEntityQueueBinding, EntityName: b.Queue, SecondaryName: b.Exchange, RoutingKey: b.Key, ChannelID: chID, Err: err}
				if cont, fatal := skipOrAbort(e); !cont {
					return skipped, fatal
				}
				ch.reopenIfClosed()
			}
		}
	}

	// 4. Recover exchange-to-exchange bindings across all channels.
	for chID, config := range topologyMap {
		ch, ok := channelMap[chID]
		if !ok {
			continue
		}
		for _, eb := range config.ExchangeBindings {
			if err := ch.ExchangeBind(eb.Destination, eb.Key, eb.Source, eb.NoWait, eb.Args); err != nil {
				Logger.Printf("failed to recover exchange binding from %s to %s on channel %d: %v", eb.Source, eb.Destination, chID, err)
				e := TopologyRecoveryEntity{EntityType: TopologyEntityExchangeBinding, EntityName: eb.Source, SecondaryName: eb.Destination, RoutingKey: eb.Key, ChannelID: chID, Err: err}
				if cont, fatal := skipOrAbort(e); !cont {
					return skipped, fatal
				}
				ch.reopenIfClosed()
			}
		}
	}

	// 5. Re-subscribe consumers across all channels.
	//
	// basic.consume is sent directly via ch.call rather than ch.Consume so that
	// the existing buffer goroutine and outer delivery channel (chan Delivery)
	// that the caller holds are reused. No new goroutines are spawned and the
	// application's delivery channel stays valid without any caller intervention.
	for _, ch := range channels {
		// Snapshot consumer configs under lock to avoid holding the lock
		// during the broker round-trips below.
		ch.consumers.Lock()
		configs := make(map[string]consumerConfig, len(ch.consumers.configs))
		for tag, config := range ch.consumers.configs {
			configs[tag] = config
		}
		ch.consumers.Unlock()

		for tag, config := range configs {
			req := &basicConsume{
				Queue:       config.Queue,
				ConsumerTag: tag,
				NoLocal:     config.NoLocal,
				NoAck:       config.AutoAck,
				Exclusive:   config.Exclusive,
				NoWait:      config.NoWait,
				Arguments:   config.Args,
			}
			res := &basicConsumeOk{}
			if err := ch.call(req, res); err != nil {
				Logger.Printf("failed to recover consumer for tag %s on queue %s on channel %d: %v", tag, config.Queue, ch.id, err)
				e := TopologyRecoveryEntity{EntityType: TopologyEntityConsumer, EntityName: tag, ChannelID: ch.id, Err: err}
				if cont, fatal := skipOrAbort(e); !cont {
					return skipped, fatal
				}
				ch.reopenIfClosed()
			}
		}
	}

	return skipped, nil
}
