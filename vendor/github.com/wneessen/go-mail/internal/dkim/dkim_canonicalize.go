// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package dkim

import (
	"bytes"
	"strings"
)

// canonicalizeHeader canonicalizes a single header line according to the
// specified canonicalization mode
func canonicalizeHeader(line string, mode Canonicalization) []byte {
	if mode == CanonicalizationSimple {
		return []byte(line)
	}
	name, val, ok := strings.Cut(line, ":")
	if !ok {
		return []byte(collapseWSP(unfold(line)))
	}
	name = strings.ToLower(strings.TrimSpace(name))
	val = strings.TrimSpace(collapseWSP(unfold(val)))
	return []byte(name + ":" + val)
}

// unfold removes CRLF folding from a header value
func unfold(s string) string {
	r := strings.NewReplacer("\r\n", "", "\n", "", "\r", "")
	return r.Replace(s)
}

// collapseWSP collapses consecutive whitespace characters into a
// single space
func collapseWSP(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	in := false
	for i := 0; i < len(value); i++ {
		if value[i] == ' ' || value[i] == '\t' {
			in = true
			continue
		}
		if in {
			builder.WriteByte(' ')
			in = false
		}
		builder.WriteByte(value[i])
	}
	if in {
		builder.WriteByte(' ')
	}
	return builder.String()
}

// canonicalizeBody canonicalizes the body of a message according to the specified
// canonicalization method
func canonicalizeBody(body []byte, mode Canonicalization) []byte {
	body = normalizeCRLF(body)

	// Simple canonicalization
	if mode == CanonicalizationSimple {
		for bytes.HasSuffix(body, []byte("\r\n\r\n")) {
			body = body[:len(body)-2]
		}
		if len(body) == 0 {
			return []byte("\r\n")
		}
		if !bytes.HasSuffix(body, []byte("\r\n")) {
			body = append(body, '\r', '\n')
		}
		return body
	}

	// Relaxed canonicalization
	out := bytes.NewBuffer(nil)
	for line := range bytes.SplitSeq(body, []byte("\r\n")) {
		cl := bytes.TrimRight(collapseWSPBytes(line), " \t")
		out.Write(cl)
		out.WriteString("\r\n")
	}
	return trimTrailingEmptyLines(out.Bytes())
}

// normalizeCRLF normalizes CRLF line endings in the given data to "\r\n"
func normalizeCRLF(data []byte) []byte {
	out := bytes.NewBuffer(nil)
	out.Grow(len(data))
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '\r':
			out.WriteString("\r\n")
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
		case '\n':
			out.WriteString("\r\n")
		default:
			out.WriteByte(data[i])
		}
	}
	return out.Bytes()
}

// trimTrailingEmptyLines removes only complete trailing CRLF pairs
// It never touches any lone CRLFs (unlike bytes.TrimRight)
// It will leave one CRLF for a non-empty body and "" for an empty one.
func trimTrailingEmptyLines(data []byte) []byte {
	for len(data) >= 4 &&
		data[len(data)-4] == '\r' && data[len(data)-3] == '\n' &&
		data[len(data)-2] == '\r' && data[len(data)-1] == '\n' {
		data = data[:len(data)-2]
	}
	if len(data) == 2 && data[0] == '\r' && data[1] == '\n' {
		return data[:0]
	}
	return data
}

// collapseWSPBytes collapses consecutive whitespace characters into a single space
func collapseWSPBytes(data []byte) []byte {
	out := bytes.NewBuffer(nil)
	out.Grow(len(data))
	in := false
	for i := range len(data) {
		if data[i] == ' ' || data[i] == '\t' {
			in = true
			continue
		}
		if in {
			out.WriteByte(' ')
			in = false
		}
		out.WriteByte(data[i])
	}
	if in {
		out.WriteByte(' ')
	}
	return out.Bytes()
}
