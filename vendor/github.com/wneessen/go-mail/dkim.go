// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/wneessen/go-mail/internal/dkim"
)

// DKIMSigner represents the DKIM signing configuration.
//
// This type is an alias for dkim.Signer and holds the configuration used to apply DomainKeys
// Identified Mail (DKIM) signatures to outgoing email messages, including the signing domain,
// selector, private key, and canonicalization settings.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc6376
type DKIMSigner = dkim.Signer

// Canonicalization represents a DKIM canonicalization algorithm.
//
// This type is an alias for dkim.Canonicalization and identifies the method used to normalize the
// header and body of a message before signing or verification. Canonicalization determines how
// tolerant the signature is to modifications introduced in transit, such as whitespace changes.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc6376#section-3.4
type Canonicalization = dkim.Canonicalization

// CanonicalizationSimple and CanonicalizationRelaxed are the supported DKIM canonicalization modes.
//
// These constants enumerate the canonicalization algorithms defined by DKIM. The "simple" algorithm
// tolerates almost no modification to the message, while the "relaxed" algorithm permits common,
// insignificant changes such as whitespace normalization and header case folding.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc6376#section-3.4
const (
	CanonicalizationSimple  Canonicalization = dkim.CanonicalizationSimple
	CanonicalizationRelaxed Canonicalization = dkim.CanonicalizationRelaxed
)

// NewDKIMSigner creates a new DKIMSigner.
//
// This function constructs and returns a DKIMSigner configured to sign outgoing email messages
// using DomainKeys Identified Mail (DKIM). It associates the signer with the provided domain and
// selector, which together locate the public key in DNS, and uses the supplied private key to
// generate cryptographic signatures over message headers and body.
//
// Parameters:
//   - domain: The signing domain published in DNS (the "d=" tag).
//   - selector: The selector used to locate the public key in DNS (the "s=" tag).
//   - privKey: The private key used to produce the DKIM signature.
//
// Returns:
//   - A pointer to a configured DKIMSigner ready to sign messages.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc6376
func NewDKIMSigner(domain, selector string, privKey crypto.Signer) *DKIMSigner {
	return dkim.NewSigner(domain, selector, privKey)
}

// PrivKeyFromPEM decodes a PEM-encoded private key into a crypto.Signer.
//
// This function parses a PEM-encoded RSA or Ed25519 private key block and returns it as a
// crypto.Signer. It accepts both PKCS#1 ("RSA PRIVATE KEY") and PKCS#8 ("PRIVATE KEY") encodings,
// decoding the PEM data, dispatching on the block type, and validating that the contained key is a
// supported type. The returned crypto.Signer can be passed directly to NewDKIMSigner to construct
// a DKIMSigner. It returns an error if no valid PEM data is found, if parsing fails, or if the key
// type is unsupported.
//
// Parameters:
//   - pemBytes: The PEM-encoded private key data, in either PKCS#1 or PKCS#8 format.
//
// Returns:
//   - A crypto.Signer wrapping the decoded private key, suitable for use with NewDKIMSigner.
//   - An error if the PEM data is invalid, cannot be parsed, or contains an unsupported key type.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc8017
//   - https://datatracker.ietf.org/doc/html/rfc5208
//   - https://datatracker.ietf.org/doc/html/rfc8032
func PrivKeyFromPEM(pemBytes []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no valid PEM data found")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		case ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("unsupported key type %T", key)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type %q", block.Type)
	}
}
