// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

package smtp

import (
	"fmt"
)

// loginAuth is the type that satisfies the Auth interface for the "SMTP LOGIN" auth
type loginAuth struct {
	username, password   string
	host                 string
	respStep             uint8
	allowUnencryptedAuth bool
}

// LoginAuth returns an [Auth] that implements the LOGIN authentication
// mechanism as it is used by MS Outlook. The Auth works similar to PLAIN
// but instead of sending all in one response, the login is handled within
// 3 steps:
// - Sending AUTH LOGIN (server might responds with "Username:")
// - Sending the username (server might responds with "Password:")
// - Sending the password (server authenticates)
// This is the common approach as specified by Microsoft in their MS-XLOGIN spec.
// See: https://msopenspecs.azureedge.net/files/MS-XLOGIN/%5bMS-XLOGIN%5d.pdf
// Yet, there is also an old IETF draft for SMTP AUTH LOGIN that states for clients:
// "The contents of both challenges SHOULD be ignored.".
// See: https://datatracker.ietf.org/doc/html/draft-murchison-sasl-login-00
// Since there is no official standard RFC and we've seen different implementations
// of this mechanism (sending "Username:", "Username", "username", "User name", etc.)
// we follow the IETF-Draft and ignore any server challenge to allow compatibility
// with most mail servers/providers.
//
// LoginAuth will only send the credentials if the connection is using TLS
// or is connected to localhost. Otherwise authentication will fail with an
// error, without sending the credentials.
func LoginAuth(username, password, host string, allowUnenc bool) Auth {
	return &loginAuth{username, password, host, 0, allowUnenc}
}

// Start begins the SMTP authentication process by validating server's TLS status and hostname.
// Returns "LOGIN" on success.
func (a *loginAuth) Start(server *ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises LOGIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if !a.allowUnencryptedAuth && !server.TLS && !isLocalhost(server.Name) {
		return "", nil, ErrUnencrypted
	}
	if server.Name != a.host {
		return "", nil, ErrWrongHostname
	}
	a.respStep = 0
	return "LOGIN", nil, nil
}

// Next processes responses from the server during the SMTP authentication exchange, sending the
// username and password.
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch a.respStep {
		case 0:
			a.respStep++
			return []byte(a.username), nil
		case 1:
			a.respStep++
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedServerResponse, string(fromServer))
		}
	}
	return nil, nil
}
