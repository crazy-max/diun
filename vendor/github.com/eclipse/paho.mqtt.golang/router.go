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
	"container/list"
	"strings"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// route is a type which associates MQTT Topic strings with a
// callback to be executed upon the arrival of a message associated
// with a subscription to that topic.
type route struct {
	topic    string
	callback MessageHandler
}

// match takes a slice of strings which represent the route being tested having been split on '/'
// separators, and a slice of strings representing the topic string in the published message, similarly
// split.
// The function determines if the topic string matches the route according to the MQTT topic rules
// and returns a boolean of the outcome
func match(route []string, topic []string) bool {
	if len(route) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return route[0] == "#"
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}
	return false
}

func routeIncludesTopic(route, topic string) bool {
	return match(routeSplit(route), strings.Split(topic, "/"))
}

// removes $share and sharename when splitting the route to allow
// shared subscription routes to correctly match the topic
func routeSplit(route string) []string {
	var result []string
	if strings.HasPrefix(route, "$share") {
		result = strings.Split(route, "/")[2:]
	} else {
		result = strings.Split(route, "/")
	}
	return result
}

// match takes the topic string of the published message and does a basic compare to the
// string of the current Route, if they match it returns true
func (r *route) match(topic string) bool {
	return r.topic == topic || routeIncludesTopic(r.topic, topic)
}

type router struct {
	sync.RWMutex
	routes         *list.List
	defaultHandler MessageHandler
	messages       chan *packets.PublishPacket
}

// newRouter returns a new instance of a Router and channel which can be used to tell the Router
// to stop
func newRouter() *router {
	router := &router{routes: list.New(), messages: make(chan *packets.PublishPacket)}
	return router
}

// addRoute takes a topic string and MessageHandler callback. It looks in the current list of
// routes to see if there is already a matching Route. If there is it replaces the current
// callback with the new one. If not it add a new entry to the list of Routes.
func (r *router) addRoute(topic string, callback MessageHandler) {
	r.Lock()
	defer r.Unlock()
	for e := r.routes.Front(); e != nil; e = e.Next() {
		if e.Value.(*route).topic == topic {
			r := e.Value.(*route)
			r.callback = callback
			return
		}
	}
	r.routes.PushBack(&route{topic: topic, callback: callback})
}

// deleteRoute takes a route string, looks for a matching Route in the list of Routes. If
// found it removes the Route from the list.
func (r *router) deleteRoute(topic string) {
	r.Lock()
	defer r.Unlock()
	for e := r.routes.Front(); e != nil; e = e.Next() {
		if e.Value.(*route).topic == topic {
			r.routes.Remove(e)
			return
		}
	}
}

// setDefaultHandler assigns a default callback that will be called if no matching Route
// is found for an incoming Publish.
func (r *router) setDefaultHandler(handler MessageHandler) {
	r.Lock()
	defer r.Unlock()
	r.defaultHandler = handler
}

// matchAndDispatch takes a channel of Message pointers as input and starts a go routine that
// takes messages off the channel, matches them against the internal route list and calls the
// associated callback (or the defaultHandler, if one exists and no other route matched). If
// anything is sent down the stop channel the function will end.
func (r *router) matchAndDispatch(messages <-chan *packets.PublishPacket, order bool, client *client) <-chan *PacketAndToken {
	ackChan := make(chan *PacketAndToken) // Channel returned to caller; closed when goroutine terminates

	// In some cases message acknowledgments may come through after shutdown (connection is down etc). Where this is the
	// case we need to accept any such requests and then ignore them. Note that this is not a perfect solution, if we
	// have reconnected, and the session is still live, then the Ack really should be sent (see Issus #726)
	var ackMutex sync.RWMutex
	sendAckChan := ackChan // This will be set to nil before ackChan is closed
	sendAck := func(ack *PacketAndToken) {
		ackMutex.RLock()
		defer ackMutex.RUnlock()
		if sendAckChan != nil {
			sendAckChan <- ack
		} else {
			DEBUG.Println(ROU, "matchAndDispatch received acknowledgment after processing stopped (ACK dropped).")
		}
	}

	go func() { // Main go routine handling inbound messages
		var handlers []MessageHandler
		for message := range messages {
			// DEBUG.Println(ROU, "matchAndDispatch received message")
			sent := false
			r.RLock()
			m := messageFromPublish(message, ackFunc(sendAck, client.persist, message))
			for e := r.routes.Front(); e != nil; e = e.Next() {
				if e.Value.(*route).match(message.TopicName) {
					if order {
						handlers = append(handlers, e.Value.(*route).callback)
					} else {
						hd := e.Value.(*route).callback
						go func() {
							hd(client, m)
							if !client.options.AutoAckDisabled {
								m.Ack()
							}
						}()
					}
					sent = true
				}
			}
			if !sent {
				if r.defaultHandler != nil {
					if order {
						handlers = append(handlers, r.defaultHandler)
					} else {
						go func() {
							r.defaultHandler(client, m)
							if !client.options.AutoAckDisabled {
								m.Ack()
							}
						}()
					}
				} else {
					DEBUG.Println(ROU, "matchAndDispatch received message and no handler was available. Message will NOT be acknowledged.")
				}
			}
			r.RUnlock()
			if order {
				for _, handler := range handlers {
					handler(client, m)
					if !client.options.AutoAckDisabled {
						m.Ack()
					}
				}
				handlers = handlers[:0]
			}
			// DEBUG.Println(ROU, "matchAndDispatch handled message")
		}
		ackMutex.Lock()
		sendAckChan = nil
		ackMutex.Unlock()
		close(ackChan) // as sendAckChan is now nil nothing further will be sent on this
		DEBUG.Println(ROU, "matchAndDispatch exiting")
	}()
	return ackChan
}
