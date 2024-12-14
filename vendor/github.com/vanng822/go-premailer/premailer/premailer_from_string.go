package premailer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NewPremailerFromString take in a document in string format
// and create a goquery.Document
// and then create and Premailer instance.
func NewPremailerFromString(doc string, options *Options) (Premailer, error) {
	read := strings.NewReader(doc)
	d, err := goquery.NewDocumentFromReader(read)
	if err != nil {
		return nil, err
	}
	return NewPremailer(d, options), nil
}
