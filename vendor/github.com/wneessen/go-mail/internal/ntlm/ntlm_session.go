// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
	"slices"
	"time"
)

// NTLMv2Session represents a session using the NTLMv2 protocol.
type NTLMv2Session struct {
	user     string
	password string
	domain   string

	negotiateMessage          *NegotiateMessage
	challengeMessage          *ChallengeMessage
	clientChallenge           []byte
	responseKeyNT             []byte
	ntChallengeResponse       []byte
	lmChallengeResponse       []byte
	sessionBaseKey            []byte
	exportedSessionKey        []byte
	encryptedRandomSessionKey []byte
}

// NewNTLMv2Session returns a new NTLMv2Session.
func NewNTLMv2Session() *NTLMv2Session {
	return &NTLMv2Session{}
}

// SetUserInfo sets the user information for the NTLMv2Session.
func (s *NTLMv2Session) SetUserInfo(username, password, domain string) {
	s.user = username
	s.password = password
	s.domain = domain
}

// ParseChallengeMessage parses an NTLM challenge message (Type 2 message) from
// the given body
//
// See: https://curl.se/rfc/ntlm.html#theType2Message
func (s *NTLMv2Session) ParseChallengeMessage(body []byte) error {
	if len(body) < 48 {
		return ErrNTLMInvalidChallengeMessage
	}

	var err error
	challenge := new(ChallengeMessage)

	// Read NTLMSSP signature
	challenge.signature = body[0:8]
	if !bytes.Equal(challenge.signature, []byte("NTLMSSP\x00")) {
		return ErrNTLMInvalidSignature
	}

	// We expect a Type 2 message
	challenge.messageType = binary.LittleEndian.Uint32(body[8:12])
	if challenge.messageType != uint32(messageTypeChallenge) {
		return ErrNTLMInvalidMessageType
	}

	// Read the negotate flags
	challenge.negotiateFlags = readNegotiateFlagset(body[20:24])

	// Decide on the encoding, bases on the provided flags
	//
	// MS-NLMP: Unicode takes precedence, OEM is only used when UNICODE is not negotiated
	encoding := payloadEncodingOEM
	if uint32(challenge.negotiateFlags)&uint32(ntlmsspNegotiateUnicode) != 0 {
		encoding = payloadEncodingUnicode
	}

	// Read the target name (we don't use it, but it could be useful for debugging purposes)
	challenge.targetName, err = readPayload(12, body, encoding)
	if err != nil {
		return fmt.Errorf("failed to read body payload: %w", err)
	}

	// Read the server challenge
	challenge.serverChallenge = body[24:32]

	// 8 reserved bytes
	challenge.reserved = body[32:40]

	// Extract the target info (Attribute-Value pairs)
	targetInfo, err := readPayload(40, body, payloadEncodingByte)
	if err != nil {
		return fmt.Errorf("failed to read body payload: %w", err)
	}
	if challenge.targetInfo, err = readAVPairs(targetInfo.payload); err != nil {
		return fmt.Errorf("failed to read Attribute-Value pairs: %w", err)
	}

	// Read the version (we discard that information, though)
	offset := 48
	if uint32(challenge.negotiateFlags)&uint32(ntlmsspNegotiateVersion) != 0 {
		if len(body) < offset+8 {
			return ErrNTLMInvalidChallengeMessage
		}
		offset += 8
	}

	// Read the payload
	challenge.payload = body[offset:]

	// Fill the session with the challenge message and required data
	s.challengeMessage = challenge
	s.clientChallenge = randomBytes(8)
	s.responseKeyNT = ntlmv2Hash(s.user, s.password, s.domain)

	// Compute the expected response
	timestamp := timeToWindowsFileTime(time.Now())
	s.computeExpectedResponses(timestamp, challenge.targetInfo)

	return s.computeEncryptedSessionKey()
}

// GenerateAuthenticateMessage generates an NTLMv2 AuthenticateMessage (Type 3 message).
//
// See: https://curl.se/rfc/ntlm.html#theType3Message
func (s *NTLMv2Session) GenerateAuthenticateMessage() (*AuthenticateMessage, error) {
	lmChallangeResp, err := createBytePayload(s.lmChallengeResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create LM challenge response payload: %w", err)
	}
	ntChallangeResp, err := createBytePayload(s.ntChallengeResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create NTLM challenge response payload: %w", err)
	}
	encryptedRandomSessionKey, err := createBytePayload(s.encryptedRandomSessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypted random session key payload: %w", err)
	}
	domainPayload, err := createStringPayload(s.domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain payload: %w", err)
	}
	usernamePayload, err := createStringPayload(s.user)
	if err != nil {
		return nil, fmt.Errorf("failed to create username payload: %w", err)
	}
	workstationPayload, _ := createStringPayload("")
	return &AuthenticateMessage{
		signature:                 []byte("NTLMSSP\x00"),
		messageType:               3,
		lmChallengeResponse:       lmChallangeResp,
		ntChallengeResponseFields: ntChallangeResp,
		domainname:                domainPayload,
		username:                  usernamePayload,
		workstation:               workstationPayload,
		encryptedRandomSessionKey: encryptedRandomSessionKey,
		negotiateFlags:            s.challengeMessage.negotiateFlags,
	}, nil
}

// computeExpectedResponses computes the expected NTLMv2 challenge responses (LMv2 and NTLMv2).
func (s *NTLMv2Session) computeExpectedResponses(timestamp []byte, avPairs *avPairs) {
	// temp = RespType || HiRespType || ZeroByte(6) || Timestamp || ClientChallenge ||
	// ZeroByte(4) || TargetInfo || ZeroByte(4)
	//
	// See: https://curl.se/rfc/ntlm.html#theNtlmv2Response
	temp := slices.Concat(
		[]byte{0x01, 0x01, 0, 0, 0, 0, 0, 0},
		timestamp,
		s.clientChallenge,
		make([]byte, 4),
		avPairs.bytes(),
		make([]byte, 4),
	)

	// NTLMv2: both the NTLMv2 and LMv2 responses are keyed by the same
	// 16-byte "NTLMv2 hash", so one keyed HMAC-MD5 instance suffices.
	//
	// See: https://curl.se/rfc/ntlm.html#theNtlmv2Response
	mac := hmac.New(md5.New, s.responseKeyNT)
	mac.Write(s.challengeMessage.serverChallenge)
	mac.Write(temp)
	ntProofStr := mac.Sum(nil)

	s.ntChallengeResponse = slices.Concat(ntProofStr, temp)

	mac.Reset()
	mac.Write(s.challengeMessage.serverChallenge)
	mac.Write(s.clientChallenge)
	s.lmChallengeResponse = slices.Concat(mac.Sum(nil), s.clientChallenge)

	mac.Reset()
	mac.Write(ntProofStr)
	s.sessionBaseKey = mac.Sum(nil)

	if avPairs.find(msvAVTimestamp) != nil {
		s.lmChallengeResponse = make([]byte, 24)
	}
}

// computeEncryptedSessionKey computes the encrypted session key for NTLMv2 key exchange.
func (s *NTLMv2Session) computeEncryptedSessionKey() error {
	if uint32(s.challengeMessage.negotiateFlags)&uint32(ntlmsspNegotiateKeyExchange) == 0 {
		s.exportedSessionKey = s.sessionBaseKey
		s.encryptedRandomSessionKey = nil
		return nil
	}

	s.exportedSessionKey = randomBytes(16)
	cipher, err := rc4.NewCipher(s.sessionBaseKey)
	if err != nil {
		return fmt.Errorf("failed to create RC4 cipher: %w", err)
	}

	encrypted := make([]byte, len(s.exportedSessionKey))
	cipher.XORKeyStream(encrypted, s.exportedSessionKey)
	s.encryptedRandomSessionKey = encrypted

	return nil
}
