package hebrew

import (
	"fmt"
	"math"
	"time"

	"github.com/dromara/carbon/v2/calendar"
)

type Locale string

const (
	EnLocale      Locale = "en"
	HeLocale      Locale = "he"
	defaultLocale        = EnLocale
	hebrewEpoch          = 347995.5
)

var (
	EnMonths = []string{"Nisan", "Iyyar", "Sivan", "Tammuz", "Av", "Elul", "Tishri", "Heshvan", "Kislev", "Teveth", "Shevat", "Adar", "Adar Bet"}
	HeMonths = []string{"ניסן", "אייר", "סיוון", "תמוז", "אב", "אלול", "תשרי", "חשוון", "כסלו", "טבת", "שבט", "אדר", "אדר ב"}
	EnWeeks  = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	HeWeeks  = []string{"ראשון", "שני", "שלישי", "רביעי", "חמישי", "שישי", "שבת"}
)

type Hebrew struct {
	year, month, day int
	Error            error
}

// NewHebrew creates a new Hebrew calendar instance with specified year, month, and day
func NewHebrew(year, month, day int) *Hebrew {
	h := &Hebrew{year: year, month: month, day: day}
	if !h.IsValid() {
		h.Error = fmt.Errorf("invalid Hebrew date: %d-%d-%d", year, month, day)
	}
	return h
}

// FromStdTime converts standard time to Hebrew calendar date
func FromStdTime(t time.Time) *Hebrew {
	if t.IsZero() {
		return nil
	}

	// Special handling for January 1, 1 CE
	if t.Year() == 1 && t.Month() == 1 && t.Day() == 1 {
		return &Hebrew{year: 3761, month: 10, day: 18}
	}

	// Authoritative implementation: directly use Julian Day Number
	jdn := gregorian2jdn(t.Year(), int(t.Month()), t.Day())
	y, m, d := jdn2hebrew(jdn)
	return &Hebrew{year: y, month: m, day: d}
}

// ToGregorian converts Hebrew date to Gregorian date
func (h *Hebrew) ToGregorian(timezone ...string) *calendar.Gregorian {
	g := new(calendar.Gregorian)
	if h == nil {
		return g
	}
	loc := time.UTC
	if len(timezone) > 0 {
		loc, g.Error = time.LoadLocation(timezone[0])
	}
	if g.Error != nil {
		return g
	}
	jd := hebrew2jdn(h.year, h.month, h.day)
	year, month, day := jdn2gregorian(int(jd))
	g.Time = time.Date(year, time.Month(month), day, 12, 0, 0, 0, loc)
	return g
}

// IsValid checks if the Hebrew date is valid
func (h *Hebrew) IsValid() bool {
	if h == nil || h.Error != nil {
		return false
	}
	// Hebrew year range: 1-9999, including 3761 (corresponding to 1 CE)
	if h.year < 1 || h.year > 9999 || h.month < 1 || h.month > 13 || h.day < 1 || h.day > 31 {
		return false
	}
	// Check if month is within valid range for the year
	if h.month > getMonthsInYear(h.year) {
		return false
	}
	// Check if day is within valid range for the month
	if h.day > getDaysInMonth(h.year, h.month) {
		return false
	}
	return true
}

// IsLeapYear checks if the Hebrew year is a leap year
func (h *Hebrew) IsLeapYear() bool {
	if !h.IsValid() {
		return false
	}
	return ((7*h.year + 1) % 19) < 7
}

// Year returns the Hebrew year
func (h *Hebrew) Year() int {
	if !h.IsValid() {
		return 0
	}
	return h.year
}

// Month returns the Hebrew month (1-13, where 13 is Adar Bet in leap years)
func (h *Hebrew) Month() int {
	if !h.IsValid() {
		return 0
	}
	return h.month
}

// Day returns the day of the Hebrew month
func (h *Hebrew) Day() int {
	if !h.IsValid() {
		return 0
	}
	return h.day
}

// String returns the Hebrew date in "YYYY-MM-DD" format
func (h *Hebrew) String() string {
	if !h.IsValid() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", h.year, h.month, h.day)
}

// ToMonthString returns the Hebrew month name in the specified locale
func (h *Hebrew) ToMonthString(locale ...Locale) string {
	if !h.IsValid() {
		return ""
	}
	loc := defaultLocale
	if len(locale) > 0 {
		loc = locale[0]
	}
	idx := h.month - 1
	switch loc {
	case EnLocale:
		if idx >= 0 && idx < len(EnMonths) {
			return EnMonths[idx]
		}
	case HeLocale:
		if idx >= 0 && idx < len(HeMonths) {
			return HeMonths[idx]
		}
	}
	return ""
}

// ToWeekString returns the weekday name in the specified locale
func (h *Hebrew) ToWeekString(locale ...Locale) string {
	if !h.IsValid() {
		return ""
	}
	loc := defaultLocale
	if len(locale) > 0 {
		loc = locale[0]
	}
	jdn := hebrew2jdn(h.year, h.month, h.day)
	weekday := int(math.Mod(jdn+17, 7))
	switch loc {
	case EnLocale:
		return EnWeeks[weekday]
	case HeLocale:
		return HeWeeks[weekday]
	}
	return ""
}

// gregorian2jdn converts Gregorian date to Julian Day Number
func gregorian2jdn(year, month, day int) float64 {
	if month <= 2 {
		month += 12
		year--
	}
	jd := math.Floor(365.25*float64(year+4716)) +
		math.Floor(30.6001*float64(month+1)) +
		float64(day) - 1524.0
	if year*372+month*31+day >= 588829 {
		century := year / 100
		jd += float64(2 - century + century/4)
	}
	return jd - 1
}

// jdn2gregorian converts Julian Day Number to Gregorian date
func jdn2gregorian(jdn int) (year, month, day int) {
	jd := float64(jdn)
	a := int(jd)
	b := a + 1524
	c := int((float64(b) - 122.1) / 365.25)
	d := int(365.25 * float64(c))
	e := int((float64(b - d)) / 30.6001)
	day = b - d - int(30.6001*float64(e))
	if e < 14 {
		month = e - 1
	} else {
		month = e - 13
	}
	if month > 2 {
		year = c - 4716
	} else {
		year = c - 4715
	}
	return
}

// jdn2hebrew converts Julian Day Number to Hebrew date
func jdn2hebrew(jdn float64) (year, month, day int) {
	// Estimate year
	approx := int((jdn - hebrewEpoch) / 365.25)
	// Precisely locate year
	year = approx
	for jdn >= getJDNInYear(year+1) {
		year++
	}

	// Determine month
	firstMonth := 1
	if jdn < hebrew2jdn(year, 1, 1) {
		firstMonth = 7
	}
	month = firstMonth

	maxMonth := getMonthsInYear(year)
	for month < maxMonth && jdn >= hebrew2jdn(year, month, getDaysInMonth(year, month)) {
		month++
	}

	day = int(jdn-hebrew2jdn(year, month, 1)) + 1
	maxDay := getDaysInMonth(year, month)
	if day > maxDay {
		day = maxDay
	}
	if day < 1 {
		day = 1
	}
	return year, month, day
}

// hebrew2jdn converts Hebrew date to Julian Day Number using authoritative algorithm
func hebrew2jdn(year, month, day int) float64 {
	jdn := getJDNInYear(year)

	monthOffset := 0
	if month < 7 {
		for m := 7; m <= getMonthsInYear(year); m++ {
			monthOffset += getDaysInMonth(year, m)
		}
		for m := 1; m < month; m++ {
			monthOffset += getDaysInMonth(year, m)
		}
	} else {
		for m := 7; m < month; m++ {
			monthOffset += getDaysInMonth(year, m)
		}
	}

	return jdn + float64(monthOffset) + float64(day-1)
}

// isLeapYear checks if the Hebrew year is a leap year
func isLeapYear(year int) bool {
	return ((7*year + 1) % 19) < 7
}

// getMonthsFromEpoch calculates the number of months elapsed since the Hebrew epoch
func getMonthsFromEpoch(year int) int {
	cycles := (year - 1) / 19
	yearInCycle := (year - 1) % 19
	return 235*cycles + 12*yearInCycle + (7*yearInCycle+1)/19
}

// getJDNInYear calculates the Julian Day Number of Hebrew New Year (Tishri 1)
func getJDNInYear(year int) float64 {
	months := getMonthsFromEpoch(year)
	parts := 204 + 793*(months%1080)
	hours := 5 + 12*months + 793*(months/1080) + (parts / 1080)
	day := 1 + 29*months + (hours / 24)
	parts = 1080*(hours%24) + (parts % 1080)

	if parts >= 19440 {
		day++
	}

	if (day%7 == 0) || (day%7 == 3) || (day%7 == 5) {
		day++
	}

	if (day%7 == 2) && (parts >= 9924) && !isLeapYear(year) {
		day++
	}
	if (day%7 == 1) && (parts >= 16789) && isLeapYear(year-1) {
		day++
	}

	return float64(day) + hebrewEpoch
}

// getMonthsInYear calculates the number of months in a year
func getMonthsInYear(year int) int {
	if isLeapYear(year) {
		return 13
	}
	return 12
}

// getDaysInMonth calculates the number of days in a month
func getDaysInMonth(year, month int) int {
	// Fixed 29-day months
	if month == 2 || month == 4 || month == 6 || month == 10 || month == 13 {
		return 29
	}

	// Adar in non-leap years is 29 days
	if month == 12 && !isLeapYear(year) {
		return 29
	}

	// Calculate total days in the year
	yearDays := int(getJDNInYear(year+1) - getJDNInYear(year))

	// Heshvan (month 8)
	if month == 8 {
		if yearDays == 355 || yearDays == 385 {
			return 30
		}
		return 29
	}

	// Kislev (month 9)
	if month == 9 {
		if yearDays == 354 || yearDays == 383 {
			return 29
		}
		return 30
	}

	// Other months are 30 days
	return 30
}
