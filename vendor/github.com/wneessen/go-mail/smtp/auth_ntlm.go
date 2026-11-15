// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package smtp

import (
	"fmt"

	"github.com/wneessen/go-mail/internal/ntlm"
)

// ntlmAuth implements the NTLM authentication mechanism for SMTP.
// It holds the necessary information to perform NTLM authentication.
type ntlmAuth struct {
	// session is the NTLM client session used for authentication.
	session *ntlm.NTLMv2Session

	// UserName is the username for NTLM authentication.
	UserName string

	// Password is the password for NTLM authentication.
	Password string

	// Domain is the domain name for NTLM authentication.
	Domain string
}

// Start initiates the NTLM authentication process for SMTP.
// It checks if the server supports NTLM, creates an NTLM client session,
// and generates the initial negotiation message.
//
// Parameters:
//   - info: A pointer to ServerInfo containing information about the SMTP server,
//     including supported authentication methods.
//
// Returns:
//   - proto: A string indicating the authentication protocol used ("NTLM").
//   - toServer: A byte slice containing the NTLM negotiate message to be sent to the server.
//   - err: An error if the NTLM authentication method is not supported or if there's an issue
//     creating the NTLM session or generating the negotiation message.
func (a *ntlmAuth) Start(info *ServerInfo) (proto string, toServer []byte, err error) {
	// Initialize a NTLMv2 client session and assign username, password and domain
	a.session = ntlm.NewNTLMv2Session()
	a.session.SetUserInfo(a.UserName, a.Password, a.Domain)

	// Initialize a NTLMSSP Type 1 message, which holds the negotiation information
	// for the NTLM authentication.
	//
	// See: https://curl.se/rfc/ntlm.html#theType1Message
	negotiate, err := a.session.GenerateNegotiateMessage()
	if err != nil {
		return "NTLM", nil, err
	}
	negoBytes, err := negotiate.Bytes()
	if err != nil {
		return "NTLM", nil, err
	}
	return "NTLM", negoBytes, nil
}

// Next continues the NTLM authentication process by parsing the Type 2 message from the
// server followed by generating and sending the final Type 3 authentication message.
//
// Parameters:
//   - challengeBytes: A byte slice containing the NTLM challenge message from the server.
//   - more: A boolean indicating whether more authentication steps are expected.
//
// Returns:
//   - toServer: A byte slice containing the NTLM authenticate message to be sent to the server,
//     or nil if no further authentication steps are needed.
//   - err: An error if there's an issue parsing the challenge, processing it, or generating
//     the "authenticate" message. Returns nil if successful or if no further steps are needed.
func (a *ntlmAuth) Next(challengeBytes []byte, more bool) (toServer []byte, err error) {
	if more {
		// Process the server challenge message (Type 2 message)
		//
		// See: https://curl.se/rfc/ntlm.html#theType2Message
		if err := a.session.ParseChallengeMessage(challengeBytes); err != nil {
			return nil, fmt.Errorf("failed to parse challenge message: %w", err)
		}

		// Generate the authentication message (Type 3 message) and return it to the server
		//
		// See: https://curl.se/rfc/ntlm.html#theType3Message
		authenticate, err := a.session.GenerateAuthenticateMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to generate authentication message: %w", err)
		}
		return authenticate.Bytes(), nil
	}
	// no further authentication steps are needed, return nil
	return nil, nil
}

// NTLMAuth returns an [Auth] that implements the NTLM authentication
// mechanism using the NTLMSSP protocol.
func NTLMAuth(user, pass, domain string) Auth {
	return &ntlmAuth{
		UserName: user,
		Password: pass,
		Domain:   domain,
	}
}
