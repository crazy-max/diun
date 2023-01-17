package registry

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/mod/semver"
)

// SortTags sorts tags list
func SortTags(tags []string, sortTag SortTag) []string {
	switch sortTag {
	case SortTagReverse:
		for i := len(tags)/2 - 1; i >= 0; i-- {
			opp := len(tags) - 1 - i
			tags[i], tags[opp] = tags[opp], tags[i]
		}
		return tags
	case SortTagLexicographical:
		sort.Strings(tags)
		return tags
	case SortTagSemver:
		semverIsh := func(s string) string {
			s = strings.TrimLeftFunc(s, func(r rune) bool {
				return !unicode.IsNumber(r)
			})
			if vt := fmt.Sprintf("v%s", s); semver.IsValid(vt) {
				return vt
			}
			return ""
		}
		sort.Slice(tags, func(i, j int) bool {
			if c := semver.Compare(semverIsh(tags[i]), semverIsh(tags[j])); c > 0 {
				return true
			} else if c < 0 {
				return false
			}
			if c := strings.Count(tags[i], ".") - strings.Count(tags[j], "."); c > 0 {
				return true
			} else if c < 0 {
				return false
			}

			return strings.Compare(tags[i], tags[j]) < 0
		})
		return tags
	default:
		return tags
	}
}

// SortTag holds sort tag type
type SortTag string

// SortTag constants
const (
	SortTagDefault         = SortTag("default")
	SortTagReverse         = SortTag("reverse")
	SortTagLexicographical = SortTag("lexicographical")
	SortTagSemver          = SortTag("semver")
)

// SortTagTypes is the list of available sort tag types
var SortTagTypes = []SortTag{
	SortTagDefault,
	SortTagReverse,
	SortTagLexicographical,
	SortTagSemver,
}

// Valid checks sort tag type is valid
func (st *SortTag) Valid() bool {
	return st.OneOf(SortTagTypes)
}

// OneOf checks if sort type is one of the values in the list
func (st *SortTag) OneOf(stl []SortTag) bool {
	for _, n := range stl {
		if n == *st {
			return true
		}
	}
	return false
}
