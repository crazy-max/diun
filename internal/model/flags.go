package model

// Flags holds flags from command line
type Flags struct {
	Cfgfile  string
	Populate bool
	Timezone string
	LogLevel string
	LogJson  bool
	RunOnce  bool
	Docker   bool
}
