package css

import (
	"strings"

	"github.com/gorilla/css/scanner"
)

type blockParserContext struct {
	State        State
	NowProperty  string
	NowValue     string
	NowImportant int
}

// ParseBlock take a string of a css block,
// parses it and returns a map of css style declarations.
func ParseBlock(csstext string) []*CSSStyleDeclaration {
	s := scanner.New(csstext)
	return parseBlock(s)
}

func parseBlock(s *scanner.Scanner) []*CSSStyleDeclaration {
	/* block       : '{' S* [ any | block | ATKEYWORD S* | ';' S* ]* '}' S*;
	property    : IDENT;
	value       : [ any | block | ATKEYWORD S* ]+;
	any         : [ IDENT | NUMBER | PERCENTAGE | DIMENSION | STRING
	              | DELIM | URI | HASH | UNICODE-RANGE | INCLUDES
	              | DASHMATCH | ':' | FUNCTION S* [any|unused]* ')'
	              | '(' S* [any|unused]* ')' | '[' S* [any|unused]* ']'
	              ] S*;
	*/
	decls := make([]*CSSStyleDeclaration, 0)

	context := &blockParserContext{
		State:        STATE_NONE,
		NowProperty:  "",
		NowValue:     "",
		NowImportant: 0,
	}

	for {
		token := s.Next()

		//fmt.Printf("BLOCK(%d): %s:'%s'\n", context.State, token.Type.String(), token.Value)

		if token.Type == scanner.TokenError {
			break
		}

		if token.Type == scanner.TokenEOF {
			if context.State == STATE_VALUE {
				// we are ending without ; or }
				// this can happen when we parse only css declaration
				decl := NewCSSStyleDeclaration(context.NowProperty, strings.TrimSpace(context.NowValue), context.NowImportant)
				decls = append(decls, decl)
			}
			break
		}

		switch token.Type {

		case scanner.TokenS:
			if context.State == STATE_VALUE {
				context.NowValue += token.Value
			}
		case scanner.TokenIdent:
			if context.State == STATE_NONE {
				context.State = STATE_PROPERTY
				context.NowProperty = strings.TrimSpace(token.Value)
				break
			}
			if token.Value == "important" {
				context.NowImportant = 1
			} else {
				context.NowValue += token.Value
			}
		case scanner.TokenChar:
			if context.State == STATE_NONE {
				if token.Value == "{" {
					break
				}
			}
			if context.State == STATE_PROPERTY {
				if token.Value == ":" {
					context.State = STATE_VALUE
				}
				// CHAR and STATE_PROPERTY but not : then weird
				// break to ignore it
				break
			}
			// should be no state or value
			if token.Value == ";" {
				decl := NewCSSStyleDeclaration(context.NowProperty, strings.TrimSpace(context.NowValue), context.NowImportant)
				decls = append(decls, decl)
				context.NowProperty = ""
				context.NowValue = ""
				context.NowImportant = 0
				context.State = STATE_NONE
			} else if token.Value == "}" { // last property in a block can have optional ;
				if context.State == STATE_VALUE {
					// only valid if state is still VALUE, could be ;}
					decl := NewCSSStyleDeclaration(context.NowProperty, strings.TrimSpace(context.NowValue), context.NowImportant)
					decls = append(decls, decl)
				}
				// we are done
				return decls
			} else if token.Value != "!" {
				context.NowValue += token.Value
			}
			break

		// any
		case scanner.TokenNumber:
			fallthrough
		case scanner.TokenPercentage:
			fallthrough
		case scanner.TokenDimension:
			fallthrough
		case scanner.TokenString:
			fallthrough
		case scanner.TokenURI:
			fallthrough
		case scanner.TokenHash:
			fallthrough
		case scanner.TokenUnicodeRange:
			fallthrough
		case scanner.TokenIncludes:
			fallthrough
		case scanner.TokenDashMatch:
			fallthrough
		case scanner.TokenFunction:
			fallthrough
		case scanner.TokenSubstringMatch:
			context.NowValue += token.Value
		}
	}

	return decls
}
