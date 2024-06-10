package registry

import (
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/distribution/reference"
	"github.com/pkg/errors"
)

func ImageReference(name string) (types.ImageReference, error) {
	ref, err := namedReference(name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse reference")
	}
	refStr := ref.String()
	if !strings.HasPrefix(refStr, "//") {
		refStr = fmt.Sprintf("//%s", refStr)
	}
	return docker.ParseReference(refStr)
}

func namedReference(name string) (reference.Named, error) {
	name = strings.TrimPrefix(name, "//")

	ref, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing normalized named %q", name)
	}

	if _, ok := ref.(reference.Named); !ok {
		return nil, errors.Errorf("%q is not a named reference", name)
	}

	if _, hasTag := ref.(reference.NamedTagged); hasTag {
		ref, err = normalizeTaggedDigestedNamed(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "normalizing tagged digested name %q", name)
		}
	} else if _, hasDigest := ref.(reference.Digested); hasDigest {
		ref = reference.TrimNamed(ref)
	}

	return reference.TagNameOnly(ref), nil
}

// normalizeTaggedDigestedNamed strips the digest off the specified named
// reference if it is tagged and digested.
func normalizeTaggedDigestedNamed(named reference.Named) (reference.Named, error) {
	_, isDigested := named.(reference.Digested)
	if !isDigested {
		return named, nil
	}
	tag, isTagged := named.(reference.NamedTagged)
	if !isTagged {
		return named, nil
	}
	// strip off the tag and digest
	newNamed := reference.TrimNamed(named)
	// re-add the tag
	newNamed, err := reference.WithTag(newNamed, tag.Tag())
	if err != nil {
		return named, err
	}
	return newNamed, nil
}
