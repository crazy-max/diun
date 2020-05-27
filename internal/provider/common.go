package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/crazy-max/diun/v3/internal/model"
)

// ValidateContainerImage returns a standard image through Docker labels
func ValidateContainerImage(image string, labels map[string]string, watchByDef bool) (img model.Image, err error) {
	if i := strings.Index(image, "@sha256:"); i > 0 {
		image = image[:i]
	}
	img = model.Image{
		Name: image,
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
		case "diun.regopts_id":
			img.RegOptsID = value
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
		}
	}

	return img, nil
}
