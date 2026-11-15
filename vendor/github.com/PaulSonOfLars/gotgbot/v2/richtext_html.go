package gotgbot

import (
	"fmt"
	"html"
	"strings"
)

// RichTextHTML renders r as an HTML string, using standard HTML tags for each
// RichText node type. The output is suitable for embedding in an HTML document.
func RichTextHTML(rt RichText) string {
	var r renderCtx
	r.renderTextHTML(rt)
	return r.sb.String()
}

func (r *renderCtx) renderTextHTML(rt RichText) {
	if r == nil {
		return
	}
	switch v := rt.(type) {
	case RichTextString:
		r.sb.WriteString(strings.ReplaceAll(html.EscapeString(string(v)), "\n", "<br/>"))
	case RichTextArray:
		for _, child := range v {
			r.renderTextHTML(child)
		}

	// Formatting wrappers - recurse into .Text
	case RichTextBold:
		r.sb.WriteString("<b>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</b>")
	case RichTextItalic:
		r.sb.WriteString("<i>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</i>")
	case RichTextUnderline:
		r.sb.WriteString("<u>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</u>")
	case RichTextStrikethrough:
		r.sb.WriteString("<s>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</s>")
	case RichTextSpoiler:
		r.sb.WriteString(`<tg-spoiler>`)
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</tg-spoiler>")
	case RichTextMarked:
		r.sb.WriteString("<mark>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</mark>")
	case RichTextSubscript:
		r.sb.WriteString("<sub>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</sub>")
	case RichTextSuperscript:
		r.sb.WriteString("<sup>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</sup>")
	case RichTextCode:
		r.sb.WriteString("<code>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</code>")

	// Links and references
	case RichTextUrl:
		fmt.Fprintf(&r.sb, `<a href="%s">`, html.EscapeString(v.Url))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")
	case RichTextEmailAddress:
		fmt.Fprintf(&r.sb, `<a href="mailto:%s">`, html.EscapeString(v.EmailAddress))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")
	case RichTextPhoneNumber:
		fmt.Fprintf(&r.sb, `<a href="tel:%s">`, html.EscapeString(v.PhoneNumber))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")
	case RichTextTextMention:
		fmt.Fprintf(&r.sb, `<a href="tg://user?id=%d">`, v.User.Id)
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")
	case RichTextMention:
		r.renderTextHTML(v.Text)
	case RichTextAnchor:
		fmt.Fprintf(&r.sb, `<a name="%s"></a>`, html.EscapeString(v.Name))
	case RichTextAnchorLink:
		fmt.Fprintf(&r.sb, `<a href="#%s">`, html.EscapeString(v.AnchorName))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")
	case RichTextReference:
		fmt.Fprintf(&r.sb, `<tg-reference name="%s">`, html.EscapeString(v.Name))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</tg-reference>")
	case RichTextReferenceLink:
		fmt.Fprintf(&r.sb, `<a href="#%s">`, html.EscapeString(v.ReferenceName))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</a>")

	// Entities with no wrapping text child
	case RichTextHashtag:
		r.renderTextHTML(v.Text)
	case RichTextCashtag:
		r.renderTextHTML(v.Text)
	case RichTextBotCommand:
		r.renderTextHTML(v.Text)
	case RichTextButton:
		r.renderRichMessageButton(v.Button)
	case RichTextBankCardNumber:
		r.renderTextHTML(v.Text)
	case RichTextDateTime:
		fmt.Fprintf(&r.sb, `<tg-time unix="%d" format="%s">`, v.UnixTime, html.EscapeString(v.DateTimeFormat))
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</tg-time>")
	case RichTextCustomEmoji:
		fmt.Fprintf(&r.sb, `<tg-emoji emoji-id="%s">%s</tg-emoji>`, html.EscapeString(v.CustomEmojiId), html.EscapeString(v.AlternativeText))
	case RichTextMathematicalExpression:
		fmt.Fprintf(&r.sb, `<tg-math>%s</tg-math>`, html.EscapeString(v.Expression))
	}
}

func (r *renderCtx) renderRichMessageButton(btn RichMessageButton) {
	var args []string
	if btn.Style != "" {
		args = append(args, "style=\""+html.EscapeString(btn.Style)+"\"")
	}

	btnType := "disabled"
	switch {
	case btn.Url != "":
		btnType = "url"
		args = append(args, "url=\""+html.EscapeString(btn.Url)+"\"")
	case btn.CallbackData != "":
		btnType = "callback_data"
		args = append(args, "data=\""+html.EscapeString(btn.CallbackData)+"\"")
	case btn.WebApp != nil:
		btnType = "web_app"
		args = append(args, "url=\""+html.EscapeString(btn.WebApp.Url)+"\"")
	case btn.LoginUrl != nil:
		btnType = "login_url"
		args = append(args, "url=\""+html.EscapeString(btn.LoginUrl.Url)+"\"")
		if btn.LoginUrl.ForwardText != "" {
			args = append(args, "forward_text=\""+html.EscapeString(btn.LoginUrl.ForwardText)+"\"")
		}
		if btn.LoginUrl.RequestWriteAccess {
			args = append(args, "request_write_access")
		}
	case btn.SwitchInlineQuery != nil:
		btnType = "switch_inline_query"
		args = append(args, "query=\""+html.EscapeString(*btn.SwitchInlineQuery)+"\"")
	case btn.SwitchInlineQueryCurrentChat != nil:
		btnType = "switch_inline_query_current_chat"
		args = append(args, "query=\""+html.EscapeString(*btn.SwitchInlineQueryCurrentChat)+"\"")
	case btn.SwitchInlineQueryChosenChat != nil:
		btnType = "switch_inline_query_chosen_chat"
		args = append(args, "query=\""+html.EscapeString(btn.SwitchInlineQueryChosenChat.Query)+"\"")
		if btn.SwitchInlineQueryChosenChat.AllowUserChats {
			args = append(args, "allow_user_chats")
		}
		if btn.SwitchInlineQueryChosenChat.AllowBotChats {
			args = append(args, "allow_bot_chats")
		}
		if btn.SwitchInlineQueryChosenChat.AllowGroupChats {
			args = append(args, "allow_group_chats")
		}
		if btn.SwitchInlineQueryChosenChat.AllowChannelChats {
			args = append(args, "allow_channel_chats")
		}
	case btn.CopyText != nil:
		btnType = "copy_text"
		args = append(args, "text=\""+html.EscapeString(btn.CopyText.Text)+"\"")
	default:
		// unknown
	}

	fmt.Fprintf(&r.sb, `<tg-button type="%s" %s>`, btnType, strings.Join(args, " "))
	r.renderTextHTML(btn.Text)
	r.sb.WriteString("</tg-button>\n")
}

// RichBlockHTML renders a RichBlock and its descendants as an HTML fragment.
func RichBlockHTML(b RichBlock) (string, []InputRichMessageMedia) {
	var r renderCtx
	r.renderBlockHTML(b)
	return r.sb.String(), r.media
}

func (r *renderCtx) renderBlockHTML(b RichBlock) {
	switch v := b.(type) {
	case RichBlockParagraph:
		r.sb.WriteString("<p>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</p>\n")
	case RichBlockSectionHeading:
		tag := fmt.Sprintf("h%d", v.Size)
		r.sb.WriteString("<" + tag + ">")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</" + tag + ">\n")
	case RichBlockPreformatted:
		if v.Language != "" {
			fmt.Fprintf(&r.sb, `<pre><code class="language-%s">`, html.EscapeString(v.Language))
		} else {
			r.sb.WriteString("<pre><code>")
		}
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</code></pre>\n")
	case RichBlockFooter:
		r.sb.WriteString("<footer>")
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</footer>\n")
	case RichBlockBlockQuotation:
		r.sb.WriteString("<blockquote>\n")
		for _, child := range v.Blocks {
			r.renderBlockHTML(child)
		}
		if v.Credit != nil {
			r.sb.WriteString("<cite>")
			r.renderTextHTML(v.Credit)
			r.sb.WriteString("</cite>\n")
		}
		r.sb.WriteString("</blockquote>\n")
	case RichBlockExpandableBlockQuotation:
		r.sb.WriteString("<blockquote expandable>\n")
		r.renderTextHTML(v.Text)
		if v.Credit != nil {
			r.sb.WriteString("<cite>")
			r.renderTextHTML(v.Credit)
			r.sb.WriteString("</cite>\n")
		}
		r.sb.WriteString("</blockquote>\n")
	case RichBlockPullQuotation:
		r.sb.WriteString(`<aside>`)
		r.renderTextHTML(v.Text)
		if v.Credit != nil {
			r.sb.WriteString("<cite>")
			r.renderTextHTML(v.Credit)
			r.sb.WriteString("</cite>")
		}
		r.sb.WriteString("</aside>\n")
	case RichBlockDetails:
		r.sb.WriteString("<details")
		if v.IsOpen {
			r.sb.WriteString(" open")
		}
		r.sb.WriteString(">\n<summary>")
		r.renderTextHTML(v.Summary)
		r.sb.WriteString("</summary>\n")
		for _, child := range v.Blocks {
			r.renderBlockHTML(child)
		}
		r.sb.WriteString("</details>\n")
	case RichBlockList:
		if len(v.Items) == 0 {
			return
		}

		// Detect ordered vs unordered from the first item.
		first := v.Items[0]
		tag := "ul"
		listAttrs := ""
		if first.Value > 0 {
			tag = "ol"
		}
		if first.Value > 1 {
			listAttrs += fmt.Sprintf(" start=\"%d\"", first.Value)
		}
		if first.Type != "" && first.Type != "1" {
			listAttrs += fmt.Sprintf(" type=\"%s\"", first.Type)
		}
		if isReversed(v.Items) {
			listAttrs += " reversed"
		}

		fmt.Fprintf(&r.sb, "<%s%s>\n", tag, listAttrs)
		for _, item := range v.Items {
			itemAttrs := ""
			if tag == "ol" && item.Value > 1 {
				itemAttrs += fmt.Sprintf(" value=\"%d\"", item.Value)
			}
			if tag == "ol" && item.Type != "" && item.Type != "1" {
				itemAttrs += fmt.Sprintf(" type=\"%s\"", item.Type)
			}

			fmt.Fprintf(&r.sb, "<li%s>", itemAttrs)
			if item.HasCheckbox {
				if item.IsChecked {
					r.sb.WriteString(`<input type="checkbox" checked> `)
				} else {
					r.sb.WriteString(`<input type="checkbox"> `)
				}
			}
			if len(item.Blocks) > 0 {
				r.sb.WriteString("\n")
				for _, child := range item.Blocks {
					r.renderBlockHTML(child)
				}
			}
			r.sb.WriteString("</li>\n")
		}
		fmt.Fprintf(&r.sb, "</%s>\n", tag)
	case RichBlockTable:
		attrs := ""
		if v.IsBordered {
			attrs += " bordered"
		}
		if v.IsStriped {
			attrs += " striped"
		}
		if v.IsCompact {
			attrs += " compact"
		}

		fmt.Fprintf(&r.sb, "<table%s>\n", attrs)

		if v.Caption != nil {
			fmt.Fprintf(&r.sb, "<caption>")
			r.renderTextHTML(v.Caption)
			fmt.Fprintf(&r.sb, "</caption>\n")
		}

		for _, row := range v.Cells {
			r.sb.WriteString("<tr>\n")
			for _, cell := range row {
				tag := "td"
				if cell.IsHeader {
					tag = "th"
				}

				attrs := ""
				if cell.Align != "" &&
					// headers are "center" by default.
					// non-headers are "left" by default.
					((cell.IsHeader && cell.Align != "center") ||
						(!cell.IsHeader && cell.Align != "left")) {
					attrs += fmt.Sprintf(" align=\"%s\"", cell.Align)
				}
				if cell.Valign != "" && cell.Valign != "middle" {
					attrs += fmt.Sprintf(" valign=\"%s\"", cell.Valign)
				}
				if cell.Colspan > 1 {
					attrs += fmt.Sprintf(" colspan=\"%d\"", cell.Colspan)
				}
				if cell.Rowspan > 1 {
					attrs += fmt.Sprintf(" rowspan=\"%d\"", cell.Rowspan)
				}
				fmt.Fprintf(&r.sb, "<%s%s>", tag, attrs)
				if cell.Text != nil {
					r.renderTextHTML(cell.Text)
				}
				fmt.Fprintf(&r.sb, "</%s>\n", tag)
			}
			r.sb.WriteString("</tr>\n")
		}
		r.sb.WriteString("</table>\n")
	case RichBlockCollage:
		r.sb.WriteString("<tg-collage>\n")
		for _, child := range v.Blocks {
			r.renderBlockHTML(child)
		}
		r.renderCaptionHTML(v.Caption)
		r.sb.WriteString("</tg-collage>\n")
	case RichBlockSlideshow:
		r.sb.WriteString("<tg-slideshow>\n")
		for _, child := range v.Blocks {
			r.renderBlockHTML(child)
		}
		r.renderCaptionHTML(v.Caption)
		r.sb.WriteString("</tg-slideshow>\n")
	case RichBlockThinking:
		r.sb.WriteString(`<tg-thinking>`)
		r.renderTextHTML(v.Text)
		r.sb.WriteString("</tg-thinking>\n")
	case RichBlockMathematicalExpression:
		fmt.Fprintf(&r.sb, "<tg-math-block>%s</tg-math-block>\n", html.EscapeString(v.Expression))
	case RichBlockDivider:
		r.sb.WriteString("<hr/>\n")
	case RichBlockAnchor:
		fmt.Fprintf(&r.sb, `<a name="%s"></a>`+"\n", html.EscapeString(v.Name))
	case RichBlockPhoto:
		attrs := ""
		if v.HasSpoiler {
			attrs += " tg-spoiler"
		}

		tgAnimation := fmt.Sprintf("<img src=\"%s\"%s></img>\n",
			r.addFile("photo", InputMediaPhoto{Media: InputFileByID(v.Photo[len(v.Photo)-1].FileId)}), attrs)

		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgAnimation)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgAnimation)
		}
	case RichBlockAnimation:
		attrs := ""
		if v.HasSpoiler {
			attrs += " tg-spoiler"
		}
		tgAnimation := fmt.Sprintf("<video src=\"%s\"%s></video>\n",
			r.addFile("video", InputMediaAnimation{Media: InputFileByID(v.Animation.FileId)}), attrs)
		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgAnimation)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgAnimation)
		}
	case RichBlockVideo:
		attrs := ""
		if v.HasSpoiler {
			attrs += " tg-spoiler"
		}
		tgVideo := fmt.Sprintf("<video src=\"%s\"%s></video>\n",
			r.addFile("video", InputMediaVideo{Media: InputFileByID(v.Video.FileId)}), attrs)
		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgVideo)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgVideo)
		}
	case RichBlockAudio:
		tgAudio := fmt.Sprintf("<audio src=\"%s\"></audio>\n",
			r.addFile("audio", InputMediaAudio{Media: InputFileByID(v.Audio.FileId)}))
		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgAudio)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgAudio)
		}
	case RichBlockVoiceNote:
		tgAudio := fmt.Sprintf("<audio src=\"%s\"></audio>\n", r.addFile("audio", InputMediaVoiceNote{Type: "voice_note", Media: InputFileByID(v.VoiceNote.FileId)}))
		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgAudio)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgAudio)
		}
	case RichBlockDocument:
		tgDoc := fmt.Sprintf("<tg-document src=\"%s\"></tg-document>\n", r.addFile("document", InputMediaDocument{Media: InputFileByID(v.Document.FileId)}))
		if v.Caption != nil {
			r.sb.WriteString("<figure>\n")
			r.sb.WriteString(tgDoc)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>\n")
		} else {
			r.sb.WriteString(tgDoc)
		}
	case RichBlockMap:
		tgMap := fmt.Sprintf(
			`<tg-map lat="%g" long="%g" zoom="%d"/>`,
			v.Location.Latitude, v.Location.Longitude, v.Zoom)
		if v.Caption != nil {
			r.sb.WriteString("<figure>")
			r.sb.WriteString(tgMap)
			r.renderCaptionHTML(v.Caption)
			r.sb.WriteString("</figure>")
		} else {
			r.sb.WriteString(tgMap)
		}
		r.sb.WriteString("\n")
	case RichBlockButtons:
		var args string
		if v.Align != "" {
			args = " align=\"" + v.Align + "\""
		}

		r.sb.WriteString("<tg-button-row" + args + ">\n")
		for _, btn := range v.Buttons {
			r.renderRichMessageButton(btn)
		}

		r.sb.WriteString("</tg-button-row>\n")
	}
}

func (r *renderCtx) renderCaptionHTML(cap *RichBlockCaption) {
	if cap == nil {
		return
	}
	r.sb.WriteString("<figcaption>")
	r.renderTextHTML(cap.Text)
	if cap.Credit != nil {
		r.sb.WriteString("<cite>")
		r.renderTextHTML(cap.Credit)
		r.sb.WriteString("</cite>")
	}
	r.sb.WriteString("</figcaption>\n")
}
