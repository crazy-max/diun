package file

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Client represents an active file provider object
type Client struct {
	*provider.Client
	item   model.PrdFile
	logger zerolog.Logger
}

// New creates new file provider instance
func New(item model.PrdFile) *provider.Client {
	return &provider.Client{
		Handler: &Client{
			item:   item,
			logger: log.With().Str("provider", "file").Logger(),
		},
	}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	images := c.loadImages()
	if len(images) == 0 {
		return []model.Job{}
	}

	c.logger.Info().Msgf("Found %d image(s) to analyze", len(images))
	var list []model.Job
	for _, elt := range images {
		list = append(list, model.Job{
			Provider: "file",
			Image:    elt,
		})
	}

	return list
}

func (c *Client) loadImages() []model.Image {
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
	case len(c.item.Directory) > 0:
		fileList, err := ioutil.ReadDir(c.item.Directory)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Unable to read directory %s", c.item.Directory)
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
			files = append(files, filepath.Join(c.item.Directory, file.Name()))
		}
	case len(c.item.Filename) > 0:
		files = append(files, c.item.Filename)
	}

	return files
}
