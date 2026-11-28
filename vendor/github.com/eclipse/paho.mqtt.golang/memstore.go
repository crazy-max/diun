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
	"log/slog"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// MemoryStore implements the store interface to provide a "persistence"
// mechanism wholly stored in memory. This is only useful for
// as long as the client instance exists.
type MemoryStore struct {
	sync.RWMutex
	messages map[string]packets.ControlPacket
	opened   bool
	logger   *slog.Logger
}

// NewMemoryStore returns a pointer to a new instance of
// MemoryStore, the instance is not initialized and ready to
// use until Open() has been called on it.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		messages: make(map[string]packets.ControlPacket),
		opened:   false,
		logger:   noopSLogger,
	}
	return store
}

// NewMemoryStoreEx returns a pointer to a new instance of MemoryStore with a custom logger.
func NewMemoryStoreEx(logger *slog.Logger) *MemoryStore {
	if logger == nil {
		logger = noopSLogger
	}
	store := &MemoryStore{
		messages: make(map[string]packets.ControlPacket),
		opened:   false,
		logger:   logger,
	}
	return store
}

// Open initializes a MemoryStore instance.
func (store *MemoryStore) Open() {
	store.Lock()
	defer store.Unlock()
	store.opened = true
	store.logger.Debug("memorystore initialized", slog.String("component", string(STR)))
}

// Put takes a key and a pointer to a Message and stores the
// message without copying it. The packet remains shared; callers must follow
// the ownership requirements documented by Store.Put.
func (store *MemoryStore) Put(key string, message packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return
	}
	store.messages[key] = message
}

// Get takes a key and looks in the store for a matching Message
// returning either the Message pointer or nil. The returned packet is shared;
// callers must follow the ownership requirements documented by Store.Get.
func (store *MemoryStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return nil
	}
	mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		store.logger.Warn("memorystore get: message not found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	} else {
		store.logger.Debug("memorystore get: message found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	}
	return m
}

// All returns a slice of strings containing all the keys currently
// in the MemoryStore.
func (store *MemoryStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return nil
	}
	var keys []string
	for k := range store.messages {
		keys = append(keys, k)
	}
	return keys
}

// Del takes a key, searches the MemoryStore and if the key is found
// deletes the Message pointer associated with it.
func (store *MemoryStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to use memory store, but not open", slog.String("component", string(STR)))
		return
	}
	mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		store.logger.Info("memorystore del: message not found", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	} else {
		delete(store.messages, key)
		store.logger.Debug("memorystore del: message was deleted", slog.Uint64("messageID", uint64(mid)), slog.String("component", string(STR)))
	}
}

// Close will disallow modifications to the state of the store.
func (store *MemoryStore) Close() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to close memory store, but not open", slog.String("component", string(STR)))
		return
	}
	store.opened = false
	store.logger.Debug("memorystore closed", slog.String("component", string(STR)))
}

// Reset eliminates all persisted message data in the store.
func (store *MemoryStore) Reset() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to reset memory store, but not open", slog.String("component", string(STR)))
	}
	store.messages = make(map[string]packets.ControlPacket)
	store.logger.Info("memorystore wiped", slog.String("component", string(STR)))
}
