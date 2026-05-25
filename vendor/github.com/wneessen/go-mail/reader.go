// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"io"
)

// Reader is a type that implements the io.Reader interface for a Msg.
//
// This struct represents a reader that reads from a byte slice buffer. It keeps track of the
// current read position (offset) and any initialization error. The buffer holds the data to be
// read from the message.
type Reader struct {
	buffer []byte // contents are the bytes buffer[offset : len(buffer)]
	offset int    // read at &buffer[offset], write at &buffer[len(buffer)]
	err    error  // initialization error
}

// Error returns an error if the Reader err field is not nil.
//
// This function checks the Reader's err field and returns it if it is not nil. If no error
// occurred during initialization, it returns nil.
//
// Returns:
//   - The error stored in the err field, or nil if no error is present.
func (r *Reader) Error() error {
	return r.err
}

// Read reads the content of the Msg buffer into the provided payload to satisfy the io.Reader interface.
//
// This function reads data from the Reader's buffer into the provided byte slice (payload).
// It checks for errors or an empty buffer and resets the Reader if necessary. If no data is available,
// it returns io.EOF. Otherwise, it copies the content from the buffer into the payload and updates
// the read offset.
//
// Parameters:
//   - payload: A byte slice where the data will be copied.
//
// Returns:
//   - n: The number of bytes copied into the payload.
//   - err: An error if any issues occurred during the read operation or io.EOF if the buffer is empty.
func (r *Reader) Read(payload []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.empty() || r.buffer == nil {
		r.Reset()
		if len(payload) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(payload, r.buffer[r.offset:])
	r.offset += n
	return n, err
}

// Reset resets the Reader buffer to be empty, but it retains the underlying storage for future use.
//
// This function clears the Reader's buffer by setting its length to 0 and resets the read offset
// to the beginning. The underlying storage is retained, allowing future writes to reuse the buffer.
func (r *Reader) Reset() {
	r.buffer = r.buffer[:0]
	r.offset = 0
}

// empty reports whether the unread portion of the Reader buffer is empty.
//
// This function checks if the unread portion of the Reader's buffer is empty by comparing
// the length of the buffer to the current read offset.
//
// Returns:
//   - true if the unread portion is empty, false otherwise.
func (r *Reader) empty() bool { return len(r.buffer) <= r.offset }
