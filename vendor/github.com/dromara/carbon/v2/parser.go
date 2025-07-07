package carbon

import (
	"fmt"
	"time"
)

// Parse parses a time string as a Carbon instance by default layouts.
//
// Note: it doesn't support parsing timestamp string.
func Parse(value string, timezone ...string) *Carbon {
	if value == "" {
		return &Carbon{isEmpty: true}
	}
	var (
		tz  string
		tt  StdTime
		loc *Location
		err error
	)
	if len(timezone) > 0 {
		tz = timezone[0]
	} else {
		tz = DefaultTimezone
	}
	if loc, err = parseTimezone(tz); err != nil {
		return &Carbon{Error: err}
	}
	switch value {
	case "now":
		return Now().SetLocation(loc)
	case "yesterday":
		return Yesterday().SetLocation(loc)
	case "tomorrow":
		return Tomorrow().SetLocation(loc)
	}
	c := NewCarbon().SetLocation(loc)
	for _, layout := range defaultLayouts {
		if tt, err = time.ParseInLocation(layout, value, loc); err == nil {
			c.time = tt
			c.currentLayout = layout
			return c
		}
	}
	c.Error = ErrFailedParse(value)
	return c
}

// ParseByLayout parses a time string as a Carbon instance by a confirmed layout.
//
// Note: it doesn't support parsing timestamp string.
func ParseByLayout(value, layout string, timezone ...string) *Carbon {
	if value == "" {
		return &Carbon{isEmpty: true}
	}
	if layout == "" {
		return &Carbon{Error: ErrEmptyLayout()}
	}
	var (
		tz  string
		tt  StdTime
		loc *Location
		err error
	)
	if len(timezone) > 0 {
		tz = timezone[0]
	} else {
		tz = DefaultTimezone
	}
	if loc, err = parseTimezone(tz); err != nil {
		return &Carbon{Error: err}
	}

	if tt, err = time.ParseInLocation(layout, value, loc); err != nil {
		return &Carbon{Error: fmt.Errorf("%w: %w", ErrMismatchedLayout(value, layout), err)}
	}

	c := NewCarbon()
	c.loc = loc
	c.time = tt
	c.currentLayout = layout
	return c
}

// ParseByFormat parses a time string as a Carbon instance by a confirmed format.
//
// Note: If the letter used conflicts with the format sign, please use the escape character "\" to escape the letter
func ParseByFormat(value, format string, timezone ...string) *Carbon {
	if value == "" {
		return &Carbon{isEmpty: true}
	}
	if format == "" {
		return &Carbon{Error: ErrEmptyFormat()}
	}
	c := ParseByLayout(value, format2layout(format), timezone...)
	if c.HasError() {
		c.Error = fmt.Errorf("%w: %w", ErrMismatchedFormat(value, format), c.Error)
	}
	return c
}

// ParseByLayouts parses a time string as a Carbon instance by multiple fuzzy layouts.
//
// Note: it doesn't support parsing timestamp string.
func ParseByLayouts(value string, layouts []string, timezone ...string) *Carbon {
	if value == "" {
		return &Carbon{isEmpty: true}
	}
	if len(layouts) == 0 {
		return &Carbon{Error: ErrEmptyLayout()}
	}
	var (
		tz  string
		tt  StdTime
		loc *Location
		err error
	)
	if len(timezone) > 0 {
		tz = timezone[0]
	} else {
		tz = DefaultTimezone
	}
	if loc, err = parseTimezone(tz); err != nil {
		return &Carbon{Error: err}
	}
	c := NewCarbon().SetLocation(loc)
	for _, layout := range layouts {
		if tt, err = time.ParseInLocation(layout, value, loc); err == nil {
			c.time = tt
			c.currentLayout = layout
			return c
		}
	}
	c.Error = ErrFailedParse(value)
	return c
}

// ParseByFormats parses a time string as a Carbon instance by multiple fuzzy formats.
//
// Note: it doesn't support parsing timestamp string.
func ParseByFormats(value string, formats []string, timezone ...string) *Carbon {
	if value == "" {
		return &Carbon{isEmpty: true}
	}
	if len(formats) == 0 {
		return &Carbon{Error: ErrEmptyFormat()}
	}
	var (
		tz  string
		err error
	)
	if len(timezone) > 0 {
		tz = timezone[0]
	} else {
		tz = DefaultTimezone
	}
	if _, err = parseTimezone(tz); err != nil {
		return &Carbon{Error: err}
	}
	var layouts []string
	for _, v := range formats {
		layouts = append(layouts, format2layout(v))
	}
	return ParseByLayouts(value, layouts, tz)
}
