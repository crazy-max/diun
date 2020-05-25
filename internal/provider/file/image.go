package file

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/crazy-max/diun/internal/model"
	"gopkg.in/yaml.v2"
)

func (c *Client) listFileImage() []model.Image {
	var images []model.Image

	files := c.getFiles()
	if len(files) == 0 {
		return []model.Image{}
	}

	for _, file := range files {
		var items []model.Image
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Unable to read config file %s", file)
			continue
		}
		if err := yaml.UnmarshalStrict(bytes, &items); err != nil {
			c.logger.Error().Err(err).Msgf("Unable to decode into struct %s", file)
			continue
		}
		images = append(images, items...)
	}

	return images
}

func (c *Client) getFiles() []string {
	var files []string

	switch {
	case len(c.config.Directory) > 0:
		fileList, err := ioutil.ReadDir(c.config.Directory)
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
