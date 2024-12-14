package css

import (
	"strings"

	"github.com/gorilla/css/scanner"
)

func newImportRule(statement string) *CSSRule {
	statement = strings.TrimSpace(statement)
	if statement != "" {
		rule := NewRule(IMPORT_RULE)
		rule.Style.SelectorText = statement
		return rule
	}

	return nil
}

func parseImport(s *scanner.Scanner) *CSSRule {
	/*
		Syntax:
		@import url;                      or
		@import url list-of-media-queries;

		Example:
		@import url("fineprint.css") print;
		@import url("bluish.css") projection, tv;
		@import 'custom.css';
		@import url("chrome://communicator/skin/");
		@import "common.css" screen, projection;
		@import url('landscape.css') screen and (orientation:landscape);

	*/

	var statement string
	for {
		token := s.Next()

		//fmt.Printf("Import: %s:'%s'\n", token.Type.String(), token.Value)

		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			return nil
		}
		// take everything for now
		switch token.Type {
		case scanner.TokenChar:
			if token.Value == ";" {
				return newImportRule(statement)
			}
			statement += token.Value
		default:
			statement += token.Value
		}
	}
}
