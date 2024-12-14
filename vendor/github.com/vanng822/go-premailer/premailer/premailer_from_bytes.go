package premailer

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

// NewPremailerFromBytes take in a document in byte
// and create a goquery.Document
// and then create and Premailer instance.
func NewPremailerFromBytes(doc []byte, options *Options) (Premailer, error) {
	read := bytes.NewReader(doc)
	d, err := goquery.NewDocumentFromReader(read)
	if err != nil {
		return nil, err
	}
	return NewPremailer(d, options), nil
}
