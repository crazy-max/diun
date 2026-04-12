package types

type CallbackState int

const (
	CallbackUndef CallbackState = iota
	CallbackSkipped
	CallbackStarted
	CallbackActive
	CallbackFinished
	CallbackArchived
)

type CallbackKind int

const (
	CallbackManifest CallbackKind = iota
	CallbackBlob
)

func (k CallbackKind) String() string {
	switch k {
	case CallbackBlob:
		return "blob"
	case CallbackManifest:
		return "manifest"
	}
	return "unknown"
}
