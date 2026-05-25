// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	// ErrNoFromAddress indicates that the FROM address is not set, which is required.
	ErrNoFromAddress = errors.New("no FROM address set")

	// ErrNoRcptAddresses indicates that no recipient addresses have been set.
	ErrNoRcptAddresses = errors.New("no recipient addresses set")
)

const (
	// errTplExecuteFailed indicates that the execution of a template has failed, including the underlying error.
	errTplExecuteFailed = "failed to execute template: %w"

	// errTplPointerNil indicates that a template pointer is nil, which prevents further template execution or
	// processing.
	errTplPointerNil = "template pointer is nil"

	// errParseMailAddr indicates that parsing of a mail address has failed, including the problematic address
	// and error.
	errParseMailAddr = "failed to parse mail address %q: %w"
)

const (
	// NoPGP indicates that a message should not be treated as PGP encrypted or signed and is the default value
	// for a message
	NoPGP PGPType = iota

	// PGPEncrypt indicates that a message should be treated as PGP encrypted. This works closely together with
	// the corresponding go-mail-middleware.
	PGPEncrypt

	// PGPSignature indicates that a message should be treated as PGP signed. This works closely together with
	// the corresponding go-mail-middleware.
	PGPSignature
)

// MiddlewareType is a type wrapper for a string. It describes the type of the Middleware and needs to be
// returned by the Middleware.Type method to satisfy the Middleware interface.
type MiddlewareType string

// Middleware represents the interface for modifying or handling email messages. A Middleware allows the user to
// alter a Msg before it is finally processed. Multiple Middleware can be applied to a Msg.
//
// Type returns a unique MiddlewareType. It describes the type of Middleware and makes sure that
// a Middleware is only applied once.
// Handle performs all the processing to the Msg. It always needs to return a Msg back.
type Middleware interface {
	Handle(*Msg) *Msg
	Type() MiddlewareType
}

// PGPType is a type wrapper for an int, representing a type of PGP encryption or signature.
type PGPType int

// Msg represents an email message with various headers, attachments, and encoding settings.
//
// The Msg is the central part of go-mail. It provided a lot of methods that you would expect in a mail
// user agent (MUA). Msg satisfies the io.WriterTo and io.Reader interfaces.
type Msg struct {
	// addrHeader holds a mapping between AddrHeader keys and their corresponding slices of mail.Address pointers.
	addrHeader map[AddrHeader][]*mail.Address

	// attachments holds a list of File pointers that represent files either as attachments or embeds files in
	// a Msg.
	attachments []*File

	// boundary represents the delimiter for separating parts in a multipart message.
	boundary string

	// charset represents the Charset of the Msg.
	//
	// By default we set CharsetUTF8 for a Msg unless overridden by a corresponding MsgOption.
	charset Charset

	// embeds contains a slice of File pointers representing the embedded files in a Msg.
	embeds []*File

	// encoder is a mime.WordEncoder used to encode strings (such as email headers) using a specified
	// Encoding.
	encoder mime.WordEncoder

	// encoding specifies the type of Encoding used for email messages and/or parts.
	encoding Encoding

	// genHeader is a map where the keys are email headers (of type Header) and the values are slices of strings
	// representing header values.
	genHeader map[Header][]string

	// headerCount is an indicate for how many headers have been written during the mail rendering process.
	// This count can be helpful to identify where the mail header ends and the mail body starts
	headerCount int

	// isDelivered indicates whether the Msg has been delivered.
	isDelivered bool

	// middlewares is a slice of Middleware used for modifying or handling messages before they are processed.
	//
	// middlewares are processed in FIFO order.
	middlewares []Middleware

	// mimever represents the MIME version used in a Msg.
	mimever MIMEVersion

	// multiPartBoundary holds the rendered boundary strings for consistent boundary rendering
	// in case a Msg is rendered several times
	multiPartBoundary map[MIMEType]string

	// parts is a slice that holds pointers to Part structures, which represent different parts of a Msg.
	parts []*Part

	// preformHeader maps Header types to their already preformatted string values.
	//
	// Preformatted Header values will not be affected by automatic line breaks.
	preformHeader map[Header]string

	// pgptype indicates that a message has a PGPType assigned and therefore will generate
	// different Content-Type settings in the msgWriter.
	pgptype PGPType

	// serverResponse holds the response from the sending server after the mail has been
	// successfully queued
	serverResponse string

	// sendError represents an error encountered during the process of sending a Msg during the
	// Client.Send operation.
	//
	// sendError will hold an error of type SendError.
	sendError error

	// noDefaultUserAgent indicates whether the default User-Agent will be omitted for the Msg when it is
	// being sent.
	//
	// This can be useful in scenarios where headers are conditionally passed based on receipt - i. e. SMTP proxies.
	noDefaultUserAgent bool

	// sMIME holds a SMIME type to sign a Msg using S/MIME
	sMIME *SMIME
}

// SendmailPath is the default system path to the sendmail binary - at least on standard Unix-like OS.
const SendmailPath = "/usr/sbin/sendmail"

// MsgOption is a function type that modifies a Msg instance during its creation or initialization.
type MsgOption func(*Msg)

// NewMsg creates a new email message with optional MsgOption functions that customize various aspects
// of the message.
//
// This function initializes a new Msg instance with default values for address headers, character set,
// encoding, general headers, and MIME version. It then applies any provided MsgOption functions to
// customize the message according to the user's needs. If an option is nil, it will be ignored.
// After applying the options, the function sets the appropriate MIME WordEncoder for the message.
//
// Parameters:
//   - opts: A variadic list of MsgOption functions that can be used to customize the Msg instance.
//
// Returns:
//   - A pointer to the newly created Msg instance.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5321
func NewMsg(opts ...MsgOption) *Msg {
	msg := &Msg{
		addrHeader:        make(map[AddrHeader][]*mail.Address),
		charset:           CharsetUTF8,
		encoding:          EncodingQP,
		genHeader:         make(map[Header][]string),
		preformHeader:     make(map[Header]string),
		multiPartBoundary: make(map[MIMEType]string),
		mimever:           MIME10,
	}

	// Override defaults with optionally provided MsgOption functions.
	for _, option := range opts {
		if option == nil {
			continue
		}
		option(msg)
	}

	// Set the matcing mime.WordEncoder for the Msg
	msg.setEncoder()

	return msg
}

// WithCharset sets the Charset type for a Msg during its creation or initialization.
//
// This MsgOption function allows you to specify the character set to be used in the email message.
// The charset defines how the text in the message is encoded and interpreted by the email client.
// This option should be called when creating a new Msg instance to ensure that the desired charset
// is set correctly.
//
// Parameters:
//   - charset: The Charset value that specifies the desired character set for the Msg.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2047#section-5
func WithCharset(charset Charset) MsgOption {
	return func(m *Msg) {
		m.charset = charset
	}
}

// WithEncoding sets the Encoding type for a Msg during its creation or initialization.
//
// This MsgOption function allows you to specify the encoding type to be used in the email message.
// The encoding defines how the message content is encoded, which affects how it is transmitted
// and decoded by email clients. This option should be called when creating a new Msg instance to
// ensure that the desired encoding is set correctly.
//
// Parameters:
//   - encoding: The Encoding value that specifies the desired encoding type for the Msg.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2047#section-6
func WithEncoding(encoding Encoding) MsgOption {
	return func(m *Msg) {
		m.encoding = encoding
	}
}

// WithMIMEVersion sets the MIMEVersion type for a Msg during its creation or initialization.
//
// Note that in the context of email, MIME Version 1.0 is the only officially standardized and
// supported version. While MIME has been updated and extended over time via various RFCs, these
// updates and extensions do not introduce new MIME versions; they refine or add features within
// the framework of MIME 1.0. Therefore, there should be no reason to ever use this MsgOption.
//
// Parameters:
//   - version: The MIMEVersion value that specifies the desired MIME version for the Msg.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc1521
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2049
func WithMIMEVersion(version MIMEVersion) MsgOption {
	return func(m *Msg) {
		m.mimever = version
	}
}

// WithBoundary sets the boundary of a Msg to the provided string value during its creation or
// initialization.
//
// NOTE: By default, random MIME boundaries are created. This option should only be used if
// a specific boundary is required for the email message. Using a predefined boundary will only
// work with messages that hold a single multipart part. Using a predefined boundary with several
// multipart parts will render the mail unreadable to the mail client.
//
// Parameters:
//   - boundary: The string value that specifies the desired boundary for the Msg.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
func WithBoundary(boundary string) MsgOption {
	return func(m *Msg) {
		m.boundary = boundary
	}
}

// WithMiddleware adds the given Middleware to the end of the list of the Client middlewares slice.
// Middleware are processed in FIFO order.
//
// This MsgOption function allows you to specify custom middleware that will be applied during the
// message handling process. Middleware can be used to modify the message, perform logging, or
// implement additional functionality as the message flows through the system. Each middleware
// is executed in the order it was added.
//
// Parameters:
//   - middleware: The Middleware to be added to the list for processing.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
func WithMiddleware(middleware Middleware) MsgOption {
	return func(m *Msg) {
		m.middlewares = append(m.middlewares, middleware)
	}
}

// WithPGPType sets the PGP type for the Msg during its creation or initialization, determining
// the encryption or signature method.
//
// This MsgOption function allows you to specify the PGP (Pretty Good Privacy) type to be used
// for securing the message. The chosen PGP type influences how the message is encrypted or
// signed, ensuring confidentiality and integrity of the content. This option should be called
// when creating a new Msg instance to set the desired PGP type appropriately.
//
// Parameters:
//   - pgptype: The PGPType value that specifies the desired PGP type for the Msg.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc4880
func WithPGPType(pgptype PGPType) MsgOption {
	return func(m *Msg) {
		m.pgptype = pgptype
	}
}

// WithNoDefaultUserAgent disables the inclusion of a default User-Agent header in the Msg during
// its creation or initialization.
//
// This MsgOption function allows you to customize the Msg instance by omitting the default
// User-Agent header, which is typically included to provide information about the software
// sending the email. This option can be useful when you want to have more control over the
// headers included in the message, such as when sending from a custom application or for
// privacy reasons.
//
// Returns:
//   - A MsgOption function that can be used to customize the Msg instance.
func WithNoDefaultUserAgent() MsgOption {
	return func(m *Msg) {
		m.noDefaultUserAgent = true
	}
}

// SetCharset sets or overrides the currently set encoding charset of the Msg.
//
// This method allows you to specify a character set for the email message. The charset is
// important for ensuring that the content of the message is correctly interpreted by
// mail clients. Common charset values include UTF-8, ISO-8859-1, and others. If a charset
// is not explicitly set, CharsetUTF8 is used as default.
//
// Parameters:
//   - charset: The Charset value to set for the Msg, determining the encoding used for the message content.
func (m *Msg) SetCharset(charset Charset) {
	m.charset = charset
}

// SetEncoding sets or overrides the currently set Encoding of the Msg.
//
// This method allows you to specify the encoding type for the email message. The encoding
// determines how the message content is represented and can affect the size and compatibility
// of the email. Common encoding types include Base64 and Quoted-Printable. Setting a new
// encoding may also adjust how the message content is processed and transmitted.
//
// Parameters:
//   - encoding: The Encoding value to set for the Msg, determining the method used to encode the
//     message content.
func (m *Msg) SetEncoding(encoding Encoding) {
	m.encoding = encoding
	m.setEncoder()
}

// SetBoundary sets or overrides the currently set boundary of the Msg.
//
// This method allows you to specify a custom boundary string for the MIME message. The
// boundary is used to separate different parts of the message, especially when dealing
// with multipart messages.
//
// NOTE: By default, random MIME boundaries are created. This option should only be used if
// a specific boundary is required for the email message. Using a predefined boundary will only
// work with messages that hold a single multipart part. Using a predefined boundary with several
// multipart parts will render the mail unreadable to the mail client.
//
// Parameters:
//   - boundary: The string value representing the boundary to set for the Msg, used in
//     multipart messages to delimit different sections.
func (m *Msg) SetBoundary(boundary string) {
	m.boundary = boundary
}

// SetMIMEVersion sets or overrides the currently set MIME version of the Msg.
//
// In the context of email, MIME Version 1.0 is the only officially standardized and
// supported version. Although MIME has been updated and extended over time through
// various RFCs, these updates do not introduce new MIME versions; they refine or add
// features within the framework of MIME 1.0. Therefore, there is generally no need to
// use this function to set a different MIME version.
//
// Parameters:
//   - version: The MIMEVersion value to set for the Msg, which determines the MIME
//     version used in the email message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc1521
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2049
func (m *Msg) SetMIMEVersion(version MIMEVersion) {
	m.mimever = version
}

// SetPGPType sets or overrides the currently set PGP type for the Msg, determining the
// encryption or signature method.
//
// This method allows you to specify the PGP type that will be used when encrypting or
// signing the message. Different PGP types correspond to various encryption and signing
// algorithms, and selecting the appropriate type is essential for ensuring the security
// and integrity of the message content.
//
// Parameters:
//   - pgptype: The PGPType value to set for the Msg, which determines the encryption
//     or signature method used for the email message.
func (m *Msg) SetPGPType(pgptype PGPType) {
	m.pgptype = pgptype
}

// Encoding returns the currently set Encoding of the Msg as a string.
//
// This method retrieves the encoding type that is currently applied to the message. The
// encoding type determines how the message content is encoded for transmission. Common
// encoding types include quoted-printable and base64, and the returned string will reflect
// the specific encoding method in use.
//
// Returns:
//   - A string representation of the current Encoding of the Msg.
func (m *Msg) Encoding() string {
	return m.encoding.String()
}

// Charset returns the currently set Charset of the Msg as a string.
//
// This method retrieves the character set that is currently applied to the message. The
// charset defines the encoding for the text content of the message, ensuring that
// characters are displayed correctly across different email clients and platforms. The
// returned string will reflect the specific charset in use, such as UTF-8 or ISO-8859-1.
//
// Returns:
//   - A string representation of the current Charset of the Msg.
func (m *Msg) Charset() string {
	return m.charset.String()
}

// SetHeader sets a generic header field of the Msg.
//
// Deprecated: This method only exists for compatibility reasons. Please use SetGenHeader
// instead. For adding address headers like "To:" or "From", use SetAddrHeader instead.
//
// This method allows you to set a header field for the message, providing the header name
// and its corresponding values. However, it is recommended to utilize the newer methods
// for better clarity and functionality. Using SetGenHeader or SetAddrHeader is preferred
// for more specific header types, ensuring proper handling of the message headers.
//
// Parameters:
//   - header: The header field to set in the Msg.
//   - values: One or more string values to associate with the header field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3
//   - https://datatracker.ietf.org/doc/html/rfc2047
func (m *Msg) SetHeader(header Header, values ...string) {
	m.SetGenHeader(header, values...)
}

// SetGenHeader sets a generic header field of the Msg to the provided list of values.
//
// This method is intended for setting generic headers in the email message. It takes a
// header name and a variadic list of string values, encoding them as necessary before
// storing them in the message's internal header map.
//
// Note: For adding email address-related headers (like "To:", "From", "Cc", etc.),
// use SetAddrHeader instead to ensure proper formatting and validation.
//
// Parameters:
//   - header: The header field to set in the Msg.
//   - values: One or more string values to associate with the header field.
//
// This method ensures that all values are appropriately encoded for email transmission,
// adhering to the necessary standards.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3
//   - https://datatracker.ietf.org/doc/html/rfc2047
func (m *Msg) SetGenHeader(header Header, values ...string) {
	if m.genHeader == nil {
		m.genHeader = make(map[Header][]string)
	}
	for i, val := range values {
		values[i] = m.encodeString(val)
	}
	m.genHeader[header] = values
}

// SetHeaderPreformatted sets a generic header field of the Msg, which content is already preformatted.
//
// Deprecated: This method only exists for compatibility reasons. Please use
// SetGenHeaderPreformatted instead for setting preformatted generic header fields.
//
// Parameters:
//   - header: The header field to set in the Msg.
//   - value: The preformatted string value to associate with the header field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3
//   - https://datatracker.ietf.org/doc/html/rfc2047
func (m *Msg) SetHeaderPreformatted(header Header, value string) {
	m.SetGenHeaderPreformatted(header, value)
}

// SetGenHeaderPreformatted sets a generic header field of the Msg which content is already preformatted.
//
// This method does not take a slice of values but only a single value. The reason for this is that we do not
// perform any content alteration on these kinds of headers and expect the user to have already taken care of
// any kind of formatting required for the header.
//
// Note: This method should be used only as a last resort. Since the user is responsible for the formatting of
// the message header, we cannot guarantee any compliance with RFC 2822. It is advised to use SetGenHeader
// instead for general header fields.
//
// Parameters:
//   - header: The header field to set in the Msg.
//   - value: The preformatted string value to associate with the header field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2822
func (m *Msg) SetGenHeaderPreformatted(header Header, value string) {
	if m.preformHeader == nil {
		m.preformHeader = make(map[Header]string)
	}
	m.preformHeader[header] = value
}

// SetAddrHeader sets the specified AddrHeader for the Msg to the given values.
//
// Addresses are parsed according to RFC 5322. If parsing any of the provided values fails,
// an error is returned. If you cannot guarantee that all provided values are valid, you can
// use SetAddrHeaderIgnoreInvalid instead, which will silently skip any parsing errors.
//
// This method allows you to set address-related headers for the message, ensuring that the
// provided addresses are properly formatted and parsed. Using this method helps maintain the
// integrity of the email addresses within the message.
//
// Parameters:
//   - header: The AddrHeader to set in the Msg (e.g., "From", "To", "Cc", "Bcc").
//   - values: One or more string values representing the email addresses to associate with
//     the specified header.
//
// Returns:
//   - An error if parsing the address according to RFC 5322 fails
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) SetAddrHeader(header AddrHeader, values ...string) error {
	if m.addrHeader == nil {
		m.addrHeader = make(map[AddrHeader][]*mail.Address)
	}
	var addresses []*mail.Address
	for _, addrVal := range values {
		address, err := mail.ParseAddress(addrVal)
		if err != nil {
			return fmt.Errorf(errParseMailAddr, addrVal, err)
		}
		addresses = append(addresses, address)
	}
	switch header {
	case HeaderFrom:
		if len(addresses) > 0 {
			m.addrHeader[header] = []*mail.Address{addresses[0]}
		}
	default:
		m.addrHeader[header] = addresses
	}
	return nil
}

// SetAddrHeaderFromMailAddress sets the specified AddrHeader for the Msg to the given mail.Address values.
//
// This method allows you to set address-related headers for the message, with mail.Address instances
// as input. Using this method helps maintain the integrity of the email addresses within the message.
//
// Since we expect the mail.Address instances to be already parsed according to RFC 5322, this method
// will not attempt to perform any sanity checks except of nil pointers and therefore no error will
// be returned. Nil pointers will be silently ignored.
//
// Parameters:
//   - header: The AddrHeader to set in the Msg (e.g., "From", "To", "Cc", "Bcc").
//   - addresses: One or more mail.Address pointers representing the email addresses to associate with
//     the specified header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) SetAddrHeaderFromMailAddress(header AddrHeader, values ...*mail.Address) {
	if m.addrHeader == nil {
		m.addrHeader = make(map[AddrHeader][]*mail.Address)
	}

	var addresses []*mail.Address
	for _, addrVal := range values {
		if addrVal == nil {
			continue
		}
		addresses = append(addresses, addrVal)
	}

	switch header {
	case HeaderEnvelopeFrom, HeaderFrom:
		if len(addresses) > 0 {
			m.addrHeader[header] = []*mail.Address{addresses[0]}
		}
	case HeaderReplyTo:
		if len(addresses) > 0 {
			m.addrHeader[header] = addresses
		}
	default:
		m.addrHeader[header] = addresses
	}
}

// SetAddrHeaderIgnoreInvalid sets the specified AddrHeader for the Msg to the given values.
//
// Addresses are parsed according to RFC 5322. If parsing of any of the provided values fails,
// the error is ignored and the address is omitted from the address list.
//
// This method allows for setting address headers while ignoring invalid addresses. It is useful
// in scenarios where you want to ensure that only valid addresses are included without halting
// execution due to parsing errors.
//
// Parameters:
//   - header: The AddrHeader field to set in the Msg.
//   - values: One or more string values representing email addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) SetAddrHeaderIgnoreInvalid(header AddrHeader, values ...string) {
	if m.addrHeader == nil {
		m.addrHeader = make(map[AddrHeader][]*mail.Address)
	}
	var addresses []*mail.Address
	for _, addrVal := range values {
		address, err := mail.ParseAddress(m.encodeString(addrVal))
		if err != nil {
			continue
		}
		addresses = append(addresses, address)
	}
	switch header {
	case HeaderFrom:
		if len(addresses) > 0 {
			m.addrHeader[header] = []*mail.Address{addresses[0]}
		}
	default:
		m.addrHeader[header] = addresses
	}
}

// EnvelopeFrom sets the envelope from address for the Msg.
//
// The HeaderEnvelopeFrom address is generally not included in the mail body but only used by the
// Client for communication with the SMTP server. If the Msg has no "FROM" address set in the
// mail body, the msgWriter will try to use the envelope from address if it has been set for the Msg.
// The provided address is validated according to RFC 5322 and will return an error if the validation fails.
//
// Parameters:
//   - from: The envelope from address to set in the Msg.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) EnvelopeFrom(from string) error {
	return m.SetAddrHeader(HeaderEnvelopeFrom, from)
}

// EnvelopeFromFormat sets the provided name and mail address as HeaderEnvelopeFrom for the Msg.
//
// The HeaderEnvelopeFrom address is generally not included in the mail body but only used by the
// Client for communication with the SMTP server. If the Msg has no "FROM" address set in the mail
// body, the msgWriter will try to use the envelope from address if it has been set for the Msg.
// The provided name and address are validated according to RFC 5322 and will return an error if
// the validation fails.
//
// Parameters:
//   - name: The name to associate with the envelope from address.
//   - addr: The mail address to set as the envelope from address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) EnvelopeFromFormat(name, addr string) error {
	return m.SetAddrHeader(HeaderEnvelopeFrom, fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// EnvelopeFromMailAddress sets the "FROM" address in the mail body for the Msg using a mail.Address instance.
//
// The HeaderEnvelopeFrom address is generally not included in the mail body but only used by the
// Client for communication with the SMTP server. If the Msg has no "FROM" address set in the mail
// body, the msgWriter will try to use the envelope from address if it has been set for the Msg.
// The provided name and address are validated according to RFC 5322 and will return an error if
// the validation fails.
//
// Parameters:
//   - addr: The address as mail.Address instance to be set as envelope from address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) EnvelopeFromMailAddress(addr *mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderEnvelopeFrom, addr)
}

// From sets the "FROM" address in the mail body for the Msg.
//
// The "FROM" address is included in the mail body and indicates the sender of the message to
// the recipient. This address is visible in the email client and is typically displayed to the
// recipient. If the "FROM" address is not set, the msgWriter may attempt to use the envelope
// from address (if available) for sending. The provided address is validated according to RFC
// 5322 and will return an error if the validation fails.
//
// Parameters:
//   - from: The "FROM" address to set in the mail body.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) From(from string) error {
	return m.SetAddrHeader(HeaderFrom, from)
}

// FromMailAddress sets the "FROM" address in the mail body for the Msg using a mail.Address instance.
//
// The "FROM" address is included in the mail body and indicates the sender of the message to
// the recipient. This address is visible in the email client and is typically displayed to the
// recipient. If the "FROM" address is not set, the msgWriter may attempt to use the envelope
// from address (if available) for sending.
//
// Parameters:
//   - from: The "FROM" address to set in the mail body as *mail.Address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) FromMailAddress(from *mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderFrom, from)
}

// FromFormat sets the provided name and mail address as the "FROM" address in the mail body for the Msg.
//
// The "FROM" address is included in the mail body and indicates the sender of the message to
// the recipient, and is visible in the email client. If the "FROM" address is not explicitly
// set, the msgWriter may use the envelope from address (if provided) when sending the message.
// The provided name and address are validated according to RFC 5322 and will return an error
// if the validation fails.
//
// Parameters:
//   - name: The name of the sender to include in the "FROM" address.
//   - addr: The email address of the sender to include in the "FROM" address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) FromFormat(name, addr string) error {
	return m.SetAddrHeader(HeaderFrom, fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// To sets one or more "TO" addresses in the mail body for the Msg.
//
// The "TO" address specifies the primary recipient(s) of the message and is included in the mail body.
// This address is visible to the recipient and any other recipients of the message. Multiple "TO" addresses
// can be set by passing them as variadic arguments to this method. Each provided address is validated
// according to RFC 5322, and an error will be returned if ANY validation fails.
//
// Parameters:
//   - rcpts: One or more recipient email addresses to include in the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) To(rcpts ...string) error {
	return m.SetAddrHeader(HeaderTo, rcpts...)
}

// ToMailAddress sets one or more "TO" addresses in the mail body for the Msg.
//
// The "TO" address specifies the primary recipient(s) of the message and is included in the mail body.
// This address is visible to the recipient and any other recipients of the message. Multiple "TO" addresses
// can be set by passing them as variadic arguments to this method.
//
// Parameters:
//   - rcpts: One or more recipient email addresses as mail.Address instance to include
//     in the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) ToMailAddress(rcpts ...*mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderTo, rcpts...)
}

// AddTo adds a single "TO" address to the existing list of recipients in the mail body for the Msg.
//
// This method allows you to add a single recipient to the "TO" field without replacing any previously set
// "TO" addresses. The "TO" address specifies the primary recipient(s) of the message and is visible in the mail
// client. The provided address is validated according to RFC 5322, and an error will be returned if the
// validation fails.
//
// Parameters:
//   - rcpt: The recipient email address to add to the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddTo(rcpt string) error {
	return m.addAddr(HeaderTo, rcpt)
}

// AddToMailAddress adds a single "TO" address to the existing list of recipients in the mail body for the Msg.
//
// This method allows you to add a single recipient to the "TO" field without replacing any previously set
// "TO" addresses. The "TO" address specifies the primary recipient(s) of the message and is visible in the mail
// client. Since the provided mail.Address has already been validated, no further validation is performed in
// this method and the values are used as given.
//
// Parameters:
//   - rcpt: The recipient email address as *mail.Address to add to the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddToMailAddress(rcpt *mail.Address) {
	addresses := append(m.addrHeader[HeaderTo], rcpt)
	m.SetAddrHeaderFromMailAddress(HeaderTo, addresses...)
}

// AddToFormat adds a single "TO" address with the provided name and email to the existing list of recipients
// in the mail body for the Msg.
//
// This method allows you to add a recipient's name and email address to the "TO" field without replacing any
// previously set "TO" addresses. The "TO" address specifies the primary recipient(s) of the message and is
// visible in the mail client. The provided name and address are validated according to RFC 5322, and an error
// will be returned if the validation fails.
//
// Parameters:
//   - name: The name of the recipient to add to the "TO" field.
//   - addr: The email address of the recipient to add to the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddToFormat(name, addr string) error {
	return m.addAddr(HeaderTo, fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// ToIgnoreInvalid sets one or more "TO" addresses in the mail body for the Msg, ignoring any invalid addresses.
//
// This method allows you to add multiple "TO" recipients to the message body. Unlike the standard `To` method,
// any invalid addresses are ignored, and no error is returned for those addresses. Valid addresses will still be
// included in the "TO" field, which is visible in the recipient's mail client. Use this method with caution if
// address validation is critical. Invalid addresses are determined according to RFC 5322.
//
// Parameters:
//   - rcpts: One or more recipient addresses to add to the "TO" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) ToIgnoreInvalid(rcpts ...string) {
	m.SetAddrHeaderIgnoreInvalid(HeaderTo, rcpts...)
}

// ToFromString takes a string of comma-separated email addresses, validates each, and sets them as the
// "TO" addresses for the Msg.
//
// This method allows you to pass a single string containing multiple email addresses separated by commas.
// Each address is validated according to RFC 5322 and set as a recipient in the "TO" field. If any validation
// fails, an error will be returned. The addresses are visible in the mail body and displayed to recipients in
// the mail client. Any "TO" address applied previously will be overwritten.
//
// Parameters:
//   - rcpts: A string containing multiple recipient addresses separated by commas.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) ToFromString(rcpts string) error {
	src := strings.Split(rcpts, ",")
	var dst []string
	for _, address := range src {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		dst = append(dst, address)
	}
	return m.To(dst...)
}

// Cc sets one or more "CC" (carbon copy) addresses in the mail body for the Msg.
//
// The "CC" address specifies secondary recipient(s) of the message, and is included in the mail body.
// These addresses are visible to all recipients, including those listed in the "TO" and other "CC" fields.
// Multiple "CC" addresses can be set by passing them as variadic arguments to this method. Each provided
// address is validated according to RFC 5322, and an error will be returned if ANY validation fails.
//
// Parameters:
//   - rcpts: One or more recipient addresses to be included in the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) Cc(rcpts ...string) error {
	return m.SetAddrHeader(HeaderCc, rcpts...)
}

// CcMailAddress sets one or more "CC" (carbon copy) addresses in the mail body for the Msg.
//
// The "CC" address specifies secondary recipient(s) of the message, and is included in the mail body.
// This address is visible to the recipient and any other recipients of the message. Multiple "CC" addresses
// can be set by passing them as variadic arguments to this method.
//
// Parameters:
//   - rcpts: One or more recipient email addresses as mail.Address instance to include
//     in the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) CcMailAddress(rcpts ...*mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderCc, rcpts...)
}

// AddCc adds a single "CC" (carbon copy) address to the existing list of "CC" recipients in the mail body
// for the Msg.
//
// This method allows you to add a single recipient to the "CC" field without replacing any previously set "CC"
// addresses. The "CC" address specifies secondary recipient(s) and is visible to all recipients, including those
// in the "TO" field. The provided address is validated according to RFC 5322, and an error will be returned if
// the validation fails.
//
// Parameters:
//   - rcpt: The recipient address to be added to the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddCc(rcpt string) error {
	return m.addAddr(HeaderCc, rcpt)
}

// AddCcMailAddress adds a single "CC" address to the existing list of recipients in the mail body for the Msg.
//
// This method allows you to add a single recipient to the "CC" field without replacing any previously set "CC"
// addresses. The "CC" address specifies secondary recipient(s) and is visible to all recipients, including those
// in the "CC" field. Since the provided mail.Address has already been validated, no further validation is
// performed in this method and the values are used as given.
//
// Parameters:
//   - rcpt: The recipient email address as *mail.Address to add to the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddCcMailAddress(rcpt *mail.Address) {
	addresses := append(m.addrHeader[HeaderCc], rcpt)
	m.SetAddrHeaderFromMailAddress(HeaderCc, addresses...)
}

// AddCcFormat adds a single "CC" (carbon copy) address with the provided name and email to the existing list
// of "CC" recipients in the mail body for the Msg.
//
// This method allows you to add a recipient's name and email address to the "CC" field without replacing any
// previously set "CC" addresses. The "CC" address specifies secondary recipient(s) and is visible to all
// recipients, including those in the "TO" field. The provided name and address are validated according to
// RFC 5322, and an error will be returned if the validation fails.
//
// Parameters:
//   - name: The name of the recipient to be added to the "CC" field.
//   - addr: The email address of the recipient to be added to the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddCcFormat(name, addr string) error {
	return m.addAddr(HeaderCc, fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// CcIgnoreInvalid sets one or more "CC" (carbon copy) addresses in the mail body for the Msg, ignoring any
// invalid addresses.
//
// This method allows you to add multiple "CC" recipients to the message body. Unlike the standard `Cc` method,
// any invalid addresses are ignored, and no error is returned for those addresses. Valid addresses will still
// be included in the "CC" field, which is visible to all recipients in the mail client. Use this method with
// caution if address validation is critical, as invalid addresses are determined according to RFC 5322.
//
// Parameters:
//   - rcpts: One or more recipient email addresses to be added to the "CC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) CcIgnoreInvalid(rcpts ...string) {
	m.SetAddrHeaderIgnoreInvalid(HeaderCc, rcpts...)
}

// CcFromString takes a string of comma-separated email addresses, validates each, and sets them as the "CC"
// addresses for the Msg.
//
// This method allows you to pass a single string containing multiple email addresses separated by commas.
// Each address is validated according to RFC 5322 and set as a recipient in the "CC" field. If any validation
// fails, an error will be returned. The addresses are visible in the mail body and displayed to recipients
// in the mail client.
//
// Parameters:
//   - rcpts: A string containing multiple email addresses separated by commas.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) CcFromString(rcpts string) error {
	src := strings.Split(rcpts, ",")
	var dst []string
	for _, address := range src {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		dst = append(dst, address)
	}
	return m.Cc(dst...)
}

// Bcc sets one or more "BCC" (blind carbon copy) addresses in the mail body for the Msg.
//
// The "BCC" address specifies recipient(s) of the message who will receive a copy without other recipients
// being aware of it. These addresses are not visible in the mail body or to any other recipients, ensuring
// the privacy of BCC'd recipients. Multiple "BCC" addresses can be set by passing them as variadic arguments
// to this method. Each provided address is validated according to RFC 5322, and an error will be returned
// if ANY validation fails.
//
// Parameters:
//   - rcpts: One or more string values representing the BCC addresses to set in the Msg.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) Bcc(rcpts ...string) error {
	return m.SetAddrHeader(HeaderBcc, rcpts...)
}

// BccMailAddress sets one or more "BCC" (blind carbon copy) addresses in the mail body for the Msg.
//
// The "BCC" address specifies recipient(s) of the message who will receive a copy without other recipients
// being aware of it. These addresses are not visible in the mail body or to any other recipients, ensuring
// the privacy of BCC'd recipients. Multiple "BCC" addresses can be set by passing them as variadic arguments
// arguments to this method.
//
// Parameters:
//   - rcpts: One or more recipient email addresses as mail.Address instance to include in the "BCC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) BccMailAddress(rcpts ...*mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderBcc, rcpts...)
}

// AddBcc adds a single "BCC" (blind carbon copy) address to the existing list of "BCC" recipients in the mail
// body for the Msg.
//
// This method allows you to add a single recipient to the "BCC" field without replacing any previously set
// "BCC" addresses. The "BCC" address specifies recipient(s) of the message who will receive a copy without other
// recipients being aware of it. The provided address is validated according to RFC 5322, and an error will be
// returned if the validation fails.
//
// Parameters:
//   - rcpt: The BCC address to add to the existing list of recipients in the Msg.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddBcc(rcpt string) error {
	return m.addAddr(HeaderBcc, rcpt)
}

// AddBccMailAddress adds a single "BCC" address to the existing list of recipients in the mail body for the Msg.
//
// This method allows you to add a single recipient to the "BCC" field without replacing any previously set
// "BCC" addresses. The "BCC" address specifies recipient(s) of the message who will receive a copy without other
// recipients being aware of it.
//
// Parameters:
//   - rcpt: The recipient email address as *mail.Address to add to the "BCC" field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddBccMailAddress(rcpt *mail.Address) {
	addresses := append(m.addrHeader[HeaderBcc], rcpt)
	m.SetAddrHeaderFromMailAddress(HeaderBcc, addresses...)
}

// AddBccFormat adds a single "BCC" (blind carbon copy) address with the provided name and email to the existing
// list of "BCC" recipients in the mail body for the Msg.
//
// This method allows you to add a recipient's name and email address to the "BCC" field without replacing
// any previously set "BCC" addresses. The "BCC" address specifies recipient(s) of the message who will receive
// a copy without other recipients being aware of it. The provided name and address are validated according to
// RFC 5322, and an error will be returned if the validation fails.
//
// Parameters:
//   - name: The name of the recipient to add to the BCC field.
//   - addr: The email address of the recipient to add to the BCC field.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) AddBccFormat(name, addr string) error {
	return m.addAddr(HeaderBcc, fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// BccIgnoreInvalid sets one or more "BCC" (blind carbon copy) addresses in the mail body for the Msg,
// ignoring any invalid addresses.
//
// This method allows you to add multiple "BCC" recipients to the message body. Unlike the standard `Bcc`
// method, any invalid addresses are ignored, and no error is returned for those addresses. Valid addresses
// will still be included in the "BCC" field, which ensures the privacy of the BCC'd recipients. Use this method
// with caution if address validation is critical, as invalid addresses are determined according to RFC 5322.
//
// Parameters:
//   - rcpts: One or more string values representing the BCC email addresses to set.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) BccIgnoreInvalid(rcpts ...string) {
	m.SetAddrHeaderIgnoreInvalid(HeaderBcc, rcpts...)
}

// BccFromString takes a string of comma-separated email addresses, validates each, and sets them as the "BCC"
// addresses for the Msg.
//
// This method allows you to pass a single string containing multiple email addresses separated by commas.
// Each address is validated according to RFC 5322 and set as a recipient in the "BCC" field. If any validation
// fails, an error will be returned. The addresses are not visible in the mail body and ensure the privacy of
// BCC'd recipients.
//
// Parameters:
//   - rcpts: A string of comma-separated email addresses to set as BCC recipients.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) BccFromString(rcpts string) error {
	src := strings.Split(rcpts, ",")
	var dst []string
	for _, address := range src {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		dst = append(dst, address)
	}
	return m.Bcc(dst...)
}

// ReplyTo sets the "Reply-To" address for the Msg, specifying where replies should be sent.
//
// This method takes a single email address as input and attempts to parse it. If the address is valid, it sets
// the "Reply-To" header in the message. The "Reply-To" address can be different from the "From" address,
// allowing the sender to specify an alternate address for responses. If the provided address cannot be parsed,
// an error will be returned, indicating the parsing failure.
//
// Parameters:
//   - addr: The email address to set as the "Reply-To" address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) ReplyTo(addr string) error {
	return m.SetAddrHeader(HeaderReplyTo, addr)
}

// ReplyToMailAddress sets one or more "Reply-To" addresses for the Msg, specifying where replies should be sent.
//
// The "Reply-To" address can be different from the "From" address, allowing the sender to specify an alternate
// address for responses.
//
// Parameters:
//   - addr: The mail.Address instance to set as the "Reply-To" address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) ReplyToMailAddress(addrs ...*mail.Address) {
	m.SetAddrHeaderFromMailAddress(HeaderReplyTo, addrs...)
}

// ReplyToFormat sets the "Reply-To" address for the Msg using the provided name and email address, specifying
// where replies should be sent.
//
// This method formats the name and email address into a single "Reply-To" header. If the formatted address is valid,
// it sets the "Reply-To" header in the message. This allows the sender to specify a display name along with the
// reply address, providing clarity for recipients. If the constructed address cannot be parsed, an error will
// be returned, indicating the parsing failure.
//
// Parameters:
//   - name: The display name associated with the reply address.
//   - addr: The email address to set as the "Reply-To" address.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) ReplyToFormat(name, addr string) error {
	return m.ReplyTo(fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// Subject sets the "Subject" header for the Msg, specifying the topic of the message.
//
// This method takes a single string as input and sets it as the "Subject" of the email. The subject line provides
// a brief summary of the content of the message, allowing recipients to quickly understand its purpose.
//
// Parameters:
//   - subj: The subject line of the email.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.5
func (m *Msg) Subject(subj string) {
	m.SetGenHeader(HeaderSubject, subj)
}

// SetMessageID generates and sets a unique "Message-ID" header for the Msg.
//
// This method creates a "Message-ID" string using a randomly generated string and the hostname of the machine.
// The generated ID helps uniquely identify the message in email systems, facilitating tracking and preventing
// duplication. If the hostname cannot be retrieved, it defaults to "localhost.localdomain".
//
// The generated Message-ID follows the format
// "<randomString@hostname>".
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.4
func (m *Msg) SetMessageID() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost.localdomain"
	}
	// We have 64 possible characters, which for a 22 character string, provides approx. 132 bits of entropy.
	randString, _ := randomStringSecure(22)
	m.SetMessageIDWithValue(fmt.Sprintf("%s@%s", randString, hostname))
}

// GetMessageID retrieves the "Message-ID" header from the Msg.
//
// This method checks if a "Message-ID" has been set in the message's generated headers. If a valid "Message-ID"
// exists in the Msg, it returns the first occurrence of the header. If the "Message-ID" has not been set or
// is empty, it returns an empty string. This allows other components to access the unique identifier for the
// message, which is useful for tracking and referencing in email systems.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.4
func (m *Msg) GetMessageID() string {
	if msgidheader, ok := m.genHeader[HeaderMessageID]; ok {
		if len(msgidheader) > 0 {
			return msgidheader[0]
		}
	}
	return ""
}

// SetMessageIDWithValue sets the "Message-ID" header for the Msg using the provided messageID string.
//
// This method formats the input messageID by enclosing it in angle brackets ("<>") and sets it as the "Message-ID"
// header in the message. The "Message-ID" is a unique identifier for the email, helping email clients and servers
// to track and reference the message. There are no validations performed on the input messageID, so it should
// be in a suitable format for use as a Message-ID.
//
// Parameters:
//   - messageID: The string to set as the "Message-ID" in the message header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.4
func (m *Msg) SetMessageIDWithValue(messageID string) {
	m.SetGenHeader(HeaderMessageID, fmt.Sprintf("<%s>", messageID))
}

// SetBulk sets the "Precedence: bulk" and "X-Auto-Response-Suppress: All" headers for the Msg,
// which are recommended for automated emails such as out-of-office replies.
//
// The "Precedence: bulk" header indicates that the message is a bulk email, and the "X-Auto-Response-Suppress: All"
// header instructs mail servers and clients to suppress automatic responses to this message.
// This is particularly useful for reducing unnecessary replies to automated notifications or replies.
//
// References:
//   - https://www.rfc-editor.org/rfc/rfc2076#section-3.9
//   - https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcmail/ced68690-498a-4567-9d14-5c01f974d8b1#Appendix_A_Target_51
func (m *Msg) SetBulk() {
	m.SetGenHeader(HeaderPrecedence, "bulk")
	m.SetGenHeader(HeaderXAutoResponseSuppress, "All")
}

// SetDate sets the "Date" header for the Msg to the current time in a valid RFC 1123 format.
//
// This method retrieves the current time and formats it according to RFC 1123, ensuring that the "Date"
// header is compliant with email standards. The "Date" header indicates when the message was created,
// providing recipients with context for the timing of the email.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.3
//   - https://datatracker.ietf.org/doc/html/rfc1123
func (m *Msg) SetDate() {
	m.SetDateWithValue(time.Now())
}

// SetDateWithValue sets the "Date" header for the Msg using the provided time value in a valid RFC 1123 format.
//
// This method takes a `time.Time` value as input and formats it according to RFC 1123, ensuring that the "Date"
// header is compliant with email standards. The "Date" header indicates when the message was created,
// providing recipients with context for the timing of the email. This allows for setting a custom date
// rather than using the current time.
//
// Parameters:
//   - timeVal: The time value used to set the "Date" header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.3
//   - https://datatracker.ietf.org/doc/html/rfc1123
func (m *Msg) SetDateWithValue(timeVal time.Time) {
	m.SetGenHeader(HeaderDate, timeVal.Format(time.RFC1123Z))
}

// SetImportance sets the "Importance" and "Priority" headers for the Msg to the specified Importance level.
//
// This method adjusts the email's importance based on the provided Importance value. If the importance level
// is set to `ImportanceNormal`, no headers are modified. Otherwise, it sets the "Importance", "Priority",
// "X-Priority", and "X-MSMail-Priority" headers accordingly, providing email clients with information on
// how to prioritize the message. This allows the sender to indicate the significance of the email to recipients.
//
// Parameters:
//   - importance: The Importance value that determines the priority of the email message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2156
func (m *Msg) SetImportance(importance Importance) {
	if importance == ImportanceNormal {
		return
	}
	m.SetGenHeader(HeaderImportance, importance.String())
	m.SetGenHeader(HeaderPriority, importance.NumString())
	m.SetGenHeader(HeaderXPriority, importance.XPrioString())
	m.SetGenHeader(HeaderXMSMailPriority, importance.NumString())
}

// SetOrganization sets the "Organization" header for the Msg to the specified organization string.
//
// This method allows you to specify the organization associated with the email sender. The "Organization"
// header provides recipients with information about the organization that is sending the message.
// This can help establish context and credibility for the email communication.
//
// Parameters:
//   - org: The name of the organization to be set in the "Organization" header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.4
func (m *Msg) SetOrganization(org string) {
	m.SetGenHeader(HeaderOrganization, org)
}

// SetUserAgent sets the "User-Agent" and "X-Mailer" headers for the Msg to the specified user agent string.
//
// This method allows you to specify the user agent or mailer software used to send the email.
// The "User-Agent" and "X-Mailer" headers provide recipients with information about the email client
// or application that generated the message. This can be useful for identifying the source of the email,
// particularly for troubleshooting or filtering purposes.
//
// Parameters:
//   - userAgent: The user agent or mailer software to be set in the "User-Agent" and "X-Mailer" headers.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.7
func (m *Msg) SetUserAgent(userAgent string) {
	m.SetGenHeader(HeaderUserAgent, userAgent)
	m.SetGenHeader(HeaderXMailer, userAgent)
}

// IsDelivered indicates whether the Msg has been delivered.
//
// This method checks the internal state of the message to determine if it has been successfully
// delivered. It returns true if the message is marked as delivered and false otherwise.
// This can be useful for tracking the status of the email communication.
//
// Returns:
//   - A boolean value indicating the delivery status of the message (true if delivered, false otherwise).
func (m *Msg) IsDelivered() bool {
	return m.isDelivered
}

// RequestMDNTo adds the "Disposition-Notification-To" header to the Msg to request a Message Disposition
// Notification (MDN) from the receiving end, as specified in RFC 8098.
//
// This method allows you to provide a list of recipient addresses to receive the MDN.
// Each address is validated according to RFC 5322 standards. If ANY address is invalid, an error
// will be returned indicating the parsing failure. If the "Disposition-Notification-To" header
// is already set, it will be updated with the new list of addresses.
//
// Parameters:
//   - rcpts: One or more recipient email addresses to request the MDN from.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc8098
func (m *Msg) RequestMDNTo(rcpts ...string) error {
	if m.genHeader == nil {
		m.genHeader = make(map[Header][]string)
	}
	var addresses []string
	for _, addrVal := range rcpts {
		address, err := mail.ParseAddress(addrVal)
		if err != nil {
			return fmt.Errorf(errParseMailAddr, addrVal, err)
		}
		addresses = append(addresses, address.String())
	}
	m.genHeader[HeaderDispositionNotificationTo] = addresses
	return nil
}

// RequestMDNToFormat adds the "Disposition-Notification-To" header to the Msg to request a Message Disposition
// Notification (MDN) from the receiving end, as specified in RFC 8098.
//
// This method allows you to provide a recipient address along with a name, formatting it appropriately.
// Address validation is performed according to RFC 5322 standards. If the provided address is invalid,
// an error will be returned. This method internally calls RequestMDNTo to handle the actual setting of the header.
//
// Parameters:
//   - name: The name of the recipient for the MDN request.
//   - addr: The email address of the recipient for the MDN request.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc8098
func (m *Msg) RequestMDNToFormat(name, addr string) error {
	return m.RequestMDNTo(fmt.Sprintf(`%s <%s>`, name, addr))
}

// RequestMDNAddTo adds an additional recipient to the "Disposition-Notification-To" header for the Msg.
//
// This method allows you to append a new recipient address to the existing list of recipients for the
// MDN. The provided address is validated according to RFC 5322 standards. If the address is invalid,
// an error will be returned indicating the parsing failure. If the "Disposition-Notification-To"
// header is already set, the new recipient will be added to the existing list.
//
// Parameters:
//   - rcpt: The recipient email address to add to the "Disposition-Notification-To" header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc8098
func (m *Msg) RequestMDNAddTo(rcpt string) error {
	address, err := mail.ParseAddress(rcpt)
	if err != nil {
		return fmt.Errorf(errParseMailAddr, rcpt, err)
	}
	var addresses []string
	if current, ok := m.genHeader[HeaderDispositionNotificationTo]; ok {
		addresses = current
	}
	addresses = append(addresses, address.String())
	m.genHeader[HeaderDispositionNotificationTo] = addresses
	return nil
}

// RequestMDNAddToFormat adds an additional formatted recipient to the "Disposition-Notification-To"
// header for the Msg.
//
// This method allows you to specify a recipient address along with a name, formatting it appropriately
// before adding it to the existing list of recipients for the MDN. The formatted address is validated
// according to RFC 5322 standards. If the provided address is invalid, an error will be returned.
// This method internally calls RequestMDNAddTo to handle the actual addition of the recipient.
//
// Parameters:
//   - name: The name of the recipient to add to the "Disposition-Notification-To" header.
//   - addr: The email address of the recipient to add to the "Disposition-Notification-To" header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc8098
func (m *Msg) RequestMDNAddToFormat(name, addr string) error {
	return m.RequestMDNAddTo(fmt.Sprintf(`"%s" <%s>`, name, addr))
}

// GetSender returns the currently set envelope "FROM" address for the Msg. If no envelope
// "FROM" address is set, it will use the first "FROM" address from the mail body. If the
// useFullAddr parameter is true, it will return the full address string, including the name
// if it is set.
//
// If neither the envelope "FROM" nor the body "FROM" addresses are available, it will return
// an error indicating that no "FROM" address is present.
//
// Parameters:
//   - useFullAddr: A boolean indicating whether to return the full address string (including
//     the name) or just the email address.
//
// Returns:
//   - The sender's address as a string and an error if applicable.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) GetSender(useFullAddr bool) (string, error) {
	from, ok := m.addrHeader[HeaderEnvelopeFrom]
	if !ok || len(from) == 0 {
		from, ok = m.addrHeader[HeaderFrom]
		if !ok || len(from) == 0 {
			return "", ErrNoFromAddress
		}
	}

	addr := *from[0]
	if !useFullAddr {
		return mailAddressStringWithoutName(addr), nil
	}
	return addr.String(), nil
}

// GetRecipients returns a list of the currently set "TO", "CC", and "BCC" addresses for the Msg.
//
// This method aggregates recipients from the "TO", "CC", and "BCC" headers and returns them as a
// slice of strings. If no recipients are found in these headers, it will return an error indicating
// that no recipient addresses are present.
//
// Returns:
//   - A slice of strings containing the recipients' addresses and an error if applicable.
//   - If there are no recipient addresses set, it will return an error indicating no recipient
//     addresses are available.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetRecipients() ([]string, error) {
	var rcpts []string
	for _, addressType := range []AddrHeader{HeaderTo, HeaderCc, HeaderBcc} {
		addresses, ok := m.addrHeader[addressType]
		if !ok || len(addresses) == 0 {
			continue
		}
		for _, r := range addresses {
			rcpts = append(rcpts, mailAddressStringWithoutName(*r))
		}
	}
	if len(rcpts) <= 0 {
		return rcpts, ErrNoRcptAddresses
	}
	return rcpts, nil
}

// GetAddrHeader returns the content of the requested address header for the Msg.
//
// This method retrieves the addresses associated with the specified address header. It returns a
// slice of pointers to mail.Address structures representing the addresses found in the header.
// If the requested header does not exist or contains no addresses, it will return nil.
//
// Parameters:
//   - header: The AddrHeader enum value indicating which address header to retrieve (e.g., "TO",
//     "CC", "BCC", etc.).
//
// Returns:
//   - A slice of pointers to mail.Address structures containing the addresses from the specified
//     header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6
func (m *Msg) GetAddrHeader(header AddrHeader) []*mail.Address {
	return m.addrHeader[header]
}

// GetAddrHeaderString returns the address strings of the requested address header for the Msg.
//
// This method retrieves the addresses associated with the specified address header and returns them
// as a slice of strings. Each address is formatted as a string, which includes both the name (if
// available) and the email address. If the requested header does not exist or contains no addresses,
// it will return an empty slice.
//
// Parameters:
//   - header: The AddrHeader enum value indicating which address header to retrieve (e.g., "TO",
//     "CC", "BCC", etc.).
//
// Returns:
//   - A slice of strings containing the formatted addresses from the specified header.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6
func (m *Msg) GetAddrHeaderString(header AddrHeader) []string {
	var addresses []string
	for _, mh := range m.addrHeader[header] {
		addresses = append(addresses, mh.String())
	}
	return addresses
}

// GetFrom returns the content of the "From" address header of the Msg.
//
// This method retrieves the list of email addresses set in the "From" header of the message.
// It returns a slice of pointers to `mail.Address` objects representing the sender(s) of the email.
//
// Returns:
//   - A slice of `*mail.Address` containing the "From" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) GetFrom() []*mail.Address {
	return m.GetAddrHeader(HeaderFrom)
}

// GetFromString returns the content of the "From" address header of the Msg as a string slice.
//
// This method retrieves the list of email addresses set in the "From" header of the message
// and returns them as a slice of strings, with each entry representing a formatted email address.
//
// Returns:
//   - A slice of strings containing the "From" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.2
func (m *Msg) GetFromString() []string {
	return m.GetAddrHeaderString(HeaderFrom)
}

// GetTo returns the content of the "To" address header of the Msg.
//
// This method retrieves the list of email addresses set in the "To" header of the message.
// It returns a slice of pointers to `mail.Address` objects representing the primary recipient(s) of the email.
//
// Returns:
//   - A slice of `*mail.Address` containing the "To" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetTo() []*mail.Address {
	return m.GetAddrHeader(HeaderTo)
}

// GetToString returns the content of the "To" address header of the Msg as a string slice.
//
// This method retrieves the list of email addresses set in the "To" header of the message
// and returns them as a slice of strings, with each entry representing a formatted email address.
//
// Returns:
//   - A slice of strings containing the "To" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetToString() []string {
	return m.GetAddrHeaderString(HeaderTo)
}

// GetCc returns the content of the "Cc" address header of the Msg.
//
// This method retrieves the list of email addresses set in the "Cc" (carbon copy) header of the message.
// It returns a slice of pointers to `mail.Address` objects representing the secondary recipient(s) of the email.
//
// Returns:
//   - A slice of `*mail.Address` containing the "Cc" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetCc() []*mail.Address {
	return m.GetAddrHeader(HeaderCc)
}

// GetCcString returns the content of the "Cc" address header of the Msg as a string slice.
//
// This method retrieves the list of email addresses set in the "Cc" (carbon copy) header of the message
// and returns them as a slice of strings, with each entry representing a formatted email address.
//
// Returns:
//   - A slice of strings containing the "Cc" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetCcString() []string {
	return m.GetAddrHeaderString(HeaderCc)
}

// GetBcc returns the content of the "Bcc" address header of the Msg.
//
// This method retrieves the list of email addresses set in the "Bcc" (blind carbon copy) header of the message.
// It returns a slice of pointers to `mail.Address` objects representing the Bcc recipient(s) of the email.
//
// Returns:
//   - A slice of `*mail.Address` containing the "Bcc" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetBcc() []*mail.Address {
	return m.GetAddrHeader(HeaderBcc)
}

// GetBccString returns the content of the "Bcc" address header of the Msg as a string slice.
//
// This method retrieves the list of email addresses set in the "Bcc" (blind carbon copy) header of the message
// and returns them as a slice of strings, with each entry representing a formatted email address.
//
// Returns:
//   - A slice of strings containing the "Bcc" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetBccString() []string {
	return m.GetAddrHeaderString(HeaderBcc)
}

// GetGenHeader returns the content of the requested generic header of the Msg.
//
// This method retrieves the list of string values associated with the specified generic header of the message.
// It returns a slice of strings representing the header's values.
//
// Parameters:
//   - header: The Header field whose values are being retrieved.
//
// Returns:
//   - A slice of strings containing the values of the specified generic header.
func (m *Msg) GetGenHeader(header Header) []string {
	return m.genHeader[header]
}

// GetReplyTo returns the content of the "ReplyTo" address header of the Msg.
//
// This method retrieves the list of email addresses set in the "ReplyTo" header of the message.
// It returns a slice of pointers to `mail.Address` objects representing the return path(s) of the email.
//
// Returns:
//   - A slice of `*mail.Address` containing the "ReplyTo" header addresses.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.3
func (m *Msg) GetReplyTo() []*mail.Address {
	return m.GetAddrHeader(HeaderReplyTo)
}

// GetParts returns the message parts of the Msg.
//
// This method retrieves the list of parts that make up the email message. Each part may represent
// a different section of the email, such as a plain text body, HTML body, or attachments.
//
// Returns:
//   - A slice of Part pointers representing the message parts of the email.
func (m *Msg) GetParts() []*Part {
	return m.parts
}

// GetAttachments returns the attachments of the Msg.
//
// This method retrieves the list of files that have been attached to the email message.
// Each attachment includes details about the file, such as its name, content type, and data.
//
// Returns:
//   - A slice of File pointers representing the attachments of the email.
func (m *Msg) GetAttachments() []*File {
	return m.attachments
}

// GetBoundary returns the boundary of the Msg.
//
// This method retrieves the MIME boundary that is used to separate different parts of the message,
// particularly in multipart emails. The boundary helps to differentiate between various sections
// such as plain text, HTML content, and attachments.
//
// NOTE: By default, random MIME boundaries are created. Using a predefined boundary will only
// work with messages that hold a single multipart part. Using a predefined boundary with several
// multipart parts will render the mail unreadable to the mail client.
//
// Returns:
//   - A string representing the boundary of the message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.1
func (m *Msg) GetBoundary() string {
	return m.boundary
}

// ServerResponse returns the server's response after queuing the mail.
//
// This function retrieves the value of m.serverResponse, which typically contains information
// such as the queue ID returned by the mail server once a message has been queued. Unfortunately
// different mail server software returns different server responses, therefore you have to
// parse the output yourself.
//
// Returns:
//   - The server response string, usually containing the queue ID or status.
func (m *Msg) ServerResponse() string {
	return m.serverResponse
}

// SetAttachments sets the attachments of the message.
//
// This method allows you to specify the attachments for the message by providing a slice of File pointers.
// Each file represents an attachment that will be included in the email.
//
// Parameters:
//   - files: A slice of pointers to File structures representing the attachments to set for the message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) SetAttachments(files []*File) {
	m.attachments = files
}

// SetAttachements sets the attachments of the message.
//
// Deprecated: use SetAttachments instead.
func (m *Msg) SetAttachements(files []*File) {
	m.SetAttachments(files)
}

// UnsetAllAttachments unsets the attachments of the message.
//
// This method removes all attachments from the message by setting the attachments to nil, effectively
// clearing any previously set attachments.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) UnsetAllAttachments() {
	m.attachments = nil
}

// GetEmbeds returns the embedded files of the Msg.
//
// This method retrieves the list of files that have been embedded in the message. Embeds are typically
// images or other media files that are referenced directly in the content of the email, such as inline
// images in HTML emails.
//
// Returns:
//   - A slice of pointers to File structures representing the embedded files in the message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) GetEmbeds() []*File {
	return m.embeds
}

// SetEmbeds sets the embedded files of the message.
//
// This method allows you to specify the files to be embedded in the message by providing a slice of File pointers.
// Embedded files, such as images or media, are typically used for inline content in HTML emails.
//
// Parameters:
//   - files: A slice of pointers to File structures representing the embedded files to set for the message.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) SetEmbeds(files []*File) {
	m.embeds = files
}

// UnsetAllEmbeds unsets the embedded files of the message.
//
// This method removes all embedded files from the message by setting the embeds to nil, effectively
// clearing any previously set embedded files.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) UnsetAllEmbeds() {
	m.embeds = nil
}

// UnsetAllParts unsets the embeds and attachments of the message.
//
// This method removes all embedded files and attachments from the message by unsetting both the
// embeds and attachments, effectively clearing all previously set message parts.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) UnsetAllParts() {
	m.UnsetAllAttachments()
	m.UnsetAllEmbeds()
}

// SetBodyString sets the body of the message.
//
// This method sets the body of the message using the provided content type and string content. The body can
// be set as plain text, HTML, or other formats based on the specified content type. Optional part settings
// can be passed through PartOption to customize the message body further.
//
// Parameters:
//   - contentType: The ContentType of the body (e.g., plain text, HTML).
//   - content: The string content to set as the body of the message.
//   - opts: Optional parameters for customizing the body part.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyString(contentType ContentType, content string, opts ...PartOption) {
	buffer := bytes.NewBufferString(content)
	writeFunc := writeFuncFromBuffer(buffer)
	m.SetBodyWriter(contentType, writeFunc, opts...)
}

// SetBodyWriter sets the body of the message.
//
// This method sets the body of the message using a write function, allowing content to be written
// directly to the body. The content type determines the format (e.g., plain text, HTML).
// Optional part settings can be provided via PartOption to customize the body further.
//
// Parameters:
//   - contentType: The ContentType of the body (e.g., plain text, HTML).
//   - writeFunc: A function that writes content to an io.Writer and returns the number of bytes written
//     and an error, if any.
//   - opts: Optional parameters for customizing the body part.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyWriter(
	contentType ContentType, writeFunc func(io.Writer) (int64, error),
	opts ...PartOption,
) {
	p := m.newPart(contentType, opts...)
	p.writeFunc = writeFunc
	m.parts = []*Part{p}
}

// AddAlternativeString sets the alternative body of the message.
//
// This method adds an alternative representation of the message body using the specified content type
// and string content. This is typically used to provide both plain text and HTML versions of the email.
// Optional part settings can be provided via PartOption to further customize the message.
//
// Parameters:
//   - contentType: The content type of the alternative body (e.g., plain text, HTML).
//   - content: The string content to set as the alternative body.
//   - opts: Optional parameters for customizing the alternative body part.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeString(contentType ContentType, content string, opts ...PartOption) {
	buffer := bytes.NewBufferString(content)
	writeFunc := writeFuncFromBuffer(buffer)
	m.AddAlternativeWriter(contentType, writeFunc, opts...)
}

// AddAlternativeWriter sets the alternative body of the message.
//
// This method adds an alternative representation of the message body using a write function, allowing
// content to be written directly to the body. This is typically used to provide different formats, such
// as plain text and HTML. Optional part settings can be provided via PartOption to customize the message part.
//
// Parameters:
//   - contentType: The content type of the alternative body (e.g., plain text, HTML).
//   - writeFunc: A function that writes content to an io.Writer and returns the number of bytes written and
//     an error, if any.
//   - opts: Optional parameters for customizing the alternative body part.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeWriter(
	contentType ContentType, writeFunc func(io.Writer) (int64, error),
	opts ...PartOption,
) {
	part := m.newPart(contentType, opts...)
	part.writeFunc = writeFunc
	m.parts = append(m.parts, part)
}

// AttachFile adds an attachment File to the Msg.
//
// This method attaches a file to the message by specifying the file name. The file is retrieved from the
// filesystem and added to the list of attachments. Optional FileOption parameters can be provided to customize
// the attachment, such as setting its content type or encoding.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - opts: Optional parameters for customizing the attachment.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachFile(name string, opts ...FileOption) {
	file := fileFromFS(name)
	if file == nil {
		return
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
}

// AttachReader adds an attachment File via io.Reader to the Msg.
//
// This method allows you to attach a file to the message using an io.Reader. It reads all data from the
// io.Reader into memory before attaching the file, which may not be suitable for large data sources.
// For larger files, it is recommended to use AttachFile or AttachReadSeeker instead.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - reader: The io.Reader providing the file data to be attached.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the file could not be read from the io.Reader, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachReader(name string, reader io.Reader, opts ...FileOption) error {
	file, err := fileFromReader(name, reader)
	if err != nil {
		return err
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// AttachReadSeeker adds an attachment File via io.ReadSeeker to the Msg.
//
// This method allows you to attach a file to the message using an io.ReadSeeker, which is more efficient
// for larger files compared to AttachReader, as it allows for seeking through the data without needing
// to load the entire content into memory.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - reader: The io.ReadSeeker providing the file data to be attached.
//   - opts: Optional parameters for customizing the attachment.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachReadSeeker(name string, reader io.ReadSeeker, opts ...FileOption) {
	file := fileFromReadSeeker(name, reader)
	m.attachments = m.appendFile(m.attachments, file, opts...)
}

// AttachFromEmbedFS adds an attachment File from an embed.FS to the Msg.
//
// This method allows you to attach a file from an embedded filesystem (embed.FS) to the message.
// The file is retrieved from the provided embed.FS and attached to the email. If the embedded filesystem
// is nil or the file cannot be retrieved, an error will be returned.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - fs: A pointer to the embed.FS from which the file will be retrieved.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the embed.FS is nil or the file cannot be retrieved, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachFromEmbedFS(name string, fs *embed.FS, opts ...FileOption) error {
	if fs == nil {
		return errors.New("embed.FS must not be nil")
	}
	return m.AttachFromIOFS(name, *fs, opts...)
}

// AttachFromIOFS attaches a file from a generic file system to the message.
//
// This function retrieves a file by name from an fs.FS instance, processes it, and appends it to the
// message's attachment collection. Additional file options can be provided for further customization.
//
// Parameters:
//   - name: The name of the file to retrieve from the file system.
//   - iofs: The file system (must not be nil).
//   - opts: Optional file options to customize the attachment process.
//
// Returns:
//   - An error if the file cannot be retrieved, the fs.FS is nil, or any other issue occurs.
func (m *Msg) AttachFromIOFS(name string, iofs fs.FS, opts ...FileOption) error {
	if iofs == nil {
		return errors.New("fs.FS must not be nil")
	}
	file, err := fileFromIOFS(name, iofs)
	if err != nil {
		return err
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// EmbedFile adds an embedded File to the Msg.
//
// This method embeds a file from the filesystem directly into the email message. The embedded file,
// typically an image or media file, can be referenced within the email's content (such as inline in HTML).
// If the file is not found or cannot be loaded, it will not be added.
//
// Parameters:
//   - name: The name of the file to be embedded.
//   - opts: Optional parameters for customizing the embedded file.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedFile(name string, opts ...FileOption) {
	file := fileFromFS(name)
	if file == nil {
		return
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
}

// EmbedReader adds an embedded File from an io.Reader to the Msg.
//
// This method embeds a file into the email message by reading its content from an io.Reader.
// It reads all data into memory before embedding the file, which may not be efficient for large data sources.
// For larger files, it is recommended to use EmbedFile or EmbedReadSeeker instead.
//
// Parameters:
//   - name: The name of the file to be embedded.
//   - reader: The io.Reader providing the file data to be embedded.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the file could not be read from the io.Reader, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedReader(name string, reader io.Reader, opts ...FileOption) error {
	file, err := fileFromReader(name, reader)
	if err != nil {
		return err
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// EmbedReadSeeker adds an embedded File from an io.ReadSeeker to the Msg.
//
// This method embeds a file into the email message by reading its content from an io.ReadSeeker.
// Using io.ReadSeeker allows for more efficient handling of large files since it can seek through the data
// without loading the entire content into memory.
//
// Parameters:
//   - name: The name of the file to be embedded.
//   - reader: The io.ReadSeeker providing the file data to be embedded.
//   - opts: Optional parameters for customizing the embedded file.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedReadSeeker(name string, reader io.ReadSeeker, opts ...FileOption) {
	file := fileFromReadSeeker(name, reader)
	m.embeds = m.appendFile(m.embeds, file, opts...)
}

// EmbedFromEmbedFS adds an embedded File from an embed.FS to the Msg.
//
// This method embeds a file from an embedded filesystem (embed.FS) into the email message. If the
// embedded filesystem is nil or the file cannot be retrieved, an error will be returned.
//
// Parameters:
//   - name: The name of the file to be embedded.
//   - fs: A pointer to the embed.FS from which the file will be retrieved.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the embed.FS is nil or the file cannot be retrieved, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedFromEmbedFS(name string, fs *embed.FS, opts ...FileOption) error {
	if fs == nil {
		return errors.New("embed.FS must not be nil")
	}
	return m.EmbedFromIOFS(name, *fs, opts...)
}

// EmbedFromIOFS embeds a file from a generic file system into the message.
//
// This function retrieves a file by name from an fs.FS instance, processes it, and appends it to the
// message's embed collection. Additional file options can be provided for further customization.
//
// Parameters:
//   - name: The name of the file to retrieve from the file system.
//   - iofs: The file system (must not be nil).
//   - opts: Optional file options to customize the embedding process.
//
// Returns:
//   - An error if the file cannot be retrieved, the fs.FS is nil, or any other issue occurs.
func (m *Msg) EmbedFromIOFS(name string, iofs fs.FS, opts ...FileOption) error {
	if iofs == nil {
		return errors.New("fs.FS must not be nil")
	}
	file, err := fileFromIOFS(name, iofs)
	if err != nil {
		return err
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// Reset resets all headers, body parts, attachments, and embeds of the Msg.
//
// This method clears all address headers, attachments, embeds, generic headers, and body parts of the message.
// However, it preserves the existing encoding, charset, boundary, and other message-level settings.
// Use this method to reset the message content while keeping certain configurations intact.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) Reset() {
	m.addrHeader = make(map[AddrHeader][]*mail.Address)
	m.attachments = nil
	m.embeds = nil
	m.genHeader = make(map[Header][]string)
	m.parts = nil
}

// ApplyMiddlewares applies the list of middlewares to a Msg.
//
// This method sequentially applies each middleware function in the list to the message (in FIFO order).
// The middleware functions can modify the message, such as adding headers or altering its content.
// The message is passed through each middleware in order, and the modified message is returned.
//
// Parameters:
//   - msg: The Msg object to which the middlewares will be applied.
//
// Returns:
//   - The modified Msg after all middleware functions have been applied.
func (m *Msg) applyMiddlewares(msg *Msg) *Msg {
	for _, middleware := range m.middlewares {
		msg = middleware.Handle(msg)
	}
	return msg
}

// WriteTo writes the formatted Msg into the given io.Writer and satisfies the io.WriterTo interface.
//
// This method writes the email message, including its headers, body, and attachments, to the provided
// io.Writer. It applies any middlewares to the message before writing it. The total number of bytes
// written and any error encountered during the writing process are returned.
//
// Parameters:
//   - writer: The io.Writer to which the formatted message will be written.
//
// Returns:
//   - The total number of bytes written.
//   - An error if any occurred during the writing process, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) WriteTo(writer io.Writer) (int64, error) {
	mw := &msgWriter{writer: writer, charset: m.charset, encoder: m.encoder}
	msg := m.applyMiddlewares(m)

	if m.hasSMIME() {
		if err := m.signMessage(); err != nil {
			return 0, err
		}
	}

	mw.writeMsg(msg)
	m.headerCount = 0
	return mw.bytesWritten, mw.err
}

// WriteToSkipMiddleware writes the formatted Msg into the given io.Writer, but skips the specified
// middleware type.
//
// This method writes the email message to the provided io.Writer after applying all middlewares,
// except for the specified middleware type, which will be skipped. It temporarily removes the
// middleware of the given type, writes the message, and then restores the original middleware list.
//
// Parameters:
//   - writer: The io.Writer to which the formatted message will be written.
//   - middleWareType: The MiddlewareType that should be skipped during the writing process.
//
// Returns:
//   - The total number of bytes written.
//   - An error if any occurred during the writing process, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) WriteToSkipMiddleware(writer io.Writer, middleWareType MiddlewareType) (int64, error) {
	var origMiddlewares, middlewares []Middleware
	origMiddlewares = m.middlewares
	for i := range m.middlewares {
		if m.middlewares[i].Type() == middleWareType {
			continue
		}
		middlewares = append(middlewares, m.middlewares[i])
	}
	m.middlewares = middlewares
	mw := &msgWriter{writer: writer, charset: m.charset, encoder: m.encoder}
	mw.writeMsg(m.applyMiddlewares(m))
	m.middlewares = origMiddlewares
	return mw.bytesWritten, mw.err
}

// Write is an alias method to WriteTo for compatibility reasons.
//
// This method provides a backward-compatible way to write the formatted Msg to the provided io.Writer
// by calling the WriteTo method. It writes the email message, including headers, body, and attachments,
// to the io.Writer and returns the number of bytes written and any error encountered.
//
// Parameters:
//   - writer: The io.Writer to which the formatted message will be written.
//
// Returns:
//   - The total number of bytes written.
//   - An error if any occurred during the writing process, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) Write(writer io.Writer) (int64, error) {
	return m.WriteTo(writer)
}

// WriteToFile stores the Msg as a file on disk. It will try to create the given filename,
// and if the file already exists, it will be overwritten.
//
// This method writes the email message, including its headers, body, and attachments, to a file on disk.
// If the file cannot be created or an error occurs during writing, an error is returned.
//
// Parameters:
//   - name: The name of the file to be created or overwritten.
//
// Returns:
//   - An error if the file cannot be created or if writing to the file fails, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) WriteToFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = file.Close() }()
	_, err = m.WriteTo(file)
	if err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}
	return file.Close()
}

// WriteToSendmail returns WriteToSendmailWithCommand with a default sendmail path.
//
// This method sends the email message using the default sendmail path. It calls WriteToSendmailWithCommand
// using the standard SendmailPath. If sending via sendmail fails, an error is returned.
//
// Returns:
//   - An error if sending the message via sendmail fails, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5321
func (m *Msg) WriteToSendmail() error {
	return m.WriteToSendmailWithCommand(SendmailPath)
}

// WriteToSendmailWithCommand returns WriteToSendmailWithContext with a default timeout
// of 5 seconds and a given sendmail path.
//
// This method sends the email message using the provided sendmail path, with a default timeout of 5 seconds.
// It creates a context with the specified timeout and then calls WriteToSendmailWithContext to send the message.
//
// Parameters:
//   - sendmailPath: The path to the sendmail executable to be used for sending the message.
//
// Returns:
//   - An error if sending the message via sendmail fails, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5321
func (m *Msg) WriteToSendmailWithCommand(sendmailPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return m.WriteToSendmailWithContext(ctx, sendmailPath)
}

// WriteToSendmailWithContext opens a pipe to the local sendmail binary and tries to send the
// email through it. It takes a context.Context, the path to the sendmail binary, and additional
// arguments for the sendmail binary as parameters.
//
// This method establishes a pipe to the sendmail executable using the provided context and arguments.
// It writes the email message to the sendmail process via STDIN. If any errors occur during the
// communication with the sendmail binary, they will be captured and returned.
//
// Parameters:
//   - ctx: The context to control the timeout and cancellation of the sendmail process.
//   - sendmailPath: The path to the sendmail executable.
//   - args: Additional arguments for the sendmail binary.
//
// Returns:
//   - An error if sending the message via sendmail fails, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5321
func (m *Msg) WriteToSendmailWithContext(ctx context.Context, sendmailPath string, args ...string) error {
	cmdCtx := exec.CommandContext(ctx, sendmailPath)
	cmdCtx.Args = append(cmdCtx.Args, "-oi", "-t")
	cmdCtx.Args = append(cmdCtx.Args, args...)

	stdErr, err := cmdCtx.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to set STDERR pipe: %w", err)
	}

	stdIn, err := cmdCtx.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to set STDIN pipe: %w", err)
	}
	if stdErr == nil || stdIn == nil {
		return fmt.Errorf("received nil for STDERR or STDIN pipe")
	}

	// Start the execution and write to STDIN
	if err = cmdCtx.Start(); err != nil {
		return fmt.Errorf("could not start sendmail execution: %w", err)
	}
	_, err = m.WriteTo(stdIn)
	if err != nil {
		if !errors.Is(err, syscall.EPIPE) {
			return fmt.Errorf("failed to write mail to buffer: %w", err)
		}
	}

	// Close STDIN and wait for completion or cancellation of the sendmail executable
	if err = stdIn.Close(); err != nil {
		return fmt.Errorf("failed to close STDIN pipe: %w", err)
	}

	// Read the stderr pipe for possible errors
	sendmailErr, err := io.ReadAll(stdErr)
	if err != nil {
		return fmt.Errorf("failed to read STDERR pipe: %w", err)
	}
	if len(sendmailErr) > 0 {
		return fmt.Errorf("sendmail command failed: %s", string(sendmailErr))
	}

	if err = cmdCtx.Wait(); err != nil {
		return fmt.Errorf("sendmail command execution failed: %w", err)
	}

	return nil
}

// NewReader returns a Reader type that satisfies the io.Reader interface.
//
// This method creates a new Reader for the Msg, capturing the current state of the message.
// Any subsequent changes made to the Msg after creating the Reader will not be reflected in the Reader's buffer.
// To reflect these changes in the Reader, you must call Msg.UpdateReader to update the Reader's content with
// the current state of the Msg.
//
// Returns:
//   - A pointer to a Reader, which allows the Msg to be read as a stream of bytes.
//
// IMPORTANT: Any changes made to the Msg after creating the Reader will not be reflected in the Reader unless
// Msg.UpdateReader is called.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) NewReader() *Reader {
	reader := &Reader{}
	buffer := bytes.NewBuffer(nil)
	_, err := m.Write(buffer)
	if err != nil {
		reader.err = fmt.Errorf("failed to write Msg to Reader buffer: %w", err)
	}
	reader.buffer = buffer.Bytes()
	return reader
}

// UpdateReader updates a Reader with the current content of the Msg and resets the
// Reader's position to the start.
//
// This method rewrites the content of the provided Reader to reflect any changes made to the Msg.
// It resets the Reader's position to the beginning and updates the buffer with the latest message content.
//
// Parameters:
//   - reader: A pointer to the Reader that will be updated with the Msg's current content.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) UpdateReader(reader *Reader) {
	buffer := bytes.Buffer{}
	_, err := m.Write(&buffer)
	reader.Reset()
	reader.buffer = buffer.Bytes()
	reader.err = err
}

// HasSendError returns true if the Msg experienced an error during message delivery
// and the sendError field of the Msg is not nil.
//
// This method checks whether the message has encountered a delivery error by verifying if the
// sendError field is populated.
//
// Returns:
//   - A boolean value indicating whether a send error occurred (true if an error is present).
func (m *Msg) HasSendError() bool {
	return m.sendError != nil
}

// SendErrorIsTemp returns true if the Msg experienced a delivery error, and the corresponding
// error was of a temporary nature, meaning it can be retried later.
//
// This method checks whether the encountered sendError is a temporary error that can be retried.
// It uses the errors.As function to determine if the error is of type SendError and checks if
// the error is marked as temporary.
//
// Returns:
//   - A boolean value indicating whether the send error is temporary (true if the error is temporary).
func (m *Msg) SendErrorIsTemp() bool {
	var err *SendError
	if errors.As(m.sendError, &err) && err != nil {
		return err.isTemp
	}
	return false
}

// SendError returns the sendError field of the Msg.
//
// This method retrieves the error that occurred during the message delivery process, if any.
// It returns the sendError field, which holds the error encountered during sending.
//
// Returns:
//   - The error encountered during message delivery, or nil if no error occurred.
func (m *Msg) SendError() error {
	return m.sendError
}

// addAddr adds an additional address to the given addrHeader of the Msg.
//
// This method appends an email address to the specified address header (such as "To", "Cc", or "Bcc")
// without overwriting existing addresses. It first collects the current addresses in the header, then
// adds the new address and updates the header.
//
// Parameters:
//   - header: The AddrHeader (e.g., HeaderTo, HeaderCc) to which the address will be added.
//   - addr: The email address to add to the specified header.
//
// Returns:
//   - An error if the address cannot be added, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func (m *Msg) addAddr(header AddrHeader, addr string) error {
	var addresses []string
	for _, address := range m.addrHeader[header] {
		addresses = append(addresses, address.String())
	}
	addresses = append(addresses, addr)
	return m.SetAddrHeader(header, addresses...)
}

// SignWithKeypair configures the Msg to be signed with S/MIME using RSA or ECDSA public-/private keypair.
//
// This function sets up S/MIME signing for the Msg by associating it with the provided private key,
// certificate, and intermediate certificate.
//
// Parameters:
//   - privateKey: The RSA private key used for signing.
//   - certificate: The x509 certificate associated with the private key.
//   - intermediateCert: An optional intermediate x509 certificate for chain validation.
//
// Returns:
//   - An error if any issue occurred while configuring S/MIME signing; otherwise nil.
func (m *Msg) SignWithKeypair(privateKey crypto.PrivateKey, certificate *x509.Certificate,
	intermediateCert *x509.Certificate,
) error {
	smime, err := newSMIME(privateKey, certificate, intermediateCert)
	if err != nil {
		return err
	}
	m.sMIME = smime
	return nil
}

// SignWithTLSCertificate signs the Msg with the provided *tls.Certificate.
//
// This function configures the Msg for S/MIME signing using the private key and certificates
// from the provided TLS certificate. It supports both RSA and ECDSA private keys.
//
// Parameters:
//   - keyPairTlS: The *tls.Certificate containing the private key and associated certificate chain.
//
// Returns:
//   - An error if any issue occurred during parsing, signing configuration, or unsupported private key type.
func (m *Msg) SignWithTLSCertificate(keyPairTLS *tls.Certificate) error {
	if keyPairTLS == nil {
		return fmt.Errorf("keyPairTLS cannot be nil")
	}

	var intermediateCert *x509.Certificate
	var err error
	if len(keyPairTLS.Certificate) > 1 {
		intermediateCert, err = x509.ParseCertificate(keyPairTLS.Certificate[1])
		if err != nil {
			return fmt.Errorf("failed to parse intermediate certificate: %w", err)
		}
	}

	leafCertificate, err := getLeafCertificate(keyPairTLS)
	if err != nil {
		return fmt.Errorf("failed to get leaf certificate: %w", err)
	}

	switch keyPairTLS.PrivateKey.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		return m.SignWithKeypair(keyPairTLS.PrivateKey, leafCertificate, intermediateCert)
	default:
		return fmt.Errorf("unsupported private key type: %T", keyPairTLS.PrivateKey)
	}
}

// appendFile adds a File to the Msg, either as an attachment or an embed.
//
// This method appends a File to the list of files (attachments or embeds) for the message. It applies
// optional FileOption functions to customize the file properties before adding it. If no files are
// already present, a new list is created.
//
// Parameters:
//   - files: The current list of files (either attachments or embeds).
//   - file: The File to be added.
//   - opts: Optional FileOption functions to customize the file.
//
// Returns:
//   - A slice of File pointers representing the updated list of files.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) appendFile(files []*File, file *File, opts ...FileOption) []*File {
	// Override defaults with optionally provided FileOption functions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(file)
	}

	if files == nil {
		return []*File{file}
	}

	return append(files, file)
}

// encodeString encodes a string based on the configured message encoder and the corresponding
// charset for the Msg.
//
// This method encodes the provided string using the message's charset and encoder settings.
// The encoding ensures that the string is properly formatted according to the message's
// character encoding (e.g., UTF-8, ISO-8859-1).
//
// Parameters:
//   - str: The string to be encoded.
//
// Returns:
//   - The encoded string based on the message's charset and encoder.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2047
func (m *Msg) encodeString(str string) string {
	return m.encoder.Encode(string(m.charset), str)
}

// hasAlt returns true if the Msg has more than one part.
//
// This method checks whether the message contains more than one part, indicating that
// the message has alternative content (e.g., both plain text and HTML parts). It ignores
// any parts marked as deleted and returns true only if more than one valid part exists
// and no PGP type is set.
//
// Returns:
//   - A boolean value indicating whether the message has multiple parts (true if more than one part exists).
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) hasAlt() bool {
	count := 0
	for _, part := range m.parts {
		if !part.isDeleted && !part.smime {
			count++
		}
	}
	return count > 1 && m.pgptype == 0
}

// hasMixed returns true if the Msg has mixed parts.
//
// This method checks whether the message contains mixed content, such as attachments along with
// message parts (e.g., text or HTML). A message is considered to have mixed parts if there are both
// attachments and message parts, or if there are multiple attachments.
//
// Returns:
//   - A boolean value indicating whether the message has mixed parts.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.3
func (m *Msg) hasMixed() bool {
	return m.pgptype == 0 && ((len(m.parts) > 0 && len(m.attachments) > 0) || len(m.attachments) > 1)
}

// hasSMIME determines if the Msg should be signed with S/MIME.
//
// This function checks whether the Msg has SMIME type assigned.
//
// Returns:
//   - true if the Msg has SMIME type assigned and should be signed with S/MIME; otherwise false.
func (m *Msg) hasSMIME() bool {
	return m.sMIME != nil
}

// isSMIMEInProgress checks whether an S/MIME signing operation is currently in progress.
//
// This function verifies if the S/MIME configuration exists and if the signing process is active.
//
// Returns:
//   - true if an S/MIME signing operation is in progress; otherwise false.
func (m *Msg) isSMIMEInProgress() bool {
	return m.sMIME != nil && m.sMIME.inProgress
}

// hasRelated returns true if the Msg has related parts.
//
// This method checks whether the message contains related parts, such as inline embedded files
// (e.g., images) that are referenced within the message body. A message is considered to have
// related parts if there are both message parts and embedded files, or if there are multiple embedded files.
//
// Returns:
//   - A boolean value indicating whether the message has related parts.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2387
func (m *Msg) hasRelated() bool {
	return m.pgptype == 0 && ((len(m.parts) > 0 && len(m.embeds) > 0) || len(m.embeds) > 1)
}

// hasPGPType returns true if the Msg should be treated as a PGP-encoded message.
//
// This method checks whether the message is configured to be treated as a PGP-encoded message by examining
// the pgptype field. If the PGP type is set to a value greater than 0, the message is considered PGP-encoded.
//
// Returns:
//   - A boolean value indicating whether the message is PGP-encoded.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc4880
func (m *Msg) hasPGPType() bool {
	return m.pgptype > 0
}

// newPart returns a new Part for the Msg.
//
// This method creates a new Part for the message with the specified content type,
// using the message's current charset and encoding settings. Optional PartOption
// functions can be applied to customize the Part further.
//
// Parameters:
//   - contentType: The content type for the new Part (e.g., text/plain, text/html).
//   - opts: Optional PartOption functions to customize the Part.
//
// Returns:
//   - A pointer to the newly created Part structure.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) newPart(contentType ContentType, opts ...PartOption) *Part {
	p := &Part{
		contentType: contentType,
		charset:     m.charset,
		encoding:    m.encoding,
	}

	// Override defaults with optionally provided MsgOption functions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(p)
	}

	return p
}

// setEncoder creates a new mime.WordEncoder based on the encoding setting of the message.
//
// This method sets the message's encoder by creating a new mime.WordEncoder that matches the
// current encoding setting (e.g., quoted-printable or base64). The encoder is used to encode
// message headers and body content appropriately.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2047
func (m *Msg) setEncoder() {
	m.encoder = getEncoder(m.encoding)
}

// checkUserAgent checks if a User-Agent or X-Mailer header is set, and if not, sets a default version string.
//
// This method ensures that the message includes a User-Agent and X-Mailer header, unless the noDefaultUserAgent
// flag is set. If neither of these headers is present, a default User-Agent string with the current library
// version is added.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.7
func (m *Msg) checkUserAgent() {
	if m.noDefaultUserAgent {
		return
	}
	_, uaok := m.genHeader[HeaderUserAgent]
	_, xmok := m.genHeader[HeaderXMailer]
	if !uaok && !xmok {
		m.SetUserAgent(fmt.Sprintf("go-mail v%s // https://github.com/wneessen/go-mail",
			VERSION))
	}
}

// addDefaultHeader sets default headers if they haven't been set before.
//
// This method ensures that essential headers such as "Date", "Message-ID", and "MIME-Version" are set
// in the message. If these headers are not already present, they will be set to default values.
// The "Date" and "Message-ID" headers are generated, and the "MIME-Version" is set to the message's current setting.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.1 (Date)
//   - https://datatracker.ietf.org/doc/html/rfc5322#section-3.6.4 (Message-ID)
//   - https://datatracker.ietf.org/doc/html/rfc2045#section-4 (MIME-Version)
func (m *Msg) addDefaultHeader() {
	if _, ok := m.genHeader[HeaderDate]; !ok {
		m.SetDate()
	}
	if _, ok := m.genHeader[HeaderMessageID]; !ok {
		m.SetMessageID()
	}
	m.SetGenHeader(HeaderMIMEVersion, string(m.mimever))
}

// fileFromIOFS returns a File pointer from a given file in the provided fs.FS.
//
// This method retrieves a file from the provided io/fs (fs.FS) and returns a File structure
// that can be used as an attachment or embed in the email message. The file's content is read when
// writing to an io.Writer, and the file is identified by its base name.
//
// Parameters:
//   - name: The name of the file to retrieve from the embedded filesystem.
//   - fs: An instance that satisfies the fs.FS interface
//
// Returns:
//   - A pointer to the File structure representing the embedded file.
//   - An error if the file cannot be opened or read from the embedded filesystem.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromIOFS(name string, iofs fs.FS) (*File, error) {
	if iofs == nil {
		return nil, errors.New("fs.FS is nil")
	}

	_, err := iofs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file from fs.FS: %w", err)
	}
	return &File{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		Writer: func(writer io.Writer) (int64, error) {
			file, ferr := iofs.Open(name)
			if ferr != nil {
				return 0, fmt.Errorf("failed to open file from fs.FS: %w", ferr)
			}
			numBytes, ferr := io.Copy(writer, file)
			if ferr != nil {
				_ = file.Close()
				return numBytes, fmt.Errorf("failed to copy file from fs.FS to io.Writer: %w", ferr)
			}
			return numBytes, file.Close()
		},
	}, nil
}

// fileFromFS returns a File pointer from a given file in the system's file system.
//
// This method retrieves a file from the system's file system and returns a File structure
// that can be used as an attachment or embed in the email message. The file is identified
// by its base name, and its content is read when writing to an io.Writer.
//
// Parameters:
//   - name: The name of the file to retrieve from the system's file system.
//
// Returns:
//   - A pointer to the File structure representing the file from the system's file system.
//   - Nil if the file does not exist or cannot be accessed.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromFS(name string) *File {
	_, err := os.Stat(name)
	if err != nil {
		return nil
	}

	return &File{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		Writer: func(writer io.Writer) (int64, error) {
			file, err := os.Open(name)
			if err != nil {
				return 0, err
			}
			numBytes, err := io.Copy(writer, file)
			if err != nil {
				_ = file.Close()
				return numBytes, fmt.Errorf("failed to copy file to io.Writer: %w", err)
			}
			return numBytes, file.Close()
		},
	}
}

// fileFromReader returns a File pointer from a given io.Reader.
//
// This method reads all data from the provided io.Reader and creates a File structure
// that can be used as an attachment or embed in the email message. The file's content
// is stored in memory and written to an io.Writer when needed.
//
// Parameters:
//   - name: The name of the file to be represented by the reader's content.
//   - reader: The io.Reader from which the file content will be read.
//
// Returns:
//   - A pointer to the File structure representing the content of the io.Reader.
//   - An error if the content cannot be read from the io.Reader.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromReader(name string, reader io.Reader) (*File, error) {
	d, err := io.ReadAll(reader)
	if err != nil {
		return &File{}, err
	}
	byteReader := bytes.NewReader(d)
	return &File{
		Name:   name,
		Header: make(map[string][]string),
		Writer: func(writer io.Writer) (int64, error) {
			readBytes, copyErr := io.Copy(writer, byteReader)
			if copyErr != nil {
				return readBytes, copyErr
			}
			_, copyErr = byteReader.Seek(0, io.SeekStart)
			return readBytes, copyErr
		},
	}, nil
}

// fileFromReadSeeker returns a File pointer from a given io.ReadSeeker.
//
// This method creates a File structure from an io.ReadSeeker, allowing efficient handling of file content
// by seeking and reading from the source without fully loading it into memory. The content is written
// to an io.Writer when needed, and the reader's position is reset to the start after writing.
//
// Parameters:
//   - name: The name of the file to be represented by the io.ReadSeeker.
//   - reader: The io.ReadSeeker from which the file content will be read.
//
// Returns:
//   - A pointer to the File structure representing the content of the io.ReadSeeker.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromReadSeeker(name string, reader io.ReadSeeker) *File {
	return &File{
		Name:   name,
		Header: make(map[string][]string),
		Writer: func(writer io.Writer) (int64, error) {
			readBytes, err := io.Copy(writer, reader)
			if err != nil {
				return readBytes, err
			}
			_, err = reader.Seek(0, io.SeekStart)
			return readBytes, err
		},
	}
}

// getEncoder creates a new mime.WordEncoder based on the encoding setting of the message.
//
// This function returns a mime.WordEncoder based on the specified encoding (e.g., quoted-printable or base64).
// The encoder is used for encoding message headers and body content according to the chosen encoding standard.
//
// Parameters:
//   - enc: The Encoding type for the message (e.g., EncodingQP for quoted-printable or EncodingB64 for base64).
//
// Returns:
//   - A mime.WordEncoder based on the encoding setting.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2047
func getEncoder(enc Encoding) mime.WordEncoder {
	switch enc {
	case EncodingQP:
		return mime.QEncoding
	case EncodingB64:
		return mime.BEncoding
	default:
		return mime.QEncoding
	}
}

// signMessage signs the message with S/MIME and attaches the signature as a new part.
//
// This function removes any existing S/MIME parts to prevent duplicate signatures, renders an
// unsigned version of the message, and then signs it. The resulting signature is appended to the
// message as a new S/MIME signature part.
//
// Returns:
//   - An error if any step in the signing process fails, such as finding the message body position
//     or generating the signature.
func (m *Msg) signMessage() error {
	// To avoid attaching double signatures (i. e. if WriteTo is called multiple times)
	// we remove any present smime part before signing the mail
	parts := make([]*Part, 0)
	for _, part := range m.parts {
		if !part.smime {
			parts = append(parts, part)
		}
	}
	m.parts = parts

	// We need to indicate that we are in the signing process
	m.sMIME.inProgress = true
	defer func() {
		m.sMIME.inProgress = false
	}()

	// We render an unsigned version of the mail into a buffer so we can use it for
	// the S/MIME signature
	buf := bytes.NewBuffer(nil)
	mw := &msgWriter{writer: buf, charset: m.charset, encoder: m.encoder}
	mw.writeMsg(m)

	// Since we only want to sign the message body, we need to find the position within
	// the mail body from where we start reading.
	linecount := 0
	pos := 0
	for linecount < m.headerCount {
		nextIndex := bytes.Index(buf.Bytes()[pos:], []byte("\r\n"))
		if nextIndex == -1 {
			return errors.New("unable to find message body starting index within rendered message")
		}
		pos += nextIndex + 2
		linecount++
	}

	// Sign the message and attach a new smime signature part to the mail
	signedMessage, err := m.sMIME.signMessage(buf.Bytes()[pos:])
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}
	signaturePart := m.newPart(TypeSMIMESigned, WithPartEncoding(EncodingB64), WithSMIMESigning())
	signaturePart.SetContent(signedMessage)
	m.parts = append(m.parts, signaturePart)

	return nil
}

// writeFuncFromBuffer converts a byte buffer into a writeFunc, which is commonly required by go-mail.
//
// This function wraps a byte buffer into a write function that can be used to write the buffer's content
// to an io.Writer. It returns a function that writes the buffer's content to the given writer and returns
// the number of bytes written and any error that occurred during writing.
//
// Parameters:
//   - buffer: A pointer to the bytes.Buffer containing the data to be written.
//
// Returns:
//   - A function that writes the buffer's content to an io.Writer, returning the number of bytes written
//     and any error encountered during the write operation.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc5322
func writeFuncFromBuffer(buffer *bytes.Buffer) func(io.Writer) (int64, error) {
	writeFunc := func(w io.Writer) (int64, error) {
		numBytes, err := w.Write(buffer.Bytes())
		return int64(numBytes), err
	}
	return writeFunc
}

func mailAddressStringWithoutName(addr mail.Address) string {
	addr.Name = ""
	return addr.String()
}
