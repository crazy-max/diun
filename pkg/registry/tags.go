package registry

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/crazy-max/diun/v4/internal/matcher"
	"github.com/pkg/errors"
	"go.podman.io/image/v5/docker"
	"golang.org/x/mod/semver"
)

// normalizeSemver strips non-numeric leading characters and returns a valid
// semver string (with "v" prefix), or an empty string if the tag cannot be
// interpreted as semver.
func normalizeSemver(tag string) string {
	s := strings.TrimLeftFunc(tag, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	if vt := fmt.Sprintf("v%s", s); semver.IsValid(vt) {
		return vt
	}
	return ""
}

var generatedArtifactTagRe = regexp.MustCompile(`^sha256-[a-f0-9]{64}(?:\.(?:att|sbom|sig))?$`)

// Tags holds information about image tags.
type Tags struct {
	List         []string
	NotIncluded  int
	Excluded     int
	Artifacts    int
	OlderOrEqual int
	Total        int
}

// TagsOptions holds docker tags image options
type TagsOptions struct {
	Image   Image
	Max     int
	Sort    SortTag
	Include []string
	Exclude []string
	// MinSemver, when non-empty, restricts the list to tags whose semver is
	// strictly greater than this value.  Non-semver tags are silently dropped.
	MinSemver string
	// IncludePrereleases controls whether pre-release tags (e.g. -rc.1, -alpha)
	// are kept when MinSemver filtering is active.
	IncludePrereleases bool
}

// Tags returns tags of a Docker repository
func (c *Client) Tags(opts TagsOptions) (*Tags, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgRef, err := ImageReference(opts.Image.String())
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse reference")
	}

	tags, err := docker.GetRepositoryTags(ctx, c.sysCtx, imgRef)
	if err != nil {
		return nil, err
	}

	res := &Tags{
		NotIncluded: 0,
		Excluded:    0,
		Total:       len(tags),
	}

	// Sort tags
	tags = SortTags(tags, opts.Sort)

	// Resolve minimum semver once (empty string means no filtering)
	minSemver := normalizeSemver(opts.MinSemver)

	// Filter
	for _, tag := range tags {
		if generatedArtifactTagRe.MatchString(tag) {
			res.Artifacts++
			continue
		} else if !matcher.IsIncluded(tag, opts.Include) {
			res.NotIncluded++
			continue
		} else if matcher.IsExcluded(tag, opts.Exclude) {
			res.Excluded++
			continue
		}

		if minSemver != "" {
			tagSemver := normalizeSemver(tag)
			if tagSemver == "" {
				// Not a semver tag — skip when newer-only mode is active
				res.OlderOrEqual++
				continue
			}
			if !opts.IncludePrereleases && semver.Prerelease(tagSemver) != "" {
				res.OlderOrEqual++
				continue
			}
			if semver.Compare(tagSemver, minSemver) <= 0 {
				res.OlderOrEqual++
				continue
			}
		}

		res.List = append(res.List, tag)
	}

	if opts.Max > 0 && len(res.List) >= opts.Max {
		res.List = res.List[:opts.Max]
	}

	return res, nil
}

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
