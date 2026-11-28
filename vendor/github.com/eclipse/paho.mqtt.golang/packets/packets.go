/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Allan Stockdill-Mander
 */

package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// ControlPacket defines the interface for structs intended to hold
// decoded MQTT packets, either from being read or before being
// written
type ControlPacket interface {
	// Write encodes the packet without modifying it (note that `RemainingLength` is not updated).
	Write(io.Writer) error
	Unpack(io.Reader) error
	String() string
	Details() Details
}

// PacketNames maps the constants for each of the MQTT packet types
// to a string representation of their name.
var PacketNames = map[uint8]string{
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
}

// Below are the constants assigned to each of the MQTT packet types
const (
	Connect     = 1
	Connack     = 2
	Publish     = 3
	Puback      = 4
	Pubrec      = 5
	Pubrel      = 6
	Pubcomp     = 7
	Subscribe   = 8
	Suback      = 9
	Unsubscribe = 10
	Unsuback    = 11
	Pingreq     = 12
	Pingresp    = 13
	Disconnect  = 14
)

// Below are the const definitions for error codes returned by
// Connect()
const (
	Accepted                        = 0x00
	ErrRefusedBadProtocolVersion    = 0x01
	ErrRefusedIDRejected            = 0x02
	ErrRefusedServerUnavailable     = 0x03
	ErrRefusedBadUsernameOrPassword = 0x04
	ErrRefusedNotAuthorised         = 0x05
	ErrNetworkError                 = 0xFE
	ErrProtocolViolation            = 0xFF
)

// ConnackReturnCodes is a map of the error codes constants for Connect()
// to a string representation of the error
var ConnackReturnCodes = map[uint8]string{
	0:   "Connection Accepted",
	1:   "Connection Refused: Bad Protocol Version",
	2:   "Connection Refused: Client Identifier Rejected",
	3:   "Connection Refused: Server Unavailable",
	4:   "Connection Refused: Username or Password in unknown format",
	5:   "Connection Refused: Not Authorised",
	254: "Connection Error",
	255: "Connection Refused: Protocol Violation",
}

var (
	ErrorRefusedBadProtocolVersion    = errors.New("unacceptable protocol version")
	ErrorRefusedIDRejected            = errors.New("identifier rejected")
	ErrorRefusedServerUnavailable     = errors.New("server Unavailable")
	ErrorRefusedBadUsernameOrPassword = errors.New("bad user name or password")
	ErrorRefusedNotAuthorised         = errors.New("not Authorized")
	ErrorNetworkError                 = errors.New("network Error")
	ErrorProtocolViolation            = errors.New("protocol Violation")
	ErrPacketTooLarge                 = errors.New("incoming packet exceeds configured maximum size")
	ErrUnexpectedBodyData             = errors.New("unexpected data at end of packet body")
)

// ConnErrors is a map of the errors codes constants for Connect()
// to a Go error
var ConnErrors = map[byte]error{
	Accepted:                        nil,
	ErrRefusedBadProtocolVersion:    ErrorRefusedBadProtocolVersion,
	ErrRefusedIDRejected:            ErrorRefusedIDRejected,
	ErrRefusedServerUnavailable:     ErrorRefusedServerUnavailable,
	ErrRefusedBadUsernameOrPassword: ErrorRefusedBadUsernameOrPassword,
	ErrRefusedNotAuthorised:         ErrorRefusedNotAuthorised,
	ErrNetworkError:                 ErrorNetworkError,
	ErrProtocolViolation:            ErrorProtocolViolation,
}

// fixedBodyLength maps the packet types whose MQTT 3.1.1 body length is fixed by
// the specification to that length. A larger RemainingLength cannot describe a valid
// packet of that type, so it is rejected before the body is allocated; a smaller one
// is left to the field decoders, which report the truncated read.
var fixedBodyLength = map[byte]int{
	Connack:    2,
	Puback:     2,
	Pubrec:     2,
	Pubrel:     2,
	Pubcomp:    2,
	Unsuback:   2,
	Pingreq:    0,
	Pingresp:   0,
	Disconnect: 0,
}

// ReadPacket takes an instance of an io.Reader (such as net.Conn) and attempts
// to read an MQTT packet from the stream. It returns a ControlPacket
// representing the decoded MQTT packet and an error. One of these returns will
// always be nil, a nil ControlPacket indicating an error occurred.
func ReadPacket(r io.Reader) (ControlPacket, error) {
	return ReadPacketWithLimit(r, 0)
}

// ReadPacketWithLimit behaves like ReadPacket, but rejects packets whose MQTT
// Remaining Length exceeds maxPacketSize before allocating a payload buffer. A
// maxPacketSize of 0 disables the limit.
func ReadPacketWithLimit(r io.Reader, maxPacketSize uint32) (ControlPacket, error) {
	var fh FixedHeader
	b := make([]byte, 1)

	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	err = fh.unpack(b[0], r)
	if err != nil {
		return nil, err
	}
	if maxPacketSize > 0 && uint32(fh.RemainingLength) > maxPacketSize {
		return nil, fmt.Errorf("%w: remaining length %d exceeds limit %d", ErrPacketTooLarge, fh.RemainingLength, maxPacketSize)
	}
	if want, ok := fixedBodyLength[fh.MessageType]; ok && fh.RemainingLength > want {
		return nil, fmt.Errorf("%w: %s declares Remaining Length %d, expected %d",
			ErrUnexpectedBodyData, PacketNames[fh.MessageType], fh.RemainingLength, want)
	}

	cp, err := NewControlPacketWithHeader(fh)
	if err != nil {
		return nil, err
	}

	packetBytes := make([]byte, fh.RemainingLength)
	n, err := io.ReadFull(r, packetBytes)
	if err != nil {
		return nil, err
	}
	if n != fh.RemainingLength {
		return nil, errors.New("failed to read expected data")
	}

	body := bytes.NewBuffer(packetBytes)
	if err = cp.Unpack(body); err != nil {
		return nil, err
	}
	// MQTT 3.1.1 fixes the length of several packet bodies, so bytes left after the
	// packet's fields mean the Remaining Length (section 2.2.3) does not match the
	// packet type. Checking the buffer here (Unpack must always consume the whole
	// body) avoids probing the reader inside every packet and never consumes from
	// a caller's stream.
	if body.Len() != 0 {
		return nil, fmt.Errorf(
			"%w: %s has %d unread body bytes",
			ErrUnexpectedBodyData,
			PacketNames[fh.MessageType],
			body.Len(),
		)
	}
	return cp, nil
}

// NewControlPacket is used to create a new ControlPacket of the type specified
// by packetType, this is usually done by reference to the packet type constants
// defined in packets.go. The newly created ControlPacket is empty and a pointer
// is returned.
func NewControlPacket(packetType byte) ControlPacket {
	switch packetType {
	case Connect:
		return &ConnectPacket{FixedHeader: FixedHeader{MessageType: Connect}}
	case Connack:
		return &ConnackPacket{FixedHeader: FixedHeader{MessageType: Connack}}
	case Disconnect:
		return &DisconnectPacket{FixedHeader: FixedHeader{MessageType: Disconnect}}
	case Publish:
		return &PublishPacket{FixedHeader: FixedHeader{MessageType: Publish}}
	case Puback:
		return &PubackPacket{FixedHeader: FixedHeader{MessageType: Puback}}
	case Pubrec:
		return &PubrecPacket{FixedHeader: FixedHeader{MessageType: Pubrec}}
	case Pubrel:
		return &PubrelPacket{FixedHeader: FixedHeader{MessageType: Pubrel, Qos: 1}}
	case Pubcomp:
		return &PubcompPacket{FixedHeader: FixedHeader{MessageType: Pubcomp}}
	case Subscribe:
		return &SubscribePacket{FixedHeader: FixedHeader{MessageType: Subscribe, Qos: 1}}
	case Suback:
		return &SubackPacket{FixedHeader: FixedHeader{MessageType: Suback}}
	case Unsubscribe:
		return &UnsubscribePacket{FixedHeader: FixedHeader{MessageType: Unsubscribe, Qos: 1}}
	case Unsuback:
		return &UnsubackPacket{FixedHeader: FixedHeader{MessageType: Unsuback}}
	case Pingreq:
		return &PingreqPacket{FixedHeader: FixedHeader{MessageType: Pingreq}}
	case Pingresp:
		return &PingrespPacket{FixedHeader: FixedHeader{MessageType: Pingresp}}
	}
	return nil
}

// NewControlPacketWithHeader is used to create a new ControlPacket of the type
// specified within the FixedHeader that is passed to the function.
// The newly created ControlPacket is empty and a pointer is returned.
func NewControlPacketWithHeader(fh FixedHeader) (ControlPacket, error) {
	switch fh.MessageType {
	case Connect:
		return &ConnectPacket{FixedHeader: fh}, nil
	case Connack:
		return &ConnackPacket{FixedHeader: fh}, nil
	case Disconnect:
		return &DisconnectPacket{FixedHeader: fh}, nil
	case Publish:
		return &PublishPacket{FixedHeader: fh}, nil
	case Puback:
		return &PubackPacket{FixedHeader: fh}, nil
	case Pubrec:
		return &PubrecPacket{FixedHeader: fh}, nil
	case Pubrel:
		return &PubrelPacket{FixedHeader: fh}, nil
	case Pubcomp:
		return &PubcompPacket{FixedHeader: fh}, nil
	case Subscribe:
		return &SubscribePacket{FixedHeader: fh}, nil
	case Suback:
		return &SubackPacket{FixedHeader: fh}, nil
	case Unsubscribe:
		return &UnsubscribePacket{FixedHeader: fh}, nil
	case Unsuback:
		return &UnsubackPacket{FixedHeader: fh}, nil
	case Pingreq:
		return &PingreqPacket{FixedHeader: fh}, nil
	case Pingresp:
		return &PingrespPacket{FixedHeader: fh}, nil
	}
	return nil, fmt.Errorf("unsupported packet type 0x%x", fh.MessageType)
}

// Details struct returned by the Details() function called on
// ControlPackets to present details of the Qos and MessageID
// of the ControlPacket
type Details struct {
	Qos       byte
	MessageID uint16
}

// FixedHeader is a struct to hold the decoded information from
// the fixed header of an MQTT ControlPacket
type FixedHeader struct {
	MessageType byte
	Dup         bool
	Qos         byte
	Retain      bool
	// RemainingLength is populated when decoding a packet from the wire.
	// Write calculates the encoded length independently and leaves this field unchanged.
	RemainingLength int
}

func (fh FixedHeader) String() string {
	return fmt.Sprintf("%s: dup: %t qos: %d retain: %t rLength: %d", PacketNames[fh.MessageType], fh.Dup, fh.Qos, fh.Retain, fh.RemainingLength)
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

func (fh FixedHeader) pack(remainingLength int) (bytes.Buffer, error) {
	var header bytes.Buffer
	header.WriteByte(fh.MessageType<<4 | boolToByte(fh.Dup)<<3 | fh.Qos<<1 | boolToByte(fh.Retain))
	l, err := encodeLength(remainingLength)
	if err != nil {
		return header, err
	}
	header.Write(l)
	return header, nil
}

func (fh *FixedHeader) unpack(typeAndFlags byte, r io.Reader) error {
	fh.MessageType = typeAndFlags >> 4
	fh.Dup = (typeAndFlags>>3)&0x01 > 0
	fh.Qos = (typeAndFlags >> 1) & 0x03
	fh.Retain = typeAndFlags&0x01 > 0

	var err error
	fh.RemainingLength, err = decodeLength(r)
	return err
}

func decodeByte(b io.Reader) (byte, error) {
	num := make([]byte, 1)
	_, err := io.ReadFull(b, num)
	if err != nil {
		return 0, err
	}

	return num[0], nil
}

func decodeUint16(b io.Reader) (uint16, error) {
	num := make([]byte, 2)
	_, err := io.ReadFull(b, num)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(num), nil
}

func encodeUint16(num uint16) []byte {
	bytesResult := make([]byte, 2)
	binary.BigEndian.PutUint16(bytesResult, num)
	return bytesResult
}

func encodeString(field string) []byte {
	return encodeBytes([]byte(field))
}

// validateString checks that field is a valid MQTT UTF-8 string as defined by
// section 1.5.3 of the MQTT 3.1.1 spec: it must be well-formed UTF-8 and must
// not include the null character U+0000 [MQTT-1.5.3-1, MQTT-1.5.3-2]. Non
// characters (which the spec only recommends against) and U+FEFF (which a
// parser must preserve, [MQTT-1.5.3-3]) are accepted.
//
// Error messages intentionally do not include the field contents: strings may
// originate from the network, and echoing a peer-controlled value could
// interact with error handling that matches on specific error text.
func validateString(field string) error {
	if !utf8.ValidString(field) {
		return fmt.Errorf("invalid UTF-8 string (length %d)", len(field))
	}
	for i := 0; i < len(field); i++ {
		if field[i] == 0 {
			return fmt.Errorf("invalid MQTT string (length %d): must not include the null character U+0000", len(field))
		}
	}
	// encodeBytes truncates fields to 65,535 bytes. If the truncation point
	// falls inside a multi-byte character the packet is written successfully
	// but cannot be decoded again (a persisted message would be archived as
	// corrupt instead of replayed), so require the prefix that would be
	// retained to still be a valid UTF-8 string.
	if len(field) > 65535 && !utf8.ValidString(field[:65535]) {
		return fmt.Errorf("invalid UTF-8 string (length %d): the first 65535 bytes retained after truncation are not valid UTF-8", len(field))
	}
	return nil
}

// encodeStringChecked encodes field, failing if it is not a valid MQTT UTF-8
// string rather than writing a malformed packet.
func encodeStringChecked(field string) ([]byte, error) {
	if err := validateString(field); err != nil {
		return nil, err
	}
	return encodeString(field), nil
}

func decodeString(b io.Reader) (string, error) {
	buf, err := decodeBytes(b)
	if err != nil {
		return "", err
	}
	field := string(buf)
	if err := validateString(field); err != nil {
		return "", err
	}
	return field, nil
}

func decodeBytes(b io.Reader) ([]byte, error) {
	fieldLength, err := decodeUint16(b)
	if err != nil {
		return nil, err
	}

	field := make([]byte, fieldLength)
	_, err = io.ReadFull(b, field)
	if err != nil {
		return nil, err
	}

	return field, nil
}

func encodeBytes(field []byte) []byte {
	// Attempting to encode more than 65,535 bytes would lead to an unexpected 16-bit length and extra data written
	// (which would be parsed as later parts of the message). The safest option is to truncate.
	if len(field) > 65535 {
		field = field[0:65535]
	}
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
}

// encodeLength encodes the "Remaining Length" as per 2.2.3 in the spec
func encodeLength(length int) ([]byte, error) {
	// Spec states max packet size is 268,435,455 bytes. Sending length outside this range is invalid.
	if length < 0 || length > 268435455 {
		return nil, errors.New("invalid packet length")
	}
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength, nil
}

// decodeLength decodes the "Remaining Length" as per 2.2.3 in the spec
func decodeLength(r io.Reader) (int, error) {
	var rLength uint32
	var multiplier uint32
	b := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, b)
		if err != nil {
			return 0, err
		}

		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
		if multiplier >= 27 {
			return 0, errors.New("malformed remaining length") // maximum of 4 bytes may be used (see example in spec)
		}
	}
	return int(rLength), nil
}
