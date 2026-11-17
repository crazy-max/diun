// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package attachment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"

	"go.mau.fi/util/exerrors"
	"go.mau.fi/util/random"

	"maunium.net/go/mautrix/crypto/floe"
)

type FLOEFile EncryptedFile

var FLOEParams = floe.GCM256_IV256_4K()

func NewFLOEFile() *FLOEFile {
	key := random.Bytes(32)
	encryptor := exerrors.Must(floe.NewFloeEncryptor(FLOEParams, key, nil))
	return &FLOEFile{
		Version:    VersionBeeperFLOE,
		FLOEKey:    base64.RawStdEncoding.EncodeToString(key),
		FLOEHeader: base64.RawStdEncoding.EncodeToString(encryptor.Header),

		decoded: &decodedKeys{encryptor: encryptor, decryptor: encryptor.ToDecryptor()},
	}
}

func (ff *FLOEFile) ToStandard() *EncryptedFile {
	return (*EncryptedFile)(ff)
}

func (ff *FLOEFile) decodeKeys() error {
	if ff.Version != VersionBeeperFLOE {
		return fmt.Errorf("attachments.FLOEFile: %w: %s", ErrUnsupportedVersion, ff.Version)
	}
	if ff.decoded != nil {
		return nil
	}
	key, err := base64.RawStdEncoding.DecodeString(ff.FLOEKey)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}
	header, err := base64.RawStdEncoding.DecodeString(ff.FLOEHeader)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFLOEHeader, err)
	}
	decryptor, err := floe.NewFloeDecryptor(floe.GCM256_IV256_4K(), key, nil, header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFLOEInit, err)
	}
	encryptor := decryptor.ToEncryptor()
	encryptor.Header = header
	var shaHash [sha256.Size]byte
	if len(ff.Hashes.SHA256) != 0 {
		shaHashBytes, err := base64.RawStdEncoding.DecodeString(ff.Hashes.SHA256)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidHash, err)
		} else if len(shaHashBytes) != sha256.Size {
			return fmt.Errorf("%w: invalid length %d", ErrInvalidHash, len(shaHashBytes))
		}
		shaHash = [sha256.Size]byte(shaHashBytes)
	}
	ff.decoded = &decodedKeys{
		sha256:    shaHash,
		encryptor: encryptor,
		decryptor: decryptor,
	}
	return nil
}

func (ff *FLOEFile) makeStream(reader io.Reader, encrypt bool) *FLOEStream {
	var hasher hash.Hash
	if encrypt || ff.Hashes.SHA256 != "" {
		hasher = sha256.New()
	}
	return &FLOEStream{
		file:    ff,
		PrepErr: ff.decodeKeys(),
		hash:    hasher,
		input:   reader,
		encrypt: encrypt,
	}
}

func (ff *FLOEFile) Encrypt(data []byte) ([]byte, error) {
	chunkCount := int(math.Ceil(float64(len(data)) / float64(ff.decoded.encryptor.Params.PtSegLen())))
	buf := bytes.NewBuffer(make([]byte, 0, len(data)+chunkCount*ff.decoded.encryptor.Params.SegOverhead()))
	stream := ff.makeStream(bytes.NewReader(data), true)
	_, err := stream.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), stream.Close()
}

func (ff *FLOEFile) DecryptInPlace(data []byte) (output []byte, err error) {
	stream := ff.makeStream(nil, false)
	if stream.PrepErr != nil {
		return nil, stream.PrepErr
	}
	chunkSize := int(ff.decoded.decryptor.Params.EncSegLen)
	var writePtr int
	for i := 0; i < len(data); i += chunkSize {
		err = stream.doChunk(data[i:min(i+chunkSize, len(data))], i+chunkSize >= len(data))
		if err != nil {
			return
		}
		writePtr += copy(data[writePtr:], stream.pending)
	}
	clear(data[writePtr:])
	return data[:writePtr], stream.Close()
}

func (ff *FLOEFile) DecryptStream(reader io.Reader) *FLOEStream {
	return ff.makeStream(reader, false)
}

func (ff *FLOEFile) EncryptStream(reader io.Reader) *FLOEStream {
	return ff.makeStream(reader, true)
}

type FLOEStream struct {
	PrepErr error
	file    *FLOEFile
	hash    hash.Hash
	input   io.Reader
	buf     []byte
	pending []byte
	started bool
	done    bool
	encrypt bool
}

var (
	_ io.ReadSeekCloser = (*FLOEStream)(nil)
	_ io.WriterTo       = (*FLOEStream)(nil)
)

func (f *FLOEStream) start() error {
	if f.started {
		return nil
	}
	if f.PrepErr != nil {
		return f.PrepErr
	}
	bufLen := f.file.decoded.decryptor.Params.EncSegLen + 1
	if f.encrypt {
		bufLen -= uint32(f.file.decoded.encryptor.Params.SegOverhead())
	}
	f.buf = make([]byte, bufLen)
	f.started = true
	return nil
}

func (f *FLOEStream) nextChunk() error {
	if f.done {
		return fmt.Errorf("attachments.FLOEStream: nextChunk called after done")
	}
	startRead := 0
	if f.started {
		startRead = 1
		f.buf[0] = f.buf[len(f.buf)-1]
	} else if err := f.start(); err != nil {
		return err
	}
	n, err := io.ReadFull(f.input, f.buf[startRead:])
	isEOF := errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
	if err != nil && !isEOF {
		return err
	} else if errors.Is(err, io.EOF) && n == 0 && startRead == 0 {
		if f.encrypt {
			return f.doChunk([]byte{}, true)
		}
		// io.ReadFull returns plain EOF for empty input, but ErrUnexpectedEOF is more appropriate
		// if decryption input is empty.
		return io.ErrUnexpectedEOF
	}
	chunk := f.buf[:len(f.buf)-1]
	if isEOF {
		chunk = chunk[:n+startRead]
	}
	return f.doChunk(chunk, isEOF)
}

func (f *FLOEStream) doChunk(chunk []byte, isLast bool) (err error) {
	if f.done {
		return fmt.Errorf("attachments.FLOEStream: doChunk called after done")
	} else if err = f.start(); err != nil {
		return err
	}
	f.done = isLast
	if f.encrypt {
		if isLast {
			f.pending, err = f.file.decoded.encryptor.EncryptLastSegment(chunk)
		} else {
			f.pending, err = f.file.decoded.encryptor.EncryptSegment(chunk)
		}
		if f.hash != nil && err == nil {
			f.hash.Write(f.pending)
		}
	} else {
		if f.hash != nil {
			f.hash.Write(chunk)
		}
		if isLast {
			f.pending, err = f.file.decoded.decryptor.DecryptLastSegment(chunk)
		} else {
			f.pending, err = f.file.decoded.decryptor.DecryptSegment(chunk)
		}
	}
	return err
}

func (f *FLOEStream) Seek(offset int64, whence int) (n int64, err error) {
	if offset != 0 || whence != io.SeekStart {
		// TODO in non-hashing mode, this could actually allow somewhat arbitrary seeking
		return 0, fmt.Errorf("attachments.FLOEStream: only seeking to the beginning is supported")
	}
	if !f.started {
		return
	}
	seeker, ok := f.input.(io.ReadSeeker)
	if !ok {
		return 0, fmt.Errorf("attachments.FLOEStream: source reader (%T) is not an io.ReadSeeker", f.input)
	}
	n, err = seeker.Seek(offset, whence)
	if err != nil {
		return
	}
	if f.hash != nil {
		f.hash.Reset()
	}
	f.started = false
	f.done = false
	f.pending = nil
	clear(f.buf)
	f.file.decoded.encryptor.Seek(0)
	f.file.decoded.decryptor.Seek(0)
	return
}

func (f *FLOEStream) Read(p []byte) (n int, err error) {
	if len(f.pending) == 0 {
		if f.done {
			return 0, io.EOF
		}
		err = f.nextChunk()
		if err != nil {
			return
		}
	}
	n = copy(p, f.pending)
	f.pending = f.pending[n:]
	if len(f.pending) == 0 && f.done {
		err = io.EOF
	}
	return
}

func (f *FLOEStream) WriteTo(w io.Writer) (n int64, err error) {
	var m int
	if len(f.pending) > 0 {
		m, err = w.Write(f.pending)
		if err == nil && m < len(f.pending) {
			err = io.ErrShortWrite
		}
		if err != nil {
			return 0, err
		}
		n += int64(m)
	}
	for !f.done {
		err = f.nextChunk()
		if err != nil {
			return
		}
		m, err = w.Write(f.pending)
		if err == nil && m < len(f.pending) {
			err = io.ErrShortWrite
		}
		if err != nil {
			return
		}
		n += int64(m)
	}
	return
}

func (f *FLOEStream) Close() error {
	closer, ok := f.input.(io.Closer)
	if ok {
		err := closer.Close()
		if err != nil {
			return err
		}
	}
	if !f.started {
		return fmt.Errorf("attachments.FLOEStream: closed before started")
	} else if !f.done {
		return fmt.Errorf("attachments.FLOEStream: closed before completed")
	}
	if f.hash != nil {
		if f.encrypt {
			finalHash := f.hash.Sum(nil)
			f.file.decoded.sha256 = [32]byte(finalHash)
			f.file.Hashes.SHA256 = base64.RawStdEncoding.EncodeToString(finalHash)
		} else if !hmac.Equal(f.hash.Sum(nil), f.file.decoded.sha256[:]) {
			return ErrHashMismatch
		}
	}
	return nil
}
