// Package calendar is part of the carbon package.
package calendar

import (
	"fmt"
	"time"
)

var InvalidDateError = func() error {
	return fmt.Errorf("invalid gregorian date, please make sure the date is valid")
}

// month constants
// 月份常量
const (
	January   = "January"   // 一月
	February  = "February"  // 二月
	March     = "March"     // 三月
	April     = "April"     // 四月
	May       = "May"       // 五月
	June      = "June"      // 六月
	July      = "July"      // 七月
	August    = "August"    // 八月
	September = "September" // 九月
	October   = "October"   // 十月
	November  = "November"  // 十一月
	December  = "December"  // 十二月
)

// week constants
// 星期常量
const (
	Monday    = "Monday"    // 周一
	Tuesday   = "Tuesday"   // 周二
	Wednesday = "Wednesday" // 周三
	Thursday  = "Thursday"  // 周四
	Friday    = "Friday"    // 周五
	Saturday  = "Saturday"  // 周六
	Sunday    = "Sunday"    // 周日
)

// number constants
// 数字常量
const (
	YearsPerMillennium = 1000   // 每千年1000年
	YearsPerCentury    = 100    // 每世纪100年
	YearsPerDecade     = 10     // 每十年10年
	QuartersPerYear    = 4      // 每年4个季度
	MonthsPerYear      = 12     // 每年12月
	MonthsPerQuarter   = 3      // 每季度3月
	WeeksPerNormalYear = 52     // 每常规年52周
	weeksPerLongYear   = 53     // 每长年53周
	WeeksPerMonth      = 4      // 每月4周
	DaysPerLeapYear    = 366    // 每闰年366天
	DaysPerNormalYear  = 365    // 每常规年365天
	DaysPerWeek        = 7      // 每周7天
	HoursPerWeek       = 168    // 每周168小时
	HoursPerDay        = 24     // 每天24小时
	MinutesPerDay      = 1440   // 每天1440分钟
	MinutesPerHour     = 60     // 每小时60分钟
	SecondsPerWeek     = 604800 // 每周604800秒
	SecondsPerDay      = 86400  // 每天86400秒
	SecondsPerHour     = 3600   // 每小时3600秒
	SecondsPerMinute   = 60     // 每分钟60秒
)

// layout constants
// 布局模板常量
const (
	AtomLayout     = RFC3339Layout
	ANSICLayout    = time.ANSIC
	CookieLayout   = "Monday, 02-Jan-2006 15:04:05 MST"
	KitchenLayout  = time.Kitchen
	RssLayout      = time.RFC1123Z
	RubyDateLayout = time.RubyDate
	UnixDateLayout = time.UnixDate
	W3cLayout      = RFC3339Layout

	RFC1036Layout      = "Mon, 02 Jan 06 15:04:05 -0700"
	RFC1123Layout      = time.RFC1123
	RFC1123ZLayout     = time.RFC1123Z
	RFC2822Layout      = time.RFC1123Z
	RFC3339Layout      = "2006-01-02T15:04:05Z07:00"
	RFC3339MilliLayout = "2006-01-02T15:04:05.999Z07:00"
	RFC3339MicroLayout = "2006-01-02T15:04:05.999999Z07:00"
	RFC3339NanoLayout  = "2006-01-02T15:04:05.999999999Z07:00"
	RFC7231Layout      = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC822Layout       = time.RFC822
	RFC822ZLayout      = time.RFC822Z
	RFC850Layout       = time.RFC850

	ISO8601Layout      = "2006-01-02T15:04:05-07:00"
	ISO8601MilliLayout = "2006-01-02T15:04:05.999-07:00"
	ISO8601MicroLayout = "2006-01-02T15:04:05.999999-07:00"
	ISO8601NanoLayout  = "2006-01-02T15:04:05.999999999-07:00"

	DayDateTimeLayout        = "Mon, Jan 2, 2006 3:04 PM"
	DateTimeLayout           = "2006-01-02 15:04:05"
	DateTimeMilliLayout      = "2006-01-02 15:04:05.999"
	DateTimeMicroLayout      = "2006-01-02 15:04:05.999999"
	DateTimeNanoLayout       = "2006-01-02 15:04:05.999999999"
	ShortDateTimeLayout      = "20060102150405"
	ShortDateTimeMilliLayout = "20060102150405.999"
	ShortDateTimeMicroLayout = "20060102150405.999999"
	ShortDateTimeNanoLayout  = "20060102150405.999999999"

	DateLayout           = "2006-01-02"
	DateMilliLayout      = "2006-01-02.999"
	DateMicroLayout      = "2006-01-02.999999"
	DateNanoLayout       = "2006-01-02.999999999"
	ShortDateLayout      = "20060102"
	ShortDateMilliLayout = "20060102.999"
	ShortDateMicroLayout = "20060102.999999"
	ShortDateNanoLayout  = "20060102.999999999"

	TimeLayout           = "15:04:05"
	TimeMilliLayout      = "15:04:05.999"
	TimeMicroLayout      = "15:04:05.999999"
	TimeNanoLayout       = "15:04:05.999999999"
	ShortTimeLayout      = "150405"
	ShortTimeMilliLayout = "150405.999"
	ShortTimeMicroLayout = "150405.999999"
	ShortTimeNanoLayout  = "150405.999999999"
)

// Gregorian defines a Gregorian struct.
// 定义 Gregorian 结构体
type Gregorian struct {
	Time  time.Time
	Error error
}

// NewGregorian returns a new Gregorian instance.
// 初始化 Gregorian 结构体
func NewGregorian(t time.Time) (g Gregorian) {
	if t.IsZero() {
		return
	}
	g.Time = t
	return
}

// MaxValue returns a Gregorian instance for the greatest supported date.
// 返回 Gregorian 的最大值
func MaxValue() Gregorian {
	return NewGregorian(time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC))
}

// MinValue returns a Gregorian instance for the lowest supported date.
// 返回 Gregorian 的最小值
func MinValue() Gregorian {
	return NewGregorian(time.Date(-9998, 1, 1, 0, 0, 0, 0, time.UTC))
}

// Date gets gregorian year, month, and day like 2020, 8, 5.
// 获取公历年、月、日
func (g Gregorian) Date() (year, month, day int) {
	if g.IsZero() {
		return 0, 0, 0
	}
	var tm time.Month
	year, tm, day = g.Time.Date()
	month = int(tm)
	return
}

// Clock gets gregorian hour, minute, and second like 13, 14, 15.
// 获取公历时、分、秒
func (g Gregorian) Clock() (hour, minute, second int) {
	if g.IsZero() {
		return 0, 0, 0
	}
	return g.Time.Clock()
}

// Year gets gregorian year like 2020.
// 获取公历年
func (g Gregorian) Year() int {
	if g.IsZero() {
		return 0
	}
	return g.Time.Year()
}

// Month gets gregorian month like 8.
// 获取公历月，如 8
func (g Gregorian) Month() int {
	if g.IsZero() {
		return 0
	}
	return int(g.Time.Month())
}

// Week gets gregorian week day like 0.
// 获取周
func (g Gregorian) Week() int {
	if g.IsZero() {
		return 0
	}
	return int(g.Time.Weekday())
}

// Day gets gregorian day like 5.
// 获取公历日，如 0
func (g Gregorian) Day() int {
	if g.IsZero() {
		return 0
	}
	return g.Time.Day()
}

// Hour gets gregorian hour like 13.
// 获取公历小时，如 13
func (g Gregorian) Hour() int {
	if g.IsZero() {
		return 0
	}
	return g.Time.Hour()
}

// Minute gets gregorian minute like 14.
// 获取公历分钟数，如 14
func (g Gregorian) Minute() int {
	if g.IsZero() {
		return 0
	}
	return g.Time.Minute()
}

// Second gets gregorian second like 15.
// 获取公历秒数，如 15
func (g Gregorian) Second() int {
	if g.IsZero() {
		return 0
	}
	return g.Time.Second()
}

// Location gets gregorian timezone information.
// 获取时区信息
func (g Gregorian) Location() *time.Location {
	return g.Time.Location()
}

// String implements Stringer interface and outputs a string in YYYY-MM-DD HH::ii::ss format like "2019-12-07 00:00:00".
// 实现 Stringer 接口, 输出 YYYY-MM-DD HH::ii::ss 格式字符串，如 "2019-12-07 00:00:00"
func (g Gregorian) String() string {
	if g.IsZero() {
		return ""
	}
	return g.Time.Format(DateTimeLayout)
}

// IsZero reports whether is zero time.
// 是否是零值时间
func (g Gregorian) IsZero() bool {
	return g.Time.IsZero()
}

// IsValid reports whether is a valid gregorian date.
// 是否是有效的年份
func (g Gregorian) IsValid() bool {
	if g.Year() >= MinValue().Year() && g.Year() <= MaxValue().Year() && g.Month() >= MinValue().Month() && g.Month() <= MaxValue().Month() && g.Day() >= MinValue().Day() && g.Day() <= MaxValue().Day() {
		return true
	}
	return false
}

// IsLeapYear reports whether is a leap year.
// 是否是闰年
func (g Gregorian) IsLeapYear() bool {
	if g.IsZero() {
		return false
	}
	year := g.Year()
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return true
	}
	return false
}
