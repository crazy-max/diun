// Copyright (c) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package olm

var SetPickleKeyImpl func(key []byte)

// SetPickleKey sets the global pickle key used when encoding structs with Gob or JSON.
func SetPickleKey(key []byte) {
	SetPickleKeyImpl(key)
}
