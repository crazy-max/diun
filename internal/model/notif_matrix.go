package model

// NotifMatrix holds Matrix notification configuration details
type NotifMatrix struct {
	HomeserverURL string             `yaml:"homeserverURL,omitempty" json:"homeserverURL,omitempty" validate:"required"`
	User          string             `yaml:"user,omitempty" json:"user,omitempty" validate:"omitempty"`
	UserFile      string             `yaml:"userFile,omitempty" json:"userFile,omitempty" validate:"omitempty,file"`
	Password      string             `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile  string             `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	RoomID        string             `yaml:"roomID,omitempty" json:"roomID,omitempty" validate:"required"`
	MsgType       NotifMatrixMsgType `yaml:"msgType,omitempty" json:"msgType,omitempty" validate:"required,oneof=notice text"`
	TemplateBody  string             `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// NotifMatrix message type constants
const (
	NotifMatrixMsgTypeNotice = NotifMatrixMsgType("notice")
	NotifMatrixMsgTypeText   = NotifMatrixMsgType("text")
)

// NotifMatrixMsgType holds message type
type NotifMatrixMsgType string

// GetDefaults gets the default values
func (s *NotifMatrix) GetDefaults() *NotifMatrix {
	n := &NotifMatrix{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifMatrix) SetDefaults() {
	s.HomeserverURL = "https://matrix.org"
	s.MsgType = NotifMatrixMsgTypeNotice
	s.TemplateBody = NotifDefaultTemplateBody
}
