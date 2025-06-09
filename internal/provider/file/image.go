package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/platforms"
	"github.com/crazy-max/diun/v4/internal/model"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	yaml "gopkg.in/yaml.v2"
)

func (c *Client) listFileImage() []model.Image {
	var images model.ImageList

	files := c.getFiles()
	if len(files) == 0 {
		return []model.Image{}
	}

	for _, file := range files {
		var items []model.Image
		bytes, err := os.ReadFile(file)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Unable to read config file %s", file)
			continue
		}
		if err := yaml.UnmarshalStrict(bytes, &items); err != nil {
			c.logger.Error().Err(err).Msgf("Unable to decode into struct %s", file)
			continue
		}

		for _, item := range items {
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

			// Set default ExcludeTags
			if len(item.ExcludeTags) == 0 {
				item.ExcludeTags = c.defaults.ExcludeTags
			}

			images = append(images, item)
		}
	}

	return images.Dedupe()
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
