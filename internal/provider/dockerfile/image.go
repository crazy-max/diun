package dockerfile

import (
	"reflect"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/dockerfile"
	"github.com/crazy-max/diun/v4/pkg/utl"
)

func (c *Client) listExtImage() (list []model.Image) {
	for _, filename := range c.listDockerfiles(c.config.Patterns) {
		dfile, err := dockerfile.New(dockerfile.Options{
			Filename: filename,
		})
		if err != nil {
			c.logger.Warn().Err(err).Msg("Cannot create dockerfile client")
			continue
		}
		fromImages, err := dfile.FromImages()
		if err != nil {
			c.logger.Warn().Err(err).Msg("Cannot extract images")
			continue
		}
		for _, fromImage := range fromImages {
			c.logger.Debug().
				Str("dfile_image", fromImage.Name).
				Str("dfile_code", fromImage.Code).
				Interface("dfile_comments", fromImage.Comments).
				Int("dfile_line", fromImage.Line).
				Msg("Validate image")
			image, err := provider.ValidateImage(fromImage.Name, nil, c.extractLabels(fromImage.Comments), true, *c.imageDefaults)
			if err != nil {
				c.logger.Error().Err(err).
					Str("dfile_image", fromImage.Name).
					Str("dfile_code", fromImage.Code).
					Interface("dfile_comments", fromImage.Comments).
					Int("dfile_line", fromImage.Line).
					Msg("Invalid image")
				continue
			} else if reflect.DeepEqual(image, model.Image{}) {
				c.logger.Debug().
					Str("dfile_image", fromImage.Name).
					Str("dfile_code", fromImage.Code).
					Interface("dfile_comments", fromImage.Comments).
					Int("dfile_line", fromImage.Line).
					Msg("Watch disabled")
				continue
			}
			list = append(list, image)
		}
	}
	return
}

func (c *Client) listDockerfiles(patterns []string) (dfiles []string) {
	if len(patterns) == 0 {
		patterns = []string{"./Dockerfile"}
	}
	for _, pattern := range patterns {
		matches, err := doublestar.Glob(pattern)
		if err != nil {
			c.logger.Warn().Err(err).Msgf("No Dockerfile found for %s", pattern)
			continue
		}
		for _, dfile := range matches {
			if utl.Contains(dfiles, dfile) {
				continue
			}
			dfiles = append(dfiles, dfile)
		}
	}
	return
}

func (c *Client) extractLabels(comments []string) map[string]string {
	labels := map[string]string{}
	if len(comments) == 0 {
		return labels
	}
	for _, comment := range comments {
		if !strings.HasPrefix(comment, "diun.") {
			continue
		}
		kvp := strings.SplitN(comment, "=", 2)
		if len(kvp) == 2 {
			labels[kvp[0]] = kvp[1]
		}
	}
	return labels
}
