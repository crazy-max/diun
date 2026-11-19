// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package strparse is used to parse strings
package strparse

import (
	"fmt"

	"github.com/regclient/regclient/types/errs"
)

// SplitCSKV splits a comma separated key=value list into a map
func SplitCSKV(s string) (map[string]string, error) {
	state := "key"
	key := ""
	val := ""
	result := map[string]string{}
	procKV := func() {
		if key != "" {
			result[key] = val
		}
		state = "key"
		key = ""
		val = ""
	}
	for _, c := range s {
		switch state {
		case "key":
			switch c {
			case '"':
				state = "keyQuote"
			case '\\':
				state = "keyEscape"
			case '=':
				state = "val"
			case ',':
				procKV()
			default:
				key = key + string(c)
			}
		case "keyQuote":
			switch c {
			case '"':
				state = "key"
			case '\\':
				state = "keyEscapeQuote"
			default:
				key = key + string(c)
			}
		case "keyEscape":
			key = key + string(c)
			state = "key"
		case "keyEscapeQuote":
			key = key + string(c)
			state = "keyQuote"
		case "val":
			switch c {
			case '"':
				state = "valQuote"
			case ',':
				procKV()
			case '\\':
				state = "valEscape"
			default:
				val = val + string(c)
			}
		case "valQuote":
			switch c {
			case '"':
				state = "val"
			case '\\':
				state = "valEscapeQuote"
			default:
				val = val + string(c)
			}
		case "valEscape":
			val = val + string(c)
			state = "val"
		case "valEscapeQuote":
			val = val + string(c)
			state = "valQuote"
		default:
			return nil, fmt.Errorf("unhandled state: %s", state)
		}
	}
	switch state {
	case "val", "key":
		procKV()
	default:
		return nil, fmt.Errorf("string parsing failed, end state: %s%.0w", state, errs.ErrParsingFailed)
	}
	return result, nil
}
