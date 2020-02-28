package registry

import (
	"github.com/containers/image/v5/docker"
	"github.com/crazy-max/diun/pkg/utl"
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
	Include []string
	Exclude []string
}

// Tags returns tags of a Docker repository
func (c *Client) Tags(opts TagsOptions) (*Tags, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgCls, err := c.newImage(ctx, opts.Image.String())
	if err != nil {
		return nil, err
	}
	defer imgCls.Close()

	tags, err := docker.GetRepositoryTags(ctx, c.sysCtx, imgCls.Reference())
	if err != nil {
		return nil, err
	}

	res := &Tags{
		NotIncluded: 0,
		Excluded:    0,
		Total:       len(tags),
	}

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

	// Reverse order (latest tags first)
	for i := len(res.List)/2 - 1; i >= 0; i-- {
		opp := len(res.List) - 1 - i
		res.List[i], res.List[opp] = res.List[opp], res.List[i]
	}

	if opts.Max > 0 && len(res.List) >= opts.Max {
		res.List = res.List[:opts.Max]
	}

	return res, nil
}
