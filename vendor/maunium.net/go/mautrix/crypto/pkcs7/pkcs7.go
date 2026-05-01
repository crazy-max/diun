// Copyright (c) 2024 Sumner Evans
// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package pkcs7

import (
	"errors"
	"fmt"
)

// Pad implements PKCS#7 padding as defined in [RFC2315]. It pads the data to
// the given blockSize in the range [1, 255]. This is normally used in AES-CBC
// encryption.
//
// [RFC2315]: https://www.ietf.org/rfc/rfc2315.txt
func Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	output := make([]byte, len(data)+padding)
	copy(output, data)
	for i := len(data); i < len(output); i++ {
		output[i] = byte(padding)
	}
	return output
}

var (
	ErrEmptyData      = errors.New("pkcs7: empty data")
	ErrInvalidPadding = errors.New("pkcs7: invalid padding")
)

// Unpad implements PKCS#7 unpadding as defined in [RFC2315]. It unpads the
// data by reading the padding amount from the last byte of the data. This is
// normally used in AES-CBC decryption.
//
// [RFC2315]: https://www.ietf.org/rfc/rfc2315.txt
func Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, ErrEmptyData
	}
	unpadding := data[length-1]
	if unpadding == 0 || int(unpadding) > length {
		return nil, fmt.Errorf("%w: length %d", ErrInvalidPadding, unpadding)
	}
	for _, b := range data[length-int(unpadding) : length-1] {
		if b != unpadding {
			return nil, fmt.Errorf("%w: got byte %d (expected only %d)", ErrInvalidPadding, b, unpadding)
		}
	}
	return data[:length-int(unpadding)], nil
}
