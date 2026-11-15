package gotgbot

import (
	"html"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

var mdMap = map[string]string{
	"bold":   "*",
	"italic": "_",
	"code":   "`",
}

var mdV2Map = map[string]string{
	"bold":                  "*",
	"italic":                "_",
	"code":                  "`",
	"pre":                   "```",
	"underline":             "__",
	"strikethrough":         "~",
	"spoiler":               "||",
	"blockquote":            ">",
	"expandable_blockquote": "**>",
}

var htmlMap = map[string]string{
	"bold":                  "b",
	"italic":                "i",
	"code":                  "code",
	"pre":                   "pre",
	"underline":             "u",
	"strikethrough":         "s",
	"spoiler":               "span class=\"tg-spoiler\"",
	"blockquote":            "blockquote",
	"expandable_blockquote": "blockquote expandable",
}

// OriginalMD gets the original markdown formatting of a message text.
func (m Message) OriginalMD() string {
	return getOrigMsgMD(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalMDV2 gets the original markdownV2 formatting of a message text.
func (m Message) OriginalMDV2() string {
	return getOrigMsgMDV2(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalHTML gets the original HTML formatting of a message text.
func (m Message) OriginalHTML() string {
	return getOrigMsgHTML(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalCaptionMD gets the original markdown formatting of a message caption.
func (m Message) OriginalCaptionMD() string {
	return getOrigMsgMD(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalCaptionMDV2 gets the original markdownV2 formatting of a message caption.
func (m Message) OriginalCaptionMDV2() string {
	return getOrigMsgMDV2(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalCaptionHTML gets the original HTML formatting of a message caption.
func (m Message) OriginalCaptionHTML() string {
	return getOrigMsgHTML(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalTextMD gets the original markdown formatting of a message text or caption.
func (m Message) OriginalTextMD() string {
	return getOrigMsgMD(utf16.Encode([]rune(m.GetText())), m.GetEntities())
}

// OriginalTextMDV2 gets the original markdownV2 formatting of a message text or caption.
func (m Message) OriginalTextMDV2() string {
	return getOrigMsgMDV2(utf16.Encode([]rune(m.GetText())), m.GetEntities())
}

// OriginalTextHTML gets the original HTML formatting of a message text caption.
func (m Message) OriginalTextHTML() string {
	return getOrigMsgHTML(utf16.Encode([]rune(m.GetText())), m.GetEntities())
}

// GetRichMessage returns the RichMessage representation of the message - converting any MessageEntity types to the
// new RichText/RichBlock syntax, for consistent usage.
// NOTE: Does not include media (since documents cannot be expressed in RichText, it would be inconsistent).
func (m Message) GetRichMessage() *RichMessage {
	if m.RichMessage != nil {
		return m.RichMessage
	}
	return &RichMessage{Blocks: entitiesToRichBlocks(m.GetText(), m.GetEntities())}
}

// Does not support nesting. only look at upper entities.
func getOrigMsgMD(utf16Data []uint16, ents []MessageEntity) string {
	out := strings.Builder{}
	prev := int64(0)
	for _, ent := range getUpperEntities(ents) {
		newPrev := ent.Offset + ent.Length
		prevText := string(utf16.Decode(utf16Data[prev:ent.Offset]))

		text := utf16.Decode(utf16Data[ent.Offset:newPrev])
		pre, cleanCntnt, post := splitEdgeWhitespace(string(text), ent)
		cleanCntntRune := []rune(cleanCntnt)

		switch ent.Type {
		case "bold", "italic", "code":
			out.WriteString(prevText + pre + mdMap[ent.Type] + escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
		case "pre":
			if ent.Language == "" {
				out.WriteString(prevText + pre + mdMap[ent.Type] + escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
			} else {
				out.WriteString(prevText + pre + mdMap[ent.Type] + ent.Language + "\n" + escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
			}
		case "text_mention":
			out.WriteString(prevText + pre + "[" + escapeContainedMDV1(cleanCntntRune, []rune("[]()")) + "](tg://user?id=" + strconv.FormatInt(ent.User.Id, 10) + ")" + post)
		case "text_link":
			out.WriteString(prevText + pre + "[" + escapeContainedMDV1(cleanCntntRune, []rune("[]()")) + "](" + ent.Url + ")" + post)
		default:
			continue
		}
		prev = newPrev
	}

	out.WriteString(string(utf16.Decode(utf16Data[prev:])))
	return out.String()
}

func getOrigMsgHTML(utf16Data []uint16, ents []MessageEntity) string {
	if len(ents) == 0 {
		return html.EscapeString(string(utf16.Decode(utf16Data)))
	}

	bd := strings.Builder{}
	prev := int64(0)
	for _, e := range getUpperEntities(ents) {
		data, end := fillNestedHTML(utf16Data, e, prev, getChildEntities(e, ents))
		bd.WriteString(data)
		prev = end
	}

	bd.WriteString(html.EscapeString(string(utf16.Decode(utf16Data[prev:]))))
	return bd.String()
}

func getOrigMsgMDV2(utf16Data []uint16, ents []MessageEntity) string {
	if len(ents) == 0 {
		return string(utf16.Decode(utf16Data))
	}

	bd := strings.Builder{}
	prev := int64(0)
	for _, e := range getUpperEntities(ents) {
		data, end := fillNestedMarkdownV2(utf16Data, e, prev, getChildEntities(e, ents))
		bd.WriteString(data)
		prev = end
	}

	bd.WriteString(string(utf16.Decode(utf16Data[prev:])))
	return bd.String()
}

func fillNestedHTML(data []uint16, ent MessageEntity, start int64, entities []MessageEntity) (string, int64) {
	entEnd := ent.Offset + ent.Length
	if len(entities) == 0 || entEnd < entities[0].Offset {
		// no nesting; just return straight away and move to next.
		return writeFinalHTML(data, ent, start, html.EscapeString(string(utf16.Decode(data[ent.Offset:entEnd])))), entEnd
	}
	subPrev := ent.Offset
	subEnd := ent.Offset
	bd := strings.Builder{}
	for _, e := range getUpperEntities(entities) {
		if e.Offset < subEnd || e == ent {
			continue
		}
		if e.Offset >= entEnd {
			break
		}

		out, end := fillNestedHTML(data, e, subPrev, getChildEntities(e, entities))
		bd.WriteString(out)
		subPrev = end
	}

	bd.WriteString(html.EscapeString(string(utf16.Decode(data[subPrev:entEnd]))))

	return writeFinalHTML(data, ent, start, bd.String()), entEnd
}

func fillNestedMarkdownV2(data []uint16, ent MessageEntity, start int64, entities []MessageEntity) (string, int64) {
	entEnd := ent.Offset + ent.Length
	if len(entities) == 0 || entEnd < entities[0].Offset {
		// no nesting; just return straight away and move to next.
		return writeFinalMarkdownV2(data, ent, start, string(utf16.Decode(data[ent.Offset:entEnd]))), entEnd
	}
	subPrev := ent.Offset
	subEnd := ent.Offset
	bd := strings.Builder{}
	for _, e := range getUpperEntities(entities) {
		if e.Offset < subEnd || e == ent {
			continue
		}
		if e.Offset >= entEnd {
			break
		}

		out, end := fillNestedMarkdownV2(data, e, subPrev, getChildEntities(e, entities))
		bd.WriteString(out)
		subPrev = end
	}

	bd.WriteString(string(utf16.Decode(data[subPrev:entEnd])))

	return writeFinalMarkdownV2(data, ent, start, bd.String()), entEnd
}

func writeFinalHTML(data []uint16, ent MessageEntity, start int64, cntnt string) string {
	prevText := html.EscapeString(string(utf16.Decode(data[start:ent.Offset])))
	switch ent.Type {
	case "bold", "italic", "code", "underline", "strikethrough", "spoiler":
		// <b>bold</b>, <strong>bold</strong>
		// <i>italic</i>, <em>italic</em>
		// <u>underline</u>, <ins>underline</ins>
		// <s>strikethrough</s>, <strike>strikethrough</strike>, <del>strikethrough</del>
		// <span class="tg-spoiler">spoiler</span>, <tg-spoiler>spoiler</tg-spoiler>
		// <b>bold <i>italic bold <s>italic bold strikethrough <span class="tg-spoiler">italic bold strikethrough spoiler</span></s> <u>underline italic bold</u></i> bold</b>
		// <code>inline fixed-width code</code>
		return prevText + "<" + htmlMap[ent.Type] + ">" + cntnt + "</" + closeHTMLTag(htmlMap[ent.Type]) + ">"
	case "pre":
		// <pre>pre-formatted fixed-width code block</pre>
		// <pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>

		// <pre>text</pre>
		if ent.Language == "" {
			return prevText + "<pre>" + cntnt + "</pre>"
		}
		// <pre><code class="lang">text</code></pre>
		return prevText + `<pre><code class="` + ent.Language + `">` + cntnt + "</code></pre>"
	case "custom_emoji":
		// <tg-emoji emoji-id="5368324170671202286">👍</tg-emoji>
		return prevText + `<tg-emoji emoji-id="` + ent.CustomEmojiId + `">` + cntnt + "</tg-emoji>"
	case "date_time":
		// <tg-time unix="1647531900" format="wDT">22:45 tomorrow</tg-time>
		// <tg-time unix="1647531900" format="t">22:45 tomorrow</tg-time>
		// <tg-time unix="1647531900" format="r">22:45 tomorrow</tg-time>
		// <tg-time unix="1647531900">22:45 tomorrow</tg-time>
		if ent.DateTimeFormat != "" {
			return prevText + `<tg-time unix="` + strconv.FormatInt(ent.UnixTime, 10) + `" format="` + ent.DateTimeFormat + `">` + cntnt + "</tg-time>"
		}
		return prevText + `<tg-time unix="` + strconv.FormatInt(ent.UnixTime, 10) + `">` + cntnt + "</tg-time>"
	case "text_mention":
		// <a href="tg://user?id=123456789">inline mention of a user</a>
		return prevText + `<a href="tg://user?id=` + strconv.FormatInt(ent.User.Id, 10) + `">` + cntnt + "</a>"
	case "text_link":
		// <a href="http://www.example.com/">inline URL</a>
		return prevText + `<a href="` + ent.Url + `">` + cntnt + "</a>"
	case "blockquote":
		// <blockquote>Block quotation started\nBlock quotation continued\nThe last line of the block quotation</blockquote>
		return prevText + `<blockquote>` + cntnt + "</blockquote>"
	case "expandable_blockquote":
		// <blockquote expandable>Expandable block quotation started\nExpandable block quotation continued\nExpandable block quotation continued\nHidden by default part of the block quotation started\nExpandable block quotation continued\nThe last line of the block quotation</blockquote>
		return prevText + `<blockquote expandable>` + cntnt + "</blockquote>"
	default:
		return prevText + cntnt
	}
}

// closeHTMLTag makes sure to generate the correct HTML closing tag for a given opening tag.
func closeHTMLTag(s string) string {
	if !strings.HasPrefix(s, "span") {
		return s
	}
	return "span"
}

func writeFinalMarkdownV2(data []uint16, ent MessageEntity, start int64, cntnt string) string {
	prevText := string(utf16.Decode(data[start:ent.Offset]))
	pre, cleanCntnt, post := splitEdgeWhitespace(cntnt, ent)
	switch ent.Type {
	case "bold", "italic", "code", "underline", "strikethrough", "spoiler":
		// *bold \*text*
		// _italic \*text_
		// __underline__
		// ~strikethrough~
		// ||spoiler||
		// *bold _italic bold ~italic bold strikethrough ||italic bold strikethrough spoiler||~ __underline italic bold___ bold*
		// `inline fixed-width code`
		return prevText + pre + mdV2Map[ent.Type] + cleanCntnt + mdV2Map[ent.Type] + post
	case "pre":
		// ```
		// pre-formatted fixed-width code block
		// ```
		// ```python
		// pre-formatted fixed-width code block written in the Python programming language
		// ```
		if ent.Language == "" {
			return prevText + pre + "```\n" + cleanCntnt + "```" + post
		}
		return prevText + pre + "```" + ent.Language + "\n" + cleanCntnt + "```" + post
	case "custom_emoji":
		// Yes, custom emoji have a weird little ! at the front
		// https://core.telegram.org/bots/api#markdownv2-style
		// ![👍](tg://emoji?id=5368324170671202286)
		return prevText + pre + "![" + cleanCntnt + "](tg://emoji?id=" + ent.CustomEmojiId + ")" + post
	case "date_time":
		// ![22:45 tomorrow](tg://time?unix=1647531900&format=wDT)
		// ![22:45 tomorrow](tg://time?unix=1647531900&format=t)
		// ![22:45 tomorrow](tg://time?unix=1647531900&format=r)
		// ![22:45 tomorrow](tg://time?unix=1647531900)
		if ent.DateTimeFormat != "" {
			return prevText + pre + "![" + cleanCntnt + "](tg://time?unix=" + strconv.FormatInt(ent.UnixTime, 10) + "&format=" + ent.DateTimeFormat + ")" + post
		}
		return prevText + pre + "![" + cleanCntnt + "](tg://time?unix=" + strconv.FormatInt(ent.UnixTime, 10) + ")" + post
	case "text_mention":
		// [inline mention of a user](tg://user?id=123456789)
		return prevText + pre + "[" + cleanCntnt + "](tg://user?id=" + strconv.FormatInt(ent.User.Id, 10) + ")" + post
	case "text_link":
		// [inline URL](http://www.example.com/)
		return prevText + pre + "[" + cleanCntnt + "](" + ent.Url + ")" + post
	case "blockquote":
		// >Block quotation started
		// >Block quotation continued
		// >Block quotation continued
		// >Block quotation continued
		// >The last line of the block quotation
		return prevText + pre + ">" + strings.Join(strings.Split(cleanCntnt, "\n"), "\n>") + post
	case "expandable_blockquote":
		// **>The expandable block quotation started right after the previous block quotation
		// >It is separated from the previous block quotation by an empty bold entity
		// >Expandable block quotation continued
		// >Hidden by default part of the expandable block quotation started
		// >Expandable block quotation continued
		// >The last line of the expandable block quotation with the expandability mark||
		return prevText + pre + "**>" + strings.Join(strings.Split(cleanCntnt, "\n"), "\n>") + "||" + post
	default:
		return prevText + cntnt
	}
}

func getUpperEntities(ents []MessageEntity) []MessageEntity {
	prev := int64(0)
	uppers := make([]MessageEntity, 0, len(ents))
	for _, e := range ents {
		if e.Offset < prev {
			continue
		}
		uppers = append(uppers, e)
		prev = e.Offset + e.Length
	}
	return uppers
}

func getChildEntities(ent MessageEntity, ents []MessageEntity) []MessageEntity {
	end := ent.Offset + ent.Length
	children := make([]MessageEntity, 0, len(ents))
	for _, e := range ents {
		if e.Offset < ent.Offset || e == ent {
			continue
		}
		if e.Offset >= end {
			break
		}
		children = append(children, e)
	}
	return children
}

func splitEdgeWhitespace(text string, ent MessageEntity) (pre string, cntnt string, post string) {
	keepNewLines := ent.Type == "pre"

	bd := strings.Builder{}
	rText := []rune(text)
	for i := 0; i < len(rText) && unicode.IsSpace(rText[i]) && (!keepNewLines || rText[i] != '\n'); i++ {
		bd.WriteRune(rText[i])
	}
	pre = bd.String()

	text = strings.TrimPrefix(text, pre)
	bd.Reset()
	for i := len(rText) - 1; i >= 0 && unicode.IsSpace(rText[i]); i-- {
		bd.WriteRune(rText[i])
	}
	post = bd.String()
	return pre, strings.TrimSuffix(text, post), post
}

func escapeContainedMDV1(data []rune, mdType []rune) string {
	out := strings.Builder{}
	for _, x := range data {
		if slices.Contains(mdType, x) {
			out.WriteRune('\\')
		}
		out.WriteRune(x)
	}
	return out.String()
}

type utf16Text struct {
	units []uint16
}

func newUTF16Text(s string) utf16Text {
	return utf16Text{units: utf16.Encode([]rune(s))}
}

func (t utf16Text) slice(start, end int64) string {
	s, e := int(start), int(end)
	if s < 0 {
		s = 0
	}
	if e > len(t.units) {
		e = len(t.units)
	}
	if s >= e {
		return ""
	}
	return string(utf16.Decode(t.units[s:e]))
}

var blockLevelEntityTypes = map[string]bool{
	"pre":                   true,
	"blockquote":            true,
	"expandable_blockquote": true,
}

// entitiesToRichBlocks converts classic text + MessageEntity input into the
// same []RichBlock shape native RichMessage uses.
// entities are assumed sorted ascending by Offset, with ties broken by descending
// Length (i.e. a wider entity starting at the same offset as a narrower one comes first) - this
// matches what getUpperEntities/getChildEntities already assume, and is how
// Telegram emits entities for properly-nested formatting.
func entitiesToRichBlocks(text string, entities []MessageEntity) []RichBlock {
	t := newUTF16Text(text)
	top := getUpperEntities(entities)

	var blocks []RichBlock
	var run []RichText
	pos := int64(0)

	flush := func() {
		if content := joinRichText(run); content != nil {
			blocks = append(blocks, RichBlockParagraph{Text: content})
		}
		run = nil
	}

	for _, e := range top {
		if gap := t.slice(pos, e.Offset); gap != "" {
			run = append(run, RichTextString(gap))
		}

		if blockLevelEntityTypes[e.Type] {
			flush()
			blocks = append(blocks, entityToRichBlock(t, entities, e))
		} else {
			run = append(run, richTextFromEntity(t, entities, e))
		}
		pos = e.Offset + e.Length
	}
	if tail := t.slice(pos, int64(len(t.units))); tail != "" {
		run = append(run, RichTextString(tail))
	}
	flush()

	return blocks
}

// richTextFromEntity converts one entity - including its nested descendants
// and the plain-text gaps between them - into a single wrapped RichText node.
func richTextFromEntity(t utf16Text, all []MessageEntity, ent MessageEntity) RichText {
	children := getUpperEntities(getChildEntities(ent, all))
	content := richTextSequence(t, ent.Offset, ent.Offset+ent.Length, children, all)
	return wrapEntity(ent, content)
}

// richTextSequence builds the RichText content for [start,end), given the
// direct-child entities within that span, filling gaps with plain text.
func richTextSequence(t utf16Text, start, end int64, directChildren, all []MessageEntity) RichText {
	var nodes []RichText
	pos := start
	for _, e := range directChildren {
		if gap := t.slice(pos, e.Offset); gap != "" {
			nodes = append(nodes, RichTextString(gap))
		}
		nodes = append(nodes, richTextFromEntity(t, all, e))
		pos = e.Offset + e.Length
	}
	if gap := t.slice(pos, end); gap != "" {
		nodes = append(nodes, RichTextString(gap))
	}
	return joinRichText(nodes)
}

func joinRichText(nodes []RichText) RichText {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		return nodes[0]
	default:
		return RichTextArray(nodes)
	}
}

func wrapEntity(e MessageEntity, content RichText) RichText {
	switch e.Type {
	case MessageEntityTypeBold:
		return RichTextBold{Text: content}
	case MessageEntityTypeItalic:
		return RichTextItalic{Text: content}
	case MessageEntityTypeUnderline:
		return RichTextUnderline{Text: content}
	case MessageEntityTypeStrikethrough:
		return RichTextStrikethrough{Text: content}
	case MessageEntityTypeSpoiler:
		return RichTextSpoiler{Text: content}
	case MessageEntityTypeCode:
		return RichTextCode{Text: content}
	case MessageEntityTypeTextLink:
		return RichTextUrl{Text: content, Url: e.Url}
	case MessageEntityTypeUrl:
		return RichTextUrl{Text: content, Url: RichTextContent(content)}
	case MessageEntityTypeMention:
		return RichTextMention{Text: content, Username: strings.TrimPrefix(RichTextContent(content), "@")}
	case MessageEntityTypeHashtag:
		return RichTextHashtag{Text: content, Hashtag: strings.TrimPrefix(RichTextContent(content), "#")}
	case MessageEntityTypeCashtag:
		return RichTextCashtag{Text: content, Cashtag: strings.TrimPrefix(RichTextContent(content), "$")}
	case MessageEntityTypeBotCommand:
		return RichTextBotCommand{Text: content, BotCommand: strings.TrimPrefix(RichTextContent(content), "/")}
	case MessageEntityTypeEmail:
		return RichTextEmailAddress{Text: content, EmailAddress: RichTextContent(content)}
	case MessageEntityTypePhoneNumber:
		return RichTextPhoneNumber{Text: content, PhoneNumber: RichTextContent(content)}
	case MessageEntityTypeTextMention:
		return RichTextTextMention{Text: content, User: *e.User}
	case MessageEntityTypeCustomEmoji:
		return RichTextCustomEmoji{CustomEmojiId: e.CustomEmojiId, AlternativeText: RichTextContent(content)}
	case MessageEntityTypeDateTime:
		return RichTextDateTime{Text: content, UnixTime: e.UnixTime, DateTimeFormat: e.DateTimeFormat}
	default:
		// No match? treat it as plain text.
		return content
	}
}

func entityToRichBlock(t utf16Text, all []MessageEntity, ent MessageEntity) RichBlock {
	children := getUpperEntities(getChildEntities(ent, all))
	inner := richTextSequence(t, ent.Offset, ent.Offset+ent.Length, children, all)

	switch ent.Type {
	case MessageEntityTypePre:
		return RichBlockPreformatted{Text: inner, Language: ent.Language}
	case MessageEntityTypeBlockquote, MessageEntityTypeExpandableBlockquote:
		// note: richtext does not support collapsible blockquotes.
		return RichBlockBlockQuotation{
			Blocks: []RichBlock{RichBlockParagraph{Text: inner}},
		}
	default:
		return RichBlockParagraph{Text: inner}
	}
}
