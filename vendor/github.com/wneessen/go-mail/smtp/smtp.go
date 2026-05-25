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

// Package smtp implements the Simple Mail Transfer Protocol as defined in RFC 5321.
// It also implements the following extensions:
//
//	8BITMIME  RFC 1652
//	AUTH      RFC 2554
//	STARTTLS  RFC 3207
//	DSN       RFC 1891
package smtp

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wneessen/go-mail/log"
)

var (

	// ErrNonTLSConnection is returned when an attempt is made to retrieve TLS state on a non-TLS connection.
	ErrNonTLSConnection = errors.New("connection is not using TLS")

	// ErrNoConnection is returned when attempting to perform an operation that requires an established
	// connection but none exists.
	ErrNoConnection = errors.New("connection is not established")
)

// A Client represents a client connection to an SMTP server.
type Client struct {
	// Text is the textproto.Conn used by the Client. It is exported to allow for clients to add extensions.
	Text *textproto.Conn

	// ErrorHandlerRegistry manages custom error handlers for SMTP host-command pairs.
	ErrorHandlerRegistry *ErrorHandlerRegistry

	// auth supported auth mechanisms
	auth []string

	// authIsActive indicates that the Client is currently during SMTP authentication
	authIsActive bool

	// keep a reference to the connection so it can be used to create a TLS connection later
	conn net.Conn

	// debug logging is enabled
	debug bool

	// didHello indicates whether we've said HELO/EHLO
	didHello bool

	// dsnmrtype defines the mail return option in case DSN is enabled
	dsnmrtype string

	// dsnrntype defines the recipient notify option in case DSN is enabled
	dsnrntype string

	// ext is a map of supported extensions
	ext map[string]string

	// helloError is the error from the hello
	helloError error

	// helloResponse is the response message from the hello
	helloResponse string

	// isConnected indicates if the Client has an active connection
	isConnected bool

	// logAuthData indicates if the Client should include SMTP authentication data in the logs
	logAuthData bool

	// localName is the name to use in HELO/EHLO
	localName string // the name to use in HELO/EHLO

	// logger will be used for debug logging
	logger log.Logger

	// mutex is used to synchronize access to shared resources, ensuring that only one goroutine can access
	// the resource at a time.
	mutex sync.RWMutex

	// skipUTF8 indicates whether the Client should skip SMTPUTF8 in "MAIL FROM" commands, even if the
	// server advertises support for SMTPUTF8.
	skipUTF8 bool

	// tls indicates whether the Client is using TLS
	tls bool

	// serverName denotes the name of the server to which the application will connect. Used for
	// identification and routing.
	serverName string
}

// Dial returns a new [Client] connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host)
}

// NewClient returns a new [Client] using an existing connection and host as a
// server name to be used when authenticating.
func NewClient(conn net.Conn, host string) (*Client, error) {
	text := textproto.NewConn(conn)
	_, _, err := text.ReadResponse(220)
	if err != nil {
		if cerr := text.Close(); cerr != nil {
			// Since we are being Go <1.20 compatible, we can't combine errorrs and
			// duplicate %w vers are not suppored. Therefore let's ignore this linting
			// error for now
			// nolint:errorlint
			return nil, fmt.Errorf("%w, %s", err, cerr)
		}
		return nil, err
	}
	c := &Client{Text: text, conn: conn, serverName: host, localName: "localhost"}
	_, c.tls = conn.(*tls.Conn)
	c.isConnected = true
	c.ErrorHandlerRegistry = NewErrorHandlerRegistry()

	return c, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	c.mutex.Lock()
	err := c.Text.Close()
	c.isConnected = false
	c.mutex.Unlock()
	return err
}

// hello runs a hello exchange if needed.
func (c *Client) hello() error {
	if !c.didHello {
		c.didHello = true
		err := c.ehlo()
		if err != nil {
			if heloErr := c.helo(); heloErr != nil {
				c.helloError = fmt.Errorf("smtp: EHLO/HELO exchange failed. EHLO response: %w, HELO response: %w",
					err, heloErr)
			}
		}
	}
	return c.helloError
}

// Hello sends a HELO or EHLO to the server as the given host name.
// Calling this method is only necessary if the client needs control
// over the host name used. The client will introduce itself as "localhost"
// automatically otherwise. If Hello is called, it must be called before
// any of the other methods.
func (c *Client) Hello(localName string) error {
	if err := validateLine(localName); err != nil {
		return err
	}
	if c.didHello {
		return errors.New("smtp: Hello called after other methods")
	}

	c.mutex.Lock()
	c.localName = localName
	c.mutex.Unlock()

	return c.hello()
}

// HelloResponse returns the message returned by the previous HELO or
// EHLO request, excluding code and features.
func (c *Client) HelloResponse() string {
	return c.helloResponse
}

// SkipSMTPUTF8 sets the Client's SkipSMTPUTF8 flag. If set to true, the Client will not
// send SMTPUTF8 in "MAIL FROM" commands, even if the server advertises support for SMTPUTF8.
func (c *Client) SkipSMTPUTF8(val bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.skipUTF8 = val
}

// cmd is a convenience function that sends a command and returns the response
func (c *Client) cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
	c.mutex.Lock()

	var logMsg []interface{}
	logMsg = args
	logFmt := format
	if c.authIsActive {
		logMsg = []interface{}{"<SMTP auth data redacted>"}
		logFmt = "%s"
	}
	c.debugLog(log.DirClientToServer, logFmt, logMsg...)

	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		c.mutex.Unlock()
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err := c.Text.ReadResponse(expectCode)
	if err != nil {
		fmtValues := strings.Split(format, " ")
		currentCmd := strings.ToLower(fmtValues[0])
		handler := c.ErrorHandlerRegistry.GetHandler(c.serverName, currentCmd)
		handledErr := handler.HandleError(c.serverName, currentCmd, c.Text, err)
		if handledErr != nil {
			c.mutex.Unlock()
			return 0, "", handledErr
		}

		// If the handler successfully recovered, we try reading the response again
		// This assumes the handler has consumed the problematic data beforehand.
		code, msg, err = c.Text.ReadResponse(expectCode)
	}

	logMsg = []interface{}{code, msg}
	if c.authIsActive && code >= 300 && code <= 400 {
		logMsg = []interface{}{code, "<SMTP auth data redacted>"}
	}
	c.debugLog(log.DirServerToClient, "%d %s", logMsg...)

	c.mutex.Unlock()
	return code, msg, err
}

// helo sends the HELO greeting to the server. It should be used only when the
// server does not support ehlo.
func (c *Client) helo() error {
	c.mutex.Lock()
	c.ext = nil
	c.mutex.Unlock()

	_, msg, err := c.cmd(250, "HELO %s", c.localName)
	c.helloResponse, _, _ = strings.Cut(msg, "\n")
	return err
}

// StartTLS sends the STARTTLS command and encrypts all further communication.
// Only servers that advertise the STARTTLS extension support this function.
func (c *Client) StartTLS(config *tls.Config) error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(220, "STARTTLS")
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.conn = tls.Client(c.conn, config)
	c.Text = textproto.NewConn(c.conn)
	c.tls = true
	c.mutex.Unlock()

	return c.ehlo()
}

// TLSConnectionState returns the client's TLS connection state.
// The return values are their zero values if [Client.StartTLS] did
// not succeed.
func (c *Client) TLSConnectionState() (state tls.ConnectionState, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	tc, ok := c.conn.(*tls.Conn)
	if !ok {
		return state, ok
	}
	state, ok = tc.ConnectionState(), true
	return state, ok
}

// Verify checks the validity of an email address on the server.
// If Verify returns nil, the address is valid. A non-nil return
// does not necessarily indicate an invalid address. Many servers
// will not verify addresses for security reasons.
func (c *Client) Verify(addr string) error {
	if err := validateLine(addr); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "VRFY %s", addr)
	return err
}

// Auth authenticates a client using the provided authentication mechanism.
// A failed authentication closes the connection.
// Only servers that advertise the AUTH extension support this function.
func (c *Client) Auth(a Auth) error {
	if err := c.hello(); err != nil {
		return err
	}

	c.mutex.Lock()
	if !c.logAuthData {
		c.authIsActive = true
	}
	c.mutex.Unlock()
	defer func() {
		c.mutex.Lock()
		if !c.logAuthData {
			c.authIsActive = false
		}
		c.mutex.Unlock()
	}()

	encoding := base64.StdEncoding
	mech, resp, err := a.Start(&ServerInfo{c.serverName, c.tls, c.auth})
	if err != nil {
		if qerr := c.Quit(); qerr != nil {
			// Since we are being Go <1.20 compatible, we can't combine errorrs and
			// duplicate %w vers are not suppored. Therefore let's ignore this linting
			// error for now
			// nolint:errorlint
			return fmt.Errorf("%w, %s", err, qerr)
		}
		return err
	}
	resp64 := make([]byte, encoding.EncodedLen(len(resp)))
	encoding.Encode(resp64, resp)
	code, msg64, err := c.cmd(0, "%s", strings.TrimSpace(fmt.Sprintf("AUTH %s %s", mech,
		resp64)))
	for err == nil {
		var msg []byte
		switch code {
		case 334:
			msg, err = encoding.DecodeString(msg64)
		case 235:
			// the last message isn't base64 because it isn't a challenge
			msg = []byte(msg64)
		default:
			err = &textproto.Error{Code: code, Msg: msg64}
		}
		if err == nil {
			resp, err = a.Next(msg, code == 334)
		}
		if err != nil {
			if mech != "XOAUTH2" {
				// abort the AUTH. Not required for XOAUTH2
				_, _, _ = c.cmd(501, "*")
			}
			_ = c.Quit()
			break
		}
		if resp == nil {
			break
		}
		resp64 = make([]byte, encoding.EncodedLen(len(resp)))
		encoding.Encode(resp64, resp)
		code, msg64, err = c.cmd(0, "%s", resp64)
	}
	return err
}

// Mail issues a MAIL command to the server using the provided email address.
// If the server supports the 8BITMIME extension, Mail adds the BODY=8BITMIME
// parameter. If the server supports the SMTPUTF8 extension, Mail adds the
// SMTPUTF8 parameter.
// This initiates a mail transaction and is followed by one or more [Client.Rcpt] calls.
func (c *Client) Mail(from string) error {
	if err := validateLine(from); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	cmdStr := "MAIL FROM:%s"

	c.mutex.RLock()
	if c.ext != nil {
		if _, ok := c.ext["8BITMIME"]; ok {
			cmdStr += " BODY=8BITMIME"
		}
		if _, ok := c.ext["SMTPUTF8"]; ok && !c.skipUTF8 {
			cmdStr += " SMTPUTF8"
		}
		_, ok := c.ext["DSN"]
		if ok && c.dsnmrtype != "" {
			cmdStr += fmt.Sprintf(" RET=%s", c.dsnmrtype)
		}
	}
	c.mutex.RUnlock()

	_, _, err := c.cmd(250, cmdStr, from)
	return err
}

// Rcpt issues a RCPT command to the server using the provided email address.
// A call to Rcpt must be preceded by a call to [Client.Mail] and may be followed by
// a [Client.Data] call or another Rcpt call.
func (c *Client) Rcpt(to string) error {
	if err := validateLine(to); err != nil {
		return err
	}

	c.mutex.RLock()
	_, ok := c.ext["DSN"]
	c.mutex.RUnlock()

	if ok && c.dsnrntype != "" {
		_, _, err := c.cmd(25, "RCPT TO:%s NOTIFY=%s", to, c.dsnrntype)
		return err
	}
	_, _, err := c.cmd(25, "RCPT TO:%s", to)
	return err
}

type DataCloser struct {
	c    *Client
	done bool
	io.WriteCloser
	response string
}

// Close releases the lock, closes the WriteCloser, waits for a response, and then returns any error encountered.
func (d *DataCloser) Close() error {
	d.c.mutex.Lock()
	_ = d.WriteCloser.Close()
	_, resp, err := d.c.Text.ReadResponse(250)
	d.response = resp
	d.done = true
	d.c.mutex.Unlock()
	return err
}

// Write writes data to the underlying WriteCloser while ensuring thread-safety by locking and unlocking a mutex.
func (d *DataCloser) Write(p []byte) (n int, err error) {
	d.c.mutex.Lock()
	n, err = d.WriteCloser.Write(p)
	d.c.mutex.Unlock()
	return n, err
}

// ServerResponse returns the response that was returned by the server after the DataCloser has
// been closed. If the DataCloser has not been closed yet, it will return an empty string.
func (d *DataCloser) ServerResponse() string {
	if !d.done {
		return ""
	}
	return d.response
}

// Data issues a DATA command to the server and returns a writer that
// can be used to write the mail headers and body. The caller should
// close the writer before calling any more methods on c. A call to
// Data must be preceded by one or more calls to [Client.Rcpt].
func (c *Client) Data() (io.WriteCloser, error) {
	_, _, err := c.cmd(354, "DATA")
	if err != nil {
		return nil, err
	}
	datacloser := &DataCloser{}

	c.mutex.Lock()
	datacloser.c = c
	datacloser.WriteCloser = c.Text.DotWriter()
	c.mutex.Unlock()

	return datacloser, nil
}

var testHookStartTLS func(*tls.Config) // nil, except for tests

// SendMail connects to the server at addr, switches to TLS if
// possible, authenticates with the optional mechanism a if possible,
// and then sends an email from address from, to addresses to, with
// message msg.
// The addr must include a port, as in "mail.example.com:smtp".
//
// The addresses in the to parameter are the SMTP RCPT addresses.
//
// The msg parameter should be an RFC 822-style email with headers
// first, a blank line, and then the message body. The lines of msg
// should be CRLF terminated. The msg headers should usually include
// fields such as "From", "To", "Subject", and "Cc".  Sending "Bcc"
// messages is accomplished by including an email address in the to
// parameter but not including it in the msg headers.
//
// The SendMail function and the net/smtp package are low-level
// mechanisms and provide no support for DKIM signing, MIME
// attachments (see the mime/multipart package), or other mail
// functionality. Higher-level packages exist outside of the standard
// library.
func SendMail(addr string, a Auth, from string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()
	if err = c.hello(); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: c.serverName}
		if testHookStartTLS != nil {
			testHookStartTLS(config)
		}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// Extension reports whether an extension is support by the server.
// The extension name is case-insensitive. If the extension is supported,
// Extension also returns a string that contains any parameters the
// server specifies for the extension.
func (c *Client) Extension(ext string) (bool, string) {
	if err := c.hello(); err != nil {
		return false, ""
	}
	if c.ext == nil {
		return false, ""
	}
	ext = strings.ToUpper(ext)

	c.mutex.RLock()
	param, ok := c.ext[ext]
	c.mutex.RUnlock()
	return ok, param
}

// Reset sends the RSET command to the server, aborting the current mail
// transaction.
func (c *Client) Reset() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "RSET")
	return err
}

// Noop sends the NOOP command to the server. It does nothing but check
// that the connection to the server is okay.
func (c *Client) Noop() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "NOOP")
	return err
}

// Quit sends the QUIT command and closes the connection to the server.
func (c *Client) Quit() error {
	// See https://github.com/golang/go/issues/70011
	_ = c.hello() // ignore error; we're quitting anyhow

	_, _, err := c.cmd(221, "QUIT")
	if err != nil {
		return err
	}
	c.mutex.Lock()
	err = c.Text.Close()
	c.isConnected = false
	c.mutex.Unlock()

	return err
}

// SetDebugLog enables the debug logging for incoming and outgoing SMTP messages
func (c *Client) SetDebugLog(v bool) {
	c.debug = v
	if v {
		if c.logger == nil {
			c.logger = log.New(os.Stderr, log.LevelDebug)
		}
		return
	}
	c.logger = nil
}

// SetLogger overrides the default log.Stdlog for the debug logging with a logger that
// satisfies the log.Logger interface
func (c *Client) SetLogger(l log.Logger) {
	if l == nil {
		return
	}
	c.mutex.Lock()
	c.logger = l
	c.mutex.Unlock()
}

// SetLogAuthData enables logging of authentication data in the Client.
func (c *Client) SetLogAuthData() {
	c.mutex.Lock()
	c.logAuthData = true
	c.mutex.Unlock()
}

// SetDSNMailReturnOption sets the DSN mail return option for the Mail method
func (c *Client) SetDSNMailReturnOption(d string) {
	c.mutex.Lock()
	c.dsnmrtype = d
	c.mutex.Unlock()
}

// SetDSNRcptNotifyOption sets the DSN recipient notify option for the Mail method
func (c *Client) SetDSNRcptNotifyOption(d string) {
	c.mutex.Lock()
	c.dsnrntype = d
	c.mutex.Unlock()
}

// HasConnection checks if the client has an active connection.
// Returns true if the `conn` field is not nil, indicating an active connection.
func (c *Client) HasConnection() bool {
	c.mutex.RLock()
	isConn := c.isConnected
	c.mutex.RUnlock()
	return isConn
}

// UpdateDeadline sets a new deadline on the SMTP connection with the specified timeout duration.
func (c *Client) UpdateDeadline(timeout time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.conn == nil {
		return errors.New("smtp: client has no connection")
	}
	if err := c.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("smtp: failed to update deadline: %w", err)
	}
	return nil
}

// GetTLSConnectionState retrieves the TLS connection state of the client's current connection.
// Returns an error if the connection is not using TLS or if the connection is not established.
func (c *Client) GetTLSConnectionState() (*tls.ConnectionState, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.isConnected {
		return nil, ErrNoConnection
	}
	if !c.tls {
		return nil, ErrNonTLSConnection
	}
	if conn, ok := c.conn.(*tls.Conn); ok {
		cstate := conn.ConnectionState()
		return &cstate, nil
	}
	return nil, errors.New("unable to retrieve TLS connection state")
}

// debugLog checks if the debug flag is set and if so logs the provided message to
// the log.Logger interface
func (c *Client) debugLog(d log.Direction, f string, a ...interface{}) {
	if c.debug {
		c.logger.Debugf(log.Log{Direction: d, Format: f, Messages: a})
	}
}

// validateLine checks to see if a line has CR or LF as per RFC 5321.
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
