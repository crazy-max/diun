// Copyright 2025 Snowflake Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package floe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
)

type FloeHash uint8
type FloeAead uint8

const (
	SHA384            FloeHash = 0
	AesGcm256         FloeAead = 0
	dekPurpose        string   = "DEK:"
	headerTagPurpose  string   = "HEADER_TAG:"
	messageKeyPurpose string   = "MESSAGE_KEY:"
)

func (h FloeHash) Length() int {
	switch h {
	case SHA384:
		return 48
	default:
		panic("Unsupported hash function")
	}
}

type FloeParams struct {
	Hash                 FloeHash
	Aead                 FloeAead
	IvLen                uint32
	EncSegLen            uint32
	overrideMaskRotation *uint64
}

func (p FloeParams) maskSegmentNumber(segmentNumber uint64) uint64 {
	actualMask := p.Aead.RotationMask()
	if p.overrideMaskRotation != nil {
		actualMask = *p.overrideMaskRotation
	}
	actualMask = (1 << actualMask) - 1
	actualMask = ^actualMask
	return segmentNumber & actualMask
}

func GCM256_IV256_4K() FloeParams {
	return FloeParams{
		Hash:      SHA384,
		Aead:      AesGcm256,
		IvLen:     32,
		EncSegLen: 4 * 1024,
	}
}

func GCM256_IV256_1M() FloeParams {
	return FloeParams{
		Hash:      SHA384,
		Aead:      AesGcm256,
		IvLen:     32,
		EncSegLen: 1024 * 1024,
	}
}

func (a FloeAead) goAead(key []byte) (*cipher.AEAD, error) {
	if len(key) != a.KeyLen() {
		return nil, errors.New("invalid key length")
	}
	// We only support AES-GCM for now
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// We only support 12 byte IVs right now
	if a.IvLen() != 12 {
		return nil, errors.New("only 12 byte IVs are supported")
	}
	// This implicitly prepends a 12 byte IV
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}
	return &aead, nil
}

func (a FloeAead) KeyLen() int {
	switch a {
	case AesGcm256:
		return 32
	default:
		panic("Unsupported AEAD")
	}
}

func (a FloeAead) RotationMask() uint64 {
	switch a {
	case AesGcm256:
		return 20
	default:
		panic("Unsupported AEAD")
	}
}

func (a FloeAead) IvLen() int {
	switch a {
	case AesGcm256:
		return 12
	default:
		panic("Unsupported AEAD")
	}
}

func (a FloeAead) TagLen() int {
	switch a {
	case AesGcm256:
		return 16
	default:
		panic("Unsupported AEAD")
	}
}

func (p FloeParams) kdf(baseKey, iv, aad []byte, purpose string, length int) ([]byte, error) {
	if len(iv) != int(p.IvLen) {
		return nil, errors.New("invalid IV length")
	}
	hkdfInfo := p.encoded()

	var hash func() hash.Hash
	switch p.Hash {
	case SHA384:
		hash = sha512.New384
	default:
		return nil, errors.New("unsupported hash function")
	}
	if len(hkdfInfo) != 10 {
		return nil, errors.New("invalid HKDF info length")
	}
	hkdfInfo = append(hkdfInfo, iv...)
	hkdfInfo = append(hkdfInfo, purpose...)
	hkdfInfo = append(hkdfInfo, aad...)
	result, err := hkdf.Expand(hash, baseKey, string(hkdfInfo), length)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p FloeParams) deriveKey(key, iv, aad []byte, segmentNumber uint64) ([]byte, error) {
	maskedSegementNumber := p.maskSegmentNumber(segmentNumber)
	// fmt.Printf("Masked segment number: %d\n", maskedSegementNumber)
	info := make([]byte, 0)
	info = append(info, dekPurpose...)
	info = binary.BigEndian.AppendUint64(info, maskedSegementNumber)
	// fmt.Printf("Key: %s, iv: %s, aad: %s, Info: %s\n",
	// 	hex.EncodeToString(key), hex.EncodeToString(iv), hex.EncodeToString(aad),
	// 	hex.EncodeToString(info))
	result, err := p.kdf(key, iv, aad, string(info), p.Aead.KeyLen())
	// fmt.Printf("Derived key: %s\n", hex.EncodeToString(result))
	return result, err
}

func (p FloeParams) PtSegLen() int {
	return int(p.EncSegLen) - p.SegOverhead()
}

func (p FloeParams) SegOverhead() int {
	return p.Aead.IvLen() + p.Aead.TagLen() + 4
}

func (p FloeParams) HeaderLen() int {
	return int(p.IvLen) + 10 + 32
}

func (p FloeParams) encoded() []byte {
	result := make([]byte, 0)
	result = append(result, byte(p.Aead))
	result = append(result, byte(p.Hash))
	result = binary.BigEndian.AppendUint32(result, p.EncSegLen)

	result = binary.BigEndian.AppendUint32(result, p.IvLen)

	return result
}

type FloeEncryptor struct {
	Header  []byte
	Params  FloeParams
	baseKey []byte
	floeIv  []byte
	aad     []byte
	counter uint64
	closed  bool
}

type FloeDecryptor struct {
	Params  FloeParams
	baseKey []byte
	floeIv  []byte
	aad     []byte
	counter uint64
	closed  bool
}

func (d FloeDecryptor) IsDone() bool {
	return d.closed
}

func NewFloeEncryptor(params FloeParams, key, aad []byte) (*FloeEncryptor, error) {
	if len(key) != params.Aead.KeyLen() {
		return nil, fmt.Errorf("key has incorrect length of %d rather than %d", len(key), params.Aead.KeyLen())
	}
	// 	iv = RND(FLOE_IV_LEN)
	iv := make([]byte, params.IvLen)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}

	// HeaderPrefix = PARAM_ENCODE(params) || iv
	headerPrefix := params.encoded()
	headerPrefix = append(headerPrefix, iv...)
	// HeaderTag = FLOE_KDF(key, iv, aad, “HEADER_TAG:”)
	headerTag, err := params.kdf(key, iv, aad, headerTagPurpose, 32)
	if err != nil {
		return nil, err
	}
	// MessageKey = FLOE_KDF(key, iv, aad, "MESSAGE_KEY:", 32)
	messageKey, err := params.kdf(key, iv, aad, messageKeyPurpose, params.Hash.Length())
	if err != nil {
		return nil, err
	}
	// Header = HeaderPrefix || iv || HeaderTag
	header := headerPrefix
	header = append(header, headerTag...)

	// State = {AeadKey, Counter = 0}
	// return (State, Header)
	return &FloeEncryptor{
		baseKey: messageKey,
		floeIv:  iv,
		aad:     aad,
		counter: 0,
		Header:  header,
		Params:  params,
	}, nil
}

func (e *FloeEncryptor) EncryptSegment(plaintext []byte) ([]byte, error) {
	if e.closed {
		return nil, errors.New("encryptor closed")
	}
	// assert len(plaintext) == ENC_SEG_LEN - AEAD_IV_LEN - AEAD_TAG_LEN
	if len(plaintext) != e.Params.PtSegLen() {
		return nil, errors.New("invalid plaintext length")
	}
	// assert State.Counter != 2^40-1 # Prevent overflow
	if e.counter == (1<<40)-1 {
		return nil, errors.New("counter overflow")
	}

	// aead_key = DERIVE_KEY(state.key, state.iv, state.aad, State.Counter)// AEAD key
	aead_key, err := e.Params.deriveKey(e.baseKey, e.floeIv, e.aad, e.counter)
	if err != nil {
		return nil, err
	}
	aeadPtr, err := e.Params.Aead.goAead(aead_key)
	if err != nil {
		return nil, err
	}
	aead := *aeadPtr

	// aead_iv = RND(AEAD_IV_LEN)
	// The IV is automatically generated and prepended now

	// aead_aad = I2BE(State.Counter) || 0x00
	aead_aad := make([]byte, 0)
	aead_aad, err = binary.Append(aead_aad, binary.BigEndian, e.counter)
	if err != nil {
		return nil, err
	}
	aead_aad = append(aead_aad, 0) // Not final segment

	// (aead_ciphertext, tag) = AEAD_ENC(State.AeadKey, aead_iv, plaintext, aead_aad)
	encrypted_segment := make([]byte, e.Params.EncSegLen)

	// IV is nil because NewGCMWithRandomNonce generates an IV internally and prepends it
	aead.Seal(encrypted_segment[:4], nil, plaintext, aead_aad)
	// EncryptedSegment = 0xFFFFFFFF || aead_iv || aead_ciphertext || tag
	binary.BigEndian.PutUint32(encrypted_segment, ^uint32(0))
	// State.Counter++
	e.counter++
	// return (State, EncryptedSegment)
	return encrypted_segment, nil
}

func (e *FloeEncryptor) EncryptLastSegment(plaintext []byte) ([]byte, error) {
	if e.closed {
		return nil, errors.New("encryptor closed")
	}
	// assert len(plaintext) >= 0
	//if len(plaintext) < 0 {
	//	return nil, errors.New("Invalid plaintext length")
	//}
	// assert len(plaintext) <= ENC_SEG_LEN - AEAD_IV_LEN - AEAD_TAG_LEN - 4
	if len(plaintext) > e.Params.PtSegLen() {
		return nil, errors.New("invalid plaintext length")
	}
	// aead_key = DERIVE_KEY(state.key, state.iv, state.aad, State.Counter)// AEAD key
	aead_key, err := e.Params.deriveKey(e.baseKey, e.floeIv, e.aad, e.counter)
	if err != nil {
		return nil, err
	}
	aeadPtr, err := e.Params.Aead.goAead(aead_key)
	if err != nil {
		return nil, err
	}
	aead := *aeadPtr

	// aead_iv = RND(AEAD_IV_LEN)
	// IV is implicitly generated and prepended

	// aead_aad = I2BE(State.Counter) || 0x01
	aead_aad := make([]byte, 0)
	aead_aad, err = binary.Append(aead_aad, binary.BigEndian, e.counter)
	if err != nil {
		return nil, err
	}
	aead_aad = append(aead_aad, 1) // Final segment

	// (aead_ciphertext, tag) = AEAD_ENC(State.AeadKey, aead_iv, plaintext, aead_aad)
	encrypted_segment := make([]byte, 4+len(plaintext)+e.Params.Aead.IvLen()+e.Params.Aead.TagLen())
	// IV is nil because NewGCMWithRandomNonce generates an IV internally and prepends it
	aead.Seal(encrypted_segment[:4], nil, plaintext, aead_aad)
	// EncryptedSegment = 0xFFFFFFFF || aead_iv || aead_ciphertext || tag
	binary.BigEndian.PutUint32(encrypted_segment, uint32(len(encrypted_segment)))

	e.closed = true
	// return (State, EncryptedSegment)
	return encrypted_segment, nil
}

func NewFloeDecryptor(params FloeParams, key, aad, header []byte) (*FloeDecryptor, error) {
	if len(key) != params.Aead.KeyLen() {
		return nil, fmt.Errorf("key has incorrect length of %d rather than %d", len(key), params.Aead.KeyLen())
	}
	// EncodedParams = PARAM_ENCODE(params)
	encodedParams := params.encoded()
	// assert len(header) == FLOE_IV_LEN + len(EncodedParams) + 32
	if len(header) != params.HeaderLen() {
		return nil, errors.New("invalid header length")
	}
	// (HeaderParams, iv, HeaderTag) = SPLIT(header, len(EncodedParams), 32)
	headerParams := header[:len(encodedParams)]
	iv := header[len(encodedParams) : len(encodedParams)+int(params.IvLen)]
	headerTag := header[len(encodedParams)+int(params.IvLen):]

	// assert HeaderParams == EncodedParams
	if string(headerParams) != string(encodedParams) {
		return nil, errors.New("Invalid header parameters: " + hex.EncodeToString(encodedParams) + " != " + hex.EncodeToString(headerParams))
	}

	// ExpectedHeaderTag = FLOE_KDF(key, iv, aad, “HEADER_TAG:”)
	expectedHeaderTag, err := params.kdf(key, iv, aad, headerTagPurpose, 32)
	if err != nil {
		return nil, err
	}
	// assert ExpectedHeaderTag == HeaderTag // Must be constant time
	if subtle.ConstantTimeCompare(expectedHeaderTag, headerTag) != 1 {
		return nil, errors.New("invalid header tag")
	}

	// MessageKey = FLOE_KDF(key, iv, aad, "MESSAGE_KEY:", 32)
	messageKey, err := params.kdf(key, iv, aad, messageKeyPurpose, params.Hash.Length())
	if err != nil {
		return nil, err
	}

	// State = {AeadKey, Counter = 0}
	// return State
	return &FloeDecryptor{
		baseKey: messageKey,
		floeIv:  iv,
		aad:     aad,
		counter: 0,
		Params:  params,
	}, nil
}

func (d *FloeDecryptor) DecryptSegment(encrypted_segment []byte) ([]byte, error) {
	if d.closed {
		return nil, errors.New("decryptor closed")
	}
	// assert len(EncryptedSegment) == ENC_SEG_LEN
	if len(encrypted_segment) != int(d.Params.EncSegLen) {
		return nil, errors.New("invalid encrypted segment length")
	}
	// assert BE2I(EncryptedSegment[:4]) == 0xFFFFFFFF
	if binary.BigEndian.Uint32(encrypted_segment[:4]) != ^uint32(0) {
		return nil, errors.New("unexpected last segment")
	}
	// assert State.Counter != 2^40-1 # Prevent overflow
	if d.counter == (1<<40)-1 {
		return nil, errors.New("counter overflow")
	}
	// aead_key = DERIVE_KEY(state.key, state.iv, state.aad, State.Counter)// AEAD key
	aead_key, err := d.Params.deriveKey(d.baseKey, d.floeIv, d.aad, d.counter)
	if err != nil {
		return nil, err
	}
	aeadPtr, err := d.Params.Aead.goAead(aead_key)
	if err != nil {
		return nil, err
	}
	aead := *aeadPtr

	// (aead_iv, aead_ciphertext, tag) = SPLIT(EncryptedSegment[4:], AEAD_IV_LEN, AEAD_TAG_LEN)
	// IV is implicit
	aead_ciphertext := encrypted_segment[4:]

	// aead_aad = I2BE(State.Counter) || 0x00
	aead_aad := make([]byte, 0)
	aead_aad, err = binary.Append(aead_aad, binary.BigEndian, d.counter)
	if err != nil {
		return nil, err
	}
	aead_aad = append(aead_aad, 0) // Not final segment

	// Plaintext = AEAD_DEC(State.AeadKey, aead_iv, aead_ciphertext, aead_aad)
	plaintext := make([]byte, d.Params.PtSegLen())
	// IV is nil because NewGCMWithRandomNonce reads the IV from the start of the ciphertext
	_, err = aead.Open(plaintext[:0], nil, aead_ciphertext, aead_aad)

	// assert Plaintext != FAIL
	if err != nil {
		return nil, err
	}
	// State.Counter++
	d.counter++
	// return (State, Plaintext)
	return plaintext, nil
}

func (d *FloeDecryptor) DecryptLastSegment(encrypted_segment []byte) ([]byte, error) {
	if d.closed {
		return nil, errors.New("decryptor closed")
	}
	// assert len(EncryptedSegment) >= AEAD_IV_LEN + AEAD_TAG_LEN + 4
	if len(encrypted_segment) < int(d.Params.Aead.IvLen())+d.Params.Aead.TagLen()+4 {
		return nil, fmt.Errorf("invalid encrypted segment length %d", len(encrypted_segment))
	}
	// assert BE2I(EncryptedSegment[:4]) == len(EncryptedSegment)
	if binary.BigEndian.Uint32(encrypted_segment[:4]) != uint32(len(encrypted_segment)) {
		return nil, fmt.Errorf("invalid last segment length: %d", binary.BigEndian.Uint32(encrypted_segment[:4]))
	}
	// assert len(EncryptedSegment) <= ENC_SEG_LEN
	if len(encrypted_segment) > int(d.Params.EncSegLen) {
		return nil, errors.New("invalid encrypted segment length 2")
	}
	aead_key, err := d.Params.deriveKey(d.baseKey, d.floeIv, d.aad, d.counter)
	if err != nil {
		return nil, err
	}
	aeadPtr, err := d.Params.Aead.goAead(aead_key)
	if err != nil {
		return nil, err
	}
	aead := *aeadPtr

	// (aead_iv, aead_ciphertext, tag) = SPLIT(EncryptedSegment[4:], AEAD_IV_LEN, AEAD_TAG_LEN)
	// IV is implicit
	aead_ciphertext := encrypted_segment[4:]

	// aead_aad = I2BE(State.Counter) || 0x01
	aead_aad := make([]byte, 0)
	aead_aad, err = binary.Append(aead_aad, binary.BigEndian, d.counter)
	if err != nil {
		return nil, err
	}
	aead_aad = append(aead_aad, 1) // Final segment

	// Plaintext = AEAD_DEC(State.AeadKey, aead_iv, aead_ciphertext, aead_aad)
	plaintext := make([]byte, len(encrypted_segment)-d.Params.SegOverhead())
	// IV is nil because NewGCMWithRandomNonce reads the IV from the start of the ciphertext
	_, err = aead.Open(plaintext[:0], nil, aead_ciphertext, aead_aad)

	// assert Plaintext != FAIL
	if err != nil {
		return nil, err
	}
	d.closed = true
	// return Plaintext
	return plaintext, nil

}

// mautrix-go additions below

func (d *FloeDecryptor) ToEncryptor() *FloeEncryptor {
	return &FloeEncryptor{
		Params:  d.Params,
		baseKey: d.baseKey,
		floeIv:  d.floeIv,
		aad:     d.aad,
	}
}

func (d *FloeDecryptor) Seek(counter uint64) {
	d.counter = counter
	d.closed = false
}

func (e *FloeEncryptor) ToDecryptor() *FloeDecryptor {
	return &FloeDecryptor{
		Params:  e.Params,
		baseKey: e.baseKey,
		floeIv:  e.floeIv,
		aad:     e.aad,
	}
}

func (e *FloeEncryptor) Seek(counter uint64) {
	e.counter = counter
	e.closed = false
}
