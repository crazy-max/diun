package premailer

// Options for controlling behaviour
type Options struct {
	// Remove class attribute from element
	// Default false
	RemoveClasses bool
	// Copy related CSS properties into HTML attributes (e.g. background-color to bgcolor)
	// Default true
	CssToAttributes bool
}

// NewOptions return an Options instance with default value
func NewOptions() *Options {
	options := &Options{}
	options.CssToAttributes = true
	return options
}
