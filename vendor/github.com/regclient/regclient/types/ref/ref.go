// Package ref is used to define references.
// References default to remote registry references (registry:port/repo:tag).
// Schemes can be included in front of the reference for different reference types.
package ref

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/regclient/regclient/types/errs"
)

const (
	dockerLibrary = "library"
	// dockerRegistry is the name resolved in docker images on Hub.
	dockerRegistry = "docker.io"
	// dockerRegistryLegacy is the name resolved in docker images on Hub.
	dockerRegistryLegacy = "index.docker.io"
	// dockerRegistryDNS is the host to connect to for Hub.
	dockerRegistryDNS = "registry-1.docker.io"
)

var (
	hostPartS = `(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?)`
	portS     = `(?:` + regexp.QuoteMeta(`:`) + `[0-9]+)`
	ipv6PartS = `(?:[0-9a-fA-F]{1,4}:){0,7}[0-9a-fA-F]{1,4}`
	ipv6S     = `(?:` + regexp.QuoteMeta(`[`) + `(?:` +
		ipv6PartS + `|` + // uncompressed
		regexp.QuoteMeta(`::`) + ipv6PartS + `|` + // prefix compressed
		ipv6PartS + regexp.QuoteMeta(`::`) + ipv6PartS + `|` + // middle compressed
		ipv6PartS + regexp.QuoteMeta(`::`) + // suffix compressed
		`)` + regexp.QuoteMeta(`]`) + `)`
	localhostS  = `localhost`
	hostDomainS = `(?:` + hostPartS + `(?:(?:` + regexp.QuoteMeta(`.`) + hostPartS + `)+` + regexp.QuoteMeta(`.`) + `?|` + regexp.QuoteMeta(`.`) + `))`
	hostUpperS  = `(?:[a-zA-Z0-9]*[A-Z][a-zA-Z0-9-]*[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[A-Z][a-zA-Z0-9]*)`
	registryS   = `(?:` +
		`(?:` + hostDomainS + `|` + hostUpperS + `|` + ipv6S + `|` + localhostS + `)` + portS + `?|` + // name with dotted domain, upper case, or IPv6 with optional port
		hostPartS + portS + // a short name with required port
		`)`
	repoPartS  = `[a-z0-9]+(?:(?:\.|_|__|-+)[a-z0-9]+)*`
	pathS      = `[/a-zA-Z0-9_\-. ~\+]+`
	tagS       = `[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}`
	digestS    = `[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}`
	schemeRE   = regexp.MustCompile(`^([a-z]+)://(.+)$`)
	registryRE = regexp.MustCompile(`^(` + registryS + `)$`)
	refRE      = regexp.MustCompile(`^(?:(` + registryS + `)` + regexp.QuoteMeta(`/`) + `)?` +
		`(` + repoPartS + `(?:` + regexp.QuoteMeta(`/`) + repoPartS + `)*)` +
		`(?:` + regexp.QuoteMeta(`:`) + `(` + tagS + `))?` +
		`(?:` + regexp.QuoteMeta(`@`) + `(` + digestS + `))?$`)
	ocidirRE = regexp.MustCompile(`^(` + pathS + `)` +
		`(?:` + regexp.QuoteMeta(`:`) + `(` + tagS + `))?` +
		`(?:` + regexp.QuoteMeta(`@`) + `(` + digestS + `))?$`)
)

// Ref is a reference to a registry/repository.
// Direct access to the contents of this struct should not be assumed.
type Ref struct {
	Scheme     string // Scheme is the type of reference, "reg" or "ocidir".
	Reference  string // Reference is the unparsed string or common name.
	Registry   string // Registry is the server for the "reg" scheme.
	Repository string // Repository is the path on the registry for the "reg" scheme.
	Tag        string // Tag is a mutable tag for a reference.
	Digest     string // Digest is an immutable hash for a reference.
	Path       string // Path is the directory of the OCI Layout for "ocidir".
}

// New returns a reference based on the scheme (defaulting to "reg").
func New(parse string) (Ref, error) {
	scheme := ""
	tail := parse
	matchScheme := schemeRE.FindStringSubmatch(parse)
	if len(matchScheme) == 3 {
		scheme = matchScheme[1]
		tail = matchScheme[2]
	}
	ret := Ref{
		Scheme:    scheme,
		Reference: parse,
	}
	switch scheme {
	case "":
		ret.Scheme = "reg"
		matchRef := refRE.FindStringSubmatch(tail)
		if len(matchRef) < 5 {
			if refRE.FindStringSubmatch(strings.ToLower(tail)) != nil {
				return Ref{}, fmt.Errorf("%w \"%s\", repo must be lowercase", errs.ErrInvalidReference, tail)
			}
			return Ref{}, fmt.Errorf("%w \"%s\"", errs.ErrInvalidReference, tail)
		}
		ret.Registry = matchRef[1]
		ret.Repository = matchRef[2]
		ret.Tag = matchRef[3]
		ret.Digest = matchRef[4]

		// handle localhost use case since it matches the regex for a repo path entry
		repoPath := strings.Split(ret.Repository, "/")
		if ret.Registry == "" && repoPath[0] == "localhost" {
			ret.Registry = repoPath[0]
			ret.Repository = strings.Join(repoPath[1:], "/")
		}
		switch ret.Registry {
		case "", dockerRegistryDNS, dockerRegistryLegacy:
			ret.Registry = dockerRegistry
		}
		if ret.Registry == dockerRegistry && !strings.Contains(ret.Repository, "/") {
			ret.Repository = dockerLibrary + "/" + ret.Repository
		}
		if ret.Tag == "" && ret.Digest == "" {
			ret.Tag = "latest"
		}
		if ret.Repository == "" {
			return Ref{}, fmt.Errorf("%w \"%s\"", errs.ErrInvalidReference, tail)
		}

	case "ocidir", "ocifile":
		matchPath := ocidirRE.FindStringSubmatch(tail)
		if len(matchPath) < 2 || matchPath[1] == "" {
			return Ref{}, fmt.Errorf("%w, invalid path for scheme \"%s\": %s", errs.ErrInvalidReference, scheme, tail)
		}
		ret.Path = matchPath[1]
		if len(matchPath) > 2 && matchPath[2] != "" {
			ret.Tag = matchPath[2]
		}
		if len(matchPath) > 3 && matchPath[3] != "" {
			ret.Digest = matchPath[3]
		}

	default:
		return Ref{}, fmt.Errorf("%w, unknown scheme \"%s\" in \"%s\"", errs.ErrInvalidReference, scheme, parse)
	}
	return ret, nil
}

// NewHost returns a Reg for a registry hostname or equivalent.
// The ocidir schema equivalent is the path.
func NewHost(parse string) (Ref, error) {
	scheme := ""
	tail := parse
	matchScheme := schemeRE.FindStringSubmatch(parse)
	if len(matchScheme) == 3 {
		scheme = matchScheme[1]
		tail = matchScheme[2]
	}
	ret := Ref{
		Scheme: scheme,
	}

	switch scheme {
	case "":
		ret.Scheme = "reg"
		matchReg := registryRE.FindStringSubmatch(tail)
		if len(matchReg) < 2 {
			return Ref{}, fmt.Errorf("%w \"%s\"", errs.ErrParsingFailed, tail)
		}
		ret.Registry = matchReg[1]
		if ret.Registry == "" {
			return Ref{}, fmt.Errorf("%w \"%s\"", errs.ErrParsingFailed, tail)
		}

	case "ocidir", "ocifile":
		matchPath := ocidirRE.FindStringSubmatch(tail)
		if len(matchPath) < 2 || matchPath[1] == "" {
			return Ref{}, fmt.Errorf("%w, invalid path for scheme \"%s\": %s", errs.ErrParsingFailed, scheme, tail)
		}
		ret.Path = matchPath[1]

	default:
		return Ref{}, fmt.Errorf("%w, unknown scheme \"%s\" in \"%s\"", errs.ErrParsingFailed, scheme, parse)
	}
	return ret, nil
}

// AddDigest returns a ref with the requested digest set.
// The tag will NOT be unset and the reference value will be reset.
func (r Ref) AddDigest(digest string) Ref {
	r.Digest = digest
	r.Reference = r.CommonName()
	return r
}

// CommonName outputs a parsable name from a reference.
func (r Ref) CommonName() string {
	cn := ""
	switch r.Scheme {
	case "reg":
		if r.Registry != "" {
			cn = r.Registry + "/"
		}
		if r.Repository == "" {
			return ""
		}
		cn = cn + r.Repository
		if r.Tag != "" {
			cn = cn + ":" + r.Tag
		}
		if r.Digest != "" {
			cn = cn + "@" + r.Digest
		}
	case "ocidir":
		cn = fmt.Sprintf("ocidir://%s", r.Path)
		if r.Tag != "" {
			cn = cn + ":" + r.Tag
		}
		if r.Digest != "" {
			cn = cn + "@" + r.Digest
		}
	}
	return cn
}

// IsSet returns true if needed values are defined for a specific reference.
func (r Ref) IsSet() bool {
	if !r.IsSetRepo() {
		return false
	}
	// Registry requires a tag or digest, OCI Layout doesn't require these.
	if r.Scheme == "reg" && r.Tag == "" && r.Digest == "" {
		return false
	}
	return true
}

// IsSetRepo returns true when the ref includes values for a specific repository.
func (r Ref) IsSetRepo() bool {
	switch r.Scheme {
	case "reg":
		if r.Registry != "" && r.Repository != "" {
			return true
		}
	case "ocidir":
		if r.Path != "" {
			return true
		}
	}
	return false
}

// IsZero returns true if ref is unset.
func (r Ref) IsZero() bool {
	if r.Scheme == "" && r.Registry == "" && r.Repository == "" && r.Path == "" && r.Tag == "" && r.Digest == "" {
		return true
	}
	return false
}

// SetDigest returns a ref with the requested digest set.
// The tag will be unset and the reference value will be reset.
func (r Ref) SetDigest(digest string) Ref {
	r.Digest = digest
	r.Tag = ""
	r.Reference = r.CommonName()
	return r
}

// SetTag returns a ref with the requested tag set.
// The digest will be unset and the reference value will be reset.
func (r Ref) SetTag(tag string) Ref {
	r.Tag = tag
	r.Digest = ""
	r.Reference = r.CommonName()
	return r
}

// ToReg converts a reference to a registry like syntax.
func (r Ref) ToReg() Ref {
	switch r.Scheme {
	case "ocidir":
		r.Scheme = "reg"
		r.Registry = "localhost"
		// clean the path to strip leading ".."
		r.Repository = path.Clean("/" + r.Path)[1:]
		r.Repository = strings.ToLower(r.Repository)
		// convert any unsupported characters to "-" in the path
		re := regexp.MustCompile(`[^/a-z0-9]+`)
		r.Repository = string(re.ReplaceAll([]byte(r.Repository), []byte("-")))
	}
	return r
}

// EqualRegistry compares the registry between two references.
func EqualRegistry(a, b Ref) bool {
	if a.Scheme != b.Scheme {
		return false
	}
	switch a.Scheme {
	case "reg":
		return a.Registry == b.Registry
	case "ocidir":
		return a.Path == b.Path
	case "":
		// both undefined
		return true
	default:
		return false
	}
}

// EqualRepository compares the repository between two references.
func EqualRepository(a, b Ref) bool {
	if a.Scheme != b.Scheme {
		return false
	}
	switch a.Scheme {
	case "reg":
		return a.Registry == b.Registry && a.Repository == b.Repository
	case "ocidir":
		return a.Path == b.Path
	case "":
		// both undefined
		return true
	default:
		return false
	}
}
