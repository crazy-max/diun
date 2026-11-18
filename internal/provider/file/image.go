package file

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/containerd/platforms"
	"github.com/crazy-max/diun/v4/internal/model"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
)

func (c *Client) listFileImage() []model.Image {
	var images []model.Image

	files := c.getFiles()
	if len(files) == 0 {
		return []model.Image{}
	}

	for _, file := range files {
		var items []model.Image
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Unable to read config file %s", file)
			continue
		}
		dec := yaml.NewDecoder(bytes.NewReader(fileBytes))
		dec.KnownFields(true)
		if err := dec.Decode(&items); err != nil {
			c.logger.Error().Err(err).Msgf("Unable to decode into struct %s", file)
			continue
		}

		for _, item := range items {
			if item.WatchRepo != nil {
				c.logger.Warn().Str("file", file).Str("img_name", item.Name).Msg("watch_repo is deprecated; use watch_mode (true: all, false: tag)")
			}
			if err := item.ResolveWatchMode(c.defaults); err != nil {
				c.logger.Error().Err(err).Str("file", file).Str("img_name", item.Name).Msg("Invalid watch mode")
				continue
			}
			// Set default WatchRepo
			if item.WatchRepo == nil {
				item.WatchRepo = c.defaults.WatchRepo
			}
			// Check NotifyOn
			if len(item.NotifyOn) == 0 {
				item.NotifyOn = c.defaults.NotifyOn
			} else {
				for _, no := range item.NotifyOn {
					if !no.Valid() {
						c.logger.Error().
							Str("file", file).
							Str("img_name", item.Name).
							Msgf("unknown notify status %q", no)
					}
				}
			}

			// Check SortType
			if item.SortTags == "" {
				item.SortTags = c.defaults.SortTags
			}
			if !item.SortTags.Valid() {
				c.logger.Error().
					Str("file", file).
					Str("img_name", item.Name).
					Msgf("unknown sort tags type %q", item.SortTags)
			}

			// Check Platform
			if item.Platform != (model.ImagePlatform{}) {
				_, err = platforms.Parse(platforms.Format(ocispecs.Platform{
					OS:           item.Platform.OS,
					Architecture: item.Platform.Arch,
					Variant:      item.Platform.Variant,
				}))
				if err != nil {
					c.logger.Error().
						Str("file", file).
						Str("img_name", item.Name).
						Msgf("cannot parse %s platform", platforms.Format(ocispecs.Platform{
							OS:           item.Platform.OS,
							Architecture: item.Platform.Arch,
							Variant:      item.Platform.Variant,
						}))
				}
			}

			// Set default MaxTags
			if item.MaxTags == 0 {
				item.MaxTags = c.defaults.MaxTags
			}

			// Set default IncludeTags
			if len(item.IncludeTags) == 0 {
				item.IncludeTags = c.defaults.IncludeTags
			}

			item.ExcludeTags = mergeTags(c.defaults.ExcludeTags, item.ExcludeTags)

			if item.AssumeImmutableTags == nil {
				item.AssumeImmutableTags = c.defaults.AssumeImmutableTags
			}

			images = append(images, item)
		}
	}

	return images
}

func (c *Client) getFiles() []string {
	var files []string

	switch {
	case len(c.config.Directory) > 0:
		fileList, err := os.ReadDir(c.config.Directory)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Unable to read directory %s", c.config.Directory)
			return files
		}
		for _, file := range fileList {
			if file.IsDir() {
				continue
			}
			switch strings.ToLower(filepath.Ext(file.Name())) {
			case ".yaml", ".yml":
				// noop
			default:
				continue
			}
			files = append(files, filepath.Join(c.config.Directory, file.Name()))
		}
	case len(c.config.Filename) > 0:
		files = append(files, c.config.Filename)
	}

	return files
}

func mergeTags(defaultTags, imageTags []string) []string {
	tags := slices.Concat(defaultTags, imageTags)
	seen := make(map[string]struct{}, len(tags))
	return slices.DeleteFunc(tags, func(tag string) bool {
		if _, ok := seen[tag]; ok {
			return true
		}
		seen[tag] = struct{}{}
		return false
	})
}
