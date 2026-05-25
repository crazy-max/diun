// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// List of SendError reasons
const (
	// ErrGetSender is returned if the Msg.GetSender method fails during a Client.Send
	ErrGetSender SendErrReason = iota

	// ErrGetRcpts is returned if the Msg.GetRecipients method fails during a Client.Send
	ErrGetRcpts

	// ErrSMTPMailFrom is returned if the Msg delivery failed when sending the MAIL FROM command
	// to the sending SMTP server
	ErrSMTPMailFrom

	// ErrSMTPRcptTo is returned if the Msg delivery failed when sending the RCPT TO command
	// to the sending SMTP server
	ErrSMTPRcptTo

	// ErrSMTPData is returned if the Msg delivery failed when sending the DATA command
	// to the sending SMTP server
	ErrSMTPData

	// ErrSMTPDataClose is returned if the Msg delivery failed when trying to close the
	// Client data writer
	ErrSMTPDataClose

	// ErrSMTPReset is returned if the Msg delivery failed when sending the RSET command
	// to the sending SMTP server
	ErrSMTPReset

	// ErrWriteContent is returned if the Msg delivery failed when sending Msg content
	// to the Client writer
	ErrWriteContent

	// ErrConnCheck is returned if the Msg delivery failed when checking if the SMTP
	// server connection is still working
	ErrConnCheck

	// ErrNoUnencoded is returned if the Msg delivery failed when the Msg is configured for
	// unencoded delivery but the server does not support this
	ErrNoUnencoded

	// ErrAmbiguous is a generalized delivery error for the SendError type that is
	// returned if the exact reason for the delivery failure is ambiguous
	ErrAmbiguous
)

// SendError is an error wrapper for delivery errors of the Msg.
//
// This struct represents an error that occurs during the delivery of a message. It holds
// details about the affected message, a list of errors, the recipient list, and whether
// the error is temporary or permanent. It also includes a reason code for the error.
type SendError struct {
	affectedMsg        *Msg
	errcode            int
	enhancedStatusCode string
	errlist            []error
	isTemp             bool
	rcpt               []string
	Reason             SendErrReason
}

// SendErrReason represents a comparable reason on why the delivery failed
type SendErrReason int

// Error implements the error interface for the SendError type.
//
// This function returns a detailed error message string for the SendError, including the
// reason for failure, list of errors, affected recipients, and the message ID of the
// affected message (if available). If the reason is unknown (greater than 10), it returns
// "unknown reason". The error message is built dynamically based on the content of the
// error list, recipient list, and message ID.
//
// Returns:
//   - A string representing the error message.
func (e *SendError) Error() string {
	if e.Reason > ErrAmbiguous {
		return "unknown reason"
	}

	var errMessage strings.Builder
	errMessage.WriteString(e.Reason.String())
	if len(e.errlist) > 0 {
		errMessage.WriteRune(':')
		for i := range e.errlist {
			errMessage.WriteRune(' ')
			errMessage.WriteString(e.errlist[i].Error())
			if i != len(e.errlist)-1 {
				errMessage.WriteString(",")
			}
		}
	}
	if len(e.rcpt) > 0 {
		errMessage.WriteString(", affected recipient(s): ")
		for i := range e.rcpt {
			errMessage.WriteString(e.rcpt[i])
			if i != len(e.rcpt)-1 {
				errMessage.WriteString(", ")
			}
		}
	}
	if e.affectedMsg != nil && e.affectedMsg.GetMessageID() != "" {
		errMessage.WriteString(", affected message ID: ")
		errMessage.WriteString(e.affectedMsg.GetMessageID())
	}

	return errMessage.String()
}

// Is implements the errors.Is functionality and compares the SendErrReason.
//
// This function allows for comparison between two errors by checking if the provided
// error matches the SendError type and, if so, compares the SendErrReason and the
// temporary status (isTemp) of both errors.
//
// Parameters:
//   - errType: The error to compare against the current SendError.
//
// Returns:
//   - true if the errors have the same reason and temporary status, false otherwise.
func (e *SendError) Is(errType error) bool {
	var t *SendError
	if errors.As(errType, &t) && t != nil {
		return e.Reason == t.Reason && e.isTemp == t.isTemp
	}
	return false
}

// IsTemp returns true if the delivery error is of a temporary nature and can be retried.
//
// This function checks whether the SendError indicates a temporary error, which suggests
// that the delivery can be retried. If the SendError is nil, it returns false.
//
// Returns:
//   - true if the error is temporary, false otherwise.
func (e *SendError) IsTemp() bool {
	if e == nil {
		return false
	}
	return e.isTemp
}

// MessageID returns the message ID of the affected Msg that caused the error.
//
// This function retrieves the message ID of the Msg associated with the SendError.
// If no message ID was set or if the SendError or Msg is nil, it returns an empty string.
//
// Returns:
//   - The message ID as a string, or an empty string if no ID is available.
func (e *SendError) MessageID() string {
	if e == nil || e.affectedMsg == nil {
		return ""
	}
	return e.affectedMsg.GetMessageID()
}

// Msg returns the pointer to the affected message that caused the error.
//
// This function retrieves the Msg associated with the SendError. If the SendError or
// the affectedMsg is nil, it returns nil.
//
// Returns:
//   - A pointer to the Msg that caused the error, or nil if not available.
func (e *SendError) Msg() *Msg {
	if e == nil || e.affectedMsg == nil {
		return nil
	}
	return e.affectedMsg
}

// EnhancedStatusCode returns the enhanced status code of the server response if the
// server supports it, as described in RFC 2034.
//
// This function retrieves the enhanced status code of an error returned by the server. This
// requires that the receiving server supports this SMTP extension as described in RFC 2034.
// Since this is the SendError interface, we only collect status codes for error responses,
// meaning 4xx or 5xx. If the server does not support the ENHANCEDSTATUSCODES extension or
// the error did not include an enhanced status code, it will return an empty string.
//
// Returns:
//   - The enhanced status code as returned by the server, or an empty string is not supported.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2034
func (e *SendError) EnhancedStatusCode() string {
	if e == nil {
		return ""
	}
	return e.enhancedStatusCode
}

// ErrorCode returns the error code of the server response.
//
// This function retrieves the error code the error returned by the server. The error code will
// start with 5 on permanent errors and with 4 on a temporary error. If the error is not returned
// by the server, but is generated by go-mail, the code will be 0.
//
// Returns:
//   - The error code as returned by the server, or 0 if not a server error.
func (e *SendError) ErrorCode() int {
	if e == nil {
		return 0
	}
	return e.errcode
}

// String satisfies the fmt.Stringer interface for the SendErrReason type.
//
// This function converts the SendErrReason into a human-readable string representation based
// on the error type. If the error reason does not match any predefined case, it returns
// "unknown reason".
//
// Returns:
//   - A string representation of the SendErrReason.
func (r SendErrReason) String() string {
	switch r {
	case ErrGetSender:
		return "getting sender address"
	case ErrGetRcpts:
		return "getting recipient addresses"
	case ErrSMTPMailFrom:
		return "sending SMTP MAIL FROM command"
	case ErrSMTPRcptTo:
		return "sending SMTP RCPT TO command"
	case ErrSMTPData:
		return "sending SMTP DATA command"
	case ErrSMTPDataClose:
		return "closing SMTP DATA writer"
	case ErrSMTPReset:
		return "sending SMTP RESET command"
	case ErrWriteContent:
		return "sending message content"
	case ErrConnCheck:
		return "checking SMTP connection"
	case ErrNoUnencoded:
		return ErrServerNoUnencoded.Error()
	case ErrAmbiguous:
		return "ambiguous reason, check Msg.SendError for message specific reasons"
	}
	return "unknown reason"
}

// isTempError checks if the given SMTP error is of a temporary nature and should be retried.
//
// This function inspects the error message and returns true if the first character of the
// error message is '4', indicating a temporary SMTP error that can be retried.
//
// Parameters:
//   - err: The error to check.
//
// Returns:
//   - true if the error is temporary, false otherwise.
func isTempError(err error) bool {
	return err.Error()[0] == '4'
}

func errorCode(err error) int {
	rootErr := errors.Unwrap(err)
	if rootErr != nil {
		err = rootErr
	}
	firstrune := err.Error()[0]
	if firstrune < 52 || firstrune > 53 {
		return 0
	}
	code := err.Error()[0:3]
	errcode, cerr := strconv.Atoi(code)
	if cerr != nil {
		return 0
	}
	return errcode
}

func enhancedStatusCode(err error, supported bool) string {
	if err == nil || !supported {
		return ""
	}
	rootErr := errors.Unwrap(err)
	if rootErr != nil {
		err = rootErr
	}
	firstrune := err.Error()[0]
	if firstrune != 50 && firstrune != 52 && firstrune != 53 {
		return ""
	}
	re, rerr := regexp.Compile(`\b([245])\.\d{1,3}\.\d{1,3}\b`)
	if rerr != nil {
		return ""
	}
	return re.FindString(err.Error())
}
