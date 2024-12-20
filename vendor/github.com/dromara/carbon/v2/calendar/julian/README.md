# Julian Day/Modified Julian Day

English | [简体中文](README.cn.md) | [日本語](README.jp.md)

#### Usage and example

##### Convert `Gregorian` calendar to `Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().JD() // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD() // 2460334.051563

// 4 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().JD(4) // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD(4) // 2460334.0516
```

##### Convert `Gregorian` calendar to `Modified Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().MJD() // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD() // 60333.551563

// 4 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().MJD(4) // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD(4) // 60333.5516
```

##### Convert `Julian Day` to `Modified Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.CreateFromJulian(2460334).Julian().MJD() // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD() // 60332.551563

// 4 decimal places are retained for precision
carbon.CreateFromJulian(2460334).Julian().MJD(4) // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD(4) // 60332.5516
```

##### Convert `Modified Julian Day` to `Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.CreateFromJulian(60333.5).Julian().JD()() // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD()() // 2460333.051563

// 4 decimal places are retained for precision
carbon.CreateFromJulian(60333.5).Julian().JD(4) // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD(4) // 2460333.0516
```

##### Convert `Julian Day`/`Modified Julian Day` to `Gregorian` calendar
```go
// Convert Julian Day to Gregorian calendar
carbon.CreateFromJulian(2460334).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(2460334.051563).ToDateTimeString() // 2024-01-24 13:14:15

// Convert Modified Julian Day to Gregorian calendar
carbon.CreateFromJulian(60333.5).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(60333.551563).ToDateTimeString() // 2024-01-24 13:14:15
```