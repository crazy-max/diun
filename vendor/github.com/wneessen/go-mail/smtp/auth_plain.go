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

// plainAuth is the type that satisfies the Auth interface for the "SMTP PLAIN" auth
type plainAuth struct {
	identity, username, password string
	host                         string
	allowUnencryptedAuth         bool
}

// PlainAuth returns an [Auth] that implements the PLAIN authentication
// mechanism as defined in RFC 4616. The returned Auth uses the given
// username and password to authenticate to host and act as identity.
// Usually identity should be the empty string, to act as username.
//
// PlainAuth will only send the credentials if the connection is using TLS
// or is connected to localhost. Otherwise authentication will fail with an
// error, without sending the credentials.
func PlainAuth(identity, username, password, host string, allowUnenc bool) Auth {
	return &plainAuth{identity, username, password, host, allowUnenc}
}

func (a *plainAuth) Start(server *ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises PLAIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if !a.allowUnencryptedAuth && !server.TLS && !isLocalhost(server.Name) {
		return "", nil, ErrUnencrypted
	}
	if server.Name != a.host {
		return "", nil, ErrWrongHostname
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, ErrUnexpectedServerChallange
	}
	return nil, nil
}
