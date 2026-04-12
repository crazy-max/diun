package registry

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/distribution/reference"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	regref "github.com/regclient/regclient/types/ref"
)

type Image struct {
	Domain  string
	Path    string
	Tag     string
	Digest  digest.Digest
	HubLink string

	named reference.Named
	opts  ParseImageOptions
}

type ParseImageOptions struct {
	Name   string
	HubTpl string
}

func ParseImage(parseOpts ParseImageOptions) (Image, error) {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(parseOpts.Name)
	if err != nil {
		return Image{}, errors.Wrapf(err, "parsing image %s failed", parseOpts.Name)
	}

	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)

	i := Image{
		opts:   parseOpts,
		named:  named,
		Domain: reference.Domain(named),
		Path:   reference.Path(named),
	}

	// Hub link
	i.HubLink, err = i.hubLink()
	if err != nil {
		return Image{}, errors.Wrap(err, fmt.Sprintf("resolving hub link for image %s failed", parseOpts.Name))
	}

	// Add the tag if there was one.
	if tagged, ok := named.(reference.Tagged); ok {
		i.Tag = tagged.Tag()
	}

	// Add the digest if there was one.
	if canonical, ok := named.(reference.Canonical); ok {
		i.Digest = canonical.Digest()
	}

	return i, nil
}

func (i Image) Name() string {
	return i.named.Name()
}

func (i Image) String() string {
	return i.named.String()
}

func (i Image) Reference() string {
	if len(i.Digest.String()) > 1 {
		return i.Digest.String()
	}
	return i.Tag
}

func (i Image) regRef() (regref.Ref, error) {
	ref, err := regref.New(i.Name())
	if err != nil {
		return regref.Ref{}, err
	}
	if i.Tag != "" {
		ref = ref.SetTag(i.Tag)
	}
	return ref, nil
}

func (i Image) hubLink() (string, error) {
	if i.opts.HubTpl != "" {
		var out bytes.Buffer
		tmpl, err := template.New("tmpl").
			Option("missingkey=error").
			Parse(i.opts.HubTpl)
		if err != nil {
			return "", err
		}
		err = tmpl.Execute(&out, i)
		return out.String(), err
	}

	switch i.Domain {
	case "docker.io":
		prefix := "r"
		path := i.Path
		if strings.HasPrefix(i.Path, "library/") {
			prefix = "_"
			path = strings.Replace(i.Path, "library/", "", 1)
		}
		return fmt.Sprintf("https://hub.docker.com/%s/%s", prefix, path), nil
	case "docker.bintray.io", "jfrog-docker-reg2.bintray.io":
		return fmt.Sprintf("https://bintray.com/jfrog/reg2/%s", strings.ReplaceAll(i.Path, "/", "%3A")), nil
	case "docker.pkg.github.com":
		return fmt.Sprintf("https://github.com/%s/packages", filepath.ToSlash(filepath.Dir(i.Path))), nil
	case "gcr.io":
		return fmt.Sprintf("https://%s/%s", i.Domain, i.Path), nil
	case "ghcr.io":
		ref := strings.Split(i.Path, "/")
		ghUser, ghPackage := ref[0], ref[1]
		return fmt.Sprintf("https://github.com/users/%s/packages/container/package/%s", ghUser, ghPackage), nil
	case "quay.io":
		return fmt.Sprintf("https://quay.io/repository/%s", i.Path), nil
	case "registry.access.redhat.com":
		return fmt.Sprintf("https://access.redhat.com/containers/#/registry.access.redhat.com/%s", i.Path), nil
	case "registry.gitlab.com":
		return fmt.Sprintf("https://gitlab.com/%s/container_registry", i.Path), nil
	default:
		return "", nil
	}
}
