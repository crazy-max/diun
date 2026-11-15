// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

// Package ntlm provides a very basic implementation of the NTLMv2 protocol. It provides everything needed to
// authenticate a SMTP authentication request. This package is not feature complete and is intented to only
// provide the means required for go-mail to perform NTLM via SMTP authentication.
package ntlm

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf16"

	"golang.org/x/crypto/md4"
)

const (
	// windowsFileTime is the number of 100-nanosecond intervals since January 1, 1601 (UTC).
	// This magic number is used as base for the Windows FILETIME representation.
	windowsFileTime = 116444736000000000

	// signature is the NTLMSSP signature used in NTLMv2 messages.
	signature = "NTLMSSP\x00"
)

// ntlmv2Hash computes the NT One-Way Function (OWF) for the given user, password,
// and user domain. It only implements the OWF v2 as described in the NTLMv2 spec.
//
// See: https://curl.se/rfc/ntlm.html#appendixD
func ntlmv2Hash(user, passwd, userDom string) []byte {
	hasher := md4.New()
	hasher.Write(utf16FromString(passwd))
	mac := hmac.New(md5.New, hasher.Sum(nil))
	mac.Write(utf16FromString(strings.ToUpper(user) + userDom))
	return mac.Sum(nil)
}

// timeToWindowsFileTime converts a time.Time to a Windows FILETIME (little-endian) byte slice.
//
// See:      https://learn.microsoft.com/en-us/windows/win32/sysinfo/converting-a-time-t-value-to-a-file-time
// See also: https://curl.se/rfc/ntlm.html#appendixD
func timeToWindowsFileTime(t time.Time) []byte {
	fileTime := uint64(t.UnixNano()/100 + windowsFileTime)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, fileTime)
	return buf
}

// utf16FromString converts a string to a UTF-16 (little-endian) byte slice.
func utf16FromString(payload string) []byte {
	encoded := utf16.Encode([]rune(payload))
	result := make([]byte, len(encoded)*2)

	for i, u := range encoded {
		binary.LittleEndian.PutUint16(result[i*2:], u)
	}
	return result
}

// toUint16 converts an int to a uint16, returning an error if the value is out of range.
func toUint16(n int) (uint16, error) {
	if n < 0 || n > math.MaxUint16 {
		return 0, fmt.Errorf("value %d out of uint16 range", n)
	}
	return uint16(n), nil
}

// toUint32 converts an int to a uint16, returning an error if the value is out of range.
func toUint32(n int) (uint32, error) {
	if n < 0 || int64(n) > math.MaxUint32 {
		return 0, fmt.Errorf("value %d out of uint32 range", n)
	}
	return uint32(n), nil
}

// randomBytes returns a slice of random bytes of the given length.
func randomBytes(length int) []byte {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return b
}
