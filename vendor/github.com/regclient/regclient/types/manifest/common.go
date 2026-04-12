package manifest

import (
	"net/http"
	"strconv"
	"strings"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	digest "github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
)

type common struct {
	r         ref.Ref
	desc      descriptor.Descriptor
	manifSet  bool
	ratelimit types.RateLimit
	rawHeader http.Header
	rawBody   []byte
}

// GetDigest returns the digest
func (m *common) GetDigest() digest.Digest {
	return m.desc.Digest
}

// GetDescriptor returns the descriptor
func (m *common) GetDescriptor() descriptor.Descriptor {
	return m.desc
}

// GetMediaType returns the media type
func (m *common) GetMediaType() string {
	return m.desc.MediaType
}

// GetRateLimit returns the rate limit when the manifest was pulled from a registry.
// This supports the headers used by Docker Hub.
func (m *common) GetRateLimit() types.RateLimit {
	return m.ratelimit
}

// GetRef returns the reference from the upstream registry
func (m *common) GetRef() ref.Ref {
	return m.r
}

// HasRateLimit indicates if the rate limit is set
func (m *common) HasRateLimit() bool {
	return m.ratelimit.Set
}

// IsList indicates if the manifest is a docker Manifest List or OCI Index
func (m *common) IsList() bool {
	switch m.desc.MediaType {
	case MediaTypeDocker2ManifestList, MediaTypeOCI1ManifestList:
		return true
	default:
		return false
	}
}

// IsSet indicates if the manifest is defined.
// A false indicates this is from a HEAD request, providing the digest, media-type, and other headers, but no body.
func (m *common) IsSet() bool {
	return m.manifSet
}

// RawBody returns the raw body from the manifest if available.
func (m *common) RawBody() ([]byte, error) {
	if len(m.rawBody) == 0 {
		return m.rawBody, errs.ErrManifestNotSet
	}
	return m.rawBody, nil
}

// RawHeaders returns any headers included when manifest was pulled from a registry.
func (m *common) RawHeaders() (http.Header, error) {
	return m.rawHeader, nil
}

func (m *common) setRateLimit(header http.Header) {
	// check for rate limit headers
	rlLimit := header.Get("RateLimit-Limit")
	rlRemain := header.Get("RateLimit-Remaining")
	rlReset := header.Get("RateLimit-Reset")
	if rlLimit != "" {
		lpSplit := strings.Split(rlLimit, ",")
		lSplit := strings.Split(lpSplit[0], ";")
		rlLimitI, err := strconv.Atoi(lSplit[0])
		if err != nil {
			m.ratelimit.Limit = 0
		} else {
			m.ratelimit.Limit = rlLimitI
		}
		if len(lSplit) > 1 {
			m.ratelimit.Policies = lpSplit
		} else if len(lpSplit) > 1 {
			m.ratelimit.Policies = lpSplit[1:]
		}
	}
	if rlRemain != "" {
		rSplit := strings.Split(rlRemain, ";")
		rlRemainI, err := strconv.Atoi(rSplit[0])
		if err != nil {
			m.ratelimit.Remain = 0
		} else {
			m.ratelimit.Remain = rlRemainI
			m.ratelimit.Set = true
		}
	}
	if rlReset != "" {
		rlResetI, err := strconv.Atoi(rlReset)
		if err != nil {
			m.ratelimit.Reset = 0
		} else {
			m.ratelimit.Reset = rlResetI
		}
	}
}
