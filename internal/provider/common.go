package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
)

// ValidateImage returns a standard image through Docker labels
func ValidateImage(image string, labels map[string]string, watchByDef bool) (img model.Image, err error) {
	if i := strings.Index(image, "@sha256:"); i > 0 {
		image = image[:i]
	}
	img = model.Image{
		Name:     image,
		NotifyOn: model.NotifyOnDefaults,
		SortTags: registry.SortTagReverse,
	}

	if enableStr, ok := labels["diun.enable"]; ok {
		enable, err := strconv.ParseBool(enableStr)
		if err != nil {
			return img, fmt.Errorf("cannot parse %s value of label diun.enable", enableStr)
		}
		if !enable {
			return model.Image{}, nil
		}
	} else if !watchByDef {
		return model.Image{}, nil
	}

	for key, value := range labels {
		switch key {
		case "diun.regopt":
			img.RegOpt = value
		case "diun.watch_repo":
			if img.WatchRepo, err = strconv.ParseBool(value); err != nil {
				return img, fmt.Errorf("cannot parse %s value of label %s", value, key)
			}
		case "diun.notify_on":
			if len(value) == 0 {
				break
			}
			img.NotifyOn = []model.NotifyOn{}
			for _, no := range strings.Split(value, ";") {
				notifyOn := model.NotifyOn(no)
				if !notifyOn.Valid() {
					return img, fmt.Errorf("unknown notify status %q", value)
				}
				img.NotifyOn = append(img.NotifyOn, notifyOn)
			}
		case "diun.sort_tags":
			if value == "" {
				break
			}
			sortTags := registry.SortTag(value)
			if !sortTags.Valid() {
				return img, fmt.Errorf("unknown sort tags type %q", value)
			}
			img.SortTags = sortTags
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
		case "diun.hub_link":
			img.HubLink = value
		case "diun.platform":
			platform, err := platforms.Parse(value)
			if err != nil {
				return img, fmt.Errorf("cannot parse %s platform of label %s", value, key)
			}
			img.Platform = model.ImagePlatform{
				OS:      platform.OS,
				Arch:    platform.Architecture,
				Variant: platform.Variant,
			}
		}
	}

	return img, nil
}
