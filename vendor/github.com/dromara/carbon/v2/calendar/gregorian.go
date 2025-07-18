// Package calendar is part of the carbon package.
package calendar

import (
	"time"
)

// Gregorian defines a Gregorian struct.
type Gregorian struct {
	Time  time.Time
	Error error
}

// String implements "Stringer" interface.
func (g *Gregorian) String() string {
	if g == nil {
		return ""
	}
	if g.Time.IsZero() {
		return ""
	}
	return g.Time.String()
}

func (g *Gregorian) IsLeapYear() bool {
	if g == nil || g.Error != nil {
		return false
	}
	year := g.Time.Year()
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return true
	}
	return false
}
