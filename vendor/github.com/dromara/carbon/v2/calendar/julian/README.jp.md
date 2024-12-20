# 儒略の日/簡略儒略の日

日本語 | [English](README.md) | [简体中文](README.cn.md)

#### 使い方の例

##### `西暦` を `儒略日` に変換する
```go
// デフォルトの保持 6 ビット小数点精度
carbon.Parse("2024-01-24 12:00:00").Julian().JD() // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD() // 2460334.051563

// 4 ビット小数点精度の保持
carbon.Parse("2024-01-24 12:00:00").Julian().JD(4) // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD(4) // 2460334.0516
```

##### `西暦` を `簡略儒略日` に変換する
```go
// デフォルトの保持 6 ビット小数点精度
carbon.Parse("2024-01-24 12:00:00").Julian().MJD() // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD() // 60333.551563

// 4 ビット小数点精度の保持
carbon.Parse("2024-01-24 12:00:00").Julian().MJD(4) // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD(4) // 60333.5516
```

##### `儒略日` を `簡略儒略日` に変換する
```go
// デフォルトの保持 6 ビット小数点精度
carbon.CreateFromJulian(2460334).Julian().MJD() // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD() // 60332.551563

// 4 ビット小数点精度の保持
carbon.CreateFromJulian(2460334).Julian().MJD(4) // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD(4) // 60332.5516
```

##### `簡略儒略日` を `儒略日` に変換する
```go
// デフォルトの保持 6 ビット小数点精度
carbon.CreateFromJulian(60333.5).Julian().JD()() // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD()() // 2460333.051563

// 4 ビット小数点精度の保持
carbon.CreateFromJulian(60333.5).Julian().JD(4) // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD(4) // 2460333.0516
```

##### `儒略日`/`簡略儒略日` を `公历` に変換する
```go
// 儒略日 を 公历 に変換する
carbon.CreateFromJulian(2460334).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(2460334.051563).ToDateTimeString() // 2024-01-24 13:14:15

// 簡略儒略日 を 公历 に変換する
carbon.CreateFromJulian(60333.5).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(60333.551563).ToDateTimeString() // 2024-01-24 13:14:15
```