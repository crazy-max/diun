// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

package smtp

type xoauth2Auth struct {
	username, token string
}

// XOAuth2Auth returns an [Auth] that implements the XOAuth2 authentication
// mechanism as defined in the following specs:
//
// https://developers.google.com/gmail/imap/xoauth2-protocol
// https://learn.microsoft.com/en-us/exchange/client-developer/legacy-protocols/how-to-authenticate-an-imap-pop-smtp-application-by-using-oauth
func XOAuth2Auth(username, token string) Auth {
	return &xoauth2Auth{username, token}
}

func (a *xoauth2Auth) Start(_ *ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte("user=" + a.username + "\x01" + "auth=Bearer " + a.token + "\x01\x01"), nil
}

func (a *xoauth2Auth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		return []byte(""), nil
	}
	return nil, nil
}
