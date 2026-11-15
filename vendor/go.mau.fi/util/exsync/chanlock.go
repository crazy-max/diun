// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exsync

import (
	"context"
	"errors"
)

type ChanMutex chan empty

func NewChanMutex() ChanMutex {
	return make(ChanMutex, 1)
}

func (cm ChanMutex) Lock(ctx context.Context) bool {
	if cm == nil {
		panic(errors.New("exsync.ChanMutex: Lock called on uninitialized mutex"))
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case cm <- emptyVal:
		return true
	case <-ctx.Done():
		return false
	}
}

func (cm ChanMutex) LockNoCtx() {
	if cm == nil {
		panic(errors.New("exsync.ChanMutex: LockNoCtx called on uninitialized mutex"))
	}
	cm <- emptyVal
}

func (cm ChanMutex) TryLock() bool {
	if cm == nil {
		panic(errors.New("exsync.ChanMutex: TryLock called on uninitialized mutex"))
	}
	select {
	case cm <- emptyVal:
		return true
	default:
		return false
	}
}

func (cm ChanMutex) Unlock() {
	if cm == nil {
		panic(errors.New("exsync.ChanMutex: Unlock called on uninitialized mutex"))
	}
	select {
	case <-cm:
	default:
		panic(errors.New("exsync.ChanMutex: unlock of unlocked mutex"))
	}
}

func (cm ChanMutex) IsLocked() bool {
	return len(cm) > 0
}
