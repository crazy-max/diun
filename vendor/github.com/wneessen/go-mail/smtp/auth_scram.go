// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

package smtp

import (
	"bytes"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/secure/precis"
)

// scramAuth represents a SCRAM (Salted Challenge Response Authentication Mechanism) client and
// satisfies the smtp.Auth interface.
type scramAuth struct {
	username, password, algorithm               string
	firstBareMsg, nonce, saltedPwd, authMessage []byte
	iterations                                  int
	h                                           func() hash.Hash
	isPlus                                      bool
	tlsConnState                                *tls.ConnectionState
	bindData                                    []byte
}

// ScramSHA1Auth creates and returns a new SCRAM-SHA-1 authentication mechanism with the given
// username and password.
func ScramSHA1Auth(username, password string) Auth {
	return &scramAuth{
		username:  username,
		password:  password,
		algorithm: "SCRAM-SHA-1",
		h:         sha1.New,
	}
}

// ScramSHA256Auth creates and returns a new SCRAM-SHA-256 authentication mechanism with the given
// username and password.
func ScramSHA256Auth(username, password string) Auth {
	return &scramAuth{
		username:  username,
		password:  password,
		algorithm: "SCRAM-SHA-256",
		h:         sha256.New,
	}
}

// ScramSHA1PlusAuth returns an Auth instance configured for SCRAM-SHA-1-PLUS authentication with
// the provided username, password, and TLS connection state.
func ScramSHA1PlusAuth(username, password string, tlsConnState *tls.ConnectionState) Auth {
	return &scramAuth{
		username:     username,
		password:     password,
		algorithm:    "SCRAM-SHA-1-PLUS",
		h:            sha1.New,
		isPlus:       true,
		tlsConnState: tlsConnState,
	}
}

// ScramSHA256PlusAuth returns an Auth instance configured for SCRAM-SHA-256-PLUS authentication with
// the provided username, password, and TLS connection state.
func ScramSHA256PlusAuth(username, password string, tlsConnState *tls.ConnectionState) Auth {
	return &scramAuth{
		username:     username,
		password:     password,
		algorithm:    "SCRAM-SHA-256-PLUS",
		h:            sha256.New,
		isPlus:       true,
		tlsConnState: tlsConnState,
	}
}

// Start initializes the SCRAM authentication process and returns the selected algorithm, nil data, and no error.
func (a *scramAuth) Start(_ *ServerInfo) (string, []byte, error) {
	return a.algorithm, nil, nil
}

// Next processes the server's challenge and returns the client's response for SCRAM authentication.
func (a *scramAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		if len(fromServer) == 0 {
			a.reset()
			return a.initialClientMessage()
		}
		switch {
		case bytes.HasPrefix(fromServer, []byte("r=")):
			resp, err := a.handleServerFirstResponse(fromServer)
			if err != nil {
				a.reset()
				return nil, err
			}
			return resp, nil
		case bytes.HasPrefix(fromServer, []byte("v=")):
			resp, err := a.handleServerValidationMessage(fromServer)
			if err != nil {
				a.reset()
				return nil, err
			}
			return resp, nil
		default:
			a.reset()
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedServerResponse, string(fromServer))
		}
	}
	return nil, nil
}

// reset clears all authentication-related properties in the scramAuth instance, effectively resetting its state.
func (a *scramAuth) reset() {
	a.nonce = nil
	a.firstBareMsg = nil
	a.saltedPwd = nil
	a.authMessage = nil
	a.iterations = 0
}

// initialClientMessage generates the initial message for SCRAM authentication, including a nonce and
// optional channel binding.
func (a *scramAuth) initialClientMessage() ([]byte, error) {
	username, err := a.normalizeUsername()
	if err != nil {
		return nil, fmt.Errorf("username normalization failed: %w", err)
	}

	nonceBuffer := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, nonceBuffer); err != nil {
		return nil, fmt.Errorf("unable to generate client secret: %w", err)
	}
	a.nonce = make([]byte, base64.StdEncoding.EncodedLen(len(nonceBuffer)))
	base64.StdEncoding.Encode(a.nonce, nonceBuffer)

	a.firstBareMsg = []byte("n=" + username + ",r=" + string(a.nonce))
	returnBytes := []byte("n,," + string(a.firstBareMsg))

	// SCRAM-SHA-X-PLUS auth requires channel binding
	if a.isPlus {
		if a.tlsConnState == nil {
			return nil, errors.New("tls connection state is required for SCRAM-SHA-X-PLUS")
		}
		bindType := "tls-unique"
		connState := a.tlsConnState
		bindData := connState.TLSUnique

		// crypto/tls: no tls-unique channel binding value for this tls connection, possibly due to missing
		// extended master key support and/or resumed connection
		// RFC9266:122 tls-unique not defined for tls 1.3 and later
		if bindData == nil || connState.Version >= tls.VersionTLS13 {
			bindType = "tls-exporter"
			bindData, err = connState.ExportKeyingMaterial("EXPORTER-Channel-Binding", []byte{}, 32)
			if err != nil {
				return nil, fmt.Errorf("unable to export keying material: %w", err)
			}
		}
		bindData = []byte("p=" + bindType + ",," + string(bindData))
		a.bindData = make([]byte, base64.StdEncoding.EncodedLen(len(bindData)))
		base64.StdEncoding.Encode(a.bindData, bindData)
		returnBytes = []byte("p=" + bindType + ",," + string(a.firstBareMsg))
	}

	return returnBytes, nil
}

// handleServerFirstResponse processes the first response from the server in SCRAM authentication.
func (a *scramAuth) handleServerFirstResponse(fromServer []byte) ([]byte, error) {
	parts := bytes.Split(fromServer, []byte(","))
	if len(parts) < 3 {
		return nil, errors.New("not enough fields in the first server response")
	}
	if !bytes.HasPrefix(parts[0], []byte("r=")) {
		return nil, errors.New("first part of the server response does not start with r=")
	}
	if !bytes.HasPrefix(parts[1], []byte("s=")) {
		return nil, errors.New("second part of the server response does not start with s=")
	}
	if !bytes.HasPrefix(parts[2], []byte("i=")) {
		return nil, errors.New("third part of the server response does not start with i=")
	}

	combinedNonce := parts[0][2:]
	if len(a.nonce) == 0 || !bytes.HasPrefix(combinedNonce, a.nonce) {
		return nil, errors.New("server nonce does not start with our nonce")
	}
	a.nonce = combinedNonce

	encodedSalt := parts[1][2:]
	salt := make([]byte, base64.StdEncoding.DecodedLen(len(encodedSalt)))
	n, err := base64.StdEncoding.Decode(salt, encodedSalt)
	if err != nil {
		return nil, fmt.Errorf("invalid encoded salt: %w", err)
	}
	salt = salt[:n]

	iterations, err := strconv.Atoi(string(parts[2][2:]))
	if err != nil {
		return nil, fmt.Errorf("invalid iterations: %w", err)
	}
	a.iterations = iterations

	password, err := a.normalizeString(a.password)
	if err != nil {
		return nil, fmt.Errorf("unable to normalize password: %w", err)
	}

	a.saltedPwd, err = pbkdf2.Key(a.h, password, salt, a.iterations, a.h().Size())
	if err != nil {
		return nil, fmt.Errorf("unable to derive key: %w", err)
	}

	msgWithoutProof := []byte("c=biws,r=" + string(a.nonce))

	// A PLUS authentication requires the channel binding data
	if a.isPlus {
		msgWithoutProof = []byte("c=" + string(a.bindData) + ",r=" + string(a.nonce))
	}

	a.authMessage = []byte(string(a.firstBareMsg) + "," + string(fromServer) + "," + string(msgWithoutProof))
	clientProof := a.computeClientProof()

	return []byte(string(msgWithoutProof) + ",p=" + string(clientProof)), nil
}

// handleServerValidationMessage verifies the server's signature during the SCRAM authentication process.
func (a *scramAuth) handleServerValidationMessage(fromServer []byte) ([]byte, error) {
	serverSignature := fromServer[2:]
	computedServerSignature := a.computeServerSignature()

	if !hmac.Equal(serverSignature, computedServerSignature) {
		return nil, errors.New("invalid server signature")
	}
	return []byte(""), nil
}

// computeHMAC generates a Hash-based Message Authentication Code (HMAC) using the specified key and message.
func (a *scramAuth) computeHMAC(key, msg []byte) []byte {
	mac := hmac.New(a.h, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

// computeHash generates a hash of the given key using the configured hashing algorithm.
func (a *scramAuth) computeHash(key []byte) []byte {
	hasher := a.h()
	hasher.Write(key)
	return hasher.Sum(nil)
}

// computeClientProof generates the client proof as part of the SCRAM authentication process.
func (a *scramAuth) computeClientProof() []byte {
	clientKey := a.computeHMAC(a.saltedPwd, []byte("Client Key"))
	storedKey := a.computeHash(clientKey)
	clientSignature := a.computeHMAC(storedKey[:], a.authMessage)
	clientProof := make([]byte, len(clientSignature))
	for i := 0; i < len(clientSignature); i++ {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
	base64.StdEncoding.Encode(buf, clientProof)
	return buf
}

// computeServerSignature returns the computed base64-encoded server signature in the SCRAM
// authentication process.
func (a *scramAuth) computeServerSignature() []byte {
	serverKey := a.computeHMAC(a.saltedPwd, []byte("Server Key"))
	serverSignature := a.computeHMAC(serverKey, a.authMessage)
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(serverSignature)))
	base64.StdEncoding.Encode(buf, serverSignature)
	return buf
}

// normalizeUsername replaces special characters in the username for SCRAM authentication
// and prepares it using the SASLprep profile as per RFC 8265, returning the normalized
// username or an error.
func (a *scramAuth) normalizeUsername() (string, error) {
	// RFC 5802 section 5.1: the characters ',' or '=' in usernames are
	// sent as '=2C' and '=3D' respectively.
	replacer := strings.NewReplacer("=", "=3D", ",", "=2C")
	username := replacer.Replace(a.username)
	// RFC 5802 section 5.1: before sending the username to the server,
	// the client SHOULD prepare the username using the "SASLprep"
	// profile [RFC4013] of the "stringprep" algorithm [RFC3454]
	// treating it as a query string (i.e., unassigned Unicode code
	// points are allowed). If the preparation of the username fails or
	// results in an empty string, the client SHOULD abort the
	// authentication exchange.
	//
	// Since RFC 8265 obsoletes RFC 4013 we use it instead.
	username, err := a.normalizeString(username)
	if err != nil {
		return "", fmt.Errorf("unable to normalize username: %w", err)
	}
	return username, nil
}

// normalizeString normalizes the input string according to the OpaqueString profile of the
// precis framework. It returns the normalized string or an error if normalization fails or
// results in an empty string.
func (a *scramAuth) normalizeString(s string) (string, error) {
	s, err := precis.OpaqueString.String(s)
	if err != nil {
		return "", fmt.Errorf("failed to normalize string: %w", err)
	}
	return s, nil
}
