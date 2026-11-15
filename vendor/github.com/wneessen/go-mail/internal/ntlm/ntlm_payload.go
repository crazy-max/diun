// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Payload represents a NTLM payload
type Payload struct {
	encoding int
	len      uint16
	maxLen   uint16
	offset   uint32
	payload  []byte
}

const (
	// payloadEncodingUnicode represents the Unicode payload encoding
	payloadEncodingUnicode = iota

	// payloadEncodingOEM represents the OEM payload encoding
	payloadEncodingOEM

	// payloadEncodingByte represents the byte payload encoding
	payloadEncodingByte
)

// ErrNTLMInvalidPayload is returned when the payload is invalid
var ErrNTLMInvalidPayload = errors.New("invalid NTLM payload")

// createBytePayload creates a Payload from the given byte slice.
func createBytePayload(payload []byte) (*Payload, error) {
	payloadLength, err := toUint16(len(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to convert payload length: %w", err)
	}
	return &Payload{
		encoding: payloadEncodingByte,
		len:      payloadLength,
		maxLen:   payloadLength,
		payload:  payload,
	}, nil
}

// createStringPayload creates a Payload from the given string.
func createStringPayload(payload string) (*Payload, error) {
	utf16Payload := utf16FromString(payload)
	payloadLength, err := toUint16(len(utf16Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to convert payload length: %w", err)
	}
	return &Payload{
		encoding: payloadEncodingUnicode,
		len:      payloadLength,
		maxLen:   payloadLength,
		payload:  utf16Payload,
	}, nil
}

// readPayload reads a payload from the given byte slice starting at startByte of type payloadType.
func readPayload(startByte int, payload []byte, payloadType int) (*Payload, error) {
	if startByte < 0 || len(payload) < startByte+8 {
		return nil, ErrNTLMInvalidPayload
	}

	p := &Payload{
		encoding: payloadType,
		len:      binary.LittleEndian.Uint16(payload[startByte : startByte+2]),
		maxLen:   binary.LittleEndian.Uint16(payload[startByte+2 : startByte+4]),
		offset:   binary.LittleEndian.Uint32(payload[startByte+4 : startByte+8]),
	}

	if p.len > 0 {
		end := uint64(p.offset) + uint64(p.len)
		if uint64(p.offset) > uint64(len(payload)) || end > uint64(len(payload)) {
			return nil, ErrNTLMInvalidPayload
		}
		p.payload = payload[p.offset:end]
	}
	return p, nil
}

// bytes returns the payload as a byte slice.
func (p *Payload) bytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint16(b[0:2], p.len)
	binary.LittleEndian.PutUint16(b[2:4], p.maxLen)
	binary.LittleEndian.PutUint32(b[4:8], p.offset)
	return b
}
