# Changelog

## 0.7.1 (2023/05/05)

* handle `time.Time` reflection (#47)

## 0.7.0 (2023/05/05)

* Go 1.20 (#43)
* Backport from paerser (#45)
  * fix: raw slice parsing
  * fix: return error when a string is use instead of a slice
  * fix: typo in error message
  * fix: allow invalid configuration
  * fix: raw slice of struct in raw type
  * feat: add .json file extension to file.Decode
  * allow decoding of nested raw typed slices
* Bump gopkg.in/yaml.v3 to v3.0.1 (#44)
* Bump github.com/BurntSushi/toml from 1.1.0 to 1.2.1 (#38 #39)
* Bump github.com/stretchr/testify from 1.8.0 to 1.8.2 (#41)

## 0.6.0 (2022/07/17)

* Go 1.18 support (#32)
* golangci-lint config (#34)
* Container based dev flow (#33)
* Drop Go 1.13 and 1.14 support (#33)
* Backport from paerser (#18)
  * fix(file): allow slice value that contains comma
  * fix: ignore tag for label decoding
* Bump github.com/BurntSushi/toml from 0.4.1 to 1.1.0 (#25)
* Bump github.com/stretchr/testify from 1.7.0 to 1.7.2 (#30)
* Bump github.com/stretchr/testify from 1.7.2 to 1.8.0 (#37)

## 0.5.0 (2021/08/21)

* Incorrect conversion between integer types (#19)
* Test against Go 1.16 and 1.17
* Backport from paerser (#18)
  * fix: invalid slice metadata.
  * fix: apply time.Second as default unit for integer json unmarshalling
  * fix: bijectivity of JSON marshal and unmarshal
  * fix: simplify MarshalJSON.
* Bump github.com/BurntSushi/toml from 0.3.1 to 0.4.1 (#17)
* Bump codecov/codecov-action from 1 to 2
* Bump github.com/stretchr/testify from 1.6.1 to 1.7.0 (#12)
* Bump gopkg.in/yaml.v2 from 2.3.0 to 2.4.0 (#11)

## 0.4.0 (2020/11/08)

* Handle raw values

## 0.3.0 (2020/08/15)

* Go 1.15 support

## 0.2.0 (2020/07/16)

* More tests
* Fix example and display configuration

## 0.1.1 (2020/07/16)

* Fix loader
* Add example

## 0.1.0 (2020/07/15)

* Initial version
