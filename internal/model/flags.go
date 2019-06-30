package model

// Flags holds flags from command line
type Flags struct {
	Cfgfile   string
	Timezone  string
	LogLevel  string
	LogJson   bool
	LogCaller bool
	Docker    bool
}
