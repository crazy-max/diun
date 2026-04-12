package registry

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
)

type Tags struct {
	List        []string
	NotIncluded int
	Excluded    int
	Total       int
}

type TagsOptions struct {
	Image   Image
	Max     int
	Sort    SortTag
	Include []string
	Exclude []string
}

func (c *Client) Tags(opts TagsOptions) (*Tags, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	regRef, err := opts.Image.regRef()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create regclient reference")
	}
	regRef = regRef.SetTag("")

	tags, err := c.regctl.TagList(ctx, regRef)
	if err != nil {
		return nil, errors.Wrap(err, "cannot list repository tags")
	}
	tagList := tags.Tags

	res := &Tags{
		NotIncluded: 0,
		Excluded:    0,
		Total:       len(tagList),
	}

	tagList = SortTags(tagList, opts.Sort)

	for _, tag := range tagList {
		if !utl.IsIncluded(tag, opts.Include) {
			res.NotIncluded++
			continue
		} else if utl.IsExcluded(tag, opts.Exclude) {
			res.Excluded++
			continue
		}
		res.List = append(res.List, tag)
	}

	if opts.Max > 0 && len(res.List) >= opts.Max {
		res.List = res.List[:opts.Max]
	}

	return res, nil
}

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

type SortTag string

const (
	SortTagDefault         = SortTag("default")
	SortTagReverse         = SortTag("reverse")
	SortTagLexicographical = SortTag("lexicographical")
	SortTagSemver          = SortTag("semver")
)

var SortTagTypes = []SortTag{
	SortTagDefault,
	SortTagReverse,
	SortTagLexicographical,
	SortTagSemver,
}

func (st *SortTag) Valid() bool {
	return st.OneOf(SortTagTypes)
}

func (st *SortTag) OneOf(stl []SortTag) bool {
	for _, n := range stl {
		if n == *st {
			return true
		}
	}
	return false
}
