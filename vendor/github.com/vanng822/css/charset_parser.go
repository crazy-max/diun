package css

import (
	"strings"

	"github.com/gorilla/css/scanner"
)

func newCharsetRule(statement string) *CSSRule {
	statement = strings.TrimSpace(statement)
	if statement != "" {
		rule := NewRule(CHARSET_RULE)
		rule.Style.SelectorText = statement
		return rule
	}

	return nil
}

func parseCharset(s *scanner.Scanner) *CSSRule {
	/*

		Syntax:
		@charset charset;

		Example:
		@charset "UTF-8";

	*/

	var statement string
	for {
		token := s.Next()

		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			return nil
		}
		// take everything for now
		switch token.Type {
		case scanner.TokenChar:
			if token.Value == ";" {
				return newCharsetRule(statement)
			}
			statement += token.Value
		default:
			statement += token.Value
		}
	}
}
