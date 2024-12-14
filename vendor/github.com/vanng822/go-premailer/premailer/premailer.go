package premailer

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/vanng822/css"
	"golang.org/x/net/html"
)

// Premailer is the inteface of Premailer
type Premailer interface {
	// Transform process and inlining css
	// It start to collect the rules in the document style tags
	// Calculate specificity and sort the rules based on that
	// It then collects the affected elements
	// And applies the rules on those
	// The leftover rules will put back into a style element
	Transform() (string, error)
}

var unmergableSelector = regexp.MustCompile("(?i)\\:{1,2}(visited|active|hover|focus|link|root|in-range|invalid|valid|after|before|selection|target|first\\-(line|letter))|^\\@")
var notSupportedSelector = regexp.MustCompile("(?i)\\:(checked|disabled|enabled|lang)")

type premailer struct {
	doc       *goquery.Document
	elIdAttr  string
	elements  map[string]*elementRules
	rules     []*styleRule
	leftover  []*css.CSSRule
	allRules  [][]*css.CSSRule
	elementId int
	processed bool
	options   *Options
}

// NewPremailer return a new instance of Premailer
// It take a Document as argument and it shouldn't be nil
func NewPremailer(doc *goquery.Document, options *Options) Premailer {
	pr := premailer{}
	pr.doc = doc
	pr.rules = make([]*styleRule, 0)
	pr.allRules = make([][]*css.CSSRule, 0)
	pr.leftover = make([]*css.CSSRule, 0)
	pr.elements = make(map[string]*elementRules)
	pr.elIdAttr = "pr-el-id"
	if options == nil {
		options = NewOptions()
	}
	pr.options = options
	return &pr
}

func (pr *premailer) sortRules() {
	ruleIndex := 0
	for ruleSetIndex, rules := range pr.allRules {
		if rules == nil {
			continue
		}
		for _, rule := range rules {
			if rule.Type != css.STYLE_RULE {
				pr.leftover = append(pr.leftover, rule)
				continue
			}
			normalStyles := make([]*css.CSSStyleDeclaration, 0)
			importantStyles := make([]*css.CSSStyleDeclaration, 0)

			for _, s := range rule.Style.Styles {
				if s.Important == 1 {
					importantStyles = append(importantStyles, s)
				} else {
					normalStyles = append(normalStyles, s)
				}
			}

			selectors := strings.Split(rule.Style.SelectorText, ",")
			for _, selector := range selectors {
				if unmergableSelector.MatchString(selector) || notSupportedSelector.MatchString(selector) {
					// cause longer css
					pr.leftover = append(pr.leftover, copyRule(selector, rule))
					continue
				}
				if strings.Contains(selector, "*") {
					// keep this?
					pr.leftover = append(pr.leftover, copyRule(selector, rule))
					continue
				}
				if len(normalStyles) > 0 {
					pr.rules = append(pr.rules, &styleRule{makeSpecificity(0, ruleSetIndex, ruleIndex, selector), selector, normalStyles})
					ruleIndex += 1
				}
				if len(importantStyles) > 0 {
					pr.rules = append(pr.rules, &styleRule{makeSpecificity(1, ruleSetIndex, ruleIndex, selector), selector, importantStyles})
					ruleIndex += 1
				}
			}
		}
	}
	sort.Sort(bySpecificity(pr.rules))
}

func (pr *premailer) collectRules() {
	var wg sync.WaitGroup
	pr.doc.Find("style:not([data-premailer='ignore'])").Each(func(_ int, s *goquery.Selection) {
		if media, exist := s.Attr("media"); exist && media != "all" {
			return
		}
		wg.Add(1)
		pr.allRules = append(pr.allRules, nil)
		go func(ruleSetIndex int) {
			defer wg.Done()
			ss := css.Parse(s.Text())
			pr.allRules[ruleSetIndex] = ss.GetCSSRuleList()
			s.ReplaceWithHtml("")
		}(len(pr.allRules) - 1)
	})
	wg.Wait()

}

func (pr *premailer) collectElements() {
	for _, rule := range pr.rules {
		pr.doc.Find(rule.selector).Each(func(_ int, s *goquery.Selection) {
			if id, exist := s.Attr(pr.elIdAttr); exist {
				pr.elements[id].rules = append(pr.elements[id].rules, rule)
			} else {
				id := strconv.Itoa(pr.elementId)
				s.SetAttr(pr.elIdAttr, id)
				rules := make([]*styleRule, 0)
				rules = append(rules, rule)
				pr.elements[id] = &elementRules{element: s, rules: rules, cssToAttributes: pr.options.CssToAttributes}
				pr.elementId += 1
			}
		})

	}
}

func (pr *premailer) applyInline() {
	for _, element := range pr.elements {
		element.inline()
		element.element.RemoveAttr(pr.elIdAttr)
		if pr.options.RemoveClasses {
			element.element.RemoveAttr("class")
		}
	}
}

func (pr *premailer) addLeftover() {
	if len(pr.leftover) > 0 {
		headNode := pr.doc.Find("head")

		styleNode := &html.Node{}
		styleNode.Type = html.ElementNode
		styleNode.Data = "style"
		styleNode.Attr = []html.Attribute{{Key: "type", Val: "text/css"}}
		cssNode := &html.Node{}
		cssData := make([]string, 0, len(pr.leftover))
		for _, rule := range pr.leftover {
			if rule.Type == css.MEDIA_RULE {
				mcssData := make([]string, 0, len(rule.Rules))
				for _, mrule := range rule.Rules {
					mcssData = append(mcssData, makeRuleImportant(mrule))
				}
				cssData = append(cssData, fmt.Sprintf("%s %s{\n%s\n}\n",
					rule.Type.Text(),
					rule.Style.SelectorText,
					strings.Join(mcssData, "\n")))
			} else {
				cssData = append(cssData, makeRuleImportant(rule))
			}
		}
		cssNode.Data = strings.Join(cssData, "")
		cssNode.Type = html.TextNode
		styleNode.AppendChild(cssNode)
		headNode.AppendNodes(styleNode)
	}
}

// Transform process and inlining css
// It start to collect the rules in the document style tags
// Calculate specificity and sort the rules based on that
// It then collects the affected elements
// And applies the rules on those
// The leftover rules will put back into a style element
func (pr *premailer) Transform() (string, error) {
	if !pr.processed {
		pr.collectRules()
		pr.sortRules()
		pr.collectElements()
		pr.applyInline()
		pr.addLeftover()
		pr.processed = true
	}
	return pr.doc.Html()
}
