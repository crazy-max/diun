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
	"errors"
	"fmt"
	"io"
)

// ErrMalformedUnsubscribe indicates that a UNSUBSCRIBE packet is malformed.
var ErrMalformedUnsubscribe = errors.New("malformed UNSUBSCRIBE received")

// UnsubscribePacket is an internal representation of the fields of the
// Unsubscribe MQTT packet
type UnsubscribePacket struct {
	FixedHeader
	MessageID uint16
	Topics    []string
}

func (u *UnsubscribePacket) String() string {
	return fmt.Sprintf("%s MessageID: %d", u.FixedHeader, u.MessageID)
}

// Validate checks packet fields against some requirements set in the MQTT spec.
// Currently, it checks UTF-8 string fields.
//
// MessageID will not be checked because validation may run before the client allocates a packet identifier.
func (u *UnsubscribePacket) Validate() error {
	for _, topic := range u.Topics {
		if err := validateString(topic); err != nil {
			return err
		}
	}
	return nil
}

func (u *UnsubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	if err := u.Validate(); err != nil {
		return err
	}
	body.Write(encodeUint16(u.MessageID))
	for _, topic := range u.Topics {
		body.Write(encodeString(topic))
	}
	packet, err := u.FixedHeader.pack(body.Len())
	if err != nil {
		return err
	}
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack decodes the details of a ControlPacket after the fixed
// header has been read
func (u *UnsubscribePacket) Unpack(b io.Reader) error {
	var err error
	u.MessageID, err = decodeUint16(b)
	if err != nil {
		return err
	}

	// The payload is one or more length-prefixed topic filters (MQTT-3.10.3-2
	// requires at least one), so decoding runs up to the end of the body. Only a
	// clean EOF on the length prefix terminates the payload; a filter whose
	// content is truncated is an error.
	u.Topics = nil // In case Unpack is called multiple times
	for {
		topicLength, err := decodeUint16(b)
		if errors.Is(err, io.EOF) {
			if len(u.Topics) == 0 {
				return fmt.Errorf("%w: no topic filters", ErrMalformedUnsubscribe)
			}
			break
		}
		if err != nil {
			return err // Includes a partially read length prefix.
		}
		if topicLength == 0 {
			return fmt.Errorf("%w: empty topic filter", ErrMalformedUnsubscribe)
		}

		topic := make([]byte, topicLength)
		if _, err := io.ReadFull(b, topic); err != nil {
			if errors.Is(err, io.EOF) {
				// The length prefix was present, so missing content
				// means the field is truncated.
				return io.ErrUnexpectedEOF
			}
			return err
		}

		if err := validateString(string(topic)); err != nil {
			return err
		}
		u.Topics = append(u.Topics, string(topic))
	}

	return nil
}

// Details returns a Details struct containing the Qos and
// MessageID of this ControlPacket
func (u *UnsubscribePacket) Details() Details {
	return Details{Qos: 1, MessageID: u.MessageID}
}
