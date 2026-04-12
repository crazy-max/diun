// Package httplink parses the Link header from HTTP responses according to RFC5988
package httplink

import (
	"fmt"
	"strings"

	"github.com/regclient/regclient/types/errs"
)

type (
	Links []Link
	Link  struct {
		URI   string
		Param map[string]string
	}
)

type charLU byte

var charLUs [256]charLU

const (
	isSpace charLU = 1 << iota
	isToken
	isAlphaNum
)

func init() {
	for c := range 256 {
		charLUs[c] = 0
		if strings.ContainsRune(" \t\r\n", rune(c)) {
			charLUs[c] |= isSpace
		}
		if (rune('a') <= rune(c) && rune(c) <= rune('z')) || (rune('A') <= rune(c) && rune(c) <= rune('Z') || (rune('0') <= rune(c) && rune(c) <= rune('9'))) {
			charLUs[c] |= isAlphaNum | isToken
		}
		if strings.ContainsRune("!#$%&'()*+-./:<=>?@[]^_`{|}~", rune(c)) {
			charLUs[c] |= isToken
		}
	}
}

// Parse reads "Link" http headers into an array of Link structs.
// Header array should be the output of resp.Header.Values("link").
func Parse(headers []string) (Links, error) {
	links := []Link{}
	for _, h := range headers {
		state := "init"
		var ub, pnb, pvb []byte
		parms := map[string]string{}
		endLink := func() {
			links = append(links, Link{
				URI:   string(ub),
				Param: parms,
			})
			// reset state
			ub, pnb, pvb = []byte{}, []byte{}, []byte{}
			parms = map[string]string{}
		}
		endParm := func() {
			if _, ok := parms[string(pnb)]; !ok {
				parms[string(pnb)] = string(pvb)
			}
			// reset parm
			pnb, pvb = []byte{}, []byte{}
		}
		for i, b := range []byte(h) {
			switch state {
			case "init":
				if b == '<' {
					state = "uriQuoted"
				} else if charLUs[b]&isToken != 0 {
					state = "uri"
					ub = append(ub, b)
				} else if charLUs[b]&isSpace != 0 || b == ',' {
					// noop
				} else {
					// unknown character
					return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
				}
			case "uri":
				// parse tokens until space or comma
				if charLUs[b]&isToken != 0 {
					ub = append(ub, b)
				} else if charLUs[b]&isSpace != 0 {
					state = "fieldSep"
				} else if b == ';' {
					state = "parmName"
				} else if b == ',' {
					state = "init"
					endLink()
				} else {
					// unknown character
					return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
				}
			case "uriQuoted":
				// parse tokens until quote
				if b == '>' {
					state = "fieldSep"
				} else {
					ub = append(ub, b)
				}
			case "fieldSep":
				if b == ';' {
					state = "parmName"
				} else if b == ',' {
					state = "init"
					endLink()
				} else if charLUs[b]&isSpace != 0 {
					// noop
				} else {
					// unknown character
					return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
				}
			case "parmName":
				if len(pnb) > 0 && b == '=' {
					state = "parmValue"
				} else if len(pnb) > 0 && b == '*' {
					state = "parmNameStar"
				} else if charLUs[b]&isAlphaNum != 0 {
					pnb = append(pnb, b)
				} else if len(pnb) == 0 && charLUs[b]&isSpace != 0 {
					// noop
				} else {
					// unknown character
					return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
				}
			case "parmNameStar":
				if b == '=' {
					state = "parmValue"
				} else {
					// unknown character
					return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
				}
			case "parmValue":
				if len(pvb) == 0 {
					if charLUs[b]&isToken != 0 {
						pvb = append(pvb, b)
					} else if b == '"' {
						state = "parmValueQuoted"
					} else {
						// unknown character
						return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
					}
				} else {
					if charLUs[b]&isToken != 0 {
						pvb = append(pvb, b)
					} else if charLUs[b]&isSpace != 0 {
						state = "fieldSep"
						endParm()
					} else if b == ';' {
						state = "parmName"
						endParm()
					} else if b == ',' {
						state = "init"
						endParm()
						endLink()
					} else {
						// unknown character
						return nil, fmt.Errorf("unknown character in position %d of %s: %w", i, h, errs.ErrParsingFailed)
					}
				}
			case "parmValueQuoted":
				if b == '"' {
					state = "fieldSep"
					endParm()
				} else {
					pvb = append(pvb, b)
				}
			}
		}
		// check for valid state at end of header
		switch state {
		case "parmValue":
			endParm()
			endLink()
		case "uri", "fieldSep":
			endLink()
		case "init":
			// noop
		default:
			return nil, fmt.Errorf("unexpected end state %s for header %s: %w", state, h, errs.ErrParsingFailed)
		}
	}

	return links, nil
}

// Get returns a link with a specific parm value, e.g. rel="next"
func (links Links) Get(parm, val string) (Link, error) {
	for _, link := range links {
		if link.Param != nil && link.Param[parm] == val {
			return link, nil
		}
	}
	return Link{}, errs.ErrNotFound
}
