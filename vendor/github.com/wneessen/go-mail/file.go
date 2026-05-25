// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"io"
	"net/textproto"
)

// FileOption is a function type used to modify properties of a File
type FileOption func(*File)

// File represents a file with properties such as content type, description, encoding, headers, name, and
// a writer function.
//
// This struct can represent either an attachment or an embedded file in a Msg, and it stores relevant
// metadata such as content type and encoding, as well as a function to write the file's content to an
// io.Writer.
type File struct {
	ContentType ContentType
	Desc        string
	Enc         Encoding
	Header      textproto.MIMEHeader
	Name        string
	Writer      func(w io.Writer) (int64, error)
}

// WithFileContentID sets the "Content-ID" header in the File's MIME headers to the specified ID.
//
// This function updates the File's MIME headers by setting the "Content-ID" to the provided string value,
// allowing the file to be referenced by this ID within the MIME structure.
//
// Parameters:
//   - id: A string representing the content ID to be set in the "Content-ID" header.
//
// Returns:
//   - A FileOption function that updates the File's "Content-ID" header.
func WithFileContentID(id string) FileOption {
	return func(f *File) {
		f.Header.Set(HeaderContentID.String(), id)
	}
}

// WithFileName sets the name of a File to the provided value.
//
// This function assigns the specified name to the File, updating its Name field.
//
// Parameters:
//   - name: A string representing the name to be assigned to the File.
//
// Returns:
//   - A FileOption function that sets the File's name.
func WithFileName(name string) FileOption {
	return func(f *File) {
		f.Name = name
	}
}

// WithFileDescription sets an optional description for the File, which is used in the Content-Description
// header of the MIME output.
//
// This function updates the File's description, allowing an additional text description to be added to
// the MIME headers for the file.
//
// Parameters:
//   - description: A string representing the description to be set in the Content-Description header.
//
// Returns:
//   - A FileOption function that sets the File's description.
func WithFileDescription(description string) FileOption {
	return func(f *File) {
		f.Desc = description
	}
}

// WithFileEncoding sets the encoding type for a File.
//
// This function allows the specification of an encoding type for the file, typically used for attachments
// or embedded files. By default, Base64 encoding should be used, but this function can override the
// default if needed.
//
// Note: Quoted-printable encoding (EncodingQP) must never be used for attachments or embeds. If EncodingQP
// is passed to this function, it will be ignored and the encoding will remain unchanged.
//
// Parameters:
//   - encoding: The Encoding type to be assigned to the File, unless it's EncodingQP.
//
// Returns:
//   - A FileOption function that sets the File's encoding.
func WithFileEncoding(encoding Encoding) FileOption {
	return func(f *File) {
		if encoding == EncodingQP {
			return
		}
		f.Enc = encoding
	}
}

// WithFileContentType sets the content type of the File.
//
// By default, the content type is guessed based on the file type, and if no matching type is identified,
// the default "application/octet-stream" is used. This FileOption allows overriding the guessed content
// type with a specific one if required.
//
// Parameters:
//   - contentType: The ContentType to be assigned to the File.
//
// Returns:
//   - A FileOption function that sets the File's content type.
func WithFileContentType(contentType ContentType) FileOption {
	return func(f *File) {
		f.ContentType = contentType
	}
}

// setHeader sets the value of a specified MIME header field for the File.
//
// This method updates the MIME headers of the File by assigning the provided value to the specified
// header field.
//
// Parameters:
//   - header: The Header field to be updated.
//   - value: A string representing the value to be set for the given header.
func (f *File) setHeader(header Header, value string) {
	f.Header.Set(string(header), value)
}

// getHeader retrieves the value of the specified MIME header field.
//
// This method returns the value of the given header and a boolean indicating whether the header was found
// in the File's MIME headers.
//
// Parameters:
//   - header: The Header field whose value is to be retrieved.
//
// Returns:
//   - A string containing the value of the header.
//   - A boolean indicating whether the header was present (true) or not (false).
func (f *File) getHeader(header Header) (string, bool) {
	v := f.Header.Get(string(header))
	return v, v != ""
}
