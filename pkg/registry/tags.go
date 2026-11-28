package registry

import (
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/crazy-max/diun/v4/internal/matcher"
	"github.com/pkg/errors"
	"go.podman.io/image/v5/docker"
	"golang.org/x/mod/semver"
)

var generatedArtifactTagRe = regexp.MustCompile(`^sha256-[a-f0-9]{64}(?:\.(?:att|sbom|sig))?$`)

// Tags holds information about image tags.
type Tags struct {
	List             []string
	NotIncluded      int
	Excluded         int
	Artifacts        int
	OutsideWatchMode int
	Total            int
}

// TagsOptions holds docker tags image options
type TagsOptions struct {
	WatchMode WatchMode // An empty WatchMode means all.
	Image     Image
	Max       int
	Sort      SortTag
	Include   []string
	Exclude   []string
}

// Tags returns tags of a Docker repository
func (c *Client) Tags(opts TagsOptions) (*Tags, error) {
	if _, err := opts.Select(nil); err != nil {
		return nil, err
	}
	if opts.WatchMode == WatchModeTag {
		return &Tags{}, nil
	}
	tags, err := c.ListTags(opts.Image)
	if err != nil {
		return nil, err
	}
	return opts.Select(tags)
}

// ListTags retrieves the unfiltered repository tags. The caller owns the slice.
func (c *Client) ListTags(image Image) ([]string, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()
	imgRef, err := ImageReference(image.String())
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse reference")
	}
	sysCtx, err := c.systemContext(ctx, image.Domain)
	if err != nil {
		return nil, err
	}
	return docker.GetRepositoryTags(ctx, sysCtx, imgRef)
}

// Select applies this watch's selection without modifying the supplied tags.
func (opts TagsOptions) Select(tags []string) (*Tags, error) {
	mode := opts.WatchMode
	if mode == "" {
		mode = WatchModeAll
	}
	if !mode.Valid() {
		return nil, errors.Errorf("unknown watch mode %q", mode)
	}
	if mode == WatchModeTag {
		return &Tags{}, nil
	}
	var baseline string
	if mode == WatchModeSemver || mode == WatchModeSemverPrereleases {
		baseline = versionTag(opts.Image.Tag)
		if baseline == "" {
			return nil, errors.Errorf("watch mode %q requires a semantic version tag (major.minor.patch), got %q", mode, opts.Image.Tag)
		}
	}

	res := &Tags{
		NotIncluded: 0,
		Excluded:    0,
		Total:       len(tags),
	}

	// Sort tags
	tags = SortTags(slices.Clone(tags), opts.Sort)

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
		if baseline != "" {
			version := versionTag(tag)
			if version == "" || semver.Compare(version, baseline) <= 0 || (mode == WatchModeSemver && semver.Prerelease(version) != "") {
				res.OutsideWatchMode++
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

// versionTag accepts full semantic versions with an optional v prefix. Sorting
// deliberately supports more tag formats than version comparisons do.
func versionTag(tag string) string {
	version := "v" + strings.TrimPrefix(tag, "v")
	core := version
	if i := strings.IndexAny(core, "-+"); i >= 0 {
		core = core[:i]
	}
	if strings.Count(core, ".") != 2 || !semver.IsValid(version) {
		return ""
	}
	return version
}

// WatchMode selects which image tags to watch.
type WatchMode string

// Watch modes.
const (
	WatchModeTag               WatchMode = "tag"
	WatchModeAll               WatchMode = "all"
	WatchModeSemver            WatchMode = "semver"
	WatchModeSemverPrereleases WatchMode = "semver+prereleases"
)

// Valid reports whether the watch mode is supported.
func (m WatchMode) Valid() bool {
	switch m {
	case WatchModeTag, WatchModeAll, WatchModeSemver, WatchModeSemverPrereleases:
		return true
	default:
		return false
	}
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
			// semver only accepts shorthand versions without suffixes.
			// Expand the version core while preserving prerelease and build suffixes.
			if i := strings.IndexAny(s, "-+"); i >= 0 {
				core, suffix := s[:i], s[i:]
				switch strings.Count(core, ".") {
				case 0:
					core += ".0.0"
				case 1:
					core += ".0"
				}
				s = core + suffix
			}
			if vt := "v" + s; semver.IsValid(vt) {
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
	return slices.Contains(stl, *st)
}
