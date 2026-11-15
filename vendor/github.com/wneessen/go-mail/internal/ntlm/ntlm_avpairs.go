// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package ntlm

import (
	"encoding/binary"
	"errors"
)

// avPairType represents the type of an Attribute-Value pair
type avPairType uint16

// avPair represents an Attribute-Value pair
type avPair struct {
	id    avPairType
	len   uint16
	value []byte
}

// avPairs represents a list of Attribute-Value pairs
type avPairs struct {
	list     []avPair
	reserved []byte
}

// Attribute-Value enumeration according to MSV1_0_AVID
//
// Only the subset used by this implementation is defined; values match the
// MSV1_0_AVID enumeration so they stay wire-compatible with the full list.
//
// See: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ne-ntifs-msv1_0_avid
const (
	// msvAVEOL indicates that this is the last AV_PAIR in the list
	msvAVEOL avPairType = 0x0000

	// msvAVNbComputerName represents the NetBIOS computer name
	msvAVNbComputerName avPairType = 0x0001

	// msvAVNbDomainName represents the NetBIOS domain name
	msvAVNbDomainName avPairType = 0x0002

	// msvAVDNSComputerName represents the DNS computer name
	msvAVDNSComputerName avPairType = 0x0003

	// msvAVDNSDomainName represents the DNS domain name
	msvAVDNSDomainName avPairType = 0x0004

	// msvAVTimestamp represents a FILETIME structure (64-bit) holding the
	// server's local time
	msvAVTimestamp avPairType = 0x0007
)

// ErrNTLMInvalidAVPair indicates that an invalid NTLM attribute-value pair was encountered
var ErrNTLMInvalidAVPair = errors.New("invalid NTLM attribute-value pair")

// AppendToAVPairs creates an AVPair with the given type and value and appends it to the list.
func (p *avPairs) appendAVPair(avType avPairType, value []byte) error {
	valLength, err := toUint16(len(value))
	if err != nil {
		return err
	}

	pair := new(avPair)
	pair.id = avType
	pair.len = valLength
	pair.value = value
	p.list = append(p.list, *pair)
	return nil
}

// readAVPair reads an avPair from the given byte slice at the given offset.
func readAVPair(data []byte, offset int) (*avPair, error) {
	if len(data) < offset+4 {
		return nil, ErrNTLMInvalidAVPair
	}
	pair := new(avPair)
	pair.id = avPairType(binary.LittleEndian.Uint16(data[offset : offset+2]))
	pair.len = binary.LittleEndian.Uint16(data[offset+2 : offset+4])
	if len(data) < offset+4+int(pair.len) {
		return nil, ErrNTLMInvalidAVPair
	}
	pair.value = data[offset+4 : offset+4+int(pair.len)]
	return pair, nil
}

// readAVPairs reads avPairs from the given byte slice.
func readAVPairs(data []byte) (*avPairs, error) {
	pairs := new(avPairs)
	offset := 0

	for {
		pair, err := readAVPair(data, offset)
		if err != nil {
			return nil, err
		}

		offset += 4 + int(pair.len)
		if offset > len(data) {
			return nil, ErrNTLMInvalidPayload
		}

		pairs.list = append(pairs.list, *pair)

		if pair.id == msvAVEOL {
			pairs.reserved = data[offset:]
			return pairs, nil
		}
	}
}

// bytes returns the avPairs as a byte slice.
func (p *avPairs) bytes() []byte {
	if p == nil {
		return nil
	}
	total := len(p.reserved)
	for i := range p.list {
		total += int(p.list[i].len) + 4
	}
	result := make([]byte, 0, total)
	for i := range p.list {
		result = append(result, p.list[i].bytes()...)
	}
	return append(result, p.reserved...)
}

// find returns the AVPair with the given type, or nil if not found.
func (p *avPairs) find(avType avPairType) *avPair {
	if p == nil {
		return nil
	}
	for i := range p.list {
		if p.list[i].id == avType {
			return &p.list[i]
		}
	}
	return nil
}

// bytes returns the avPair as a byte slice.
func (a *avPair) bytes() []byte {
	result := make([]byte, 4, int(a.len)+4)
	binary.LittleEndian.PutUint16(result[0:2], uint16(a.id))
	binary.LittleEndian.PutUint16(result[2:4], a.len)
	return append(result, a.value...)
}
