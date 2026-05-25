// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
)

type AuthData struct {
	Auth     bool
	Username string
	Password string
}

var testHookTLSConfig func() *tls.Config // nil, except for tests

// QuickSend is an all-in-one method for quickly sending simple text mails in go-mail.
//
// This method will create a new client that connects to the server at addr, switches to TLS if possible,
// authenticates with the optional AuthData provided in auth and create a new simple Msg with the provided
// subject string and message bytes as body. The message will be sent using from as sender address and will
// be delivered to every address in rcpts. QuickSend will always send as text/plain ContentType.
//
// For the SMTP authentication, if auth is not nil and AuthData.Auth is set to true, it will try to
// autodiscover the best SMTP authentication mechanism supported by the server. If auth is set to true
// but autodiscover is not able to find a suitable authentication mechanism or if the authentication
// fails, the mail delivery will fail completely.
//
// The content parameter should be an RFC 822-style email body. The lines of content should be CRLF terminated.
//
// Parameters:
//   - addr: The hostname and port of the mail server, it must include a port, as in "mail.example.com:smtp".
//   - auth: A AuthData pointer. If nil or if AuthData.Auth is set to false, not SMTP authentication will be performed.
//   - from: The from address of the sender as string.
//   - rcpts: A slice of strings of receipient addresses.
//   - subject: The subject line as string.
//   - content: A byte slice of the mail content
//
// Returns:
//   - A pointer to the generated Msg.
//   - An error if any step in the process of mail generation or delivery failed.
func QuickSend(addr string, auth *AuthData, from string, rcpts []string, subject string, content []byte) (*Msg, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port from address: %w", err)
	}
	portnum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port to int: %w", err)
	}
	client, err := NewClient(host, WithPort(portnum), WithTLSPolicy(TLSOpportunistic))
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	if auth != nil && auth.Auth {
		client.SetSMTPAuth(SMTPAuthAutoDiscover)
		client.SetUsername(auth.Username)
		client.SetPassword(auth.Password)
	}

	tlsConfig := client.tlsconfig
	if testHookTLSConfig != nil {
		tlsConfig = testHookTLSConfig()
	}
	if err = client.SetTLSConfig(tlsConfig); err != nil {
		return nil, fmt.Errorf("failed to set TLS config: %w", err)
	}

	message := NewMsg()
	if err = message.From(from); err != nil {
		return nil, fmt.Errorf("failed to set MAIL FROM address: %w", err)
	}
	if err = message.To(rcpts...); err != nil {
		return nil, fmt.Errorf("failed to set RCPT TO address: %w", err)
	}
	message.Subject(subject)
	buffer := bytes.NewBuffer(content)
	writeFunc := writeFuncFromBuffer(buffer)
	message.SetBodyWriter(TypeTextPlain, writeFunc)

	if err = client.DialAndSend(message); err != nil {
		return nil, fmt.Errorf("failed to dial and send message: %w", err)
	}
	return message, nil
}

// NewAuthData creates a new AuthData instance with the provided username and password.
//
// This function initializes an AuthData struct with authentication enabled and sets the
// username and password fields.
//
// Parameters:
//   - user: The username for authentication.
//   - pass: The password for authentication.
//
// Returns:
//   - A pointer to the initialized AuthData instance.
func NewAuthData(user, pass string) *AuthData {
	return &AuthData{
		Auth:     true,
		Username: user,
		Password: pass,
	}
}
