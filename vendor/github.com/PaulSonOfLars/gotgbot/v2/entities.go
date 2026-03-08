package gotgbot

import "unicode/utf16"

type ParsedMessageEntity struct {
	MessageEntity
	Text string `json:"text"`
}

// ParseEntity parses a single MessageEntity into a ParsedMessageEntity.
func ParseEntity(text string, entity MessageEntity) ParsedMessageEntity {
	return parseEntity(entity, utf16.Encode([]rune(text)))
}

// ParseEntities parses all MessageEntity items into a list of ParsedMessageEntity.
func ParseEntities(text string, entities []MessageEntity) (out []ParsedMessageEntity) {
	return ParseEntityTypes(text, entities, nil)
}

// ParseEntityTypes parses a subset of MessageEntity items into a list of ParsedMessageEntity.
func ParseEntityTypes(text string, entities []MessageEntity, accepted map[string]struct{}) (out []ParsedMessageEntity) {
	utf16Text := utf16.Encode([]rune(text))
	for _, ent := range entities {
		if _, ok := accepted[ent.Type]; ok || accepted == nil {
			out = append(out, parseEntity(ent, utf16Text))
		}
	}
	return out
}

// ParseEntities parses all message text entities into a list of ParsedMessageEntity.
func (m Message) ParseEntities() (out []ParsedMessageEntity) {
	return m.ParseEntityTypes(nil)
}

// ParseCaptionEntities parses all message caption entities into a list of ParsedMessageEntity.
func (m Message) ParseCaptionEntities() (out []ParsedMessageEntity) {
	return m.ParseCaptionEntityTypes(nil)
}

// ParseAnyEntities parses all message entities in the entire message into a list of ParsedMessageEntity.
func (m Message) ParseAnyEntities() (out []ParsedMessageEntity) {
	return m.ParseAnyEntityTypes(nil)
}

// ParseEntityTypes parses a subset of message text entities into a list of ParsedMessageEntity.
func (m Message) ParseEntityTypes(accepted map[string]struct{}) (out []ParsedMessageEntity) {
	return ParseEntityTypes(m.Text, m.Entities, accepted)
}

// ParseCaptionEntityTypes parses a subset of message caption entities into a list of ParsedMessageEntity.
func (m Message) ParseCaptionEntityTypes(accepted map[string]struct{}) (out []ParsedMessageEntity) {
	return ParseEntityTypes(m.Caption, m.CaptionEntities, accepted)
}

// ParseAnyEntityTypes parses a subset of all entities in the entire message into a list of ParsedMessageEntity.
func (m Message) ParseAnyEntityTypes(accepted map[string]struct{}) (out []ParsedMessageEntity) {
	out = append(out, m.ParseEntityTypes(accepted)...)
	out = append(out, m.ParseCaptionEntityTypes(accepted)...)
	if m.Checklist != nil {
		out = append(out, ParseEntityTypes(m.Checklist.Title, m.Checklist.TitleEntities, accepted)...)
		for _, t := range m.Checklist.Tasks {
			out = append(out, ParseEntityTypes(t.Text, t.TextEntities, accepted)...)
		}
	}
	if m.Game != nil {
		out = append(out, ParseEntityTypes(m.Game.Text, m.Game.TextEntities, accepted)...)
	}
	if m.Gift != nil {
		out = append(out, ParseEntityTypes(m.Gift.Text, m.Gift.Entities, accepted)...)
	}
	if m.Poll != nil {
		out = append(out, ParseEntityTypes(m.Poll.Question, m.Poll.QuestionEntities, accepted)...)
		out = append(out, ParseEntityTypes(m.Poll.Explanation, m.Poll.ExplanationEntities, accepted)...)
		for _, o := range m.Poll.Options {
			out = append(out, ParseEntityTypes(o.Text, o.TextEntities, accepted)...)
		}
	}
	// We do not check m.Quote, as this is not technically the current message, but the reply
	return out
}

// ParseEntity parses a single message text entity to populate text contents, URL, and offsets in UTF8.
func (m Message) ParseEntity(entity MessageEntity) ParsedMessageEntity {
	return ParseEntity(m.Text, entity)
}

// ParseCaptionEntity parses a single message caption entity to populate text contents, URL, and offsets in UTF8.
func (m Message) ParseCaptionEntity(entity MessageEntity) ParsedMessageEntity {
	return ParseEntity(m.Caption, entity)
}

func parseEntity(entity MessageEntity, utf16Text []uint16) ParsedMessageEntity {
	text := string(utf16.Decode(utf16Text[entity.Offset : entity.Offset+entity.Length]))

	if entity.Type == "url" {
		entity.Url = text
	}

	entity.Offset = int64(len(string(utf16.Decode(utf16Text[:entity.Offset]))))
	entity.Length = int64(len(text))

	return ParsedMessageEntity{
		MessageEntity: entity,
		Text:          text,
	}
}
