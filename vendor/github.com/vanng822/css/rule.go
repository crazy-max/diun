package css

type RuleType int

const (
	STYLE_RULE RuleType = iota
	CHARSET_RULE
	IMPORT_RULE
	MEDIA_RULE
	FONT_FACE_RULE
	PAGE_RULE
	KEYFRAMES_RULE
	WEBKIT_KEYFRAMES_RULE
	COUNTER_STYLE_RULE
)

var ruleTypeNames = map[RuleType]string{
	STYLE_RULE:            "",
	MEDIA_RULE:            "@media",
	CHARSET_RULE:          "@charset",
	IMPORT_RULE:           "@import",
	FONT_FACE_RULE:        "@font-face",
	PAGE_RULE:             "@page",
	KEYFRAMES_RULE:        "@keyframes",
	WEBKIT_KEYFRAMES_RULE: "@-webkit-keyframes",
	COUNTER_STYLE_RULE:    "@counter-style",
}

func (rt RuleType) Text() string {
	return ruleTypeNames[rt]
}

type CSSRule struct {
	Type  RuleType
	Style CSSStyleRule
	Rules []*CSSRule
}

func NewRule(ruleType RuleType) *CSSRule {
	r := &CSSRule{
		Type: ruleType,
	}
	r.Style.Styles = make([]*CSSStyleDeclaration, 0)
	r.Rules = make([]*CSSRule, 0)
	return r
}
