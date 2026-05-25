// SPDX-FileCopyrightText: Copyright 2010 The Go Authors. All rights reserved.
// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// Original net/smtp code from the Go stdlib by the Go Authors.
// Use of this source code is governed by a BSD-style
// LICENSE file that can be found in this directory.
//
// go-mail specific modifications by the go-mail Authors.
// Licensed under the MIT License.
// See [PROJECT ROOT]/LICENSES directory for more information.
//
// SPDX-License-Identifier: BSD-3-Clause AND MIT

package smtp

import "errors"

var (
	// ErrUnencrypted is an error indicating that the connection is not encrypted.
	ErrUnencrypted = errors.New("unencrypted connection")
	// ErrUnexpectedServerChallange is an error indicating that the server issued an unexpected challenge.
	ErrUnexpectedServerChallange = errors.New("unexpected server challenge")
	// ErrUnexpectedServerResponse is an error indicating that the server issued an unexpected response.
	ErrUnexpectedServerResponse = errors.New("unexpected server response")
	// ErrWrongHostname is an error indicating that the provided hostname does not match the expected value.
	ErrWrongHostname = errors.New("wrong host name")
)

// Auth is implemented by an SMTP authentication mechanism.
type Auth interface {
	// Start begins an authentication with a server.
	// It returns the name of the authentication protocol
	// and optionally data to include in the initial AUTH message
	// sent to the server.
	// If it returns a non-nil error, the SMTP client aborts
	// the authentication attempt and closes the connection.
	Start(server *ServerInfo) (proto string, toServer []byte, err error)

	// Next continues the authentication. The server has just sent
	// the fromServer data. If more is true, the server expects a
	// response, which Next should return as toServer; otherwise
	// Next should return toServer == nil.
	// If Next returns a non-nil error, the SMTP client aborts
	// the authentication attempt and closes the connection.
	Next(fromServer []byte, more bool) (toServer []byte, err error)
}

// ServerInfo records information about an SMTP server.
type ServerInfo struct {
	Name string   // SMTP server name
	TLS  bool     // using TLS, with valid certificate for Name
	Auth []string // advertised authentication mechanisms
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}
