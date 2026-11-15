// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// messageType represents the type of NTLM message
type messageType int

// ChallengeMessage represents an NTLM challenge message (Type 2 message)
type ChallengeMessage struct {
	signature       []byte
	messageType     uint32
	negotiateFlags  negotiateFlagset
	serverChallenge []byte
	reserved        []byte
	targetName      *Payload
	targetInfo      *avPairs
	payload         []byte
}

const (
	messageTypeNegotiate messageType = iota + 1
	messageTypeChallenge
	messageTypeAuthenticate
)

var (
	// ErrNTLMInvalidChallengeMessage is returned when the challenge message is invalid
	ErrNTLMInvalidChallengeMessage = errors.New("invalid challenge message")

	// ErrNTLMInvalidSignature is returned when the NTLM message signature is invalid
	ErrNTLMInvalidSignature = errors.New("invalid NTLM signature")

	// ErrNTLMInvalidMessageType is returned when the NTLM message type is invalid
	ErrNTLMInvalidMessageType = errors.New("invalid NTLM message type")
)

// CreateChallengeMessage creates an NTLM challenge message (Type 2 message) for the
// given target name and domain
func CreateChallengeMessage(clientFlags uint32, challenge []byte, targetName, domain string) ([]byte, error) {
	const headerLen = 48
	const versionLen = 8

	target := utf16FromString(targetName)
	targetInfo := new(avPairs)
	if err := targetInfo.appendAVPair(msvAVNbDomainName, utf16FromString(domain)); err != nil {
		return nil, fmt.Errorf("failed to append netbios domain name to targetinfo: %w", err)
	}
	if err := targetInfo.appendAVPair(msvAVNbComputerName, target); err != nil {
		return nil, fmt.Errorf("failed to append netbios computer name to targetinfo: %w", err)
	}
	if err := targetInfo.appendAVPair(msvAVDNSDomainName, utf16FromString(domain)); err != nil {
		return nil, fmt.Errorf("failed to append dns domain name to targetinfo: %w", err)
	}
	if err := targetInfo.appendAVPair(msvAVDNSComputerName, target); err != nil {
		return nil, fmt.Errorf("failed to append dns computer name to targetinfo: %w", err)
	}
	if err := targetInfo.appendAVPair(msvAVTimestamp, timeToWindowsFileTime(time.Now())); err != nil {
		return nil, fmt.Errorf("failed to append timestamp to targetinfo: %w", err)
	}
	if err := targetInfo.appendAVPair(msvAVEOL, nil); err != nil {
		return nil, fmt.Errorf("failed to append EOL to targetinfo: %w", err)
	}
	tiBytes := targetInfo.bytes() // compute once

	flags := negotiateFlagset(ntlmsspNegotiateUnicode | ntlmsspNegotiateNTLM | ntlmsspRequestTarget |
		ntlmsspNegotiateAlwaysSign | ntlmsspNegotiateTargetTypeServer |
		ntlmsspNegotiateTargetInfo | ntlmsspNegotiateVersion)
	if clientFlags&uint32(ntlmsspNegotiateExtendedSessionSecurity) != 0 {
		flags |= negotiateFlagset(ntlmsspNegotiateExtendedSessionSecurity)
	}
	if clientFlags&uint32(ntlmsspNegotiateKeyExchange) != 0 {
		flags |= negotiateFlagset(ntlmsspNegotiateKeyExchange)
	}

	payloadStart := uint32(headerLen + versionLen) // 56
	buffer := bytes.NewBuffer(make([]byte, 0, int(payloadStart)+len(target)+len(tiBytes)))

	// Signature + type
	buffer.Write([]byte(signature))
	if err := binary.Write(buffer, binary.LittleEndian, uint32(messageTypeChallenge)); err != nil {
		return nil, fmt.Errorf("failed to write message type: %w", err)
	}

	// TargetName security buffer  (len uint16, maxlen uint16, offset uint32)
	targetOff := payloadStart
	targetLength, err := toUint16(len(target))
	if err != nil {
		return nil, fmt.Errorf("failed to convert target name length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, targetLength); err != nil {
		return nil, fmt.Errorf("failed to write target name length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, targetLength); err != nil {
		return nil, fmt.Errorf("failed to write target name max length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, targetOff); err != nil {
		return nil, fmt.Errorf("failed to write target name offset: %w", err)
	}

	// Negotiate flags
	buffer.Write(flags.bytes())

	// Server challenge (8 bytes)
	buffer.Write(challenge)

	// Reserved (8 bytes)  <-- was missing
	buffer.Write(make([]byte, 8))

	// TargetInfo security buffer
	targetLength32, err := toUint32(len(target))
	if err != nil {
		return nil, fmt.Errorf("failed to convert target name length: %w", err)
	}
	tiOff := targetOff + targetLength32
	tiByteLength, err := toUint16(len(tiBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to convert target info length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, tiByteLength); err != nil {
		return nil, fmt.Errorf("failed to write target info length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, tiByteLength); err != nil {
		return nil, fmt.Errorf("failed to write target info max length: %w", err)
	}
	if err := binary.Write(buffer, binary.LittleEndian, tiOff); err != nil {
		return nil, fmt.Errorf("failed to write target info offset: %w", err)
	}

	// Version (8 bytes)
	buffer.Write(make([]byte, 8))

	// Payload
	buffer.Write(target)
	buffer.Write(tiBytes)

	return buffer.Bytes(), nil
}
