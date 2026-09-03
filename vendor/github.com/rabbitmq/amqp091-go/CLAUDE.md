# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## About

`amqp091-go` is the official Go AMQP 0.9.1 client maintained by the RabbitMQ core team (`github.com/rabbitmq/amqp091-go`). It is a single root package with no external runtime dependencies (only `go.uber.org/goleak` for tests).

## Commands

```bash
# Format
make fmt
make check-fmt       # check-only, no writes

# Lint
make checks          # golangci-lint (must be installed)

# Integration tests — require a running RabbitMQ on localhost:5672
make tests
make tests-docker    # spins up RabbitMQ in Docker, runs, then tears down

# Run a specific test
go test -race -v -tags integration -run TestIntegrationOpenClose

# Start / stop Dockerized RabbitMQ manually
make rabbitmq-server
make stop-rabbitmq-server
```

Integration tests use the `integration` build tag. Without it (or without a running broker), only unit tests run. The env var `RABBITMQ_RABBITMQCTL_PATH=DOCKER:<container>` (or path to a local `rabbitmqctl` executable) enables administrative/broker control tests.

## Architecture

All core code is in the root package (excluding examples under `_examples/` and generator files under `spec/`).

### Layers

```
Caller
  └─ Connection (connection.go)     TCP socket, AMQP handshake, heartbeat, frame mux
       ├─ read.go / write.go        frame (de)serialization
       ├─ recovery.go               connection/channel recovery (reconnection)
       └─ Channel (channel.go)      AMQP channel — all protocol methods
            ├─ confirms.go          publisher confirm tracking
            └─ consumers.go         consumer tag → delivery channel dispatch
```

`spec091.go` is auto-generated from the AMQP 0.9.1 spec XML. Do not hand-edit it.

### Connection

- `Dial` / `DialConfig` / `DialTLS` are the entry points; `DialConfig` is the most general.
- One **reader goroutine** (`connection.reader`) reads frames from the socket and calls `demux` to route them to channels.
- One **heartbeater goroutine** (`connection.heartbeater`) monitors activity and sends keep-alive frames.
- Channels are tracked in `Connection.channels map[uint16]*Channel`. Channel 0 is reserved for connection-level control frames.

### Channel

- Obtained via `conn.Channel()`.
- All AMQP operations (declare, bind, publish, consume, ack, transactions) are methods on `Channel`.
- RPC-style operations call `call()`, which sends a method frame and blocks on the reply. Non-RPC sends are fire-and-forget.
- Concurrent publishes from multiple goroutines are safe; the write side is mutex-protected via `Connection.sendM`.

### Frame assembly state machine

`Channel.recv` is a function pointer that acts as a state machine:

```
recvMethod  →  (method with content)  →  recvHeader  →  recvContent  →  recvMethod
           →  (method without content: dispatch immediately, stay in recvMethod)
```

Body can span multiple `frameBody` frames; `Channel` accumulates them before dispatch.

### Publisher confirms (`confirms.go`)

`Channel.Confirm(noWait)` enables confirm mode. Each subsequent publish is assigned a monotonically increasing delivery tag. The broker acknowledges with `basic.ack` / `basic.nack` frames, which may arrive out of order. `confirms.resequence()` buffers out-of-order acks and delivers them in order to all listeners. `DeferredConfirmation` provides a future-style API (`Wait`, `WaitContext`, `Acked`).

### Consumer dispatch (`consumers.go`)

`Channel.Consume` registers a consumer tag and launches a **buffer goroutine** per consumer that relays deliveries from an internal `chan *Delivery` to the application-facing `chan Delivery`. This decouples the reader goroutine from application consumption speed. Buffer goroutines nil out slice elements explicitly to aid GC under high load.

### Notify channels

All `Notify*` methods (`NotifyClose`, `NotifyBlocked`, `NotifyFlow`, `NotifyReturn`, `NotifyCancel`, `NotifyConfirm`, `NotifyPublish`, `NotifyStateChange`, `NotifyRecoveryCancel`) follow the same contract:

- The caller provides a channel (buffered recommended).
- The library writes to it (or closes it for signal-only notifications) and **closes it** when the entity shuts down, is closed, or the event occurs.
- Multiple registrations result in a broadcast — all listeners receive every event or signal.
- Reading from a closed listener channel signals shutdown or cancellation.

### Connection Recovery / Reconnection

The library supports automatic connection and channel recovery (reconnection) when a network failure occurs.

- **Enabling Recovery**: Automatic recovery is enabled by providing a non-nil `Recovery` configuration in `Config` when calling `DialConfig`. If `Recovery` is nil (the default), automatic recovery is disabled.
- **Configuration**: `Config.Recovery` contains `ReconnectionConfig` (defines `MaxRetryCount`, `RetryInterval`, and `RecoverableErrorCodes`) and `ConnectionRecovery` (an interface for the recovery implementation). If these are nil but `Recovery` is non-nil, `DefaultReconnectionConfig` and `DefaultConnectionRecovery` are used.
- **State Monitoring**: Applications can monitor recovery state transitions (e.g., from `StateActive` to `StateReconnecting` to `StateClosed`) by registering with `Connection.NotifyStateChange` or `Channel.NotifyStateChange`.
- **Cancellation**: Recovery can be canceled or aborted (e.g., when `Close()` is called during active reconnection). Applications can listen to this via `Connection.NotifyRecoveryCancel` or `Channel.NotifyRecoveryCancel`.
- **Examples**: `_examples/recovery/recovery.go` demonstrates the automatic recovery pattern, while `_examples/client/client.go` demonstrates a manual reconnecting wrapper pattern.

## Key conventions

- `*Error` (`types.go`) carries an AMQP reply code and whether the error is recoverable. Server-initiated closes arrive on `NotifyClose` channels as `*Error`.
- `Table` is `map[string]interface{}` with a restricted set of allowed value types enumerated in `types.go`.
- Mutexes follow a strict order: `Connection.m` → `Channel.m` (never the reverse) to avoid deadlock.
- `atomic.Bool` flags (`Connection.closed`, `Channel.closed`) allow lock-free early-exit checks on the hot path.
