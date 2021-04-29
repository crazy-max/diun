package dockerfile

import (
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/pkg/errors"
)

type Image struct {
	Name     string
	Code     string
	Comments []string
	Line     int
}

type Images []Image

func (is *Images) has(image Image) bool {
	for _, i := range *is {
		if i.Line == image.Line {
			return true
		}
	}
	return false
}

// FromImages returns external images found in Dockerfile
func (c *Client) FromImages() (Images, error) {
	images := Images{}

	for _, node := range c.ast.Children {
		switch node.Value {
		case command.From:
			ins, err := instructions.ParseInstruction(node)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot parse instruction")
			}
			if baseName := ins.(*instructions.Stage).BaseName; baseName != "scratch" {
				name, err := c.shlex.ProcessWordWithMap(baseName, metaArgsToMap(c.metaArgs))
				if err != nil {
					return nil, err
				}
				image := Image{
					Name:     name,
					Code:     node.Original,
					Comments: node.PrevComment,
					Line:     node.StartLine,
				}
				if c.isStageName(name) || images.has(image) {
					continue
				}
				images = append(images, image)
			}
		case command.Copy:
			cmd, err := instructions.ParseCommand(node)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot parse command")
			}
			if copyFrom := cmd.(*instructions.CopyCommand).From; copyFrom != "null" {
				name, err := c.shlex.ProcessWordWithMap(copyFrom, metaArgsToMap(c.metaArgs))
				if err != nil {
					return nil, err
				}
				image := Image{
					Name:     name,
					Code:     node.Original,
					Comments: node.PrevComment,
					Line:     node.StartLine,
				}
				if c.isStageName(name) || images.has(image) {
					continue
				}
				images = append(images, image)
			}
		case command.Run:
			cmd, err := instructions.ParseCommand(node)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot parse command")
			}
			if cmdRun, ok := cmd.(*instructions.RunCommand); ok {
				mounts := instructions.GetMounts(cmdRun)
				for _, mount := range mounts {
					if mount.Type != instructions.MountTypeBind || len(mount.From) == 0 {
						continue
					}
					name, err := c.shlex.ProcessWordWithMap(mount.From, metaArgsToMap(c.metaArgs))
					if err != nil {
						return nil, err
					}
					image := Image{
						Name:     name,
						Code:     node.Original,
						Comments: node.PrevComment,
						Line:     node.StartLine,
					}
					if c.isStageName(name) || images.has(image) {
						continue
					}
					images = append(images, image)
				}
			}
		}
	}

	return images, nil
}
