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
 *    Matt Brittan
 */

package mqtt

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

var ErrMalformedSuback = packets.ErrMalformedSuback

// ConnectMQTT takes a connected net.Conn and performs the initial MQTT handshake. Parameters are:
// conn - Connected net.Conn
// cm - Connect Packet with everything other than the protocol name/version populated (historical reasons)
// protocolVersion - The protocol version to attempt to connect with
//
// Note that, for backward compatibility, ConnectMQTT() suppresses the actual connection error (compare to connectMQTT()).
func ConnectMQTT(conn net.Conn, cm *packets.ConnectPacket, protocolVersion uint) (byte, bool) {
	logger := noopSLogger
	rc, sessionPresent, _ := connectMQTT(conn, cm, protocolVersion, logger, 0)
	return rc, sessionPresent
}

func ConnectMQTTEx(conn net.Conn, cm *packets.ConnectPacket, protocolVersion uint, logger *slog.Logger) (byte, bool) {
	if logger == nil {
		logger = noopSLogger
	}
	rc, sessionPresent, _ := connectMQTT(conn, cm, protocolVersion, logger, 0)
	return rc, sessionPresent
}

func connectMQTT(conn io.ReadWriter, cm *packets.ConnectPacket, protocolVersion uint, logger *slog.Logger, maxIncomingPacketSize uint32) (byte, bool, error) {
	switch protocolVersion {
	case 3:
		logger.Debug("Using MQTT 3.1 protocol", slog.String("component", string(CLI)))
		cm.ProtocolName = "MQIsdp"
		cm.ProtocolVersion = 3
	case 0x83:
		logger.Debug("Using MQTT 3.1b protocol", slog.String("component", string(CLI)))
		cm.ProtocolName = "MQIsdp"
		cm.ProtocolVersion = 0x83
	case 0x84:
		logger.Debug("Using MQTT 3.1.1b protocol", slog.String("component", string(CLI)))
		cm.ProtocolName = "MQTT"
		cm.ProtocolVersion = 0x84
	default:
		logger.Debug("Using MQTT 3.1.1 protocol", slog.String("component", string(CLI)))
		cm.ProtocolName = "MQTT"
		cm.ProtocolVersion = 4
	}

	if err := cm.Write(conn); err != nil {
		logger.Error("connectMQTT write error", slog.String("error", err.Error()), slog.String("component", string(CLI)))
		return packets.ErrNetworkError, false, err
	}

	rc, sessionPresent, err := verifyCONNACK(conn, logger, maxIncomingPacketSize)
	return rc, sessionPresent, err
}

// This function is only used for receiving a connack
// when the connection is first started.
// This prevents receiving incoming data while resume
// is in progress if clean session is false.
func verifyCONNACK(conn io.Reader, logger *slog.Logger, maxIncomingPacketSize uint32) (byte, bool, error) {
	logger.Debug("connect started", slog.String("component", string(NET)))

	ca, err := packets.ReadPacketWithLimit(conn, maxIncomingPacketSize)
	if err != nil {
		logger.Error("connect got error", slog.String("error", err.Error()), slog.String("component", string(NET)))
		return packets.ErrNetworkError, false, err
	}

	if ca == nil {
		logger.Error("received nil packet", slog.String("component", string(NET)))
		return packets.ErrNetworkError, false, errors.New("nil CONNACK packet")
	}

	msg, ok := ca.(*packets.ConnackPacket)
	if !ok {
		logger.Error("received msg that was not CONNACK", slog.String("component", string(NET)))
		return packets.ErrNetworkError, false, errors.New("non-CONNACK first packet received")
	}

	logger.Debug("received connack", slog.String("component", string(NET)))
	return msg.ReturnCode, msg.SessionPresent, nil
}

// inbound encapsulates the output from startIncoming.
// err  - If != nil then an error has occurred
// cp - A control packet received over the network link
type inbound struct {
	err error
	cp  packets.ControlPacket
}

// startIncoming initiates a goroutine that reads incoming messages off the wire and sends them to the channel (returned).
// If there are any issues with the network connection then the returned channel will be closed and the goroutine will exit
// (so closing the connection will terminate the goroutine)
func startIncoming(conn io.Reader, logger *slog.Logger, maxIncomingPacketSize uint32) <-chan inbound {
	var err error
	var cp packets.ControlPacket
	ibound := make(chan inbound)

	logger.Debug("incoming started", slog.String("component", string(NET)))

	go func() {
		for {
			if cp, err = packets.ReadPacketWithLimit(conn, maxIncomingPacketSize); err != nil {
				// We do not want to log the error if it is due to the network connection having been closed
				// elsewhere (i.e. after sending DisconnectPacket). Detecting this situation is the subject of
				// https://github.com/golang/go/issues/4373
				if !errors.Is(err, net.ErrClosed) {
					ibound <- inbound{err: err}
				}
				close(ibound)
				logger.Debug("incoming complete", slog.String("component", string(NET)))
				return
			}
			logger.Debug("startIncoming Received Message", slog.String("component", string(NET)))
			ibound <- inbound{cp: cp}
		}
	}()

	return ibound
}

// incomingComms encapsulates the possible output of the incomingComms routine. If err != nil then an error has occurred and
// the routine will have terminated; otherwise one of the other members should be non-nil
type incomingComms struct {
	err         error                  // If non-nil then there has been an error (ignore everything else)
	outbound    *PacketAndToken        // Packet (with token) than needs to be sent out (e.g. an acknowledgement)
	incomingPub *packets.PublishPacket // A new publish has been received; this will need to be passed on to our user
}

// startIncomingComms initiates incoming communications; this includes starting a goroutine to process incoming
// messages.
// Accepts a channel of inbound messages from the store (persisted messages); note this must be closed as soon as
// everything in the store has been sent.
// Returns a channel that will be passed any received packets; this will be closed on a network error (and inboundFromStore closed)
func startIncomingComms(conn io.Reader,
	c commsFns,
	inboundFromStore <-chan packets.ControlPacket,
	logger *slog.Logger,
) <-chan incomingComms {
	ibound := startIncoming(conn, logger, c.getMaxIncomingPacketSize()) // Start goroutine that reads from network connection
	output := make(chan incomingComms)

	logger.Debug("startIncomingComms started", slog.String("component", string(NET)))
	go func() {
		for {
			if inboundFromStore == nil && ibound == nil {
				close(output)
				logger.Debug("startIncomingComms goroutine complete", slog.String("component", string(NET)))
				return // As soon as ibound is closed we can exit (should have already processed an error)
			}
			logger.Debug("startIncomingComms logic waiting for msg on ibound", slog.String("component", string(NET)))

			var msg packets.ControlPacket
			var ok bool
			select {
			case msg, ok = <-inboundFromStore:
				if !ok {
					logger.Debug("startIncomingComms: inboundFromStore complete", slog.String("component", string(NET)))
					inboundFromStore = nil // should happen quickly as this is only for persisted messages
					continue
				}
				logger.Debug("startIncomingComms: got msg from store", slog.String("component", string(NET)))
			case ibMsg, ok := <-ibound:
				if !ok {
					logger.Debug("startIncomingComms: ibound complete", slog.String("component", string(NET)))
					ibound = nil
					continue
				}
				logger.Debug("startIncomingComms: got msg on ibound", slog.String("component", string(NET)))
				// If the inbound comms routine encounters any issues it will send us an error.
				if ibMsg.err != nil {
					output <- incomingComms{err: ibMsg.err}
					continue // Usually the channel will be closed immediately after sending an error but safer that we do not assume this
				}
				msg = ibMsg.cp

				c.persistInbound(msg)
				c.UpdateLastReceived() // Notify keepalive logic that we recently received a packet
			}

			switch m := msg.(type) {
			case *packets.PingrespPacket:
				logger.Debug("startIncomingComms: received pingresp", slog.String("component", string(NET)))
				c.pingRespReceived()
			case *packets.SubackPacket:
				logger.Debug("startIncomingComms: received suback", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				token := c.getToken(m.MessageID)

				if t, ok := token.(*SubscribeToken); ok {
					logger.Debug("startIncomingComms: granted qoss", slog.Any("returnCodes", m.ReturnCodes), slog.String("component", string(NET)))
					// [MQTT-3.8.4-5] - The SUBACK Packet sent by the Server to the Client MUST contain a return code for each Topic Filter/QoS pair
					if len(m.ReturnCodes) != len(t.subs) {
						token.setError(ErrMalformedSuback)
						c.freeID(m.MessageID)
						output <- incomingComms{err: ErrMalformedSuback} // This is a protocol error so connection should be closed
						continue
					}
					for i, qos := range m.ReturnCodes {
						t.subResult[t.subs[i]] = qos
					}
				}

				token.flowComplete()
				c.freeID(m.MessageID)
			case *packets.UnsubackPacket:
				logger.Debug("startIncomingComms: received unsuback", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				c.getToken(m.MessageID).flowComplete()
				c.freeID(m.MessageID)
			case *packets.PublishPacket:
				logger.Debug("startIncomingComms: received publish", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				output <- incomingComms{incomingPub: m}
			case *packets.PubackPacket:
				logger.Debug("startIncomingComms: received puback", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				c.getToken(m.MessageID).flowComplete()
				c.freeID(m.MessageID)
			case *packets.PubrecPacket:
				logger.Debug("startIncomingComms: received pubrec", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				prel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
				prel.MessageID = m.MessageID
				output <- incomingComms{outbound: &PacketAndToken{p: prel, t: nil}}
			case *packets.PubrelPacket:
				logger.Debug("startIncomingComms: received pubrel", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				pc := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
				pc.MessageID = m.MessageID
				c.persistOutbound(pc)
				output <- incomingComms{outbound: &PacketAndToken{p: pc, t: nil}}
			case *packets.PubcompPacket:
				logger.Debug("startIncomingComms: received pubcomp", slog.Uint64("messageID", uint64(m.MessageID)), slog.String("component", string(NET)))
				c.getToken(m.MessageID).flowComplete()
				c.freeID(m.MessageID)
			}
		}
	}()
	return output
}

// startOutgoingComms initiates a go routine to transmit outgoing packets.
// Pass in an open network connection and channels for outbound messages (including those triggered
// directly from incoming comms).
// Returns a channel that will receive details of any errors (closed when the goroutine exits)
// This function wil only terminate when all input channels are closed
func startOutgoingComms(conn net.Conn,
	c commsFns,
	oboundp <-chan *PacketAndToken,
	obound <-chan *PacketAndToken,
	oboundFromIncoming <-chan *PacketAndToken,
	logger *slog.Logger,
) <-chan error {
	errChan := make(chan error)
	logger.Debug("outgoing started", slog.String("component", string(NET)))

	go func() {
		for {
			logger.Debug("outgoing waiting for an outbound message", slog.String("component", string(NET)))

			// This goroutine will only exits when all of the input channels we receive on have been closed. This approach is taken to avoid any
			// deadlocks (if the connection goes down there are limited options as to what we can do with anything waiting on us and
			// throwing away the packets seems the best option)
			if oboundp == nil && obound == nil && oboundFromIncoming == nil {
				logger.Debug("outgoing comms stopping", slog.String("component", string(NET)))
				close(errChan)
				return
			}

			select {
			case pub, ok := <-obound:
				if !ok {
					obound = nil
					continue
				}
				msg := pub.p.(*packets.PublishPacket)
				logger.Debug("obound msg to write", slog.Uint64("messageID", uint64(msg.MessageID)), slog.String("component", string(NET)))

				writeTimeout := c.getWriteTimeOut()
				if writeTimeout > 0 {
					if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
						logger.Error("SetWriteDeadline error", slog.String("error", err.Error()), slog.String("component", string(NET)))
					}
				}

				if err := msg.Write(conn); err != nil {
					logger.Error("outgoing obound reporting error", slog.String("error", err.Error()), slog.String("component", string(NET)))
					pub.t.setError(err)
					// report error if it's not due to the connection being closed elsewhere
					if !errors.Is(err, net.ErrClosed) {
						errChan <- err
					}
					continue
				}

				if writeTimeout > 0 {
					// If we successfully wrote, we don't want the timeout to happen during an idle period
					// so we reset it to infinite.
					if err := conn.SetWriteDeadline(time.Time{}); err != nil {
						logger.Error("SetWriteDeadline to 0 error", slog.String("error", err.Error()), slog.String("component", string(NET)))
					}
				}

				if msg.Qos == 0 {
					pub.t.flowComplete()
				}
				logger.Debug("obound wrote msg", slog.Uint64("messageID", uint64(msg.MessageID)), slog.String("component", string(NET)))
			case msg, ok := <-oboundp:
				if !ok {
					oboundp = nil
					continue
				}
				logger.Debug("obound priority msg to write", slog.String("type", reflect.TypeOf(msg.p).String()), slog.Uint64("messageID", uint64(msg.p.Details().MessageID)), slog.String("component", string(NET)))
				if err := msg.p.Write(conn); err != nil {
					logger.Error("outgoing oboundp reporting error", slog.String("error", err.Error()), slog.String("component", string(NET)))
					if msg.t != nil {
						msg.t.setError(err)
					}
					errChan <- err
					continue
				}

				if _, ok := msg.p.(*packets.DisconnectPacket); ok {
					msg.t.(*DisconnectToken).flowComplete()
					logger.Debug("outbound wrote disconnect, closing connection", slog.String("component", string(NET)))
					// As per the MQTT spec "After sending a DISCONNECT Packet the Client MUST close the Network Connection"
					// Closing the connection will cause the goroutines to end in sequence (starting with incoming comms)
					_ = conn.Close()
				}
			case msg, ok := <-oboundFromIncoming: // message triggered by an inbound message (PubrecPacket or PubrelPacket)
				if !ok {
					oboundFromIncoming = nil
					continue
				}
				logger.Debug("obound from incoming msg to write", slog.String("type", reflect.TypeOf(msg.p).String()), slog.Uint64("messageID", uint64(msg.p.Details().MessageID)), slog.String("component", string(NET)))
				if err := msg.p.Write(conn); err != nil {
					logger.Error("outgoing oboundFromIncoming reporting error", slog.String("error", err.Error()), slog.String("component", string(NET)))
					if msg.t != nil {
						msg.t.setError(err)
					}
					errChan <- err
					continue
				}
			}
			c.UpdateLastSent() // Record that a packet has been received (for keepalive routine)
		}
	}()
	return errChan
}

// commsFns provide access to the client state (messageids, requesting disconnection and updating timing)
type commsFns interface {
	getToken(id uint16) tokenCompletor       // Retrieve the token for the specified messageid (if none then a dummy token must be returned)
	freeID(id uint16)                        // Release the specified messageid (clearing out of any persistent store)
	UpdateLastReceived()                     // Must be called whenever a packet is received
	UpdateLastSent()                         // Must be called whenever a packet is successfully sent
	getWriteTimeOut() time.Duration          // Return the writetimeout (or 0 if none)
	getMaxIncomingPacketSize() uint32        // Return the maximum accepted MQTT Remaining Length (or 0 if none)
	persistOutbound(m packets.ControlPacket) // add the packet to the outbound store
	persistInbound(m packets.ControlPacket)  // add the packet to the inbound store
	pingRespReceived()                       // Called when a ping response is received
}

// startComms initiates goroutines that handles communications over the network connection
// Messages will be stored (via commsFns) and deleted from the store as necessary
// It returns two channels:
//
//	packets.PublishPacket - Will receive publish packets received over the network.
//	Closed when incoming comms routines exit (on shutdown or if network link closed)
//	error - Any errors will be sent on this channel. The channel is closed when all comms routines have shut down
//
// Note: The comms routines monitoring oboundp and obound will not shutdown until those channels are both closed. Any messages received between the
// connection being closed and those channels being closed will generate errors (and nothing will be sent). That way the chance of a deadlock is
// minimised.
func startComms(conn net.Conn, // Network connection (must be active)
	c commsFns, // getters and setters to enable us to cleanly interact with client
	inboundFromStore <-chan packets.ControlPacket, // Inbound packets from the persistence store (should be closed relatively soon after startup)
	oboundp <-chan *PacketAndToken,
	obound <-chan *PacketAndToken,
	logger *slog.Logger) (
	<-chan *packets.PublishPacket, // Publishpackages received over the network
	<-chan error, // Any errors (should generally trigger a disconnect)
) {
	// Start inbound comms handler; this needs to be able to transmit messages so we start a go routine to add these to the priority outbound channel
	ibound := startIncomingComms(conn, c, inboundFromStore, logger)
	outboundFromIncoming := make(chan *PacketAndToken) // Will accept outgoing messages triggered by startIncomingComms (e.g. acknowledgements)

	// Start the outgoing handler. It is important to note that output from startIncomingComms is fed into startOutgoingComms (for ACK's)
	oboundErr := startOutgoingComms(conn, c, oboundp, obound, outboundFromIncoming, logger)
	logger.Debug("startComms started", slog.String("component", string(NET)))

	// Run up go routines to handle the output from the above comms functions - these are handled in separate
	// go routines because they can interact (e.g. ibound triggers an ACK to obound which triggers an error)
	var wg sync.WaitGroup
	wg.Add(2)

	outPublish := make(chan *packets.PublishPacket)
	outError := make(chan error)

	// Any messages received get passed to the appropriate channel
	go func() {
		for ic := range ibound {
			if ic.err != nil {
				outError <- ic.err
				continue
			}
			if ic.outbound != nil {
				outboundFromIncoming <- ic.outbound
				continue
			}
			if ic.incomingPub != nil {
				outPublish <- ic.incomingPub
				continue
			}
			logger.Error("startComms received empty incomingComms msg", slog.String("component", string(STR)))
		}
		// Close channels that will not be written to again (allowing other routines to exit)
		close(outboundFromIncoming)
		close(outPublish)
		wg.Done()
	}()

	// Any errors will be passed out to our caller
	go func() {
		for err := range oboundErr {
			outError <- err
		}
		wg.Done()
	}()

	// outError is used by both routines so can only be closed when they are both complete
	go func() {
		wg.Wait()
		close(outError)
		logger.Debug("startComms closing outError", slog.String("component", string(NET)))
	}()

	return outPublish, outError
}

// ackFunc acknowledges a packet
// WARNING sendAck may be called at any time (even after the connection is dead). At the time of writing ACK sent after
// connection loss will be dropped (this is not ideal)
func ackFunc(sendAck func(*PacketAndToken), persist Store, packet *packets.PublishPacket, logger *slog.Logger) func() {
	return func() {
		switch packet.Qos {
		case 2:
			pr := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
			pr.MessageID = packet.MessageID
			logger.Debug("putting pubrec msg on obound", slog.String("component", string(NET)))
			sendAck(&PacketAndToken{p: pr, t: nil})
			logger.Debug("done putting pubrec msg on obound", slog.String("component", string(NET)))
		case 1:
			pa := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
			pa.MessageID = packet.MessageID
			logger.Debug("putting puback msg on obound", slog.String("component", string(NET)))
			persistOutbound(persist, pa, logger) // May fail if store has been closed
			sendAck(&PacketAndToken{p: pa, t: nil})
			logger.Debug("done putting puback msg on obound", slog.String("component", string(NET)))
		case 0:
			// do nothing, since there is no need to send an ack packet back
		}
	}
}
