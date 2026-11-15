// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"bytes"
	"encoding/binary"
)

// AuthenticateMessage represents the Type 3 message in the NTLM authentication process.
type AuthenticateMessage struct {
	signature   []byte
	messageType uint32

	negotiateFlags negotiateFlagset

	lmChallengeResponse       *Payload
	ntChallengeResponseFields *Payload
	domainname                *Payload
	username                  *Payload
	workstation               *Payload
	encryptedRandomSessionKey *Payload
}

// Bytes returns the byte representation of the AuthenticateMessage type (Type 3 message)
//
// See: https://curl.se/rfc/ntlm.html#theType3Message
func (a *AuthenticateMessage) Bytes() []byte {
	// sig + type + 6 descriptors + flags + version
	//
	// See: https://curl.se/rfc/ntlm.html#theType3Message
	const headerLen = 8 + 4 + 6*8 + 4 + 8

	payloadLen := int(a.domainname.len) + int(a.username.len) + int(a.workstation.len) +
		int(a.lmChallengeResponse.len) + int(a.ntChallengeResponseFields.len) +
		int(a.encryptedRandomSessionKey.len)

	offset := uint32(headerLen)
	for _, p := range []*Payload{
		a.domainname, a.username, a.workstation, a.lmChallengeResponse,
		a.ntChallengeResponseFields, a.encryptedRandomSessionKey,
	} {
		p.offset = offset
		offset += uint32(p.len)
	}

	buf := bytes.NewBuffer(make([]byte, 0, headerLen+payloadLen))
	buf.Write(a.signature)

	var msgType [4]byte
	binary.LittleEndian.PutUint32(msgType[:], a.messageType)
	buf.Write(msgType[:])

	buf.Write(a.lmChallengeResponse.bytes())
	buf.Write(a.ntChallengeResponseFields.bytes())
	buf.Write(a.domainname.bytes())
	buf.Write(a.username.bytes())
	buf.Write(a.workstation.bytes())
	buf.Write(a.encryptedRandomSessionKey.bytes())
	buf.Write(a.negotiateFlags.bytes())

	// Even if NTLMSSP_NEGOTIATE_VERSION is set, we'll send a zero version byte, due
	// to improved privacy. The version field is purely cosmetic and only useful for
	// debugging purposes on the server end.
	buf.Write(make([]byte, 8))

	buf.Write(a.domainname.payload)
	buf.Write(a.username.payload)
	buf.Write(a.workstation.payload)
	buf.Write(a.lmChallengeResponse.payload)
	buf.Write(a.ntChallengeResponseFields.payload)
	buf.Write(a.encryptedRandomSessionKey.payload)

	return buf.Bytes()
}
