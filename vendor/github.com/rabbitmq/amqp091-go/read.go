// Copyright (c) 2021 VMware, Inc. or its affiliates. All Rights Reserved.
// Copyright (c) 2012-2021, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package amqp091

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"
)

/*
ReadFrame reads a frame from an input stream and returns an interface that can be cast into
one of the following:

	methodFrame
	PropertiesFrame
	bodyFrame
	heartbeatFrame

2.3.5  frame Details

All frames consist of a header (7 octets), a payload of arbitrary size, and a
'frame-end' octet that detects malformed frames:

	0      1         3             7                  size+7 size+8
	+------+---------+-------------+  +------------+  +-----------+
	| type | channel |     size    |  |  payload   |  | frame-end |
	+------+---------+-------------+  +------------+  +-----------+
	 octet   short         long         size octets       octet

To read a frame, we:
 1. Read the header and check the frame type and channel.
 2. Depending on the frame type, we read the payload and process it.
 3. Read the frame end octet.

In realistic implementations where performance is a concern, we would use
“read-ahead buffering” or

“gathering reads” to avoid doing three separate system calls to read a frame.
*/
func (r *reader) ReadFrame() (frame frame, err error) {
	var scratch [7]byte

	if _, err = io.ReadFull(r.r, scratch[:7]); err != nil {
		return
	}

	typ := scratch[0]
	channel := binary.BigEndian.Uint16(scratch[1:3])
	size := binary.BigEndian.Uint32(scratch[3:7])

	if r.maxFrameSize != nil {
		// max-frameHeaderSize is safe from underflow only because maxFrameSize,
		// whenever nonzero, is always negotiateFrameSize's floor (frameMinSize,
		// 4096) or higher — see the Store call at Connection.openTune.
		if max := r.maxFrameSize.Load(); max > 0 && size > (max-frameHeaderSize) {
			return nil, ErrFrameTooLarge
		}
	}

	switch typ {
	case frameMethod:
		if frame, err = r.parseMethodFrame(channel, size); err != nil {
			return
		}

	case frameHeader:
		if frame, err = r.parseHeaderFrame(channel, size); err != nil {
			return
		}

	case frameBody:
		if frame, err = r.parseBodyFrame(channel, size); err != nil {
			return nil, err
		}

	case frameHeartbeat:
		if frame, err = r.parseHeartbeatFrame(channel, size); err != nil {
			return
		}

	default:
		return nil, ErrFrame
	}

	if _, err = io.ReadFull(r.r, scratch[:1]); err != nil {
		return nil, err
	}

	if scratch[0] != frameEnd {
		return nil, ErrFrame
	}

	return
}

// readAllocChunk caps how far readBytes allocates ahead of data confirmed
// present on the wire, so a fabricated length field cannot force a large
// allocation on its own.
const readAllocChunk = 64 * 1024

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// readBytes reads exactly n bytes, growing the buffer in readAllocChunk
// increments so it never allocates far ahead of confirmed data.
func readBytes(r io.Reader, n int64) ([]byte, error) {
	if n < 0 {
		return nil, ErrSyntax
	}

	buf := make([]byte, 0, min64(n, readAllocChunk))
	for int64(len(buf)) < n {
		chunk := int(min64(n-int64(len(buf)), readAllocChunk))
		old := len(buf)

		// append grows the backing array with Go's amortized (roughly
		// doubling) strategy, so repeated chunks don't each re-copy the
		// entire buffer read so far.
		buf = append(buf, make([]byte, chunk)...)

		if _, err := io.ReadFull(r, buf[old:]); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func readShortstr(r io.Reader) (v string, err error) {
	var length uint8
	if err = binary.Read(r, binary.BigEndian, &length); err != nil {
		return
	}

	bytes, err := readBytes(r, int64(length))
	if err != nil {
		return
	}
	return string(bytes), nil
}

func readLongstr(r io.Reader) (v string, err error) {
	var length uint32
	if err = binary.Read(r, binary.BigEndian, &length); err != nil {
		return
	}

	// slices can't be longer than max int32 value
	if length > (^uint32(0) >> 1) {
		return "", ErrSyntax
	}

	bytes, err := readBytes(r, int64(length))
	if err != nil {
		return
	}
	return string(bytes), nil
}

func readDecimal(r io.Reader) (v Decimal, err error) {
	if err = binary.Read(r, binary.BigEndian, &v.Scale); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &v.Value); err != nil {
		return
	}
	return
}

func readTimestamp(r io.Reader) (v time.Time, err error) {
	var sec int64
	if err = binary.Read(r, binary.BigEndian, &sec); err != nil {
		return
	}
	return time.Unix(sec, 0), nil
}

/*
'A': any
'D': Decimal
'F': Table
'I': int32
'S': string
'T': time.Time
'V': nil
'b': int8
'B': byte
'd': float64
'f': float32
'l': int64
'i': uint32
's': int16
'u': uint16
't': bool
'x': []byte
*/
// maxFieldDepth limits how deeply 'A' and 'F' fields may nest, preventing
// unbounded recursion from overflowing the goroutine stack.
const maxFieldDepth = 32

// maxContainerElements limits how many elements/entries readArrayDepth and
// readTableDepth will accept in a single container. AMQP tables/arrays in
// practice hold at most a few hundred entries (message properties, queue
// arguments, policy definitions); this cap rejects the empty-value ('V')
// padding trick that would otherwise amplify wire bytes ~16-40x into heap via
// boxed `any` slice elements / Table map entries.
const maxContainerElements = 8192

func readField(r io.Reader) (v any, err error) {
	return readFieldDepth(r, 0)
}

func readFieldDepth(r io.Reader, depth int) (v any, err error) {
	var typ byte
	if err = binary.Read(r, binary.BigEndian, &typ); err != nil {
		return
	}

	switch typ {
	case 't':
		var value uint8
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value != 0, nil

	case 'B':
		var value [1]byte
		if _, err = io.ReadFull(r, value[0:1]); err != nil {
			return
		}
		return value[0], nil

	case 'b':
		var value int8
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 's':
		var value int16
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'I':
		var value int32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'l':
		var value int64
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'u':
		var value uint16
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'i':
		var value uint32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'f':
		var value float32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'd':
		var value float64
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'D':
		return readDecimal(r)

	case 'S':
		return readLongstr(r)

	case 'A':
		return readArrayDepth(r, depth+1)

	case 'T':
		return readTimestamp(r)

	case 'F':
		return readTableDepth(r, depth+1)

	case 'x':
		var len int32
		if err = binary.Read(r, binary.BigEndian, &len); err != nil {
			return nil, err
		}
		if len < 0 {
			return nil, ErrSyntax
		}

		value, err := readBytes(r, int64(len))
		if err != nil {
			return nil, err
		}
		return value, nil

	case 'V':
		return nil, nil
	}

	return nil, ErrSyntax
}

/*
Field tables are long strings that contain packed name-value pairs.  The
name-value pairs are encoded as short string defining the name, and octet
defining the values type and then the value itself.   The valid field types for
tables are an extension of the native integer, bit, string, and timestamp
types, and are shown in the grammar.  Multi-octet integer fields are always
held in network byte order.
*/
func readTable(r io.Reader) (table Table, err error) {
	return readTableDepth(r, 0)
}

func readTableDepth(r io.Reader, depth int) (table Table, err error) {
	if depth > maxFieldDepth {
		return nil, ErrSyntax
	}

	var nested bytes.Buffer
	var str string

	if str, err = readLongstr(r); err != nil {
		return
	}

	nested.WriteString(str)

	table = make(Table)

	for nested.Len() > 0 {
		var key string
		var value any

		if key, err = readShortstr(&nested); err != nil {
			return
		}

		if value, err = readFieldDepth(&nested, depth); err != nil {
			return
		}

		if len(table) >= maxContainerElements {
			return nil, ErrSyntax
		}

		table[key] = value
	}

	return
}

func readArray(r io.Reader) (arr []any, err error) {
	return readArrayDepth(r, 0)
}

func readArrayDepth(r io.Reader, depth int) (arr []any, err error) {
	if depth > maxFieldDepth {
		return nil, ErrSyntax
	}

	var size uint32

	if err = binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	// size is the encoded blob's byte length, not the element count of arr.
	// Cap it like readLongstr's length to reject clearly invalid values.
	if size > math.MaxInt32 {
		return nil, ErrSyntax
	}

	var (
		lim   = &io.LimitedReader{R: r, N: int64(size)}
		field any
	)

	for {
		if field, err = readFieldDepth(lim, depth); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if len(arr) >= maxContainerElements {
			return nil, ErrSyntax
		}

		arr = append(arr, field)
	}

	return arr, nil
}

// Checks if this bit mask matches the flags bitset
func hasProperty(mask uint16, prop int) bool {
	return int(mask)&prop > 0
}

func (r *reader) parseHeaderFrame(channel uint16, size uint32) (frame frame, err error) {
	hf := &headerFrame{
		ChannelId: channel,
	}

	lim := &io.LimitedReader{R: r.r, N: int64(size)}

	if err = binary.Read(lim, binary.BigEndian, &hf.ClassId); err != nil {
		return
	}

	if err = binary.Read(lim, binary.BigEndian, &hf.weight); err != nil {
		return
	}

	if err = binary.Read(lim, binary.BigEndian, &hf.Size); err != nil {
		return
	}

	var flags uint16

	if err = binary.Read(lim, binary.BigEndian, &flags); err != nil {
		return
	}

	if hasProperty(flags, flagContentType) {
		if hf.Properties.ContentType, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagContentEncoding) {
		if hf.Properties.ContentEncoding, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagHeaders) {
		if hf.Properties.Headers, err = readTable(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagDeliveryMode) {
		if err = binary.Read(lim, binary.BigEndian, &hf.Properties.DeliveryMode); err != nil {
			return
		}
	}
	if hasProperty(flags, flagPriority) {
		if err = binary.Read(lim, binary.BigEndian, &hf.Properties.Priority); err != nil {
			return
		}
	}
	if hasProperty(flags, flagCorrelationId) {
		if hf.Properties.CorrelationId, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagReplyTo) {
		if hf.Properties.ReplyTo, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagExpiration) {
		if hf.Properties.Expiration, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagMessageId) {
		if hf.Properties.MessageId, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagTimestamp) {
		if hf.Properties.Timestamp, err = readTimestamp(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagType) {
		if hf.Properties.Type, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagUserId) {
		if hf.Properties.UserId, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagAppId) {
		if hf.Properties.AppId, err = readShortstr(lim); err != nil {
			return
		}
	}
	if hasProperty(flags, flagReserved1) {
		if hf.Properties.reserved1, err = readShortstr(lim); err != nil {
			return
		}
	}

	// Drain any bytes remaining in the frame payload that were not consumed
	// by property parsing (e.g. padding added by other AMQP implementations).
	if lim.N > 0 {
		if _, err = io.CopyN(io.Discard, lim, lim.N); err != nil {
			return
		}
	}

	return hf, nil
}

func (r *reader) parseBodyFrame(channel uint16, size uint32) (frame frame, err error) {
	bf := &bodyFrame{
		ChannelId: channel,
	}

	if bf.Body, err = readBytes(r.r, int64(size)); err != nil {
		return nil, err
	}

	return bf, nil
}

var errHeartbeatPayload = errors.New("heartbeats should not have a payload")
var errHeartbeatChannel = errors.New("heartbeats must be sent on channel 0")

func (r *reader) parseHeartbeatFrame(channel uint16, size uint32) (frame frame, err error) {
	// Per the AMQP 0-9-1 spec, heartbeats are only ever sent on channel 0.
	if channel != 0 {
		return nil, errHeartbeatChannel
	}

	hf := &heartbeatFrame{
		ChannelId: channel,
	}

	if size > 0 {
		return nil, errHeartbeatPayload
	}

	return hf, nil
}
