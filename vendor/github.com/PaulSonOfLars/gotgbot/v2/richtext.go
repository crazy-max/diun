package gotgbot

import (
	"html"
	"strconv"
	"strings"
)

type renderCtx struct {
	sb      strings.Builder
	media   []InputRichMessageMedia
	counter int64 // media counter
}

func (r *renderCtx) addMedia(m InputRichMessageMedia) {
	r.media = append(r.media, m)
}

func (r *renderCtx) nextCounter() int64 {
	r.counter++
	return r.counter
}

func (r *renderCtx) addFile(strType string, v InputMedia) string {
	sendFileId := strconv.FormatInt(r.nextCounter(), 10)
	link := "tg://" + strType + "?id=" + sendFileId
	r.addMedia(InputRichMessageMedia{
		Id:    sendFileId,
		Media: v,
	})
	return link
}

func isReversed(items RichBlockListItemArray) bool {
	if len(items) == 0 {
		return false
	}
	if len(items) == 1 {
		return items[0].Value > 1
	}
	return items[0].Value > items[1].Value
}

// GetRichTextTypes gets the list of all richtext types which match the "want" function.
// Eg: m.GetRichTextTypes(IsRichTextType[RichTextString], IsRichTextType[RichTextUrl]).
func (m *Message) GetRichTextTypes(want ...func(RichText) bool) []RichText {
	if m.RichMessage == nil {
		return nil
	}

	var out []RichText
	m.RichMessage.Walk(nil, func(node RichText) bool {
		for _, w := range want {
			if w(node) {
				out = append(out, node)
				break
			}
		}
		return true
	})
	return out
}

// Checks whether an input type is a specific kind of RichText.
// Use like this: gotgbot.IsRichTextType[RichTextString](x).
func IsRichTextType[T RichText](n RichText) bool {
	_, ok := n.(T)
	return ok
}

// GetRichTexts allows for specifying a RichText type, walking over all the blocks, and extracting the result.
func GetRichTexts[T RichText](m *RichMessage) []T {
	var out []T
	m.Walk(nil, func(text RichText) bool {
		if v, ok := text.(T); ok {
			out = append(out, v)
		}
		return true
	})
	return out
}

// MapRichTexts allows for extracting specific internals of a specific RichText type.
func MapRichTexts[T RichText, R any](m *RichMessage, fn func(T) R) []R {
	var out []R
	m.Walk(nil, func(text RichText) bool {
		if v, ok := text.(T); ok {
			out = append(out, fn(v))
		}
		return true
	})
	return out
}

// GetRichBlocks allows for specifying a RichBlock type, walking over all the blocks, and extracting the result.
func GetRichBlocks[T RichBlock](m *RichMessage) []T {
	var out []T
	m.Walk(func(block RichBlock) bool {
		if v, ok := block.(T); ok {
			out = append(out, v)
		}
		return true
	}, nil)
	return out
}

// MapRichBlocks allows for extracting specific internals of a specific RichBlock type.
func MapRichBlocks[T RichBlock, R any](m *RichMessage, fn func(T) R) []R {
	var out []R
	m.Walk(func(block RichBlock) bool {
		if v, ok := block.(T); ok {
			out = append(out, fn(v))
		}
		return true
	}, nil)
	return out
}

func HasEntityType(entity []MessageEntity, entType string) bool {
	for _, v := range entity {
		if v.Type == entType {
			return true
		}
	}
	return false
}

func HasRichType(m *RichMessage, entType string) bool {
	if m == nil {
		return false
	}

	found := false
	m.Walk(func(block RichBlock) bool {
		if block.GetType() == entType {
			found = true
			return false
		}
		return true
	}, func(node RichText) bool {
		if node.GetType() == entType {
			found = true
			return false
		}
		return true
	})
	return found
}

// WalkRichText recursively visits r and every descendant, calling fn on each node.
// Returning false from fn skips that node's children (but not its siblings).
func WalkRichText(r RichText, fn func(RichText) bool) {
	if r == nil {
		return
	}
	if !fn(r) {
		return
	}
	for _, child := range r.Children() {
		WalkRichText(child, fn)
	}
}

// WalkRichBlock recursively visits b and every descendant block, calling fnBlock
// on each. For each block, all RichText fields are also walked via fnText.
// Returning false from fnBlock skips that block's children.
func WalkRichBlock(b RichBlock, fnBlock func(RichBlock) bool, fnText func(RichText) bool) {
	if b == nil {
		return
	}
	if !fnBlock(b) {
		return
	}
	for _, child := range b.RichTextChildren() {
		WalkRichText(child, fnText)
	}
	for _, child := range b.RichBlockChildren() {
		WalkRichBlock(child, fnBlock, fnText)
	}
}

// Walk visits every RichBlock and RichText node in a RichMessage.
// Pass nil for either callback to skip that class of node.
func (m RichMessage) Walk(fnBlock func(RichBlock) bool, fnText func(RichText) bool) {
	if fnBlock == nil {
		fnBlock = func(RichBlock) bool { return true }
	}
	if fnText == nil {
		fnText = func(RichText) bool { return true }
	}
	for _, block := range m.Blocks {
		WalkRichBlock(block, fnBlock, fnText)
	}
}

// =============================================================================
// RichTextContent - flat text extraction
// =============================================================================

// RichTextContent walks r and concatenates all RichTextString leaf nodes,
// producing the plain-text content with no formatting. This is equivalent to
// the generated GetText() methods but works on any RichText value including
// RichTextArray and the interface itself.
func RichTextContent(r RichText) string {
	var sb strings.Builder
	WalkRichText(r, func(node RichText) bool {
		switch v := node.(type) {
		case RichTextString:
			sb.WriteString(string(v))
		case RichTextMathematicalExpression:
			sb.WriteString(v.Expression)
		case RichTextCustomEmoji:
			sb.WriteString(v.AlternativeText)
		}
		return true
	})
	return sb.String()
}

// RichBlockContent walks b and returns its full plain-text content, descending
// into nested blocks and text nodes. Block-level elements are separated by
// newlines appropriate to their type.
func RichBlockContent(b RichBlock) string {
	var sb strings.Builder
	renderBlockText(b, &sb)
	return sb.String()
}

func renderBlockText(b RichBlock, sb *strings.Builder) {
	switch v := b.(type) {
	case RichBlockParagraph:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
	case RichBlockSectionHeading:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
	case RichBlockPreformatted:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
	case RichBlockFooter:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
	case RichBlockBlockQuotation:
		for _, child := range v.Blocks {
			renderBlockText(child, sb)
		}
		if v.Credit != nil {
			sb.WriteString(RichTextContent(v.Credit))
			sb.WriteString("\n")
		}
	case RichBlockPullQuotation:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
		if v.Credit != nil {
			sb.WriteString(RichTextContent(v.Credit))
			sb.WriteString("\n")
		}
	case RichBlockDetails:
		sb.WriteString(RichTextContent(v.Summary))
		sb.WriteString("\n")
		for _, child := range v.Blocks {
			renderBlockText(child, sb)
		}
	case RichBlockList:
		for _, item := range v.Items {
			for _, child := range item.Blocks {
				renderBlockText(child, sb)
			}
		}
		sb.WriteString("\n")
	case RichBlockTable:
		for _, row := range v.Cells {
			for j, cell := range row {
				if j > 0 {
					sb.WriteString("\t")
				}
				if cell.Text != nil {
					sb.WriteString(RichTextContent(cell.Text))
				}
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	case RichBlockCollage:
		for _, child := range v.Blocks {
			sb.WriteString(RichBlockContent(child))
		}
		captionContent(v.Caption, sb)

	case RichBlockSlideshow:
		for _, child := range v.Blocks {
			sb.WriteString(RichBlockContent(child))
		}
		captionContent(v.Caption, sb)

	case RichBlockThinking:
		sb.WriteString(RichTextContent(v.Text))
		sb.WriteString("\n")
	case RichBlockMathematicalExpression:
		sb.WriteString(v.Expression)
		sb.WriteString("\n")
	case RichBlockDivider:
		sb.WriteString("\n")
	case RichBlockAnchor:
		// No text content.
	case RichBlockPhoto:
		captionContent(v.Caption, sb)
	case RichBlockAnimation:
		captionContent(v.Caption, sb)
	case RichBlockVideo:
		captionContent(v.Caption, sb)
	case RichBlockAudio:
		captionContent(v.Caption, sb)
	case RichBlockVoiceNote:
		captionContent(v.Caption, sb)
	case RichBlockMap:
		captionContent(v.Caption, sb)
	case RichBlockButtons:
		for _, btn := range v.Buttons {
			sb.WriteString(RichTextContent(btn.Text) + "\n")
			if btn.CopyText != nil {
				sb.WriteString(btn.CopyText.Text + "\n")
			}
		}
	}
}

func captionContent(caption *RichBlockCaption, sb *strings.Builder) {
	if caption == nil {
		return
	}

	sb.WriteString(RichTextContent(caption.Text))
	if caption.Credit != nil {
		sb.WriteString("\n")
		sb.WriteString(RichTextContent(caption.Credit))
	}
	sb.WriteString("\n")
}

// =============================================================================
// Shared helpers
// =============================================================================

// mdEscape escapes characters that have special meaning in CommonMark.
var mdSpecial = strings.NewReplacer(
	`\`, `\\`,
	"`", "\\`",
	"*", `\*`,
	"_", `\_`,
	"[", `\[`,
	"]", `\]`,
	"(", `\(`,
	")", `\)`,
	"#", `\#`,
	"+", `\+`,
	"-", `\-`,
	"!", `\!`,
	"|", `\|`,
	"\n", "<br/>",
)

func mdEscape(s string) string {
	return mdSpecial.Replace(html.EscapeString(s))
}

func toRoman(num int64) string {
	values := []int64{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

	result := ""
	for i := 0; i < len(values); i++ {
		for num >= values[i] {
			result += symbols[i]
			num -= values[i]
		}
	}
	return result
}

func toLetters(num int64) string {
	if num <= 0 {
		return ""
	}

	result := ""
	for num > 0 {
		num-- // shift into 0-25 range for this digit
		letter := rune('a' + (num % 26))
		result = string(letter) + result
		num /= 26
	}
	return result
}
