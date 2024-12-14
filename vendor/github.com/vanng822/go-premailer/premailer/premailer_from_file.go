package premailer

import (
	"os"

	"github.com/PuerkitoBio/goquery"
)

// NewPremailerFromFile take an filename
// Read the content of this file
// and create a goquery.Document
// and then create and Premailer instance.
func NewPremailerFromFile(filename string, options *Options) (Premailer, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	d, err := goquery.NewDocumentFromReader(fd)
	if err != nil {
		return nil, err
	}
	return NewPremailer(d, options), nil
}
