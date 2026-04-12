// Package units is taken from https://github.com/docker/go-units
package units

// Copyright 2015 Docker, Inc.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 		https://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"fmt"
)

var (
	decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	binaryAbbrs  = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

func getSizeAndUnit(size float64, base float64, unitList []string) (float64, string) {
	i := 0
	unitsLimit := len(unitList) - 1
	for size >= base && i < unitsLimit {
		size = size / base
		i++
	}
	return size, unitList[i]
}

// CustomSize returns a human-readable approximation of a size using custom format.
func CustomSize(format string, size float64, base float64, unitList []string) string {
	size, unit := getSizeAndUnit(size, base, unitList)
	return fmt.Sprintf(format, size, unit)
}

// HumanSizeWithPrecision allows the size to be in any precision.
func HumanSizeWithPrecision(size float64, width, precision int) string {
	size, unit := getSizeAndUnit(size, 1000.0, decimapAbbrs)
	return fmt.Sprintf("%*.*f%s", width, precision, size, unit)
}

// HumanSize returns a human-readable approximation of a size
// with a width of 5 (eg. "2.746MB", "796.0KB").
func HumanSize(size float64) string {
	return HumanSizeWithPrecision(size, 5, 3)
}

// BytesSize returns a human-readable size in bytes, kibibytes,
// mebibytes, gibibytes, or tebibytes (eg. "44.2kiB", "17.6MiB").
func BytesSize(size float64) string {
	return CustomSize("%5.3f%s", size, 1024.0, binaryAbbrs)
}
