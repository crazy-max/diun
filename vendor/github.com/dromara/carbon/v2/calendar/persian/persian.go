// Package persian is part of the carbon package.
package persian

import (
	"fmt"
	"math"
	"time"

	"github.com/dromara/carbon/v2/calendar"
	"github.com/dromara/carbon/v2/calendar/julian"
)

var (
	EnMonths      = []string{"Farvardin", "Ordibehesht", "Khordad", "Tir", "Mordad", "Shahrivar", "Mehr", "Aban", "Azar", "Dey", "Bahman", "Esfand"}
	ShortEnMonths = []string{"Far", "Ord", "Kho", "Tir", "Mor", "Sha", "Meh", "Aba", "Aza", "Dey", "Bah", "Esf"}

	FaMonths      = []string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}
	ShortFaMonths = []string{"فرو", "ارد", "خرد", "تیر", "مرد", "شهر", "مهر", "آبا", "آذر", "دی", "بهم", "اسف"}

	EnWeeks      = []string{"Yekshanbeh", "Doshanbeh", "Seshanbeh", "Chaharshanbeh", "Panjshanbeh", "Jomeh", "Shanbeh"}
	ShortEnWeeks = []string{"Yek", "Dos", "Ses", "Cha", "Pan", "Jom", "Sha"}

	FaWeeks      = []string{"نجشنبه", "دوشنبه", "سه شنبه", "چهارشنبه", "پنجشنبه", "جمعه", "شنبه"}
	ShortFaWeeks = []string{"پ", "چ", "س", "د", "ی", "ش", "ج"}

	InvalidDateError = func() error {
		return fmt.Errorf("invalid persian date, please make sure the date is valid")
	}
)

// Gregorian defines a Gregorian struct.
// 定义 Gregorian 结构体
type Gregorian struct {
	calendar.Gregorian
}

// Persian defines a Persian struct.
// 定义 Persian 结构体
type Persian struct {
	year, month, day, hour, minute, second int
	Error                                  error
}

// MaxValue returns a Persian instance for the greatest supported date.
// 返回 Persian 的最大值
func MaxValue() Persian {
	return Persian{
		year:   9377,
		month:  12,
		day:    31,
		hour:   23,
		minute: 59,
		second: 59,
	}
}

// MinValue returns a Persian instance for the lowest supported date.
// 返回 Persian 的最小值
func MinValue() Persian {
	return Persian{
		year:   1,
		month:  1,
		day:    1,
		hour:   0,
		minute: 0,
		second: 0,
	}
}

// FromGregorian creates a Gregorian instance from time.Time.
// 从标准 time.Time 创建 Gregorian 实例
func FromGregorian(t time.Time) (g Gregorian) {
	if t.IsZero() {
		return
	}
	g.Time = t
	return
}

// FromPersian creates a Persian instance from persian datetime.
// 从 波斯日期 创建 Persian 实例
func FromPersian(year, month, day, hour, minute, second int) (p Persian) {
	p.year, p.month, p.day = year, month, day
	p.hour, p.minute, p.second = hour, minute, second
	if !p.IsValid() {
		p.Error = InvalidDateError()
		return
	}
	return
}

// ToPersian converts Gregorian instance to Persian instance.
// 将 Gregorian 实例转化为 Persian 实例
func (g Gregorian) ToPersian() (p Persian) {
	p.hour, p.minute, p.second = g.Hour(), g.Minute(), g.Second()
	gjdn := getGregorianJdn(g.Year(), g.Month(), g.Day())
	pjdn := getPersianJdn(475, 1, 1)

	diff := gjdn - pjdn
	div := diff / 1029983
	mod := diff % 1029983
	p.year = (2134*mod/366+2816*(mod%366)+2815)/1028522 + mod/366 + 1 + 2820*div + 474
	pjdn = getPersianJdn(p.year, 1, 1)
	fjdn := float64(gjdn - pjdn + 1)
	if fjdn <= 186 {
		p.month = int(math.Ceil(fjdn / 31.0))
	} else {
		p.month = int(math.Ceil((fjdn - 6) / 30.0))
	}
	pjdn = getPersianJdn(p.year, p.month, 1)
	p.day = gjdn - pjdn + 1
	if !p.IsValid() {
		p.Error = InvalidDateError()
		return
	}
	return
}

// ToGregorian converts Persian instance to Gregorian instance.
// 将 Persian 实例转化为 Gregorian 实例
func (p Persian) ToGregorian() (g Gregorian) {
	if !p.IsValid() {
		return
	}
	jdn := getPersianJdn(p.year, p.month, p.day)

	l := jdn + 68569
	n := 4 * l / 146097
	l = l - (146097*n+3)/4
	i := 4000 * (l + 1) / 1461001
	l = l - 1461*i/4 + 31
	j := 80 * l / 2447
	d := l - 2447*j/80
	l = j / 11
	m := j + 2 - 12*l
	y := 100*(n-49) + i + l

	g.Time = time.Date(y, time.Month(m), d, p.hour, p.minute, p.second, 0, time.Local)
	return
}

// Year gets lunar year like 2020.
// 获取年份，如 2020
func (p Persian) Year() int {
	if p.Error != nil {
		return 0
	}
	return p.year
}

// Month gets lunar month like 8.
// 获取月份，如 8
func (p Persian) Month() int {
	if p.Error != nil {
		return 0
	}
	return p.month
}

// Day gets lunar day like 5.
// 获取日，如 5
func (p Persian) Day() int {
	if p.Error != nil {
		return 0
	}
	return p.day
}

// Hour gets current hour like 13.
// 获取小时，如 13
func (p Persian) Hour() int {
	if p.Error != nil {
		return 0
	}
	return p.hour
}

// Minute gets current minute like 14.
// 获取分钟数，如 14
func (p Persian) Minute() int {
	if p.Error != nil {
		return 0
	}
	return p.minute
}

// Second gets current second like 15.
// 获取秒数，如 15
func (p Persian) Second() int {
	if p.Error != nil {
		return 0
	}
	return p.second
}

// String implements Stringer interface and outputs a string in YYYY-MM-DD HH::ii::ss format like "1402-11-11 00:00:00".
// 实现 Stringer 接口, 输出 YYYY-MM-DD HH::ii::ss 格式字符串，如 "1402-11-11 00:00:00"
func (p Persian) String() string {
	if !p.IsValid() {
		return ""
	}
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", p.year, p.month, p.day, p.hour, p.minute, p.second)
}

// ToMonthString outputs a string in persian month format like "فروردین".
// 获取完整月份字符串，如 "فروردین"
func (p Persian) ToMonthString(locale ...string) (month string) {
	if !p.IsValid() {
		return ""
	}
	loc := "en"
	if len(locale) > 0 {
		loc = locale[0]
	}
	switch loc {
	case "en":
		return EnMonths[p.month-1]
	case "fa":
		return FaMonths[p.month-1]
	}
	return ""
}

// ToShortMonthString outputs a short string in persian month format like "فروردین".
// 获取缩写月份字符串，如 "فروردین"
func (p Persian) ToShortMonthString(locale ...string) (month string) {
	if !p.IsValid() {
		return ""
	}
	loc := "en"
	if len(locale) > 0 {
		loc = locale[0]
	}
	switch loc {
	case "en":
		return ShortEnMonths[p.month-1]
	case "fa":
		return ShortFaMonths[p.month-1]
	}
	return ""
}

// ToWeekString outputs a string in week layout like "چهارشنبه".
// 输出完整星期字符串，如 "چهارشنبه"
func (p Persian) ToWeekString(locale ...string) (month string) {
	if !p.IsValid() {
		return ""
	}
	loc := "en"
	if len(locale) > 0 {
		loc = locale[0]
	}
	switch loc {
	case "en":
		return EnWeeks[p.ToGregorian().Week()]
	case "fa":
		return FaWeeks[p.ToGregorian().Week()]
	}
	return ""
}

// ToShortWeekString outputs a short string in week layout like "چهارشنبه".
// 输出缩写星期字符串，如 "چهارشنبه"
func (p Persian) ToShortWeekString(locale ...string) (month string) {
	if !p.IsValid() {
		return ""
	}
	loc := "en"
	if len(locale) > 0 {
		loc = locale[0]
	}
	switch loc {
	case "en":
		return ShortEnWeeks[p.ToGregorian().Week()]
	case "fa":
		return ShortFaWeeks[p.ToGregorian().Week()]
	}
	return ""
}

// IsValid reports whether is a valid persian date.
// 是否是有效的日期
func (p Persian) IsValid() bool {
	if p.Year() >= MinValue().year && p.Year() <= MaxValue().year && p.month >= MinValue().month && p.month <= MaxValue().month && p.day >= MinValue().day && p.day <= MaxValue().day {
		return true
	}
	return false
}

// IsLeapYear reports whether is a persian leap year.
// 是否是闰年
func (p Persian) IsLeapYear() bool {
	if !p.IsValid() {
		return false
	}
	return (25*p.year+11)%33 < 8
}

// gets Julian day number in Persian calendar
// 获取波斯历儒略日计数
func getPersianJdn(year, month, day int) int {
	year = year - 473
	if year >= 0 {
		year--
	}
	epy := 474 + (year % 2820)
	var md int
	if month <= 7 {
		md = (month - 1) * 31
	} else {
		md = (month-1)*30 + 6
	}
	return day + md + (epy*682-110)/2816 + (epy-1)*365 + year/2820*1029983 + 1948320
}

// gets Julian day number in Gregorian calendar
// 获取公历儒略日计数
func getGregorianJdn(year, month, day int) int {
	jdn := julian.FromGregorian(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)).ToJulian().JD(0)
	return int(jdn)
}
