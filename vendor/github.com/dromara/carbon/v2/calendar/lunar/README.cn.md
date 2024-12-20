# 中国农历

简体中文 | [English](README.md) | [日本語](README.jp.md)

#### 用法示例

> 目前仅支持公元 `1900` 年至 `2100` 年的 `200` 年时间跨度

##### 将 `公历` 转换成 `农历`

```go
// 获取农历生肖
carbon.Parse("2020-08-05 13:14:15").Lunar().Animal() // 鼠
// 获取农历节日
carbon.Parse("2021-02-12 13:14:15").Lunar().Festival() // 春节

// 获取农历年份
carbon.Parse("2020-08-05 13:14:15").Lunar().Year() // 2020
// 获取农历月份
carbon.Parse("2020-08-05 13:14:15").Lunar().Month() // 6
// 获取农历闰月月份
carbon.Parse("2020-08-05 13:14:15").Lunar().LeapMonth() // 4
// 获取农历日期
carbon.Parse("2020-08-05 13:14:15").Lunar().Day() // 16
// 获取农历时辰
carbon.Parse("2020-08-05 13:14:15").Lunar().Hour() // 13
// 获取农历分钟
carbon.Parse("2020-08-05 13:14:15").Lunar().Minute() // 14
// 获取农历秒数
carbon.Parse("2020-08-05 13:14:15").Lunar().Second() // 15

// 获取农历日期时间字符串
carbon.Parse("2020-08-05 13:14:15").Lunar().String() // 2020-06-16 13:14:15
fmt.Printf("%s", carbon.Parse("2020-08-05 13:14:15").Lunar()) // 2020-06-16 13:14:15
// 获取农历年字符串
carbon.Parse("2020-08-05 13:14:15").Lunar().ToYearString() // 二零二零
// 获取农历月字符串
carbon.Parse("2020-08-05 13:14:15").Lunar().ToMonthString() // 六月
// 获取农历闰月字符串
carbon.Parse("2020-04-23 13:14:15").Lunar().ToMonthString() // 闰四月
// 获取农历周字符串
carbon.Parse("2020-04-23 13:14:15").Lunar().ToWeekString() // 周四
// 获取农历天字符串
carbon.Parse("2020-08-05 13:14:15").Lunar().ToDayString() // 十六
// 获取农历日期字符串
carbon.Parse("2020-08-05 13:14:15").Lunar().ToDateString() // 二零二零年六月十六

```

##### 将 `农历` 转化成 `公历`

```go
// 将农历 二零二三年腊月十一 转化为 公历
carbon.CreateFromLunar(2023, 12, 11, 0, 0, 0, false).ToDateTimeString() // 2024-01-21 00:00:00
// 将农历 二零二三年二月十一 转化为 公历
carbon.CreateFromLunar(2023, 2, 11, 0, 0, 0, false).ToDateTimeString() // 2023-03-02 00:00:00
// 将农历 二零二三年闰二月十一 转化为 公历
carbon.CreateFromLunar(2023, 2, 11, 0, 0, 0, true).ToDateTimeString() // 2023-04-01 00:00:00
```

##### 日期判断

```go

// 是否是合法农历日期
carbon.Parse("0000-00-00 00:00:00").Lunar().IsValid() // false
carbon.Parse("2020-08-05 13:14:15").Lunar().IsValid() // true

// 是否是农历闰年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsLeapYear() // true
// 是否是农历闰月
carbon.Parse("2020-08-05 13:14:15").Lunar().IsLeapMonth() // false

// 是否是鼠年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRatYear() // true
// 是否是牛年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsOxYear() // false
// 是否是虎年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsTigerYear() // false
// 是否是兔年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRabbitYear() // false
// 是否是龙年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsDragonYear() // false
// 是否是蛇年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsSnakeYear() // false
// 是否是马年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsHorseYear() // false
// 是否是羊年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsGoatYear() // false
// 是否是猴年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsMonkeyYear() // false
// 是否是鸡年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRoosterYear() // false
// 是否是狗年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsDogYear() // false
// 是否是猪年
carbon.Parse("2020-08-05 13:14:15").Lunar().IsPigYear() // false
```