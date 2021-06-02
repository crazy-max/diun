package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/crazy-max/diun/v4/internal/model"
)

type ValidateImageOpts struct {
	Name           string
	Labels         map[string]string
	WatchByDefault bool
}

// ValidateImage returns a standard image through Docker labels
func ValidateImage(opts ValidateImageOpts) (img model.Image, err error) {
	if i := strings.Index(opts.Name, "@sha256:"); i > 0 {
		opts.Name = opts.Name[:i]
	}
	img = model.Image{
		Name: opts.Name,
	}

	if enableStr, ok := opts.Labels["diun.enable"]; ok {
		enable, err := strconv.ParseBool(enableStr)
		if err != nil {
			return img, fmt.Errorf("cannot parse %s value of label diun.enable", enableStr)
		}
		if !enable {
			return model.Image{}, nil
		}
	} else if !opts.WatchByDefault {
		return model.Image{}, nil
	}

	for key, value := range opts.Labels {
		switch key {
		case "diun.regopt":
			img.RegOpt = value
		case "diun.watch_repo":
			if img.WatchRepo, err = strconv.ParseBool(value); err != nil {
				return img, fmt.Errorf("cannot parse %s value of label %s", value, key)
			}
		case "diun.max_tags":
			if img.MaxTags, err = strconv.Atoi(value); err != nil {
				return img, fmt.Errorf("cannot parse %s value of label %s", value, key)
			}
		case "diun.include_tags":
			img.IncludeTags = strings.Split(value, ";")
		case "diun.exclude_tags":
			img.ExcludeTags = strings.Split(value, ";")
		case "diun.hub_tpl":
			img.HubTpl = value
		case "diun.platform":
			platform, err := platforms.Parse(value)
			if err != nil {
				return img, fmt.Errorf("cannot parse %s platform of label %s", value, key)
			}
			img.Platform = model.ImagePlatform{
				Os:      platform.OS,
				Arch:    platform.Architecture,
				Variant: platform.Variant,
			}
		}
	}

	return img, nil
}
