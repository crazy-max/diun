// SPDX-FileCopyrightText: Copyright 2010 The Go Authors. All rights reserved.
// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// Original net/smtp code from the Go stdlib by the Go Authors.
// Use of this source code is governed by a BSD-style
// LICENSE file that can be found in this directory.
//
// go-mail specific modifications by the go-mail Authors.
// Licensed under the MIT License.
// See [PROJECT ROOT]/LICENSES directory for more information.
//
// SPDX-License-Identifier: BSD-3-Clause AND MIT

package smtp

import "strings"

// ehlo sends the EHLO (extended hello) greeting to the server. It
// should be the preferred greeting for servers that support it.
func (c *Client) ehlo() error {
	_, msg, err := c.cmd(250, "EHLO %s", c.localName)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	ext := make(map[string]string)
	extList := strings.Split(msg, "\n")
	c.helloResponse = extList[0]
	if len(extList) > 1 {
		extList = extList[1:]
		for _, line := range extList {
			k, v, _ := strings.Cut(line, " ")
			ext[k] = v
		}
	}
	if mechs, ok := ext["AUTH"]; ok {
		c.auth = strings.Split(mechs, " ")
	}
	c.ext = ext
	return err
}
