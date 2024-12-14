package css

import (
	"fmt"
)

type CSSStyleDeclaration struct {
	Property  string
	Value     string
	Important int
}

func NewCSSStyleDeclaration(property, value string, important int) *CSSStyleDeclaration {
	return &CSSStyleDeclaration{
		Property:  property,
		Value:     value,
		Important: important,
	}
}

func (decl *CSSStyleDeclaration) Text() string {
	if decl.Important == 1 {
		return fmt.Sprintf("%s: %s !important", decl.Property, decl.Value)
	}
	return fmt.Sprintf("%s: %s", decl.Property, decl.Value)
}
