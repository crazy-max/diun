// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"errors"
	"fmt"
	"strings"
)

// SMTPAuthType is a type wrapper for a string type. It represents the type of SMTP authentication
// mechanism to be used.
type SMTPAuthType string

const (
	// SMTPAuthCramMD5 is the "CRAM-MD5" SASL authentication mechanism as described in RFC 4954.
	// https://datatracker.ietf.org/doc/html/rfc4954/
	//
	// CRAM-MD5 is not secure by modern standards. The vulnerabilities of MD5 and the lack of
	// advanced security features make it inappropriate for protecting sensitive communications
	// today.
	//
	// It was recommended to deprecate the standard in 20 November 2008. As an alternative it
	// recommends e.g. SCRAM or SASL Plain protected by TLS instead.
	//
	// https://datatracker.ietf.org/doc/html/draft-ietf-sasl-crammd5-to-historic-00.html
	SMTPAuthCramMD5 SMTPAuthType = "CRAM-MD5"

	// SMTPAuthCustom is a custom SMTP AUTH mechanism provided by the user. If a user provides
	// a custom smtp.Auth function to the Client, the Client will its smtpAuthType to this type.
	//
	// Do not use this SMTPAuthType without setting a custom smtp.Auth function on the Client.
	SMTPAuthCustom SMTPAuthType = "CUSTOM"

	// SMTPAuthLogin is the "LOGIN" SASL authentication mechanism. This authentication mechanism
	// does not have an official RFC that could be followed. There is a spec by Microsoft and an
	// IETF draft. The IETF draft is more lax than the MS spec, therefore we follow the I-D, which
	// automatically matches the MS spec.
	//
	// Since the "LOGIN" SASL authentication mechanism transmits the username and password in
	// plaintext over the internet connection, we only allow this mechanism over a TLS secured
	// connection.
	//
	// https://msopenspecs.azureedge.net/files/MS-XLOGIN/%5bMS-XLOGIN%5d.pdf
	//
	// https://datatracker.ietf.org/doc/html/draft-murchison-sasl-login-00
	SMTPAuthLogin SMTPAuthType = "LOGIN"

	// SMTPAuthLoginNoEnc is the "LOGIN" SASL authentication mechanism. This authentication mechanism
	// does not have an official RFC that could be followed. There is a spec by Microsoft and an
	// IETF draft. The IETF draft is more lax than the MS spec, therefore we follow the I-D, which
	// automatically matches the MS spec.
	//
	// Since the "LOGIN" SASL authentication mechanism transmits the username and password in
	// plaintext over the internet connection, by default we only allow this mechanism over
	// a TLS secured connection. This authentiation mechanism overrides this default and will
	// allow LOGIN authentication via an unencrypted channel. This can be useful if the
	// connection has already been secured in a different way (e. g. a SSH tunnel)
	//
	// Note: Use this authentication method with caution. If used in the wrong way, you might
	// expose your authentication information over unencrypted channels!
	//
	// https://msopenspecs.azureedge.net/files/MS-XLOGIN/%5bMS-XLOGIN%5d.pdf
	//
	// https://datatracker.ietf.org/doc/html/draft-murchison-sasl-login-00
	SMTPAuthLoginNoEnc SMTPAuthType = "LOGIN-NOENC"

	// SMTPAuthNoAuth is equivalent to performing no authentication at all. It is a convenience
	// option and should not be used. Instead, for mail servers that do no support/require
	// authentication, the Client should not be passed the WithSMTPAuth option at all.
	SMTPAuthNoAuth SMTPAuthType = "NOAUTH"

	// SMTPAuthPlain is the "PLAIN" authentication mechanism as described in RFC 4616.
	//
	// Since the "PLAIN" SASL authentication mechanism transmits the username and password in
	// plaintext over the internet connection, we only allow this mechanism over a TLS secured
	// connection.
	//
	// https://datatracker.ietf.org/doc/html/rfc4616/
	SMTPAuthPlain SMTPAuthType = "PLAIN"

	// SMTPAuthPlainNoEnc is the "PLAIN" authentication mechanism as described in RFC 4616.
	//
	// Since the "PLAIN" SASL authentication mechanism transmits the username and password in
	// plaintext over the internet connection, by default we only allow this mechanism over
	// a TLS secured connection. This authentiation mechanism overrides this default and will
	// allow PLAIN authentication via an unencrypted channel. This can be useful if the
	// connection has already been secured in a different way (e. g. a SSH tunnel)
	//
	// Note: Use this authentication method with caution. If used in the wrong way, you might
	// expose your authentication information over unencrypted channels!
	//
	// https://datatracker.ietf.org/doc/html/rfc4616/
	SMTPAuthPlainNoEnc SMTPAuthType = "PLAIN-NOENC"

	// SMTPAuthXOAUTH2 is the "XOAUTH2" SASL authentication mechanism.
	// https://developers.google.com/gmail/imap/xoauth2-protocol
	SMTPAuthXOAUTH2 SMTPAuthType = "XOAUTH2"

	// SMTPAuthSCRAMSHA1 is the "SCRAM-SHA-1" SASL authentication mechanism as described in RFC 5802.
	//
	// SCRAM-SHA-1 is still considered secure for certain applications, particularly when used as part
	// of a challenge-response authentication mechanism (as we use it). However, it is generally
	// recommended to prefer stronger alternatives like SCRAM-SHA-256(-PLUS), as SHA-1 has known
	// vulnerabilities in other contexts, although it remains effective in HMAC constructions.
	//
	// https://datatracker.ietf.org/doc/html/rfc5802
	SMTPAuthSCRAMSHA1 SMTPAuthType = "SCRAM-SHA-1"

	// SMTPAuthSCRAMSHA1PLUS is the "SCRAM-SHA-1-PLUS" SASL authentication mechanism as described in RFC 5802.
	//
	// SCRAM-SHA-X-PLUS authentication require TLS channel bindings to protect against MitM attacks and
	// to guarantee that the integrity of the transport layer is preserved throughout the authentication
	// process. Therefore we only allow this mechanism over a TLS secured connection.
	//
	// SCRAM-SHA-1-PLUS is still considered secure for certain applications, particularly when used as part
	// of a challenge-response authentication mechanism (as we use it). However, it is generally
	// recommended to prefer stronger alternatives like SCRAM-SHA-256(-PLUS), as SHA-1 has known
	// vulnerabilities in other contexts, although it remains effective in HMAC constructions.
	//
	// https://datatracker.ietf.org/doc/html/rfc5802
	SMTPAuthSCRAMSHA1PLUS SMTPAuthType = "SCRAM-SHA-1-PLUS"

	// SMTPAuthSCRAMSHA256 is the "SCRAM-SHA-256" SASL authentication mechanism as described in RFC 7677.
	//
	// https://datatracker.ietf.org/doc/html/rfc7677
	SMTPAuthSCRAMSHA256 SMTPAuthType = "SCRAM-SHA-256"

	// SMTPAuthSCRAMSHA256PLUS is the "SCRAM-SHA-256-PLUS" SASL authentication mechanism as described in RFC 7677.
	//
	// SCRAM-SHA-X-PLUS authentication require TLS channel bindings to protect against MitM attacks and
	// to guarantee that the integrity of the transport layer is preserved throughout the authentication
	// process. Therefore we only allow this mechanism over a TLS secured connection.
	//
	// https://datatracker.ietf.org/doc/html/rfc7677
	SMTPAuthSCRAMSHA256PLUS SMTPAuthType = "SCRAM-SHA-256-PLUS"

	// SMTPAuthAutoDiscover is a mechanism that dynamically discovers all authentication mechanisms
	// supported by the SMTP server and selects the strongest available one.
	//
	// This type simplifies authentication by automatically negotiating the most secure mechanism
	// offered by the server, based on a predefined security ranking. For instance, mechanisms like
	// SCRAM-SHA-256(-PLUS) or XOAUTH2 are prioritized over weaker mechanisms such as CRAM-MD5 or PLAIN.
	//
	// The negotiation process ensures that mechanisms requiring additional capabilities (e.g.,
	// SCRAM-SHA-X-PLUS with TLS channel binding) are only selected when the necessary prerequisites
	// are in place, such as an active TLS-secured connection.
	//
	// By automating mechanism selection, SMTPAuthAutoDiscover minimizes configuration effort while
	// maximizing security and compatibility with a wide range of SMTP servers.
	SMTPAuthAutoDiscover SMTPAuthType = "AUTODISCOVER"
)

// SMTP Auth related static errors
var (
	// ErrPlainAuthNotSupported is returned when the server does not support the "PLAIN" SMTP
	// authentication type.
	ErrPlainAuthNotSupported = errors.New("server does not support SMTP AUTH type: PLAIN")

	// ErrLoginAuthNotSupported is returned when the server does not support the "LOGIN" SMTP
	// authentication type.
	ErrLoginAuthNotSupported = errors.New("server does not support SMTP AUTH type: LOGIN")

	// ErrCramMD5AuthNotSupported is returned when the server does not support the "CRAM-MD5" SMTP
	// authentication type.
	ErrCramMD5AuthNotSupported = errors.New("server does not support SMTP AUTH type: CRAM-MD5")

	// ErrXOauth2AuthNotSupported is returned when the server does not support the "XOAUTH2" schema.
	ErrXOauth2AuthNotSupported = errors.New("server does not support SMTP AUTH type: XOAUTH2")

	// ErrSCRAMSHA1AuthNotSupported is returned when the server does not support the "SCRAM-SHA-1" SMTP
	// authentication type.
	ErrSCRAMSHA1AuthNotSupported = errors.New("server does not support SMTP AUTH type: SCRAM-SHA-1")

	// ErrSCRAMSHA1PLUSAuthNotSupported is returned when the server does not support the "SCRAM-SHA-1-PLUS" SMTP
	// authentication type.
	ErrSCRAMSHA1PLUSAuthNotSupported = errors.New("server does not support SMTP AUTH type: SCRAM-SHA-1-PLUS")

	// ErrSCRAMSHA256AuthNotSupported is returned when the server does not support the "SCRAM-SHA-256" SMTP
	// authentication type.
	ErrSCRAMSHA256AuthNotSupported = errors.New("server does not support SMTP AUTH type: SCRAM-SHA-256")

	// ErrSCRAMSHA256PLUSAuthNotSupported is returned when the server does not support the "SCRAM-SHA-256-PLUS" SMTP
	// authentication type.
	ErrSCRAMSHA256PLUSAuthNotSupported = errors.New("server does not support SMTP AUTH type: SCRAM-SHA-256-PLUS")

	// ErrNoSupportedAuthDiscovered is returned when the SMTP Auth AutoDiscover process fails to identify
	// any supported authentication mechanisms offered by the server.
	ErrNoSupportedAuthDiscovered = errors.New("SMTP Auth autodiscover was not able to detect a supported " +
		"authentication mechanism")
)

// UnmarshalString satisfies the fig.StringUnmarshaler interface for the SMTPAuthType type
// https://pkg.go.dev/github.com/kkyr/fig#StringUnmarshaler
func (sa *SMTPAuthType) UnmarshalString(value string) error {
	switch strings.ToLower(value) {
	case "auto", "autodiscover", "autodiscovery":
		*sa = SMTPAuthAutoDiscover
	case "cram-md5", "crammd5", "cram":
		*sa = SMTPAuthCramMD5
	case "custom":
		*sa = SMTPAuthCustom
	case "login":
		*sa = SMTPAuthLogin
	case "login-noenc":
		*sa = SMTPAuthLoginNoEnc
	case "none", "noauth", "no":
		*sa = SMTPAuthNoAuth
	case "plain":
		*sa = SMTPAuthPlain
	case "plain-noenc":
		*sa = SMTPAuthPlainNoEnc
	case "scram-sha-1", "scram-sha1", "scramsha1":
		*sa = SMTPAuthSCRAMSHA1
	case "scram-sha-1-plus", "scram-sha1-plus", "scramsha1plus":
		*sa = SMTPAuthSCRAMSHA1PLUS
	case "scram-sha-256", "scram-sha256", "scramsha256":
		*sa = SMTPAuthSCRAMSHA256
	case "scram-sha-256-plus", "scram-sha256-plus", "scramsha256plus":
		*sa = SMTPAuthSCRAMSHA256PLUS
	case "xoauth2", "oauth2":
		*sa = SMTPAuthXOAUTH2
	default:
		return fmt.Errorf("unsupported SMTP auth type: %s", value)
	}
	return nil
}
