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
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	inboundPrefix  = "i."
	outboundPrefix = "o."
)

// Store is an interface which can be used to provide implementations
// for message persistence.
// Because we may have to store distinct messages with the same
// message ID, we need a unique key for each message. This is
// possible by prepending "i." or "o." to each message id
type Store interface {
	Open()
	// Put stores a packet under the given key. Implementations may retain the
	// packet pointer rather than make a copy. Put does not transfer exclusive
	// ownership: the caller may still be using the packet. Both callers and
	// store implementations must follow the shared-data ownership requirements
	// documented by Get.
	Put(key string, message packets.ControlPacket)
	// Get returns the stored packet, or nil if the key is not found.
	// The returned packet may be shared with other users of the store; callers must not modify shared packet data.
	// Make an independent copy of fields that need to change.
	Get(key string) packets.ControlPacket
	All() []string
	Del(key string)
	Close()
	Reset()
}

// A key MUST have the form "X.[messageid]"
// where X is 'i' or 'o'
func mIDFromKey(key string) uint16 {
	s := key[2:]
	i, err := strconv.ParseUint(s, 10, 16)
	chkerr(err)
	return uint16(i)
}

// Return true if key prefix is outbound
func isKeyOutbound(key string) bool {
	return key[:2] == outboundPrefix
}

// Return true if key prefix is inbound
func isKeyInbound(key string) bool {
	return key[:2] == inboundPrefix
}

// Return a string of the form "i.[id]"
func inboundKeyFromMID(id uint16) string {
	return fmt.Sprintf("%s%d", inboundPrefix, id)
}

// Return a string of the form "o.[id]"
func outboundKeyFromMID(id uint16) string {
	return fmt.Sprintf("%s%d", outboundPrefix, id)
}

// govern which outgoing messages are persisted
func persistOutbound(s Store, m packets.ControlPacket, logger *slog.Logger) {
	switch m.Details().Qos {
	case 0:
		switch m.(type) {
		case *packets.PubackPacket, *packets.PubcompPacket:
			// Sending puback. delete matching publish
			// from ibound
			s.Del(inboundKeyFromMID(m.Details().MessageID))
		}
	case 1:
		switch m.(type) {
		case *packets.PublishPacket, *packets.PubrelPacket, *packets.SubscribePacket, *packets.UnsubscribePacket:
			// Sending publish. store in obound
			// until puback received
			s.Put(outboundKeyFromMID(m.Details().MessageID), m)
		default:
			logger.Error("Asked to persist an invalid message type", slog.String("component", string(STR)))
		}
	case 2:
		switch m.(type) {
		case *packets.PublishPacket:
			// Sending publish. store in obound
			// until pubrel received
			s.Put(outboundKeyFromMID(m.Details().MessageID), m)
		default:
			logger.Error("Asked to persist an invalid message type", slog.String("component", string(STR)))
		}
	}
}

// govern which incoming messages are persisted
func persistInbound(s Store, m packets.ControlPacket, logger *slog.Logger) {
	switch m.Details().Qos {
	case 0:
		switch m.(type) {
		case *packets.PubackPacket, *packets.SubackPacket, *packets.UnsubackPacket, *packets.PubcompPacket:
			// Received a puback. delete matching publish
			// from obound
			s.Del(outboundKeyFromMID(m.Details().MessageID))
		case *packets.PublishPacket, *packets.PubrecPacket, *packets.PingrespPacket, *packets.ConnackPacket:
		default:
			logger.Error("Asked to persist an invalid messages type", slog.String("component", string(STR)))
		}
	case 1:
		switch m.(type) {
		case *packets.PublishPacket, *packets.PubrelPacket:
			// Received a publish. store it in ibound
			// until puback sent
			s.Put(inboundKeyFromMID(m.Details().MessageID), m)
		default:
			logger.Error("Asked to persist an invalid messages type", slog.String("component", string(STR)))
		}
	case 2:
		switch m.(type) {
		case *packets.PublishPacket:
			// Received a publish. store it in ibound
			// until pubrel received
			s.Put(inboundKeyFromMID(m.Details().MessageID), m)
		default:
			logger.Error("Asked to persist an invalid messages type", slog.String("component", string(STR)))
		}
	}
}
