// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

// Charset is a type wrapper for a string representing different character encodings.
type Charset string

// ContentType is a type wrapper for a string and represents the MIME type of the content being handled.
type ContentType string

// Encoding is a type wrapper for a string and represents the type of encoding used for email messages
// and/or parts.
type Encoding string

// MIMEVersion is a type wrapper for a string nad represents the MIME version used in email messages.
type MIMEVersion string

// MIMEType is a type wrapper for a string and represents the MIME type for the Msg content or parts.
type MIMEType string

const (
	// EncodingB64 represents the Base64 encoding as specified in RFC 2045.
	//
	// https://datatracker.ietf.org/doc/html/rfc2045#section-6.8
	EncodingB64 Encoding = "base64"

	// EncodingQP represents the "quoted-printable" encoding as specified in RFC 2045.
	//
	// https://datatracker.ietf.org/doc/html/rfc2045#section-6.7
	EncodingQP Encoding = "quoted-printable"

	// EncodingUSASCII represents encoding with only US-ASCII characters (aka 7Bit)
	//
	// https://datatracker.ietf.org/doc/html/rfc2045#section-2.7
	EncodingUSASCII Encoding = "7bit"

	// NoEncoding represents 8-bit encoding for email messages as specified in RFC 6152.
	//
	// https://datatracker.ietf.org/doc/html/rfc2045#section-2.8
	//
	// https://datatracker.ietf.org/doc/html/rfc6152
	NoEncoding Encoding = "8bit"
)

const (
	// CharsetUTF7 represents the "UTF-7" charset.
	CharsetUTF7 Charset = "UTF-7"

	// CharsetUTF8 represents the "UTF-8" charset.
	CharsetUTF8 Charset = "UTF-8"

	// CharsetASCII represents the "US-ASCII" charset.
	CharsetASCII Charset = "US-ASCII"

	// CharsetISO88591 represents the "ISO-8859-1" charset.
	CharsetISO88591 Charset = "ISO-8859-1"

	// CharsetISO88592 represents the "ISO-8859-2" charset.
	CharsetISO88592 Charset = "ISO-8859-2"

	// CharsetISO88593 represents the "ISO-8859-3" charset.
	CharsetISO88593 Charset = "ISO-8859-3"

	// CharsetISO88594 represents the "ISO-8859-4" charset.
	CharsetISO88594 Charset = "ISO-8859-4"

	// CharsetISO88595 represents the "ISO-8859-5" charset.
	CharsetISO88595 Charset = "ISO-8859-5"

	// CharsetISO88596 represents the "ISO-8859-6" charset.
	CharsetISO88596 Charset = "ISO-8859-6"

	// CharsetISO88597 represents the "ISO-8859-7" charset.
	CharsetISO88597 Charset = "ISO-8859-7"

	// CharsetISO88599 represents the "ISO-8859-9" charset.
	CharsetISO88599 Charset = "ISO-8859-9"

	// CharsetISO885913 represents the "ISO-8859-13" charset.
	CharsetISO885913 Charset = "ISO-8859-13"

	// CharsetISO885914 represents the "ISO-8859-14" charset.
	CharsetISO885914 Charset = "ISO-8859-14"

	// CharsetISO885915 represents the "ISO-8859-15" charset.
	CharsetISO885915 Charset = "ISO-8859-15"

	// CharsetISO885916 represents the "ISO-8859-16" charset.
	CharsetISO885916 Charset = "ISO-8859-16"

	// CharsetISO2022JP represents the "ISO-2022-JP" charset.
	CharsetISO2022JP Charset = "ISO-2022-JP"

	// CharsetISO2022KR represents the "ISO-2022-KR" charset.
	CharsetISO2022KR Charset = "ISO-2022-KR"

	// CharsetWindows1250 represents the "windows-1250" charset.
	CharsetWindows1250 Charset = "windows-1250"

	// CharsetWindows1251 represents the "windows-1251" charset.
	CharsetWindows1251 Charset = "windows-1251"

	// CharsetWindows1252 represents the "windows-1252" charset.
	CharsetWindows1252 Charset = "windows-1252"

	// CharsetWindows1255 represents the "windows-1255" charset.
	CharsetWindows1255 Charset = "windows-1255"

	// CharsetWindows1256 represents the "windows-1256" charset.
	CharsetWindows1256 Charset = "windows-1256"

	// CharsetKOI8R represents the "KOI8-R" charset.
	CharsetKOI8R Charset = "KOI8-R"

	// CharsetKOI8U represents the "KOI8-U" charset.
	CharsetKOI8U Charset = "KOI8-U"

	// CharsetBig5 represents the "Big5" charset.
	CharsetBig5 Charset = "Big5"

	// CharsetGB18030 represents the "GB18030" charset.
	CharsetGB18030 Charset = "GB18030"

	// CharsetGB2312 represents the "GB2312" charset.
	CharsetGB2312 Charset = "GB2312"

	// CharsetTIS620 represents the "TIS-620" charset.
	CharsetTIS620 Charset = "TIS-620"

	// CharsetEUCKR represents the "EUC-KR" charset.
	CharsetEUCKR Charset = "EUC-KR"

	// CharsetShiftJIS represents the "Shift_JIS" charset.
	CharsetShiftJIS Charset = "Shift_JIS"

	// CharsetUnknown represents the "Unknown" charset.
	CharsetUnknown Charset = "Unknown"

	// CharsetGBK represents the "GBK" charset.
	CharsetGBK Charset = "GBK"
)

// MIME10 represents the MIME version "1.0" used in email messages.
const MIME10 MIMEVersion = "1.0"

const (
	// TypeAppOctetStream represents the MIME type for arbitrary binary data.
	TypeAppOctetStream ContentType = "application/octet-stream"

	// TypeMultipartAlternative represents the MIME type for a message body that can contain multiple alternative
	// formats.
	TypeMultipartAlternative ContentType = "multipart/alternative"

	// TypeMultipartMixed represents the MIME type for a multipart message containing different parts.
	TypeMultipartMixed ContentType = "multipart/mixed"

	// TypeMultipartRelated represents the MIME type for a multipart message where each part is a related file
	// or resource.
	TypeMultipartRelated ContentType = "multipart/related"

	// TypePGPSignature represents the MIME type for PGP signed messages.
	TypePGPSignature ContentType = "application/pgp-signature"

	// TypePGPEncrypted represents the MIME type for PGP encrypted messages.
	TypePGPEncrypted ContentType = "application/pgp-encrypted"

	// TypeTextHTML represents the MIME type for HTML text content.
	TypeTextHTML ContentType = "text/html"

	// TypeTextPlain represents the MIME type for plain text content.
	TypeTextPlain ContentType = "text/plain"

	// TypeSMIMESigned represents the MIME type for S/MIME singed messages.
	TypeSMIMESigned ContentType = `application/pkcs7-signature; name="smime.p7s"`
)

const (
	// MIMEAlternative MIMEType represents a MIME multipart/alternative type, used for emails with multiple versions.
	MIMEAlternative MIMEType = "alternative"

	// MIMEMixed MIMEType represents a MIME multipart/mixed type used fork emails containing different types of content.
	MIMEMixed MIMEType = "mixed"

	// MIMERelated MIMEType represents a MIME multipart/related type, used for emails with related content entities.
	MIMERelated MIMEType = "related"

	// MIMESMIMESigned MIMEType represents a MIME multipart/signed type, used for siging emails with S/MIME.
	MIMESMIMESigned MIMEType = `signed; protocol="application/pkcs7-signature"; micalg=sha-256`
)

// String satisfies the fmt.Stringer interface for the Charset type.
// It converts a Charset into a printable format.
//
// This method returns the string representation of the Charset, allowing it to be easily
// printed or logged.
//
// Returns:
//   - A string representation of the Charset.
func (c Charset) String() string {
	return string(c)
}

// String satisfies the fmt.Stringer interface for the ContentType type.
// It converts a ContentType into a printable format.
//
// This method returns the string representation of the ContentType, enabling its use
// in formatted output such as logging or displaying information to the user.
//
// Returns:
//   - A string representation of the ContentType.
func (c ContentType) String() string {
	return string(c)
}

// String satisfies the fmt.Stringer interface for the Encoding type.
// It converts an Encoding into a printable format.
//
// This method returns the string representation of the Encoding, which can be used
// for displaying or logging purposes.
//
// Returns:
//   - A string representation of the Encoding.
func (e Encoding) String() string {
	return string(e)
}

// String satisfies the fmt.Stringer interface for the MIMEType type.
// It converts an MIMEType into a printable format.
//
// This method returns the string representation of the MIMEType, which can be used
// for displaying or logging purposes.
//
// Returns:
//   - A string representation of the MIMEType.
func (e MIMEType) String() string {
	return string(e)
}
