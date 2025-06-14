package carbon

var (
	// DefaultLayout default layout
	DefaultLayout = DateTimeLayout

	// DefaultTimezone default timezone
	DefaultTimezone = UTC

	// DefaultLocale default language locale
	DefaultLocale = "en"

	// DefaultWeekStartsAt default start date of the week
	DefaultWeekStartsAt = Monday

	// DefaultWeekendDays default weekend days of the week
	DefaultWeekendDays = []Weekday{
		Saturday, Sunday,
	}
)

type Default struct {
	Layout       string
	Timezone     string
	Locale       string
	WeekStartsAt Weekday
	WeekendDays  []Weekday
}

// SetDefault sets default.
func SetDefault(d Default) {
	if d.Layout != "" {
		DefaultLayout = d.Layout
	}
	if d.Timezone != "" {
		DefaultTimezone = d.Timezone
	}
	if d.Locale != "" {
		DefaultLocale = d.Locale
	}
	if d.WeekStartsAt.String() != "" {
		DefaultWeekStartsAt = d.WeekStartsAt
	}
	if len(d.WeekendDays) > 0 {
		DefaultWeekendDays = d.WeekendDays
	}
}

// ResetDefault resets default.
func ResetDefault() {
	DefaultLayout = DateTimeLayout
	DefaultTimezone = UTC
	DefaultLocale = "en"
	DefaultWeekStartsAt = Monday
	DefaultWeekendDays = []Weekday{
		Saturday, Sunday,
	}
}
