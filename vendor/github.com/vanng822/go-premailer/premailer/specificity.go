package premailer

import (
	"regexp"
	"strings"
)

// https://developer.mozilla.org/en-US/docs/Web/CSS/Specificity
// https://developer.mozilla.org/en-US/docs/Web/CSS/Reference#Selectors

type specificity struct {
	important    int
	idCount      int
	classCount   int
	typeCount    int
	attrCount    int
	ruleSetIndex int
	ruleIndex    int
}

func (s *specificity) importantOrders() []int {
	return []int{s.important, s.idCount,
		s.classCount, s.attrCount,
		s.typeCount, s.ruleSetIndex,
		s.ruleIndex}
}

var typeSelectorRegex = regexp.MustCompile("(^|\\s)\\w")

func makeSpecificity(important, ruleSetIndex, ruleIndex int, selector string) *specificity {
	spec := specificity{}
	// determine values for priority
	if important > 0 {
		spec.important = 1
	} else {
		spec.important = 0
	}
	spec.idCount = strings.Count(selector, "#")
	spec.classCount = strings.Count(selector, ".")
	spec.attrCount = strings.Count(selector, "[")
	spec.typeCount = len(typeSelectorRegex.FindAllString(selector, -1))
	spec.ruleSetIndex = ruleSetIndex
	spec.ruleIndex = ruleIndex
	return &spec
}

type bySpecificity []*styleRule

func (bs bySpecificity) Len() int {
	return len(bs)
}
func (bs bySpecificity) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs bySpecificity) Less(i, j int) bool {
	iorders := bs[i].specificity.importantOrders()
	jorders := bs[j].specificity.importantOrders()
	for n, v := range iorders {
		if v < jorders[n] {
			return true
		}
		if v > jorders[n] {
			return false
		}
	}
	return false
}
