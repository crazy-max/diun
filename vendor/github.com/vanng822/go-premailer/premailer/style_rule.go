package premailer

import (
	"github.com/vanng822/css"
)

type styleRule struct {
	specificity *specificity
	selector    string
	styles      []*css.CSSStyleDeclaration
}
