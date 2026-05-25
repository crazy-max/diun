// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import "strings"

// Header is a type wrapper for a string and represents email header fields in a Msg.
type Header string

// AddrHeader is a type wrapper for a string and represents email address headers fields in a Msg.
type AddrHeader string

// Importance is a type wrapper for an int and represents the level of importance or priority for a Msg.
type Importance int

const (
	// HeaderContentDescription is the "Content-Description" header.
	HeaderContentDescription Header = "Content-Description"

	// HeaderContentDisposition is the "Content-Disposition" header.
	HeaderContentDisposition Header = "Content-Disposition"

	// HeaderContentID is the "Content-ID" header.
	HeaderContentID Header = "Content-ID"

	// HeaderContentLang is the "Content-Language" header.
	HeaderContentLang Header = "Content-Language"

	// HeaderContentLocation is the "Content-Location" header (RFC 2110).
	// https://datatracker.ietf.org/doc/html/rfc2110#section-4.3
	HeaderContentLocation Header = "Content-Location"

	// HeaderContentTransferEnc is the "Content-Transfer-Encoding" header.
	HeaderContentTransferEnc Header = "Content-Transfer-Encoding"

	// HeaderContentType is the "Content-Type" header.
	HeaderContentType Header = "Content-Type"

	// HeaderDate represents the "Date" field.
	// https://datatracker.ietf.org/doc/html/rfc822#section-5.1
	HeaderDate Header = "Date"

	// HeaderDispositionNotificationTo is the MDN header as described in RFC 8098.
	// https://datatracker.ietf.org/doc/html/rfc8098#section-2.1
	HeaderDispositionNotificationTo Header = "Disposition-Notification-To"

	// HeaderImportance represents the "Importance" field.
	HeaderImportance Header = "Importance"

	// HeaderInReplyTo represents the "In-Reply-To" field.
	HeaderInReplyTo Header = "In-Reply-To"

	// HeaderListUnsubscribe is the "List-Unsubscribe" header field.
	HeaderListUnsubscribe Header = "List-Unsubscribe"

	// HeaderListUnsubscribePost is the "List-Unsubscribe-Post" header field.
	HeaderListUnsubscribePost Header = "List-Unsubscribe-Post"

	// HeaderMessageID represents the "Message-ID" field for message identification.
	// https://datatracker.ietf.org/doc/html/rfc1036#section-2.1.5
	HeaderMessageID Header = "Message-ID"

	// HeaderMIMEVersion represents the "MIME-Version" field as per RFC 2045.
	// https://datatracker.ietf.org/doc/html/rfc2045#section-4
	HeaderMIMEVersion Header = "MIME-Version"

	// HeaderOrganization is the "Organization" header field.
	HeaderOrganization Header = "Organization"

	// HeaderPrecedence is the "Precedence" header field.
	HeaderPrecedence Header = "Precedence"

	// HeaderPriority represents the "Priority" field.
	HeaderPriority Header = "Priority"

	// HeaderReferences is the "References" header field.
	HeaderReferences Header = "References"

	// HeaderSubject is the "Subject" header field.
	HeaderSubject Header = "Subject"

	// HeaderUserAgent is the "User-Agent" header field.
	HeaderUserAgent Header = "User-Agent"

	// HeaderXAutoResponseSuppress is the "X-Auto-Response-Suppress" header field.
	HeaderXAutoResponseSuppress Header = "X-Auto-Response-Suppress"

	// HeaderXMailer is the "X-Mailer" header field.
	HeaderXMailer Header = "X-Mailer"

	// HeaderXMSMailPriority is the "X-MSMail-Priority" header field.
	HeaderXMSMailPriority Header = "X-MSMail-Priority"

	// HeaderXPriority is the "X-Priority" header field.
	HeaderXPriority Header = "X-Priority"
)

const (
	// HeaderBcc is the "Blind Carbon Copy" header field.
	HeaderBcc AddrHeader = "Bcc"

	// HeaderCc is the "Carbon Copy" header field.
	HeaderCc AddrHeader = "Cc"

	// HeaderEnvelopeFrom is the envelope FROM header field.
	//
	// It is generally not included in the mail body but only used by the Client for the communication with the
	// SMTP server. If the Msg has no "FROM" address set in the mail body, the msgWriter will try to use the
	// envelope from address, if this has been set for the Msg.
	HeaderEnvelopeFrom AddrHeader = "EnvelopeFrom"

	// HeaderFrom is the "From" header field.
	HeaderFrom AddrHeader = "From"

	// HeaderReplyTo is the "Reply-To" header field.
	HeaderReplyTo AddrHeader = "Reply-To"

	// HeaderTo is the "Recipient" header field.
	HeaderTo AddrHeader = "To"
)

const (
	// ImportanceLow indicates a low level of importance or priority in a Msg.
	ImportanceLow Importance = iota

	// ImportanceNormal indicates a standard level of importance or priority for a Msg.
	ImportanceNormal

	// ImportanceHigh indicates a high level of importance or priority in a Msg.
	ImportanceHigh

	// ImportanceNonUrgent indicates a non-urgent level of importance or priority in a Msg.
	ImportanceNonUrgent

	// ImportanceUrgent indicates an urgent level of importance or priority in a Msg.
	ImportanceUrgent
)

// NumString returns a numerical string representation of the Importance level.
//
// This method maps ImportanceHigh and ImportanceUrgent to "1", while ImportanceNonUrgent and ImportanceLow
// are mapped to "0". Other values return an empty string.
//
// Returns:
//   - A string representing the numerical value of the Importance level ("1" or "0"), or an empty string
//     if the Importance level is unrecognized.
func (i Importance) NumString() string {
	switch i {
	case ImportanceNonUrgent:
		return "0"
	case ImportanceLow:
		return "0"
	case ImportanceHigh:
		return "1"
	case ImportanceUrgent:
		return "1"
	default:
		return ""
	}
}

// XPrioString returns the X-Priority string representation of the Importance level.
//
// This method maps ImportanceHigh and ImportanceUrgent to "1", while ImportanceNonUrgent and ImportanceLow
// are mapped to "5". Other values return an empty string.
//
// Returns:
//   - A string representing the X-Priority value of the Importance level ("1" or "5"), or an empty string
//     if the Importance level is unrecognized.
func (i Importance) XPrioString() string {
	switch i {
	case ImportanceNonUrgent:
		return "5"
	case ImportanceLow:
		return "5"
	case ImportanceHigh:
		return "1"
	case ImportanceUrgent:
		return "1"
	default:
		return ""
	}
}

// String satisfies the fmt.Stringer interface for the Importance type and returns the string
// representation of the Importance level.
//
// This method provides a human-readable string for each Importance level.
//
// Returns:
//   - A string representing the Importance level ("non-urgent", "low", "high", or "urgent"), or an empty
//     string if the Importance level is unrecognized.
func (i Importance) String() string {
	switch i {
	case ImportanceNonUrgent:
		return "non-urgent"
	case ImportanceLow:
		return "low"
	case ImportanceHigh:
		return "high"
	case ImportanceUrgent:
		return "urgent"
	default:
		return ""
	}
}

// String satisfies the fmt.Stringer interface for the Header type and returns the string
// representation of the Header.
//
// Returns:
//   - A string representing the Header.
func (h Header) String() string {
	return string(h)
}

// String satisfies the fmt.Stringer interface for the AddrHeader type and returns the string
// representation of the AddrHeader.
//
// Returns:
//   - A string representing the AddrHeader.
func (a AddrHeader) String() string {
	return string(a)
}

// IsAddrHeader checks if the provided string is an address header. It returns true on a valid AddrHeader
// and false for any other string.
//
// Parameters:
//   - header: The string to check.
func IsAddrHeader(header string) bool {
	addrHeaderList := []string{"bcc", "cc", "envelopefrom", "from", "reply-to", "to"}
	for _, addrHeader := range addrHeaderList {
		if addrHeader == strings.ToLower(header) {
			return true
		}
	}
	return false
}
