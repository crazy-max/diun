// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

// Package dkim implements DKIM signing and verification according to RFC 6376. It provides
// support for both simple and relaxed canonicalization algorithms as well as support for
// RSA and ED25519 signature algorithms.
//
// This package only supports RSA-SHA256 and ED25519-SHA256 signature algorithms. RSA-SHA1
// is not supported since it's deprecated and prohibited by RFC 8301.
//
// See: https://datatracker.ietf.org/doc/html/rfc6376
// and: https://datatracker.ietf.org/doc/html/rfc8301#section-3.1
package dkim

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Canonicalization defines the DKIM canonicalization algorithm
type Canonicalization string

const (
	// CanonicalizationSimple is the simple canonicalization algorithm
	CanonicalizationSimple Canonicalization = "simple"

	// CanonicalizationRelaxed is the relaxed canonicalization algorithm
	CanonicalizationRelaxed Canonicalization = "relaxed"
)

// SignatureAlgo defines the DKIM signature algorithm
type SignatureAlgo string

const (
	// SignatureAlgoRSA is the RSA SHA-256 signature algorithm
	SignatureAlgoRSA SignatureAlgo = "rsa-sha256"

	// SignatureAlgoED25519 is the ED25519 SHA-256 signature algorithm
	SignatureAlgoED25519 SignatureAlgo = "ed25519-sha256"
)

var (
	// defaultHeader is the default set of headers to sign
	defaultHeader = []string{
		"From", "To", "Subject", "Date", "Message-ID",
		"Content-Type", "MIME-Version",
	}

	// defaultCanonicalization is the default canonicalization algorithm to use
	defaultCanonicalization = CanonicalizationRelaxed
)

var (
	// ErrDKIMNoDomain is returned when no DKIM domain is set
	ErrDKIMNoDomain = errors.New("a DKIM domain must be set")

	// ErrDKIMNoSelector is returned when no DKIM selector is set
	ErrDKIMNoSelector = errors.New("a DKIM selector must be set")

	// ErrDKIMNoSigner is returned when no DKIM crypto signer is provided
	ErrDKIMNoSigner = errors.New("a DKIM crypto signer must be provided")

	// ErrDKIMInvalidSigner is returned when the provided crypto signer is not
	// supported by DKIM
	ErrDKIMInvalidSigner = errors.New("the provided crypto signer is not supported by DKIM")

	// ErrDKIMMissingFrom is returned when the FROM header is missing
	ErrDKIMMissingFrom = errors.New("FROM is a required DKIM header")

	// ErrDKIMUnsupportedSigner is returned when the DKIM signer type is
	// unsupported
	ErrDKIMUnsupportedSigner = errors.New("unsupported DKIM signer type")
)

// Signer represents a DKIM signer that holds all signing parameters
type Signer struct {
	// Domain represents the DKIM Signing Domain Identifier (SDID). It is
	// d single domain name that is the mandatory payload output of DKIM
	// and that refers to the identity claiming some responsibility for
	// the message by signing it.
	//
	// Domain MUST not be empty
	//
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-2.5
	Domain string

	// Selector represents the DKIM domain selectors
	//
	// To support multiple concurrent public keys per signing domain, the
	// key namespace is subdivided using "selectors". For example,
	// selectors might indicate the names of office locations (e.g.,
	// "sanfrancisco", "coolumbeach", and "reykjavik"), the signing date
	// (e.g., "january2005", "february2005", etc.), or even an individual
	// user.
	//
	// Selector MUST not be empty
	//
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-3.1
	Selector string

	// Signer is the crypto.Signer used to sign the message.
	Signer crypto.Signer

	// NowFunc represents a function that is used to determine the signature's
	// time field (t=). It can be overridden for deterministic tests.
	NowFunc func() time.Time

	// auid represents the DKIM Agent or User Identifier (auid)
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-2.6
	//
	// A single identifier that refers to the agent or user on behalf of
	// whom the Signing Domain Identifier (SDID) has taken responsibility.
	// The auid comprises a domain name and an optional <local-part>.  The
	// domain name is the same as that used for the SDID or is a subdomain
	// of it. For DKIM processing, the domain name portion of the auid has
	// only basic domain name semantics; any possible owner-specific
	// semantics are outside the scope of DKIM.
	//
	// auid is optional and can be empty
	auid string

	// bodyCanonicalization defines the type of Canonicalization used for the
	// mail.Msg body.
	//
	// We default to CanonicalizationRelaxed for the body.
	//
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-3.4
	bodyCanonicalization Canonicalization

	// expiration is an optional expiration time of the signature.
	//
	// Signatures MAY be considered invalid if the verification time at
	// the verifier is past the expiration date. The verification
	// time should be the time that the message was first received at
	// the administrative domain of the verifier if that time is
	// reliably available; otherwise, the current time should be
	// used.
	//
	// See: https://www.rfc-editor.org/rfc/rfc6376.html#section-3.5
	expiration time.Duration

	// headerCanonicalization defines the type of Canonicalization used for the
	// mail.Msg header.
	//
	// We default to CanonicalizationRelaxed for the header.
	//
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-3.4
	headerCanonicalization Canonicalization

	// algorithm defines the type of signature algorithm used for the DKIM signature
	algorithm SignatureAlgo

	// bodyLength specifies the length of the message body that should be signed by DKIM.
	// It is highly discouraged to set this value as it means that only a portion of the
	// message body would be appropriately signed, including it allows malicious actors
	// to send phishing emails, alter content or otherwise exploit emails and still pass
	// DKIM
	bodyLength int64

	// signedHeaders holds the list of headers that are signed by the DKIM signature
	signedHeaders []string

	// oversignHeaders holds the list of headers that are oversigned by the DKIM signature
	// for improved security of the signature
	oversignHeaders []string
}

// NewSigner creates a new DKIM signer with the given domain, selector, and crypto.Signer
func NewSigner(domain, selector string, signer crypto.Signer) *Signer {
	dkimSigner := &Signer{
		Domain:   domain,
		Selector: selector,
		Signer:   signer,
	}
	return dkimSigner
}

// AUID sets the AUID (Authorized User ID) for the DKIM signature
func (s *Signer) AUID(auid string) {
	s.auid = auid
}

// BodyCanonicalization sets the canonicalization mode for the message body
func (s *Signer) BodyCanonicalization(mode Canonicalization) {
	if mode != CanonicalizationRelaxed && mode != CanonicalizationSimple {
		mode = defaultCanonicalization
	}
	s.bodyCanonicalization = mode
}

// Bodylength sets the length of the message body to be signed by DKIM (via the l= tag)
//
// It is highly discouraged to set this value as it means that only a portion of the
// message body would be appropriately signed, including it allows malicious actors
// to send phishing emails, alter content or otherwise exploit emails and still pass
// DKIM.
//
// Only use this option if you are sure what you are doing
func (s *Signer) Bodylength(length int64) {
	s.bodyLength = length
}

// ExpiresIn sets the expiration time for the DKIM signature
func (s *Signer) ExpiresIn(expiration time.Duration) {
	if expiration < 0 {
		return
	}
	s.expiration = expiration
}

// HeaderCanonicalization sets the canonicalization mode for the message headers
func (s *Signer) HeaderCanonicalization(mode Canonicalization) {
	if mode != CanonicalizationRelaxed && mode != CanonicalizationSimple {
		mode = defaultCanonicalization
	}
	s.headerCanonicalization = mode
}

// OversignHeaders sets the headers to be oversigned for the DKIM signature
func (s *Signer) OversignHeaders(headers ...string) {
	if len(headers) == 0 {
		return
	}
	s.oversignHeaders = headers
}

// SignHeaders sets the headers to be signed for the DKIM signature
func (s *Signer) SignHeaders(headers ...string) {
	if len(headers) == 0 {
		return
	}
	s.signedHeaders = headers
}

// Validate validates the DKIM configuration for required settings
func (s *Signer) ValidateConfig() error {
	switch {
	case s.Domain == "":
		return ErrDKIMNoDomain
	case s.Selector == "":
		return ErrDKIMNoSelector
	case s.Signer == nil:
		return ErrDKIMNoSigner
	}

	switch s.Signer.(type) {
	case ed25519.PrivateKey, *rsa.PrivateKey:
	default:
		return ErrDKIMInvalidSigner
	}

	for _, h := range s.effectiveHeaders() {
		if strings.EqualFold(h, "from") {
			return nil
		}
	}

	return ErrDKIMMissingFrom
}

// effectiveHeaders returns the list of headers to sign, defaulting to the standard set
// if none are configured
func (s *Signer) effectiveHeaders() []string {
	if len(s.signedHeaders) == 0 {
		return defaultHeader
	}
	return s.signedHeaders
}

// headerListForTag returns the list of headers to sign for a given tag, including any
// oversign headers
func (s *Signer) headerListForTag() []string {
	base := s.effectiveHeaders()
	if len(s.oversignHeaders) == 0 {
		return base
	}
	return append(append([]string{}, base...), s.oversignHeaders...)
}

// now returns the current time, using the injected Now function if set, or
// time.Now otherwise as default
func (s *Signer) now() time.Time {
	if s.NowFunc != nil {
		return s.NowFunc()
	}
	return time.Now()
}

// Sign returns the DKIM-Signature header line for the given raw headers and body
func (s *Signer) Sign(rawHeaders, body []byte) (string, error) {
	if err := s.ValidateConfig(); err != nil {
		return "", err
	}

	if s.algorithm == "" {
		switch s.Signer.Public().(type) {
		case ed25519.PublicKey:
			s.algorithm = SignatureAlgoED25519
		case *rsa.PublicKey:
			s.algorithm = SignatureAlgoRSA
		}
	}
	headerCanon, bc := s.headerCanonicalization, s.bodyCanonicalization
	if headerCanon == "" {
		headerCanon = CanonicalizationRelaxed
	}
	if bc == "" {
		bc = CanonicalizationRelaxed
	}

	// Create the body hash (bh=)
	cb := canonicalizeBody(body, bc)
	if s.bodyLength > 0 && int64(len(cb)) > s.bodyLength {
		cb = cb[:s.bodyLength]
	}
	bh := sha256.Sum256(cb)

	// Assemble the header tags as signature placeholder
	tags := []string{
		"v=1", // Version is always 1
		"c=" + string(headerCanon) + "/" + string(bc), // The canonicalization mode used for header/body
		"d=" + s.Domain,                                  // The signing domain
		"s=" + s.Selector,                                // The domain selector (<selector>._domainkey)
		"a=" + string(s.algorithm),                       // The selected signing algorithm
		fmt.Sprintf("t=%d", s.now().Unix()),              // The timestamp of the signature
		"h=" + strings.Join(s.headerListForTag(), ":"),   // The headers to be signed (including oversigned headers)
		"bh=" + base64.StdEncoding.EncodeToString(bh[:]), // The empty body hash
	}
	if s.expiration > 0 {
		tags = append(tags, fmt.Sprintf("x=%d", s.now().Add(s.expiration).Unix()))
	}
	if s.auid != "" {
		tags = append(tags, "i="+s.auid)
	}
	if s.bodyLength > 0 {
		tags = append(tags, fmt.Sprintf("l=%d", s.bodyLength))
	}
	tags = append(tags, "b=")
	value := strings.Join(tags, "; ")

	// Parse headers and assemble the input to signature (with empty b=)
	store := parseHeaders(rawHeaders)
	in := bytes.NewBuffer(nil)
	for _, header := range s.headerListForTag() {
		if line, ok := store.pop(header); ok {
			in.Write(canonicalizeHeader(line, headerCanon))
			in.WriteString("\r\n")
		}
	}

	// Fold the tags once with an empty b=. These exact bytes are reused for the
	// wire, so the fold points can never diverge between what we sign and what
	// we emit.
	foldedTags := foldHeader(value)

	switch headerCanon {
	case CanonicalizationSimple:
		in.WriteString("DKIM-Signature: ")
		in.WriteString(value) // unfolded, b= empty — matches the emitted line
	case CanonicalizationRelaxed:
		in.Write(canonicalizeHeader("DKIM-Signature:"+value, headerCanon))
	}

	// Both algorithms sign the SHA-256 hasher of the canonicalized headers
	//
	// See: https://datatracker.ietf.org/doc/html/rfc6376#section-3.7
	// and https://datatracker.ietf.org/doc/html/rfc8463#section-3
	var signature []byte
	var err error
	hasher := sha256.Sum256(in.Bytes())
	switch key := s.Signer.(type) {
	case ed25519.PrivateKey:
		signature = ed25519.Sign(key, hasher[:])
	default:
		signature, err = key.Sign(rand.Reader, hasher[:], crypto.SHA256)
	}
	if err != nil {
		return "", err
	}

	// Base64 encode the signature and fold it if allowed by the canonicalization mode
	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	if headerCanon == CanonicalizationRelaxed {
		// Relaxed canonicalization permits folding, which is our preferred output
		// and the default
		return "DKIM-Signature: " + appendFoldedBase64(foldedTags, signatureB64) + "\r\n", nil
	}

	// For simple canonicalization, we emit the unfolded value directly to avoid
	// issues with the signature validation
	return "DKIM-Signature: " + value + signatureB64 + "\r\n", nil
}
