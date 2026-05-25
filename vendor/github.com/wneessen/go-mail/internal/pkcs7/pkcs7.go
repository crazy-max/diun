// SPDX-FileCopyrightText: Copyright (c) 2015 Andrew Smith
// SPDX-FileCopyrightText: Copyright (c) 2017-2024 The mozilla services project (https://github.com/mozilla-services)
// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// Partially forked from https://github.com/mozilla-services/pkcs7, which in turn is also a fork
// of https://github.com/fullsailor/pkcs7.
// Use of the forked source code is, same as go-mail, governed by a MIT license.
//
// go-mail specific modifications by the go-mail Authors.
// Licensed under the MIT License.
// See [PROJECT ROOT]/LICENSES directory for more information.
//
// SPDX-License-Identifier: MIT

package pkcs7

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256" // for crypto.SHA256
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"
)

var (
	OIDData                   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	OIDSignedData             = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	OIDAttributeContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	OIDAttributeMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	OIDAttributeSigningTime   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}

	OIDDigestAlgorithmSHA256      = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	OIDDigestAlgorithmECDSASHA256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}

	OIDEncryptionAlgorithmRSASHA256 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}
)

// PKCS7 Represents a PKCS7 structure
type PKCS7 struct {
	Content      []byte
	Certificates []*x509.Certificate
	CRLs         []x509.RevocationList
	Signers      []signerInfo
}

type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,optional,tag:0"`
}

type signedData struct {
	Version                    int                        `asn1:"default:1"`
	DigestAlgorithmIdentifiers []pkix.AlgorithmIdentifier `asn1:"set"`
	ContentInfo                contentInfo
	Certificates               rawCertificates       `asn1:"optional,tag:0"`
	CRLs                       []x509.RevocationList `asn1:"optional,tag:1"`
	SignerInfos                []signerInfo          `asn1:"set"`
}

type rawCertificates struct {
	Raw asn1.RawContent
}

func (raw rawCertificates) Parse() ([]*x509.Certificate, error) {
	if len(raw.Raw) == 0 {
		return nil, nil
	}

	var val asn1.RawValue
	if _, err := asn1.Unmarshal(raw.Raw, &val); err != nil {
		return nil, err
	}

	return x509.ParseCertificates(val.Bytes)
}

type attribute struct {
	Type  asn1.ObjectIdentifier
	Value asn1.RawValue `asn1:"set"`
}

type issuerAndSerial struct {
	IssuerName   asn1.RawValue
	SerialNumber *big.Int
}

// MessageDigestMismatchError is returned when the signer data digest does not
// match the computed digest for the contained content
type MessageDigestMismatchError struct {
	ExpectedDigest []byte
	ActualDigest   []byte
}

func (err *MessageDigestMismatchError) Error() string {
	return fmt.Sprintf("pkcs7: Message digest mismatch\n\tExpected: %X\n\tActual  : %X", err.ExpectedDigest, err.ActualDigest)
}

type signerInfo struct {
	Version                   int `asn1:"default:1"`
	IssuerAndSerialNumber     issuerAndSerial
	DigestAlgorithm           pkix.AlgorithmIdentifier
	AuthenticatedAttributes   []attribute `asn1:"optional,tag:0"`
	DigestEncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedDigest           []byte
	UnauthenticatedAttributes []attribute `asn1:"optional,tag:1"`
}

// Attribute represents a key value pair attribute. Value must be marshalable byte
// `encoding/asn1`
type Attribute struct {
	Type  asn1.ObjectIdentifier
	Value interface{}
}

// SignerInfoConfig are optional values to include when adding a signer
type SignerInfoConfig struct {
	ExtraSignedAttributes   []Attribute
	ExtraUnsignedAttributes []Attribute
}

type attributes struct {
	types  []asn1.ObjectIdentifier
	values []interface{}
}

// Add adds the attribute, maintaining insertion order
func (as *attributes) Add(attrType asn1.ObjectIdentifier, value interface{}) {
	as.types = append(as.types, attrType)
	as.values = append(as.values, value)
}

type sortableAttribute struct {
	SortKey   []byte
	Attribute attribute
}

type attributeSet []sortableAttribute

func (as attributeSet) Len() int {
	return len(as)
}

func (as attributeSet) Less(i, j int) bool {
	return bytes.Compare(as[i].SortKey, as[j].SortKey) < 0
}

func (as attributeSet) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (as attributeSet) attributes() []attribute {
	attrs := make([]attribute, len(as))
	for i, attr := range as {
		attrs[i] = attr.Attribute
	}
	return attrs
}

func (as *attributes) forMarshaling() ([]attribute, error) {
	sortables := make(attributeSet, len(as.types))
	for i := range sortables {
		attrType := as.types[i]
		attrValue := as.values[i]
		asn1Value, err := asn1.Marshal(attrValue)
		if err != nil {
			return nil, err
		}
		attr := attribute{
			Type:  attrType,
			Value: asn1.RawValue{Tag: 17, IsCompound: true, Bytes: asn1Value}, // 17 == SET tag
		}
		encoded, err := asn1.Marshal(attr)
		if err != nil {
			return nil, err
		}
		sortables[i] = sortableAttribute{
			SortKey:   encoded,
			Attribute: attr,
		}
	}
	sort.Sort(sortables)
	return sortables.attributes(), nil
}

// SignedData is an opaque data structure for creating signed data payloads
type SignedData struct {
	sd            signedData
	certs         []*x509.Certificate
	data          []byte
	messageDigest []byte
	digestOid     asn1.ObjectIdentifier
}

// NewSignedData initializes a SignedData with content
func NewSignedData(data []byte) (*SignedData, error) {
	content, err := asn1.Marshal(data)
	if err != nil {
		return nil, err
	}
	ci := contentInfo{
		ContentType: OIDData,
		Content:     asn1.RawValue{Class: 2, Tag: 0, Bytes: content, IsCompound: true},
	}
	sd := signedData{
		ContentInfo: ci,
		Version:     1,
	}
	return &SignedData{sd: sd, data: data, digestOid: OIDDigestAlgorithmSHA256}, nil
}

// AddSigner is a wrapper around AddSignerChain() that adds a signer without any parent.
func (sd *SignedData) AddSigner(cert *x509.Certificate, pkey crypto.PrivateKey, config SignerInfoConfig) error {
	var parents []*x509.Certificate
	return sd.addSignerChain(cert, pkey, parents, config)
}

// addSignerChain signs attributes about the content and adds certificates
// and signers infos to the Signed Data. The certificate and private key
// of the end-entity signer are used to issue the signature, and any
// parent of that end-entity that need to be added to the list of
// certifications can be specified in the parents slice.
//
// The signature algorithm used to hash the data is the one of the end-entity
// certificate.
func (sd *SignedData) addSignerChain(ee *x509.Certificate, pkey crypto.PrivateKey, parents []*x509.Certificate, config SignerInfoConfig) error {
	// Following RFC 2315, 9.2 SignerInfo type, the distinguished name of
	// the issuer of the end-entity signer is stored in the issuerAndSerialNumber
	// section of the SignedData.SignerInfo, alongside the serial number of
	// the end-entity.
	var ias issuerAndSerial
	ias.SerialNumber = ee.SerialNumber
	if len(parents) == 0 {
		// no parent, the issuer is the end-entity cert itself
		ias.IssuerName = asn1.RawValue{FullBytes: ee.RawIssuer}
	} else {
		err := verifyPartialChain(ee, parents)
		if err != nil {
			return err
		}
		// the first parent is the issuer
		ias.IssuerName = asn1.RawValue{FullBytes: parents[0].RawSubject}
	}
	sd.sd.DigestAlgorithmIdentifiers = append(sd.sd.DigestAlgorithmIdentifiers,
		pkix.AlgorithmIdentifier{Algorithm: sd.digestOid},
	)
	h := crypto.SHA256.New()
	h.Write(sd.data)
	sd.messageDigest = h.Sum(nil)
	encryptionOid, err := getOIDForEncryptionAlgorithm(pkey, sd.digestOid)
	if err != nil {
		return err
	}
	attrs := &attributes{}
	attrs.Add(OIDAttributeContentType, sd.sd.ContentInfo.ContentType)
	attrs.Add(OIDAttributeMessageDigest, sd.messageDigest)
	attrs.Add(OIDAttributeSigningTime, time.Now().UTC())
	for _, attr := range config.ExtraSignedAttributes {
		attrs.Add(attr.Type, attr.Value)
	}
	finalAttrs, err := attrs.forMarshaling()
	if err != nil {
		return err
	}
	unsignedAttrs := &attributes{}
	for _, attr := range config.ExtraUnsignedAttributes {
		unsignedAttrs.Add(attr.Type, attr.Value)
	}
	finalUnsignedAttrs, err := unsignedAttrs.forMarshaling()
	if err != nil {
		return err
	}
	// create signature of signed attributes
	signature, err := signAttributes(finalAttrs, pkey, crypto.SHA256)
	if err != nil {
		return err
	}
	signer := signerInfo{
		AuthenticatedAttributes:   finalAttrs,
		UnauthenticatedAttributes: finalUnsignedAttrs,
		DigestAlgorithm:           pkix.AlgorithmIdentifier{Algorithm: sd.digestOid},
		DigestEncryptionAlgorithm: pkix.AlgorithmIdentifier{Algorithm: encryptionOid},
		IssuerAndSerialNumber:     ias,
		EncryptedDigest:           signature,
		Version:                   1,
	}
	sd.certs = append(sd.certs, ee)
	if len(parents) > 0 {
		sd.certs = append(sd.certs, parents...)
	}
	sd.sd.SignerInfos = append(sd.sd.SignerInfos, signer)
	return nil
}

// AddCertificate adds the certificate to the payload. Useful for parent certificates
func (sd *SignedData) AddCertificate(cert *x509.Certificate) {
	sd.certs = append(sd.certs, cert)
}

// Detach removes content from the signed data struct to make it a detached signature.
// This must be called right before Finish()
func (sd *SignedData) Detach() {
	sd.sd.ContentInfo = contentInfo{ContentType: OIDData}
}

// Finish marshals the content and its signers
func (sd *SignedData) Finish() ([]byte, error) {
	sd.sd.Certificates = marshalCertificates(sd.certs)
	inner, err := asn1.Marshal(sd.sd)
	if err != nil {
		return nil, err
	}
	outer := contentInfo{
		ContentType: OIDSignedData,
		Content:     asn1.RawValue{Class: 2, Tag: 0, Bytes: inner, IsCompound: true},
	}
	return asn1.Marshal(outer)
}

// verifyPartialChain checks that a given cert is issued by the first parent in the list,
// then continue down the path. It doesn't require the last parent to be a root CA,
// or to be trusted in any truststore. It simply verifies that the chain provided, albeit
// partial, makes sense.
func verifyPartialChain(cert *x509.Certificate, parents []*x509.Certificate) error {
	if len(parents) == 0 {
		return fmt.Errorf("pkcs7: zero parents provided to verify the signature of certificate %q", cert.Subject.CommonName)
	}
	err := cert.CheckSignatureFrom(parents[0])
	if err != nil {
		return fmt.Errorf("pkcs7: certificate signature from parent is invalid: %w", err)
	}
	if len(parents) == 1 {
		// there is no more parent to check, return
		return nil
	}
	return verifyPartialChain(parents[0], parents[1:])
}

// getOIDForEncryptionAlgorithm takes the private key type of the signer and
// the OID of a digest algorithm to return the appropriate signerInfo.DigestEncryptionAlgorithm
func getOIDForEncryptionAlgorithm(pkey crypto.PrivateKey, _ asn1.ObjectIdentifier) (asn1.ObjectIdentifier, error) {
	switch pkey.(type) {
	case *rsa.PrivateKey:
		return OIDEncryptionAlgorithmRSASHA256, nil
	case *ecdsa.PrivateKey:
		return OIDDigestAlgorithmECDSASHA256, nil
	}
	return nil, fmt.Errorf("pkcs7: cannot convert encryption algorithm to oid, unknown private key type %T", pkey)
}

// signs the DER encoded form of the attributes with the private key
func signAttributes(attrs []attribute, pkey crypto.PrivateKey, hash crypto.Hash) ([]byte, error) {
	attrBytes, err := marshalAttributes(attrs)
	if err != nil {
		return nil, err
	}
	h := hash.New()
	h.Write(attrBytes)
	hashed := h.Sum(nil)

	key, ok := pkey.(crypto.Signer)
	if !ok {
		return nil, errors.New("pkcs7: private key does not implement crypto.Signer")
	}
	return key.Sign(rand.Reader, hashed, hash)
}

// concat and wraps the certificates in the RawValue structure
func marshalCertificates(certs []*x509.Certificate) rawCertificates {
	var buf bytes.Buffer
	for _, cert := range certs {
		buf.Write(cert.Raw)
	}
	rawCerts, _ := marshalCertificateBytes(buf.Bytes())
	return rawCerts
}

// Even though, the tag & length are stripped out during marshalling the
// RawContent, we have to encode it into the RawContent. If It's missing,
// then `asn1.Marshal()` will strip out the certificate wrapper instead.
func marshalCertificateBytes(certs []byte) (rawCertificates, error) {
	val := asn1.RawValue{Bytes: certs, Class: 2, Tag: 0, IsCompound: true}
	b, err := asn1.Marshal(val)
	if err != nil {
		return rawCertificates{}, err
	}
	return rawCertificates{Raw: b}, nil
}

func marshalAttributes(attrs []attribute) ([]byte, error) {
	encodedAttributes, err := asn1.Marshal(struct {
		A []attribute `asn1:"set"`
	}{A: attrs})
	if err != nil {
		return nil, err
	}

	// Remove the leading sequence octets
	var raw asn1.RawValue
	if _, err := asn1.Unmarshal(encodedAttributes, &raw); err != nil {
		return nil, err
	}
	return raw.Bytes, nil
}
