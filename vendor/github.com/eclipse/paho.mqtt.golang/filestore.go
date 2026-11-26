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
	"io/fs"
	"log/slog"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	msgExt     = ".msg"
	tmpExt     = ".tmp"
	corruptExt = ".CORRUPT"
	// writeTestFile is the name of the temporary file written, read back and removed by Open()
	// to confirm the store directory is usable. The leading dot and reserved name avoid any
	// clash with real message files (whose keys have the form "i.<id>"/"o.<id>").
	writeTestFile = ".paho-write-test" + tmpExt
)

// FileStore implements the store interface using the filesystem to provide
// true persistence, even across client failure. This is designed to use a
// single directory per running client. If you are running multiple clients
// on the same filesystem, you will need to be careful to specify unique
// store directories for each.
type FileStore struct {
	sync.RWMutex
	directory string
	opened    bool
	logger    *slog.Logger
}

// NewFileStore will create a new FileStore which stores its messages in the
// directory provided.
func NewFileStore(directory string) *FileStore {
	store := &FileStore{
		directory: directory,
		opened:    false,
		logger:    noopSLogger,
	}
	return store
}

// NewFileStoreEx will create a new FileStore which stores its messages in the
// directory provided, using the provided logger.
func NewFileStoreEx(directory string, logger *slog.Logger) *FileStore {
	if logger == nil {
		logger = noopSLogger
	}
	store := &FileStore{
		directory: directory,
		opened:    false,
		logger:    logger,
	}
	return store
}

// Open will allow the FileStore to be used.
func (store *FileStore) Open() {
	store.Lock()
	defer store.Unlock()
	// if no store directory was specified in ClientOpts, by default use the
	// current working directory
	if store.directory == "" {
		store.directory, _ = os.Getwd()
	}

	// if store dir exists, great, otherwise, create it
	if !exists(store.directory) {
		perms := os.FileMode(0770)
		merr := os.MkdirAll(store.directory, perms)
		chkerr(merr)
	}

	// Verify that the directory can actually be written to, read from and have files removed.
	// Doing this once, at startup, means configuration problems (most commonly insufficient
	// permissions) surface immediately - when they are easy to notice and correct - rather than
	// later as a panic in the middle of an operation or as silent message loss on a long-running,
	// unattended client. This mirrors the approach taken by the v5 client. The Store interface
	// does not allow Open to return an error, so we fail fast by panicking if the directory is
	// unusable. See https://github.com/eclipse-paho/paho.mqtt.golang/issues/720.
	verifyReadWrite(store.directory)

	store.opened = true
	store.logger.Debug("store is opened", slog.String("directory", store.directory), slog.String("component", string(STR)))
}

// Close will disallow the FileStore from being used.
func (store *FileStore) Close() {
	store.Lock()
	defer store.Unlock()
	store.opened = false
	store.logger.Debug("store is closed", slog.String("component", string(STR)))
}

// Put will put a message into the store, associated with the provided
// key value.
func (store *FileStore) Put(key string, m packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		store.logger.Error("Trying to use file store, but not open", slog.String("component", string(STR)))
		return
	}
	full := fullpath(store.directory, key)
	write(store.directory, key, m)
	if !exists(full) {
		store.logger.Error("file not created", slog.String("path", full), slog.String("component", string(STR)))
	}
}

// Get will retrieve a message from the store, the one associated with
// the provided key value.
func (store *FileStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		store.logger.Error("trying to use file store, but not open", slog.String("component", string(STR)))
		return nil
	}
	filepath := fullpath(store.directory, key)
	if !exists(filepath) {
		return nil
	}
	mfile, oerr := os.Open(filepath)
	if oerr != nil {
		// The file exists but cannot be opened (for example a permissions problem after the file
		// was written by a different user). Open() verifies the directory is usable at startup, so
		// treat this as a problem with this individual file: archive it out of the way and carry on
		// rather than panicking and bringing down a (potentially long-running, unattended) client.
		// See https://github.com/eclipse-paho/paho.mqtt.golang/issues/720.
		newpath := corruptpath(store.directory, key)
		store.logger.Error("failed to open stored message; archiving and skipping", slog.String("error", oerr.Error()), slog.String("archived at", newpath), slog.String("component", string(STR)))
		if archiveErr := os.Rename(filepath, newpath); archiveErr != nil {
			store.logger.Error("failed to archive unreadable file", slog.String("error", archiveErr.Error()), slog.String("component", string(STR)))
		}
		return nil
	}
	msg, rerr := packets.ReadPacket(mfile)
	chkerr(mfile.Close())

	// Message was unreadable, return nil
	if rerr != nil {
		newpath := corruptpath(store.directory, key)
		store.logger.Info("corrupted file detected", slog.String("error", rerr.Error()), slog.String("archived at", newpath), slog.String("component", string(STR)))
		if err := os.Rename(filepath, newpath); err != nil {
			store.logger.Error("failed to archive corrupted file", slog.String("error", err.Error()), slog.String("component", string(STR)))
		}
		return nil
	}
	return msg
}

// All will provide a list of all of the keys associated with messages
// currently residing in the FileStore.
func (store *FileStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	return store.all()
}

// Del will remove the persisted message associated with the provided
// key from the FileStore.
func (store *FileStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	store.del(key)
}

// Reset will remove all persisted messages from the FileStore.
func (store *FileStore) Reset() {
	store.Lock()
	defer store.Unlock()
	store.logger.Info("FileStore Reset", slog.String("component", string(STR)))
	for _, key := range store.all() {
		store.del(key)
	}
}

// lockless
func (store *FileStore) all() []string {
	var err error
	var keys []string

	if !store.opened {
		store.logger.Error("trying to use file store, but not open", slog.String("component", string(STR)))
		return nil
	}

	entries, err := os.ReadDir(store.directory)
	chkerr(err)
	files := make(fileInfos, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		chkerr(err)
		files = append(files, info)
	}
	sort.Sort(files)
	for _, f := range files {
		store.logger.Debug("file in All()", slog.String("name", f.Name()), slog.String("component", string(STR)))
		name := f.Name()
		if len(name) < len(msgExt) || name[len(name)-len(msgExt):] != msgExt {
			store.logger.Debug("skipping file, doesn't have right extension", slog.String("name", name), slog.String("component", string(STR)))
			continue
		}
		key := name[0 : len(name)-4] // remove file extension
		keys = append(keys, key)
	}
	return keys
}

// lockless
func (store *FileStore) del(key string) {
	if !store.opened {
		store.logger.Error("trying to use file store, but not open", slog.String("component", string(STR)))
		return
	}
	store.logger.Debug("store del filepath", slog.String("directory", store.directory), slog.String("component", string(STR)))
	store.logger.Debug("store delete key", slog.String("key", key), slog.String("component", string(STR)))
	filepath := fullpath(store.directory, key)
	store.logger.Debug("path of deletion", slog.String("filepath", filepath), slog.String("component", string(STR)))
	if !exists(filepath) {
		store.logger.Info("store could not delete key", slog.String("key", key), slog.String("component", string(STR)))
		return
	}
	rerr := os.Remove(filepath)
	chkerr(rerr)
	store.logger.Debug("del msg", slog.String("key", key), slog.String("component", string(STR)))
	if exists(filepath) {
		store.logger.Error("file not deleted", slog.String("filepath", filepath), slog.String("component", string(STR)))
	}
}

func fullpath(store string, key string) string {
	p := path.Join(store, key+msgExt)
	return p
}

func tmppath(store string, key string) string {
	p := path.Join(store, key+tmpExt)
	return p
}

func corruptpath(store string, key string) string {
	p := path.Join(store, key+corruptExt)
	return p
}

// create file called "X.[messageid].tmp" located in the store
// the contents of the file is the bytes of the message, then
// rename it to "X.[messageid].msg", overwriting any existing
// message with the same id
// X will be 'i' for inbound messages, and O for outbound messages
func write(store, key string, m packets.ControlPacket) {
	temppath := tmppath(store, key)
	f, err := os.Create(temppath)
	chkerr(err)
	werr := m.Write(f)
	chkerr(werr)
	cerr := f.Close()
	chkerr(cerr)
	rerr := os.Rename(temppath, fullpath(store, key))
	chkerr(rerr)
}

// verifyReadWrite confirms that the store directory is usable by writing, reading back and then
// removing a temporary file. It panics if any of those steps fail. The Store interface does not
// allow Open to return an error, and failing fast at startup (when an operator is most likely to
// notice and fix a misconfiguration) is preferable to discovering the problem later, mid-operation,
// when it may result in lost messages.
func verifyReadWrite(directory string) {
	testpath := path.Join(directory, writeTestFile)
	if err := os.WriteFile(testpath, []byte("test"), 0600); err != nil {
		panic(fmt.Errorf("file store directory %q is not writable: %w", directory, err))
	}
	if _, err := os.ReadFile(testpath); err != nil {
		_ = os.Remove(testpath) // best effort cleanup
		panic(fmt.Errorf("file store directory %q is not readable: %w", directory, err))
	}
	if err := os.Remove(testpath); err != nil {
		panic(fmt.Errorf("file store directory %q does not permit file removal: %w", directory, err))
	}
}

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		chkerr(err)
	}
	return true
}

type fileInfos []fs.FileInfo

func (f fileInfos) Len() int {
	return len(f)
}

func (f fileInfos) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f fileInfos) Less(i, j int) bool {
	return f[i].ModTime().Before(f[j].ModTime())
}
