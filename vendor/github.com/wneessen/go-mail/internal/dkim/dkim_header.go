// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package dkim

import "strings"

const (
	maxLineLength = 76
)

// headerStore holds parsed DKIM headers, keyed by lowercased field name.
type headerStore struct {
	byName map[string][]string
}

// parseHeader parses the raw header block into a headerStore
func parseHeaders(raw []byte) *headerStore {
	store := &headerStore{
		byName: map[string][]string{},
	}
	var cur strings.Builder
	for line := range strings.SplitSeq(string(raw), "\r\n") {
		if line == "" {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			cur.WriteString("\r\n")
			cur.WriteString(line)
			continue
		}
		store.add(cur.String())
		cur.Reset()
		cur.WriteString(line)
	}
	store.add(cur.String())
	return store
}

// add adds a parsed line to the headerStore
func (vals *headerStore) add(line string) {
	if line == "" {
		return
	}
	if key, _, ok := strings.Cut(line, ":"); ok {
		key = strings.ToLower(strings.TrimSpace(key))
		vals.byName[key] = append(vals.byName[key], line)
	}
}

// pop removes and returns the bottom-most occurrence, enabling oversigning
// RFC 6376 signs duplicates bottom-up
//
// See: https://www.rfc-editor.org/info/rfc6376/#section-4.2
func (hs *headerStore) pop(name string) (string, bool) {
	key := strings.ToLower(name)
	vals := hs.byName[key]
	if len(vals) == 0 {
		return "", false
	}
	last := vals[len(vals)-1]
	hs.byName[key] = vals[:len(vals)-1]
	return last, true
}

// writeFolded writes data to builder, inserting CRLF+SPACE whenever the current
// line would exceed maxLineLength, so even a single long token (b=...) gets
// wrapped. It returns the new current line's length
func writeFolded(builder *strings.Builder, data string, length int) int {
	for len(data) > 0 {
		room := maxLineLength - length
		if room < 1 {
			builder.WriteString("\r\n ")
			length = 1
			room = maxLineLength - length
		}
		if len(data) <= room {
			builder.WriteString(data)
			length += len(data)
			return length
		}
		builder.WriteString(data[:room])
		builder.WriteString("\r\n ")
		length = 1
		data = data[room:]
	}
	return length
}

// foldHeader wraps the value to stay within the ~78-char line limit
// (RFC 5322 2.1.1). Tags are broken at "; " boundaries; any single token that
// is still too long (notably the b= base64 blob) is hard-wrapped mid-string.
// Verifiers strip folding whitespace from tag values (and b= specifically)
// before verifying, so folding anywhere is safe.
//
// See: https://www.rfc-editor.org/info/rfc5322/#section-2.1.1
func foldHeader(value string) string {
	var builder strings.Builder
	length := len("DKIM-Signature: ")

	for i, part := range strings.SplitAfter(value, "; ") {
		// Start a new line before a tag if it won't fit and we're not already
		// at the beginning of a folded line.
		if i > 0 && length > 1 && length+len(part) > maxLineLength {
			builder.WriteString("\r\n ")
			length = 1
		}
		length = writeFolded(&builder, part, length)
	}
	return builder.String()
}

// appendFoldedBase64 appends the base64 signature immediately after the
// trailing "b=" of foldedTags, wrapping onto CRLF+HTAB continuation lines as
// needed. foldedTags is emitted verbatim, so a simple-canon verifier — which
// strips the b= value (including this folding) before hashing — reconstructs
// exactly the bytes that were signed.
func appendFoldedBase64(foldedTags, sig string) string {
	var builder strings.Builder
	builder.WriteString(foldedTags)

	// Column of the current (last) line. foldedTags has no "DKIM-Signature: "
	// prefix, so account for it only when foldedTags is still on line one.
	col := len("DKIM-Signature: ") + len(foldedTags)
	if i := strings.LastIndex(foldedTags, "\r\n"); i >= 0 {
		col = len(foldedTags) - (i + 2) // bytes after last CRLF (tab counts as 1)
	}

	for len(sig) > 0 {
		room := maxLineLength - col
		if room < 1 {
			builder.WriteString("\r\n ")
			col = 1
			room = maxLineLength - col
		}
		if len(sig) <= room {
			builder.WriteString(sig)
			return builder.String()
		}
		builder.WriteString(sig[:room])
		builder.WriteString("\r\n ")
		col = 1
		sig = sig[room:]
	}
	return builder.String()
}
