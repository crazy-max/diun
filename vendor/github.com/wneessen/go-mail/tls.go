// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

// TLSPolicy is a type wrapper for an int type and describes the different TLS policies we allow.
type TLSPolicy int

const (
	// TLSMandatory requires that the connection to the server is
	// encrypting using STARTTLS. If the server does not support STARTTLS
	// the connection will be terminated with an error.
	TLSMandatory TLSPolicy = iota

	// TLSOpportunistic tries to establish an encrypted connection via the
	// STARTTLS protocol. If the server does not support this, it will fall
	// back to non-encrypted plaintext transmission.
	TLSOpportunistic

	// NoTLS forces the transaction to be not encrypted.
	NoTLS
)

// String satisfies the fmt.Stringer interface for the TLSPolicy type.
//
// This function returns a string representation of the TLSPolicy. It matches the policy
// value to predefined constants and returns the corresponding string. If the policy does
// not match any known values, it returns "UnknownPolicy".
//
// Returns:
//   - A string representing the TLSPolicy.
func (p TLSPolicy) String() string {
	switch p {
	case TLSMandatory:
		return "TLSMandatory"
	case TLSOpportunistic:
		return "TLSOpportunistic"
	case NoTLS:
		return "NoTLS"
	default:
		return "UnknownPolicy"
	}
}
