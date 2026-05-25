// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"bytes"
	"io"
)

// PartOption returns a function that can be used for grouping Part options
type PartOption func(*Part)

// Part is a part of the Msg.
//
// This struct represents a single part of a multipart message. Each part has a content type,
// charset, optional description, encoding, and a function to write its content to an io.Writer.
// It also includes a flag to mark the part as deleted.
type Part struct {
	contentType ContentType
	charset     Charset
	description string
	encoding    Encoding
	isDeleted   bool
	writeFunc   func(io.Writer) (int64, error)
	smime       bool
}

// GetContent executes the WriteFunc of the Part and returns the content as a byte slice.
//
// This function runs the part's writeFunc to write its content into a buffer and then returns
// the content as a byte slice. If an error occurs during the writing process, it is returned.
//
// Returns:
//   - A byte slice containing the part's content.
//   - An error if the writeFunc encounters an issue.
func (p *Part) GetContent() ([]byte, error) {
	var b bytes.Buffer
	if _, err := p.writeFunc(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GetCharset returns the currently set Charset of the Part.
//
// This function returns the Charset that is currently set for the Part.
//
// Returns:
//   - The Charset of the Part.
func (p *Part) GetCharset() Charset {
	return p.charset
}

// GetContentType returns the currently set ContentType of the Part.
//
// This function returns the ContentType that is currently set for the Part.
//
// Returns:
//   - The ContentType of the Part.
func (p *Part) GetContentType() ContentType {
	return p.contentType
}

// GetEncoding returns the currently set Encoding of the Part.
//
// This function returns the Encoding that is currently set for the Part.
//
// Returns:
//   - The Encoding of the Part.
func (p *Part) GetEncoding() Encoding {
	return p.encoding
}

// GetWriteFunc returns the currently set WriteFunc of the Part.
//
// This function returns the WriteFunc that is currently set for the Part, which writes
// the part's content to an io.Writer.
//
// Returns:
//   - The WriteFunc of the Part, which is a function that takes an io.Writer and returns
//     the number of bytes written and an error (if any).
func (p *Part) GetWriteFunc() func(io.Writer) (int64, error) {
	return p.writeFunc
}

// GetDescription returns the currently set Content-Description of the Part.
//
// This function returns the Content-Description that is currently set for the Part.
//
// Returns:
//   - The Content-Description of the Part as a string.
func (p *Part) GetDescription() string {
	return p.description
}

// SetContent overrides the content of the Part with the given string.
//
// This function sets the content of the Part by creating a new writeFunc that writes the
// provided string content to an io.Writer.
//
// Parameters:
//   - content: The string that will replace the current content of the Part.
func (p *Part) SetContent(content string) {
	buffer := bytes.NewBufferString(content)
	p.writeFunc = writeFuncFromBuffer(buffer)
}

// SetContentType overrides the ContentType of the Part.
//
// This function sets a new ContentType for the Part, replacing the existing one.
//
// Parameters:
//   - contentType: The new ContentType to be set for the Part.
func (p *Part) SetContentType(contentType ContentType) {
	p.contentType = contentType
}

// SetCharset overrides the Charset of the Part.
//
// This function sets a new Charset for the Part, replacing the existing one.
//
// Parameters:
//   - charset: The new Charset to be set for the Part.
func (p *Part) SetCharset(charset Charset) {
	p.charset = charset
}

// SetEncoding creates a new mime.WordEncoder based on the encoding setting of the message.
//
// This function sets a new Encoding for the Part, replacing the existing one.
//
// Parameters:
//   - encoding: The new Encoding to be set for the Part.
func (p *Part) SetEncoding(encoding Encoding) {
	p.encoding = encoding
}

// SetDescription overrides the Content-Description of the Part.
//
// This function sets a new Content-Description for the Part, replacing the existing one.
//
// Parameters:
//   - description: The new Content-Description to be set for the Part.
func (p *Part) SetDescription(description string) {
	p.description = description
}

// SetIsSMIMESigned sets the flag for signing the Part with S/MIME.
//
// This function updates the S/MIME signing flag for the Part.
//
// Parameters:
//   - smime: A boolean indicating whether the Part should be signed with S/MIME.
func (p *Part) SetIsSMIMESigned(smime bool) {
	p.smime = smime
}

// SetWriteFunc overrides the WriteFunc of the Part.
//
// This function sets a new WriteFunc for the Part, replacing the existing one. The WriteFunc
// is responsible for writing the Part's content to an io.Writer.
//
// Parameters:
//   - writeFunc: A function that writes the Part's content to an io.Writer and returns
//     the number of bytes written and an error (if any).
func (p *Part) SetWriteFunc(writeFunc func(io.Writer) (int64, error)) {
	p.writeFunc = writeFunc
}

// Delete removes the current part from the parts list of the Msg by setting the isDeleted flag to true.
//
// This function marks the Part as deleted by setting the isDeleted flag to true. The msgWriter
// will skip over this Part during processing.
func (p *Part) Delete() {
	p.isDeleted = true
}

// WithPartCharset overrides the default Part charset.
//
// This function returns a PartOption that allows the charset of a Part to be overridden
// with the specified Charset.
//
// Parameters:
//   - charset: The Charset to be set for the Part.
//
// Returns:
//   - A PartOption function that sets the Part's charset.
func WithPartCharset(charset Charset) PartOption {
	return func(p *Part) {
		p.charset = charset
	}
}

// WithPartEncoding overrides the default Part encoding.
//
// This function returns a PartOption that allows the encoding of a Part to be overridden
// with the specified Encoding.
//
// Parameters:
//   - encoding: The Encoding to be set for the Part.
//
// Returns:
//   - A PartOption function that sets the Part's encoding.
func WithPartEncoding(encoding Encoding) PartOption {
	return func(p *Part) {
		p.encoding = encoding
	}
}

// WithPartContentDescription overrides the default Part Content-Description.
//
// This function returns a PartOption that allows the Content-Description of a Part
// to be overridden with the specified description.
//
// Parameters:
//   - description: The Content-Description to be set for the Part.
//
// Returns:
//   - A PartOption function that sets the Part's Content-Description.
func WithPartContentDescription(description string) PartOption {
	return func(p *Part) {
		p.description = description
	}
}

// WithSMIMESigning enables the S/MIME signing flag for a Part.
//
// This function provides a PartOption that overrides the S/MIME signing flag to enable signing.
//
// Returns:
//   - A PartOption that sets the S/MIME signing flag to true.
func WithSMIMESigning() PartOption {
	return func(p *Part) {
		p.smime = true
	}
}
