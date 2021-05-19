package main

import (
	"context"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/crazy-max/diun/v4/pb"
	"github.com/jedib0t/go-pretty/v6/table"
)

// ImageCmd holds image command
type ImageCmd struct {
	List ImageListCmd `kong:"cmd='list',default='1',help='List images in database.'"`
}

// ImageListCmd holds image list command
type ImageListCmd struct {
	CliGlobals
	Raw bool `kong:"name='raw',default='false',help='Profiler to use.'"`
}

func (s *ImageListCmd) Run(ctx *Context) error {
	defer s.conn.Close()
	images, err := s.imageSvc.ImageList(context.Background(), &pb.ImageListRequest{})
	if err != nil {
		return err
	}

	sort.Slice(images.Image, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, images.Image[i].Name) < strings.Map(unicode.ToUpper, images.Image[j].Name)
	})

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Tag", "Updated", "Digest"})
	for _, image := range images.Image {
		t.AppendRow(table.Row{"", image.Name, image.LastTag, image.LastUpdated.AsTime().Format(time.RFC3339), image.LastDigest})
	}
	t.AppendFooter(table.Row{"Total", len(images.Image)})
	t.Render()

	return nil
}
