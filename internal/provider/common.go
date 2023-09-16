package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/imdario/mergo"
)

var (
	metadataKeyChars  = `a-zA-Z0-9_`
	metadataKeyRegexp = regexp.MustCompile(`^[` + metadataKeyChars + `]+$`)
	errInvalidLabel   = errors.New("invalid label error")
)

// ValidateImage returns a standard image through Docker labels
func ValidateImage(image string, metadata, labels map[string]string, watchByDef bool, imageDefaults model.Image) (img model.Image, err error) {
	if i := strings.Index(image, "@sha256:"); i > 0 {
		image = image[:i]
	}

	img = model.Image{
		Name: image,
	}

	if err := mergo.Merge(&img, imageDefaults); err != nil {
		return img, fmt.Errorf("failed to merge image defaults for image %s", image)
	}

	if enableStr, ok := labels["diun.enable"]; ok {
		enable, err := strconv.ParseBool(enableStr)
		if err != nil {
			return img, fmt.Errorf("cannot parse %q value of label diun.enable: %w", enableStr, errInvalidLabel)
		}
		if !enable {
			return model.Image{}, nil
		}
	} else if !watchByDef {
		return model.Image{}, nil
	}

	for key, value := range labels {
		switch {
		case key == "diun.regopt":
			img.RegOpt = value
		case key == "diun.watch_repo":
			if watchRepo, err := strconv.ParseBool(value); err == nil {
				img.WatchRepo = &watchRepo
			} else {
				return img, fmt.Errorf("cannot parse %q value of label %s: %w", value, key, errInvalidLabel)
			}
		case key == "diun.notify_on":
			if len(value) == 0 {
				break
			}
			img.NotifyOn = []model.NotifyOn{}
			for _, no := range strings.Split(value, ";") {
				notifyOn := model.NotifyOn(no)
				if !notifyOn.Valid() {
					return img, fmt.Errorf("unknown notify status %q: %w", value, errInvalidLabel)
				}
				img.NotifyOn = append(img.NotifyOn, notifyOn)
			}
		case key == "diun.sort_tags":
			if value == "" {
				break
			}
			sortTags := registry.SortTag(value)
			if !sortTags.Valid() {
				return img, fmt.Errorf("unknown sort tags type %q: %w", value, errInvalidLabel)
			}
			img.SortTags = sortTags
		case key == "diun.max_tags":
			if img.MaxTags, err = strconv.Atoi(value); err != nil {
				return img, fmt.Errorf("cannot parse %q value of label %s: %w", value, key, errInvalidLabel)
			}
		case key == "diun.include_tags":
			img.IncludeTags = strings.Split(value, ";")
		case key == "diun.exclude_tags":
			img.ExcludeTags = strings.Split(value, ";")
		case key == "diun.hub_tpl":
			img.HubTpl = value
		case key == "diun.hub_link":
			img.HubLink = value
		case key == "diun.platform":
			platform, err := platforms.Parse(value)
			if err != nil {
				return img, fmt.Errorf("cannot parse %q platform of label %s: %w", value, key, errInvalidLabel)
			}
			img.Platform = model.ImagePlatform{
				OS:      platform.OS,
				Arch:    platform.Architecture,
				Variant: platform.Variant,
			}
		case strings.HasPrefix(key, "diun.metadata."):
			mkey := strings.TrimPrefix(key, "diun.metadata.")
			if len(mkey) == 0 || len(value) == 0 {
				break
			}
			if err := validateMetadataKey(mkey); err != nil {
				return img, fmt.Errorf("invalid metadata key %q: %w: %w", mkey, err, errInvalidLabel)
			}
			if img.Metadata == nil {
				img.Metadata = map[string]string{}
			}
			img.Metadata[mkey] = value
		}
	}

	// Update provider metadata with metadata from img labels
	if err := mergo.Merge(&img.Metadata, metadata); err != nil {
		return img, fmt.Errorf("failed merging metadata: %w", err)
	}

	return img, nil
}

func validateMetadataKey(key string) error {
	if !metadataKeyRegexp.MatchString(key) {
		return fmt.Errorf("only %q are allowed", metadataKeyChars)
	}
	return nil
}
