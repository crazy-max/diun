[![Build Status](https://api.travis-ci.org/streadway/amqp.svg)](http://travis-ci.org/streadway/amqp) [![GoDoc](https://godoc.org/github.com/streadway/amqp?status.svg)](http://godoc.org/github.com/streadway/amqp)

# Go RabbitMQ Client Library (Unmaintained Fork)

## Beware of Abandonware

This repository is **NOT ACTIVELY MAINTAINED**. Consider using
a different fork instead: [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go).
In case of questions, start a discussion in that repo or [use other RabbitMQ community resources](https://rabbitmq.com/contact.html).



## Project Maturity

This project has been used in production systems for many years. As of 2022,
this repository is **NOT ACTIVELY MAINTAINED**.

This repository is **very strict** about any potential public API changes.
You may want to consider [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) which
is more willing to adapt the API.


## Supported Go Versions

This library supports two most recent Go release series, currently 1.10 and 1.11.


## Supported RabbitMQ Versions

This project supports RabbitMQ versions starting with `2.0` but primarily tested
against reasonably recent `3.x` releases. Some features and behaviours may be
server version-specific.

## Goals

Provide a functional interface that closely represents the AMQP 0.9.1 model
targeted to RabbitMQ as a server.  This includes the minimum necessary to
interact the semantics of the protocol.

## Non-goals

Things not intended to be supported.

  * Auto reconnect and re-synchronization of client and server topologies.
    * Reconnection would require understanding the error paths when the
      topology cannot be declared on reconnect.  This would require a new set
      of types and code paths that are best suited at the call-site of this
      package.  AMQP has a dynamic topology that needs all peers to agree. If
      this doesn't happen, the behavior is undefined.  Instead of producing a
      possible interface with undefined behavior, this package is designed to
      be simple for the caller to implement the necessary connection-time
      topology declaration so that reconnection is trivial and encapsulated in
      the caller's application code.
  * AMQP Protocol negotiation for forward or backward compatibility.
    * 0.9.1 is stable and widely deployed.  Versions 0.10 and 1.0 are divergent
      specifications that change the semantics and wire format of the protocol.
      We will accept patches for other protocol support but have no plans for
      implementation ourselves.
  * Anything other than PLAIN and EXTERNAL authentication mechanisms.
    * Keeping the mechanisms interface modular makes it possible to extend
      outside of this package.  If other mechanisms prove to be popular, then
      we would accept patches to include them in this package.

## Usage

See the 'examples' subdirectory for simple producers and consumers executables.
If you have a use-case in mind which isn't well-represented by the examples,
please file an issue.

## Documentation

Use [Godoc documentation](http://godoc.org/github.com/streadway/amqp) for
reference and usage.

[RabbitMQ tutorials in
Go](https://github.com/rabbitmq/rabbitmq-tutorials/tree/master/go) are also
available.

## Contributing

Pull requests are very much welcomed.  Create your pull request on a non-master
branch, make sure a test or example is included that covers your change and
your commits represent coherent changes that include a reason for the change.

To run the integration tests, make sure you have RabbitMQ running on any host,
export the environment variable `AMQP_URL=amqp://host/` and run `go test -tags
integration`.  TravisCI will also run the integration tests.

Thanks to the [community of contributors](https://github.com/streadway/amqp/graphs/contributors).

## External packages

  * [Google App Engine Dialer support](https://github.com/soundtrackyourbrand/gaeamqp)
  * [RabbitMQ examples in Go](https://github.com/rabbitmq/rabbitmq-tutorials/tree/master/go)

## License

BSD 2 clause - see LICENSE for more details.


