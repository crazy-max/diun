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

import (
	"crypto/hmac"
	"crypto/md5"
	"fmt"
)

// cramMD5Auth is the type that satisfies the Auth interface for the "SMTP CRAM_MD5" auth
type cramMD5Auth struct {
	username, secret string
}

// CRAMMD5Auth returns an [Auth] that implements the CRAM-MD5 authentication
// mechanism as defined in RFC 2195.
// The returned Auth uses the given username and secret to authenticate
// to the server using the challenge-response mechanism.
func CRAMMD5Auth(username, secret string) Auth {
	return &cramMD5Auth{username, secret}
}

func (a *cramMD5Auth) Start(_ *ServerInfo) (string, []byte, error) {
	return "CRAM-MD5", nil, nil
}

func (a *cramMD5Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		d := hmac.New(md5.New, []byte(a.secret))
		d.Write(fromServer)
		s := make([]byte, 0, d.Size())
		return fmt.Appendf(nil, "%s %x", a.username, d.Sum(s)), nil
	}
	return nil, nil
}
