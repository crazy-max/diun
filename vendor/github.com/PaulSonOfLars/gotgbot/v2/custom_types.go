package gotgbot

type RichTextString string

func (r RichTextString) GetType() string {
	return RichTextTypeString
}

func (r RichTextString) richText() {}

func (v RichTextString) Children() []RichText { return nil }

type RichTextArray []RichText

func (r RichTextArray) GetType() string {
	return "array"
}

func (r RichTextArray) richText() {}

func (v RichTextArray) Children() []RichText { return []RichText(v) }

type RichBlockArray []RichBlock

type RichBlockListItemArray []RichBlockListItem

func (v RichTextUnknown) Children() []RichText { return nil }

func (v RichBlockUnknown) RichBlockChildren() []RichBlock { return nil }

func (v RichBlockUnknown) RichTextChildren() []RichText { return nil }

// Voice note bits; technically NOT valid, since "voicenote" arent InputMedia (since they cant be edited message content).
// Need to add better support for multi-type values in the future.
func (v InputMediaVoiceNote) GetType() string {
	return "voice_note"
}

func (v InputMediaVoiceNote) GetMedia() InputFileOrString {
	return v.Media
}

func (v InputMediaVoiceNote) inputMedia() {}
