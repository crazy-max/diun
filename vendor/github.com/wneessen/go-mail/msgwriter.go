// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// MaxHeaderLength defines the maximum line length for a mail header.
	//
	// This constant follows the recommendation of RFC 2047, which suggests a maximum length of 76 characters.
	//
	// References:
	//   - https://datatracker.ietf.org/doc/html/rfc2047
	MaxHeaderLength = 76

	// MaxBodyLength defines the maximum line length for the mail body.
	//
	// This constant follows the recommendation of RFC 2047, which suggests a maximum length of 76 characters.
	//
	// References:
	//   - https://datatracker.ietf.org/doc/html/rfc2047
	MaxBodyLength = 76

	// SingleNewLine represents a single newline character sequence ("\r\n").
	//
	// This constant can be used by the msgWriter to issue a carriage return when writing mail content.
	SingleNewLine = "\r\n"

	// DoubleNewLine represents a double newline character sequence ("\r\n\r\n").
	//
	// This constant can be used by the msgWriter to indicate a new segment of the mail when writing mail content.
	DoubleNewLine = "\r\n\r\n"
)

// msgWriter handles the I/O operations for writing to the io.WriteCloser of the SMTP client.
//
// This struct keeps track of the number of bytes written, the character set used, and the depth of the
// current multipart section. It also handles encoding, error tracking, and managing multipart and part
// writers for constructing the email message body.
type msgWriter struct {
	bytesWritten    int64
	charset         Charset
	depth           int8
	encoder         mime.WordEncoder
	err             error
	multiPartWriter [4]*multipart.Writer
	partWriter      io.Writer
	writer          io.Writer
}

// Write implements the io.Writer interface for msgWriter.
//
// This method writes the provided payload to the underlying writer. It keeps track of the number of bytes
// written and handles any errors encountered during the writing process. If a previous error exists, it
// prevents further writing and returns the error.
//
// Parameters:
//   - payload: A byte slice containing the data to be written.
//
// Returns:
//   - The number of bytes successfully written.
//   - An error if the writing process fails, or if a previous error was encountered.
func (mw *msgWriter) Write(payload []byte) (int, error) {
	if mw.err != nil {
		return 0, fmt.Errorf("failed to write due to previous error: %w", mw.err)
	}

	var n int
	n, mw.err = mw.writer.Write(payload)
	mw.bytesWritten += int64(n)
	return n, mw.err
}

// writeMsg formats the message and writes it to the msgWriter's io.Writer.
//
// This method handles the process of writing the message headers and body content, including handling
// multipart structures (e.g., mixed, related, alternative), PGP types, and attachments/embeds. It sets the
// required headers (e.g., "From", "To", "Cc") and iterates over the message parts, writing them to the
// output writer.
//
// Parameters:
//   - msg: A pointer to the Msg struct containing the message data and headers to be written.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045 (Multipurpose Internet Mail Extensions - MIME)
//   - https://datatracker.ietf.org/doc/html/rfc5322 (Internet Message Format)
func (mw *msgWriter) writeMsg(msg *Msg) {
	msg.addDefaultHeader()
	msg.checkUserAgent()
	mw.writeGenHeader(msg)
	mw.writePreformattedGenHeader(msg)

	// Set the FROM header (or envelope FROM if FROM is empty)
	hasFrom := true
	from, ok := msg.addrHeader[HeaderFrom]
	if !ok || (len(from) == 0 || from == nil) {
		from, ok = msg.addrHeader[HeaderEnvelopeFrom]
		if !ok || (len(from) == 0 || from == nil) {
			hasFrom = false
		}
	}
	if hasFrom && (len(from) > 0 && from[0] != nil) {
		msg.headerCount += mw.writeHeader(Header(HeaderFrom), from[0].String())
	}

	// Set the rest of the address headers
	for _, to := range []AddrHeader{HeaderTo, HeaderCc, HeaderReplyTo} {
		if addresses, ok := msg.addrHeader[to]; ok {
			var val []string
			for _, addr := range addresses {
				if addr == nil {
					continue
				}
				val = append(val, addr.String())
			}
			msg.headerCount += mw.writeHeader(Header(to), val...)
		}
	}

	if msg.hasSMIME() && !msg.isSMIMEInProgress() {
		boundary, err := randomBoundary()
		if err != nil {
			mw.err = err
			return
		}
		mw.startMP(msg, MIMESMIMESigned, boundary)
		mw.writeString(DoubleNewLine)
	}
	if msg.hasMixed() {
		boundary := mw.getMultipartBoundary(msg, MIMEMixed)
		boundary = mw.startMP(msg, MIMEMixed, boundary)
		msg.multiPartBoundary[MIMEMixed] = boundary
		if mw.depth == 1 {
			mw.writeString(DoubleNewLine)
		}
	}
	if msg.hasRelated() {
		boundary := mw.getMultipartBoundary(msg, MIMERelated)
		boundary = mw.startMP(msg, MIMERelated, boundary)
		msg.multiPartBoundary[MIMERelated] = boundary
		if mw.depth == 1 {
			mw.writeString(DoubleNewLine)
		}
	}
	if msg.hasAlt() {
		boundary := mw.getMultipartBoundary(msg, MIMEAlternative)
		boundary = mw.startMP(msg, MIMEAlternative, boundary)
		msg.multiPartBoundary[MIMEAlternative] = boundary
		if mw.depth == 1 {
			mw.writeString(DoubleNewLine)
		}
	}
	if msg.hasPGPType() {
		switch msg.pgptype {
		case PGPEncrypt:
			mw.startMP(msg, `encrypted; protocol="application/pgp-encrypted"`,
				msg.boundary)
		case PGPSignature:
			mw.startMP(msg, `signed; protocol="application/pgp-signature";`,
				msg.boundary)
		default:
		}
		mw.writeString(DoubleNewLine)
	}

	for _, part := range msg.parts {
		if !part.isDeleted && !part.smime {
			mw.writePart(part, msg.charset)
		}
	}

	if msg.hasAlt() {
		mw.stopMP()
	}

	// Add embeds
	mw.addFiles(msg.embeds, false)
	if msg.hasRelated() {
		mw.stopMP()
	}

	// Add attachments
	mw.addFiles(msg.attachments, true)
	if msg.hasMixed() {
		mw.stopMP()
	}

	if msg.hasSMIME() && !msg.isSMIMEInProgress() {
		for _, part := range msg.parts {
			if part.smime {
				mw.writePart(part, msg.charset)
			}
		}
		mw.stopMP()
	}
}

// writeGenHeader writes out all generic headers to the msgWriter.
//
// This function extracts all generic headers from the provided Msg object, sorts them, and writes them
// to the msgWriter in alphabetical order.
//
// Parameters:
//   - msg: The Msg object containing the headers to be written.
func (mw *msgWriter) writeGenHeader(msg *Msg) {
	keys := make([]string, 0, len(msg.genHeader))
	for key := range msg.genHeader {
		keys = append(keys, string(key))
	}

	sort.Strings(keys)
	for _, key := range keys {
		msg.headerCount += mw.writeHeader(Header(key), msg.genHeader[Header(key)]...)
	}
}

// writePreformattedGenHeader writes out all preformatted generic headers to the msgWriter.
//
// This function iterates over all preformatted generic headers from the provided Msg object and writes
// them to the msgWriter in the format "key: value" followed by a newline.
//
// Parameters:
//   - msg: The Msg object containing the preformatted headers to be written.
func (mw *msgWriter) writePreformattedGenHeader(msg *Msg) {
	for key, val := range msg.preformHeader {
		line := fmt.Sprintf("%s: %s%s", key, val, SingleNewLine)
		mw.writeString(line)
		msg.headerCount += strings.Count(line, SingleNewLine)
	}
}

// startMP writes a multipart beginning.
//
// This function initializes a multipart writer for the msgWriter using the specified MIME type and
// boundary. It sets the Content-Type header to indicate the multipart type and writes the boundary
// information. If a boundary is provided, it is set explicitly; otherwise, a default boundary is
// generated. It also handles writing a new part when nested multipart structures are used.
//
// Parameters:
//   - msg: The message whose multipart headers are being written.
//   - mimeType: The MIME type of the multipart content (e.g., "mixed", "alternative").
//   - boundary: The boundary string separating different parts of the multipart message.
//
// Returns:
//   - The multipart boundary string
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (mw *msgWriter) startMP(msg *Msg, mimeType MIMEType, boundary string) string {
	multiPartWriter := multipart.NewWriter(mw)
	if boundary != "" {
		mw.err = multiPartWriter.SetBoundary(boundary)
	}

	contentType := fmt.Sprintf("multipart/%s;\r\n boundary=%s", mimeType,
		multiPartWriter.Boundary())
	mw.multiPartWriter[mw.depth] = multiPartWriter

	// Do not write Content-Type if the header was already written as part of genHeaders.
	if _, ok := msg.genHeader[HeaderContentType]; !ok {
		if mw.depth == 0 {
			mw.writeString(fmt.Sprintf("%s: %s", HeaderContentType, contentType))
		}
		if mw.depth > 0 {
			mw.newPart(map[string][]string{"Content-Type": {contentType}})
		}
	}

	mw.depth++
	return multiPartWriter.Boundary()
}

// stopMP closes the multipart.
//
// This function closes the current multipart writer if there is an active multipart structure.
// It decreases the depth level of multipart nesting.
func (mw *msgWriter) stopMP() {
	if mw.depth > 0 {
		mw.err = mw.multiPartWriter[mw.depth-1].Close()
		mw.depth--
	}
}

// getMultipartBoundary returns the appropriate multipart boundary for the given MIME type.
//
// If the Msg has a predefined boundary, it is returned. Otherwise, the function checks
// for a MIME type-specific boundary in the Msg's multiPartBoundary map. If no boundary
// is found, an empty string is returned.
//
// Parameters:
//   - msg: A pointer to the Msg containing the boundary and MIME type-specific mappings.
//   - mimetype: The MIMEType for which the boundary is being determined.
//
// Returns:
//   - A string representing the multipart boundary, or an empty string if none is found.
func (mw *msgWriter) getMultipartBoundary(msg *Msg, mimetype MIMEType) string {
	if msg.boundary != "" {
		return msg.boundary
	}
	if msg.multiPartBoundary[mimetype] != "" {
		return msg.multiPartBoundary[mimetype]
	}
	return ""
}

// addFiles adds the attachments/embeds file content to the mail body.
//
// This function iterates through the list of files, setting necessary headers for each file,
// including Content-Type, Content-Transfer-Encoding, Content-Disposition, and Content-ID
// (if the file is an embed). It determines the appropriate MIME type for each file based on
// its extension or the provided ContentType. It writes file headers and file content
// to the mail body using the appropriate encoding.
//
// Parameters:
//   - files: A slice of File objects to be added to the mail body.
//   - isAttachment: A boolean indicating whether the files are attachments (true) or embeds (false).
func (mw *msgWriter) addFiles(files []*File, isAttachment bool) {
	for _, file := range files {
		encoding := EncodingB64
		if _, ok := file.getHeader(HeaderContentType); !ok {
			mimeType := mime.TypeByExtension(filepath.Ext(file.Name))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			if file.ContentType != "" {
				mimeType = string(file.ContentType)
			}
			file.setHeader(HeaderContentType, fmt.Sprintf(`%s; name="%s"`, mimeType,
				mw.encoder.Encode(mw.charset.String(), sanitizeFilename(file.Name))))
		}

		if _, ok := file.getHeader(HeaderContentTransferEnc); !ok {
			if file.Enc != "" {
				encoding = file.Enc
			}
			file.setHeader(HeaderContentTransferEnc, string(encoding))
		}

		if file.Desc != "" {
			if _, ok := file.getHeader(HeaderContentDescription); !ok {
				file.setHeader(HeaderContentDescription, mw.encoder.Encode(mw.charset.String(), file.Desc))
			}
		}

		if _, ok := file.getHeader(HeaderContentDisposition); !ok {
			disposition := "inline"
			if isAttachment {
				disposition = "attachment"
			}
			file.setHeader(HeaderContentDisposition, fmt.Sprintf(`%s; filename="%s"`,
				disposition, mw.encoder.Encode(mw.charset.String(), sanitizeFilename(file.Name))))
		}

		if !isAttachment {
			if _, ok := file.getHeader(HeaderContentID); !ok {
				file.setHeader(HeaderContentID, fmt.Sprintf("<%s>", sanitizeFilename(file.Name)))
			}
		}
		if mw.depth == 0 {
			for header, val := range file.Header {
				mw.writeHeader(Header(header), val...)
			}
			mw.writeString(SingleNewLine)
		}
		if mw.depth > 0 {
			mw.newPart(file.Header)
		}

		if mw.err == nil {
			mw.writeBody(file.Writer, encoding)
		}
	}
}

// newPart creates a new MIME multipart io.Writer and sets the partWriter to it.
//
// This function creates a new MIME part using the provided header information and assigns it
// to the partWriter. It interacts with the current multipart writer at the specified depth
// to create the part.
//
// Parameters:
//   - header: A map containing the header fields and their corresponding values for the new part.
func (mw *msgWriter) newPart(header map[string][]string) {
	mw.partWriter, mw.err = mw.multiPartWriter[mw.depth-1].CreatePart(header)
}

// writePart writes the corresponding part to the Msg body.
//
// This function writes a MIME part to the message body, setting the appropriate headers such
// as Content-Type and Content-Transfer-Encoding. It determines the charset for the part,
// either using the part's own charset or a fallback charset if none is specified. If the part
// is at the top level (depth 0), headers are written directly. For nested parts, it creates
// a new MIME part with the provided headers.
//
// Parameters:
//   - part: The Part object containing the data to be written.
//   - charset: The Charset used as a fallback if the part does not specify one.
func (mw *msgWriter) writePart(part *Part, charset Charset) {
	partCharset := part.charset
	if partCharset.String() == "" {
		partCharset = charset
	}

	contentType := fmt.Sprintf("%s; charset=%s", part.contentType, partCharset)
	if part.smime {
		contentType = part.contentType.String()
	}
	contentTransferEnc := part.encoding.String()

	if mw.depth == 0 {
		mw.writeHeader(HeaderContentTransferEnc, contentTransferEnc)
		mw.writeHeader(HeaderContentType, contentType)
		mw.writeString(SingleNewLine)
	}
	if mw.depth > 0 {
		mimeHeader := textproto.MIMEHeader{}
		if part.description != "" {
			mimeHeader.Add(string(HeaderContentDescription), part.description)
		}
		mimeHeader.Add(string(HeaderContentTransferEnc), contentTransferEnc)
		mimeHeader.Add(string(HeaderContentType), contentType)
		mw.newPart(mimeHeader)
	}
	mw.writeBody(part.writeFunc, part.encoding)
}

// writeString writes a string into the msgWriter's io.Writer interface.
//
// This function writes the given string to the msgWriter's underlying writer. It checks for
// existing errors before performing the write operation. It also tracks the number of bytes
// written and updates the bytesWritten field accordingly.
//
// Parameters:
//   - s: The string to be written.
func (mw *msgWriter) writeString(s string) {
	if mw.err != nil {
		return
	}
	var n int
	n, mw.err = io.WriteString(mw.writer, s)
	mw.bytesWritten += int64(n)
}

// writeHeader writes a header into the msgWriter's io.Writer.
//
// This function writes a header key and its associated values to the msgWriter. It ensures
// proper formatting of long headers by inserting line breaks as needed. The header values
// are joined and split into words to ensure compliance with the maximum header length
// (MaxHeaderLength). After processing the header, it is written to the underlying writer.
//
// Parameters:
//   - key: The Header key to be written.
//   - values: A variadic parameter representing the values associated with the header.
func (mw *msgWriter) writeHeader(key Header, values ...string) int {
	lines := 0
	buffer := strings.Builder{}
	charLength := MaxHeaderLength - 2
	buffer.WriteString(string(key))
	charLength -= len(key)
	if len(values) == 0 {
		buffer.WriteString(":\r\n")
		return lines + 1
	}
	buffer.WriteString(": ")
	charLength -= 2

	fullValueStr := strings.Join(values, ", ")
	words := strings.Split(fullValueStr, " ")
	for i, val := range words {
		if charLength-len(val) <= 1 {
			buffer.WriteString(fmt.Sprintf("%s ", SingleNewLine))
			charLength = MaxHeaderLength - 3
		}
		buffer.WriteString(val)
		if i < len(words)-1 {
			buffer.WriteString(" ")
			charLength -= 1
		}
		charLength -= len(val)
	}

	bufferString := buffer.String()
	bufferString = strings.ReplaceAll(bufferString, fmt.Sprintf(" %s", SingleNewLine),
		SingleNewLine)
	mw.writeString(bufferString)
	mw.writeString("\r\n")

	lines += strings.Count(bufferString, SingleNewLine) + 1
	return lines
}

// writeBody writes an io.Reader into an io.Writer using the provided Encoding.
//
// This function writes data from an io.Reader to the underlying writer using a specified
// encoding (quoted-printable, base64, or no encoding). It handles encoding of the content
// and manages writing the encoded data to the appropriate writer, depending on the depth
// (whether the data is part of a multipart structure or not). It also tracks the number
// of bytes written and manages any errors encountered during the process.
//
// Parameters:
//   - writeFunc: A function that writes the body content to the given io.Writer.
//   - encoding: The encoding type to use when writing the content (e.g., base64, quoted-printable).
func (mw *msgWriter) writeBody(writeFunc func(io.Writer) (int64, error), encoding Encoding) {
	var writer io.Writer
	var encodedWriter io.WriteCloser
	var n int64
	var err error
	if mw.depth == 0 {
		writer = mw.writer
	}
	if mw.depth > 0 {
		writer = mw.partWriter
	}
	if writer == nil {
		return
	}
	writeBuffer := bytes.Buffer{}
	lineBreaker := base64LineBreaker{}
	lineBreaker.out = &writeBuffer

	switch encoding {
	case EncodingQP:
		encodedWriter = quotedprintable.NewWriter(&writeBuffer)
	case EncodingB64:
		encodedWriter = base64.NewEncoder(base64.StdEncoding, &lineBreaker)
	case NoEncoding:
		_, err = writeFunc(&writeBuffer)
		if err != nil {
			mw.err = fmt.Errorf("bodyWriter function: %w", err)
		}
		n, err = io.Copy(writer, &writeBuffer)
		if err != nil && mw.err == nil {
			mw.err = fmt.Errorf("bodyWriter io.Copy: %w", err)
		}
		if mw.depth == 0 {
			mw.bytesWritten += n
		}
		return
	default:
		encodedWriter = quotedprintable.NewWriter(writer)
	}

	_, err = writeFunc(encodedWriter)
	if err != nil {
		mw.err = fmt.Errorf("bodyWriter function: %w", err)
	}
	err = encodedWriter.Close()
	if err != nil && mw.err == nil {
		mw.err = fmt.Errorf("bodyWriter close encoded writer: %w", err)
	}
	err = lineBreaker.Close()
	if err != nil && mw.err == nil {
		mw.err = fmt.Errorf("bodyWriter close linebreaker: %w", err)
	}
	n, err = io.Copy(writer, &writeBuffer)
	if err != nil && mw.err == nil {
		mw.err = fmt.Errorf("bodyWriter io.Copy: %w", err)
	}

	// Since the part writer uses the WriteTo() method, we don't need to add the
	// bytes twice
	if mw.depth == 0 {
		mw.bytesWritten += n
	}
}

// sanitizeFilename sanitizes a given filename string by replacing specific unwanted characters with
// an underscore ('_').
//
// This method replaces any control character and any special character that is problematic for
// MIME headers and file systems with an underscore ('_') character.
//
// The following characters are replaced
// - Any control character (US-ASCII < 32)
// - ", /, :, <, >, ?, \, |, [DEL]
//
// Parameters:
//   - input: A string of a filename that is supposed to be sanitized
//
// Returns:
//   - A string representing the sanitized version of the filename
func sanitizeFilename(input string) string {
	var sanitized strings.Builder
	for i := 0; i < len(input); i++ {
		// We do not allow control characters in file names.
		if input[i] < 32 || input[i] == 34 || input[i] == 47 || input[i] == 58 ||
			input[i] == 60 || input[i] == 62 || input[i] == 63 || input[i] == 92 ||
			input[i] == 124 || input[i] == 127 {
			sanitized.WriteRune('_')
			continue
		}
		sanitized.WriteByte(input[i])
	}
	return sanitized.String()
}
