<!--
SPDX-FileCopyrightText: The go-mail Authors

SPDX-License-Identifier: MIT
-->

# go-mail - Easy to use, yet comprehensive library for sending mails with Go

[![GoDoc](https://godoc.org/github.com/wneessen/go-mail?status.svg)](https://pkg.go.dev/github.com/wneessen/go-mail)
[![codecov](https://codecov.io/gh/wneessen/go-mail/branch/main/graph/badge.svg?token=37KWJV03MR)](https://codecov.io/gh/wneessen/go-mail) 
[![Go Report Card](https://goreportcard.com/badge/github.com/wneessen/go-mail)](https://goreportcard.com/report/github.com/wneessen/go-mail) 
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)
[![#go-mail on Discord](https://img.shields.io/badge/Discord-%23go%E2%80%93mail-blue.svg)](https://discord.gg/ysQXkaccXk) 
[![REUSE status](https://api.reuse.software/badge/github.com/wneessen/go-mail)](https://api.reuse.software/info/github.com/wneessen/go-mail)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8701/badge)](https://www.bestpractices.dev/projects/8701)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/wneessen/go-mail/badge)](https://securityscorecards.dev/viewer/?uri=github.com/wneessen/go-mail)
<a href="https://ko-fi.com/D1D24V9IX"><img src="https://uploads-ssl.webflow.com/5c14e387dab576fe667689cf/5cbed8a4ae2b88347c06c923_BuyMeACoffee_blue.png" height="20" alt="buy ma a coffee"></a>

<p style="text-align: center"><img src="./assets/gopher2.svg" width="250" alt="go-mail logo"/></p>

The main idea of this library was to provide a simple interface for sending mails to
my [JS-Mailer](https://github.com/wneessen/js-mailer) project. It quickly evolved into a full-fledged mail library.

go-mail follows idiomatic Go style and best practice. It has a small dependency footprint by mainly relying on the
Go Standard Library and the Go extended packages. It combines a lot of functionality from the standard library to 
give easy and convenient access to mail and SMTP related tasks.

In the early days, parts of this library (especially some parts of [msgwriter.go](msgwriter.go)) had been 
forked/ported from [go-mail/mail](https://github.com/go-mail/mail) and respectively [go-gomail/gomail](https://github.com/go-gomail/gomail). Today
most of the ported code has been refactored.

The `smtp` package of go-mail has been forked from the original Go stdlib's `net/smtp` package and has then been extended 
by the go-mail team to fit the packages needs (more SMTP Auth methods, logging, concurrency-safety, etc.).

## Features

Here are some highlights of go-mail's featureset:

* [X] Very small dependency footprint (mainly Go Stdlib and Go extended packages)
* [X] Modern, idiomatic Go
* [X] Sane and secure defaults
* [X] Implicit SSL/TLS support
* [X] Explicit STARTTLS support with different policies
* [X] Makes use of contexts for a better control flow and timeout/cancelation handling
* [X] SMTP Auth support
  * [X] CRAM-MD5
  * [X] LOGIN
  * [X] PLAIN
  * [X] SCRAM-SHA-1/SCRAM-SHA-1-PLUS
  * [X] SCRAM-SHA-256/SCRAM-SHA-256-PLUS
  * [X] XOAUTH2
* [X] RFC5322 compliant mail address validation
* [X] Support for common mail header field generation (Message-ID, Date, Bulk-Precedence, Priority, etc.)
* [X] Concurrency-safe reusing the same SMTP connection to send multiple mails
* [X] Support for attachments and inline embeds (from file system, `io.Reader`, `embed.FS` or `fs.FS`)
* [X] Support for different encodings
* [X] Middleware support for 3rd-party libraries to alter mail messages
* [X] Support sending mails via a local sendmail command
* [X] Support for requestng MDNs (RFC 8098) and DSNs (RFC 1891)
* [X] DKIM signature support via [go-mail-middlware](https://github.com/wneessen/go-mail-middleware)
* [X] Message object satisfies `io.WriterTo` and `io.Reader` interfaces
* [X] Support for Go's `html/template` and `text/template` (as message body, alternative part or attachment/emebed)
* [X] Output to file support which allows storing mail messages as e. g. `.eml` files to disk to open them in a MUA
* [X] Debug logging of SMTP traffic
* [X] Custom error types for delivery errors
* [X] Custom dial-context functions for more control over the connection (proxing, DNS hooking, etc.)
* [X] Output a go-mail message as EML file and parse EML file into a go-mail message
* [X] S/MIME message signing support (RSA and ECDSA)
* [X] UNIX domain socket support
* [X] Pluggable SMTP error registry for advanced handling of non-RFC-conforming servers

go-mail works like a programatic email client and provides lots of methods and functionalities you would consider
standard in a MUA.

## Documentation
We aim for good GoDoc documenation in our library which gives you a full API reference. We also provide a more in-depth 
documentation website at [go-mail.dev](https://go-mail.dev)

## Compatibility

Go evolves quickly and introduces valuable improvements with each release. To balance adopting new features with 
ensuring stability for our users, we align our support with the official Go release policy. In practice, go-mail 
will always support the same set of Go versions that the Go team actively maintains.

Since Go provides two releases per year, this translates into roughly one year of guaranteed compatibility for 
any given Go version. We encourage users to stay current with supported Go versions to benefit from security 
updates, performance improvements, and new language features.

## Support
We have a support and general discussion channel on Discord. Find us at: [#go-mail](https://discord.gg/dbfQyC4s) alternatively find us
on the [Gophers Slack](https://gophers.slack.com) in #go-mail

## Middleware
The goal of go-mail is to keep it free from 3rd party dependencies and only focus on things a mail library should
fulfill. Yet, since version v0.2.8 we've added support for middleware on the `Msg` object, allowing 3rd parties to
alter a given mail message to their needs without relying on `go-mail` to support their specific need.

To get our users started with message middleware, we've created a collection of useful middlewares. It can be 
found in a seperate repository: [go-mail-middlware](https://github.com/wneessen/go-mail-middleware).

## Merch
Thanks to our wonderful friends at [HelloTux](https://www.hellotux.com) we can offer great go-mail merchandising. All merch articles are embroidery 
to provide the best and most long-lasting quality possible.

If you want to support the open source community and represent your favourite Go mail library with some cool drip, check out our merch shop at: 
[https://www.hellotux.com/go-mail](https://www.hellotux.com/go-mail).

## Examples

We provide example code in both our GoDocs as well as on our official Website (see [Documentation](#documentation)). For a quick start into go-mail
check out our [Getting started](https://go-mail.dev/getting-started/introduction/) guide.

## Authors/Contributors
go-mail was initially created and developed by [Winni Neessen](https://github.com/wneessen/), but over time a lot of amazing people 
contributed ot the project. Big thanks to all of them for improving the go-mail project (be it writing code, testing
code, reviewing code, writing documenation or helping to translate the website):

<a href="https://github.com/wneessen/go-mail/graphs/contributors">
  <img alt="image of contributors" src="https://contrib.rocks/image?repo=wneessen/go-mail" />
</a>

A huge thank you also goes to [Maria Letta](https://github.com/MariaLetta) for designing our super cool go-mail logo!

## Sponsors
We sincerely thank our amazing sponsors for their generous support! Your contributions do not go unnoticed and helps
keeping up the project!

* [kolaente](https://github.com/kolaente)
