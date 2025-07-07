// @Package carbon
// @Description a simple, semantic and developer-friendly time package for golang
// @Page github.com/dromara/carbon
// @Developer gouguoyin
// @Blog www.gouguoyin.com
// @Email 245629560@qq.com

// Package carbon is a simple, semantic and developer-friendly time package for golang.
package carbon

import (
	"time"
)

type StdTime = time.Time
type Weekday = time.Weekday
type Location = time.Location
type Duration = time.Duration

// Carbon defines a Carbon struct.
type Carbon struct {
	time          StdTime
	weekStartsAt  Weekday
	weekendDays   []Weekday
	loc           *Location
	lang          *Language
	currentLayout string
	isEmpty       bool
	Error         error
}

// NewCarbon returns a new Carbon instance.
func NewCarbon(stdTime ...StdTime) *Carbon {
	c := new(Carbon)
	c.lang = NewLanguage().SetLocale(DefaultLocale)
	c.weekStartsAt = DefaultWeekStartsAt
	c.weekendDays = DefaultWeekendDays
	c.currentLayout = DefaultLayout
	if len(stdTime) > 0 {
		c.time = stdTime[0]
		c.loc = c.time.Location()
		return c
	}
	c.loc, c.Error = parseTimezone(DefaultTimezone)
	return c
}

// Copy returns a copy of the Carbon instance.
func (c *Carbon) Copy() *Carbon {
	if c.IsNil() {
		return nil
	}
	return &Carbon{
		time:          time.Date(c.Year(), time.Month(c.Month()), c.Day(), c.Hour(), c.Minute(), c.Second(), c.Nanosecond(), c.loc),
		weekStartsAt:  c.weekStartsAt,
		weekendDays:   c.weekendDays,
		loc:           c.loc,
		lang:          c.lang.Copy(),
		currentLayout: c.currentLayout,
		isEmpty:       c.isEmpty,
		Error:         c.Error,
	}
}

// Sleep sleeps for the specified duration like time.Sleep.
func (c *Carbon) Sleep(d time.Duration) {
	if IsTestNow() && d > 0 {
		frozenNow.rw.Lock()
		frozenNow.testNow = frozenNow.testNow.AddDuration(d.String())
		frozenNow.rw.Unlock()
		return
	}
	time.Sleep(d)
}
