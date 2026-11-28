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
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// OrderedMemoryStore uses a map internally so the order in which All() returns packets is
// undefined. OrderedMemoryStore resolves this by storing the time the message is added
// and sorting based upon this.

// storedMessage encapsulates a message and the time it was initially stored
type storedMessage struct {
	ts  time.Time
	msg packets.ControlPacket
}

// OrderedMemoryStore implements the store interface to provide a "persistence"
// mechanism wholly stored in memory. This is only useful for
// as long as the client instance exists.
type OrderedMemoryStore struct {
	sync.RWMutex
	messages map[string]storedMessage
	opened   bool
	logger   *slog.Logger
}

// NewOrderedMemoryStore returns a pointer to a new instance of
// OrderedMemoryStore, the instance is not initialized and ready to
// use until Open() has been called on it.
func NewOrderedMemoryStore() *OrderedMemoryStore {
	store := &OrderedMemoryStore{
		messages: make(map[string]storedMessage),
		opened:   false,
		logger:   noopSLogger,
	}
	return store
}

// NewOrderedMemoryStoreEx returns a pointer to a new instance of
// OrderedMemoryStore, the instance is not initialized and ready to
// use until Open() has been called on it. The logger provided will be used
// to log messages from the store.
func NewOrderedMemoryStoreEx(logger *slog.Logger) *OrderedMemoryStore {
	if logger == nil {
		logger = noopSLogger
	}
	store := &OrderedMemoryStore{
		messages: make(map[string]storedMessage),
		opened:   false,
		logger:   logger,
	}
	return store
}

// Open initializes a OrderedMemoryStore instance.
func (store *OrderedMemoryStore) Open() {
	store.Lock()
	defer store.Unlock()
	store.opened = true
	store.logger.Debug("OrderedMemoryStore initialized", slog.String("component", string(STR)))
}

// Put takes a key and a pointer to a Message and stores the
// message.
func (store *OrderedMemoryStore) Put(key string, message packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return
	}
	store.messages[key] = storedMessage{ts: time.Now(), msg: message}
}

// Get takes a key and looks in the store for a matching Message
// returning either the Message pointer or nil.
func (store *OrderedMemoryStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return nil
	}
	mid := mIDFromKey(key)
	m, ok := store.messages[key]
	if !ok || m.msg == nil {
		store.logger.Warn("OrderedMemoryStore get: message not found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	} else {
		store.logger.Debug("OrderedMemoryStore get: message found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	}
	return m.msg
}

// All returns a slice of strings containing all the keys currently
// in the OrderedMemoryStore.
func (store *OrderedMemoryStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return nil
	}
	type tsAndKey struct {
		ts  time.Time
		key string
	}

	tsKeys := make([]tsAndKey, 0, len(store.messages))
	for k, v := range store.messages {
		tsKeys = append(tsKeys, tsAndKey{ts: v.ts, key: k})
	}
	sort.Slice(tsKeys, func(a int, b int) bool { return tsKeys[a].ts.Before(tsKeys[b].ts) })

	keys := make([]string, len(tsKeys))
	for i := range tsKeys {
		keys[i] = tsKeys[i].key
	}
	return keys
}

// Del takes a key, searches the OrderedMemoryStore and if the key is found
// deletes the Message pointer associated with it.
func (store *OrderedMemoryStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return
	}
	mid := mIDFromKey(key)
	_, ok := store.messages[key]
	if !ok {
		store.logger.Info("OrderedMemoryStore del: message not found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	} else {
		delete(store.messages, key)
		store.logger.Debug("OrderedMemoryStore del: message was deleted", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	}
}

// Close will disallow modifications to the state of the store.
func (store *OrderedMemoryStore) Close() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to close memory store, but not open", slog.String("component", string(STR)))
		return
	}
	store.opened = false
	store.logger.Debug("OrderedMemoryStore closed", slog.String("component", string(STR)))
}

// Reset eliminates all persisted message data in the store.
func (store *OrderedMemoryStore) Reset() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to reset memory store, but not open", slog.String("component", string(STR)))
	}
	store.messages = make(map[string]storedMessage)
	store.logger.Info("OrderedMemoryStore wiped", slog.String("component", string(STR)))
}
