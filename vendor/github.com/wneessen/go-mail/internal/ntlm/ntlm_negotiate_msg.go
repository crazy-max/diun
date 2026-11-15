// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	// negotiateFlag holds the individual flags for the NTLMv2 negotiate message (Type 1 message).
	negotiateFlag uint32

	// negotiateFlagset holds the flags for the NTLMv2 negotiate message (Type 1 message).
	negotiateFlagset uint32
)

// NegotiateMessage represents a NTLMv2 negotiate message (Type 1 message).
type NegotiateMessage struct {
	signature      []byte
	messageType    uint32
	negotiateFlags negotiateFlagset
	domainname     *Payload
	workstation    *Payload
}

// List of required NTLM flags. This list only holds the flags that are required for
// this package to work. Check the reference for a comprehensive list.
//
// See: https://curl.se/rfc/ntlm.html#theNtlmFlags
const (
	// ntlmsspNegotiateUnicode indicates that Unicode strings are supported for use
	// in security buffer data.
	ntlmsspNegotiateUnicode negotiateFlag = 0x00000001

	// ntlmsspNegotiateOEM indicates that OEM strings are supported for use in
	// security buffer data.
	ntlmsspNegotiateOEM negotiateFlag = 0x00000002

	// ntlmsspRequestTarget requests that the server's authentication realm be included
	// in the Type 2 message.
	ntlmsspRequestTarget negotiateFlag = 0x00000004

	// ntlmsspNegotiateLMKey indicates that the Lan Manager Session Key should be used
	// for signing and sealing authenticated communications. We do not use it in NTLMv2
	// ntlmsspNegotiateLMKey negotiateFlag = 0x00000080

	// ntlmsspNegotiateNTLM indicates that NTLM authentication is being used.
	ntlmsspNegotiateNTLM negotiateFlag = 0x00000200

	// ntlmsspNegotiateDomainSupplied is sent by the client in the Type 1 message to indicate that
	// the name of the domain in which the client workstation has membership is included in the message.
	// This is used by the server to determine whether the client is eligible for local authentication.
	ntlmsspNegotiateDomainSupplied negotiateFlag = 0x00001000

	// ntlmsspNegotiateAlwaysSign indicates that authenticated communication between the client and server
	// should be signed with a "dummy" signature.
	ntlmsspNegotiateAlwaysSign negotiateFlag = 0x00008000

	// ntlmsspNegotiateExtendedSessionSecurity indicates that the NTLM2 signing and sealing scheme
	// should be used for protecting authenticated communications.
	//
	// Note that this refers to a particular session security scheme, and is not related to the use of
	// NTLMv2 authentication. This flag can, however, have an effect on the response calculations.
	//
	// See: https://curl.se/rfc/ntlm.html#theNtlm2SessionResponse
	ntlmsspNegotiateExtendedSessionSecurity negotiateFlag = 0x00080000

	// ntlmsspNegotiateTargetTypeServer is sent by the server in the Type 2 message to indicate that
	// the target authentication realm is a server.
	ntlmsspNegotiateTargetTypeServer negotiateFlag = 0x00020000

	// ntlmsspNegotiateTargetInfo is sent by the server in the Type 2 message to indicate that it is
	// including a Target Information block in the message. The Target Information block is used in
	// the calculation of the NTLMv2 response.
	ntlmsspNegotiateTargetInfo negotiateFlag = 0x00800000

	// ntlmsspNegotiateVersion indicates the OS version (purley cosmetic and for debugging purposes)
	// We ignore this
	ntlmsspNegotiateVersion negotiateFlag = 0x02000000

	// ntlmsspNegotiate128Bit indicates that 128-bit encryption is supported.
	ntlmsspNegotiate128Bit negotiateFlag = 0x20000000

	// ntlmsspNegotiateKeyExchange indicates that the client will provide an encrypted master key in
	// the "Session Key" field of the Type 3 message.
	ntlmsspNegotiateKeyExchange negotiateFlag = 0x40000000
)

// GenerateNegotiateMessage generates a NTLMv2 negotiation message (Type 1 message).
//
// See: https://curl.se/rfc/ntlm.html#theType1Message
func (s *NTLMv2Session) GenerateNegotiateMessage() (*NegotiateMessage, error) {
	domainPayload, err := createStringPayload(s.domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain payload: %w", err)
	}
	message := &NegotiateMessage{
		signature:   []byte(signature),
		messageType: uint32(messageTypeNegotiate),
		negotiateFlags: negotiateFlagset(ntlmsspNegotiateUnicode | ntlmsspNegotiateOEM |
			ntlmsspRequestTarget | ntlmsspNegotiateNTLM | ntlmsspNegotiateDomainSupplied |
			ntlmsspNegotiate128Bit | ntlmsspNegotiateExtendedSessionSecurity),
		domainname:  domainPayload,
		workstation: new(Payload),
	}
	s.negotiateMessage = message

	return message, nil
}

// readNegotiateFlagset reads the negotiate flags from the given byte slice and returns them as
// NegotiateFlags type.
func readNegotiateFlagset(flags []byte) negotiateFlagset {
	return negotiateFlagset(binary.LittleEndian.Uint32(flags))
}

// bytes returns the byte representation of the negotiate flags.
func (f negotiateFlagset) bytes() []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(f))
	return b
}

// Bytes returns the byte representation of the NTLMv2 negotiate message (Type 1 message).
func (nm *NegotiateMessage) Bytes() ([]byte, error) {
	const headerLen = 32

	payloadLen := int(nm.domainname.len) + int(nm.workstation.len)
	buffer := bytes.NewBuffer(make([]byte, 0, headerLen+payloadLen))

	buffer.Write(nm.signature)
	if err := binary.Write(buffer, binary.LittleEndian, nm.messageType); err != nil {
		return nil, fmt.Errorf("failed to write message type: %w", err)
	}
	buffer.Write(nm.negotiateFlags.bytes())

	payloadOffset := uint16(headerLen)

	nm.domainname.offset = uint32(payloadOffset)
	payloadOffset += nm.domainname.len
	buffer.Write(nm.domainname.bytes())

	nm.workstation.offset = uint32(payloadOffset)
	buffer.Write(nm.workstation.bytes())

	buffer.Write(nm.domainname.payload)
	buffer.Write(nm.workstation.payload)

	return buffer.Bytes(), nil
}
