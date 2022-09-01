package registry

import (
	"github.com/containers/image/v5/docker"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Tags holds information about image tags.
type Tags struct {
	List        []string
	NotIncluded int
	Excluded    int
	Total       int
}

// TagsOptions holds docker tags image options
type TagsOptions struct {
	Image   Image
	Max     int
	Sort    SortTag
	Include []string
	Exclude []string
}

// Tags returns tags of a Docker repository
func (c *Client) Tags(opts TagsOptions) (*Tags, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgRef, err := ParseReference(opts.Image.String())
	if err != nil {
		return nil, errors.Wrap(err, "Cannot parse reference")
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

	// Filter tags
	tags = ExtractVersions(tags, opts.Include)

	// Sort tags
	tags = SortTags(tags, opts.Sort)

	// Filter
	for _, tag := range tags {
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
