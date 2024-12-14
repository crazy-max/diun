package css

type CSSStyleSheet struct {
	Type        string
	Media       string
	CssRuleList []*CSSRule
}

func (ss *CSSStyleSheet) GetCSSRuleList() []*CSSRule {
	return ss.CssRuleList
}
