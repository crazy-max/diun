package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/crazy-max/diun/v4/pb"
	"github.com/docker/go-units"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/tidwall/pretty"
	"google.golang.org/protobuf/encoding/protojson"
)

// ImageCmd holds image command
type ImageCmd struct {
	List    ImageListCmd    `kong:"cmd,default='1',help='List images in database.'"`
	Inspect ImageInspectCmd `kong:"cmd,help='Display information of an image in database.'"`
	Remove  ImageRemoveCmd  `kong:"cmd,help='Remove an image manifest from database.'"`
	//Prune   ImagePruneCmd   `kong:"cmd,help='Remove unused manifests from the database.'"`
}

// ImageListCmd holds image list command
type ImageListCmd struct {
	CliGlobals
	Raw bool `kong:"name='raw',default='false',help='JSON output.'"`
}

func (s *ImageListCmd) Run(ctx *Context) error {
	defer s.conn.Close()

	il, err := s.imageSvc.ImageList(context.Background(), &pb.ImageListRequest{})
	if err != nil {
		return err
	}

	sort.Slice(il.Images, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, il.Images[i].Name) < strings.Map(unicode.ToUpper, il.Images[j].Name)
	})

	if s.Raw {
		b, _ := protojson.Marshal(il)
		fmt.Println(string(pretty.Pretty(b)))
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Manifests Count", "Latest Tag", "Latest Created", "Latest Digest"})
	for _, image := range il.Images {
		t.AppendRow(table.Row{image.Name, image.ManifestsCount, image.Latest.Tag, image.Latest.Created.AsTime().Format(time.RFC3339), image.Latest.Digest})
	}
	t.AppendFooter(table.Row{"Total", len(il.Images)})
	t.Render()

	return nil
}

// ImageInspectCmd holds image inspect command
type ImageInspectCmd struct {
	CliGlobals
	Image string `kong:"arg,name='image',required,help='Image to inspect.'"`
	Raw   bool   `kong:"name='raw',default='false',help='JSON output.'"`
}

func (s *ImageInspectCmd) Run(ctx *Context) error {
	defer s.conn.Close()

	ii, err := s.imageSvc.ImageInspect(context.Background(), &pb.ImageInspectRequest{
		Name: s.Image,
	})
	if err != nil {
		return err
	}

	sort.Slice(ii.Image.Manifests, func(i, j int) bool {
		return ii.Image.Manifests[i].Created.AsTime().After(ii.Image.Manifests[j].Created.AsTime())
	})

	if s.Raw {
		b, _ := protojson.Marshal(ii)
		fmt.Println(string(pretty.Pretty(b)))
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Tag", "Created", "Digest"})
	for _, image := range ii.Image.Manifests {
		t.AppendRow(table.Row{image.Tag, image.Created.AsTime().Format(time.RFC3339), image.Digest})
	}
	t.AppendFooter(table.Row{"Total", len(ii.Image.Manifests)})
	t.Render()

	return nil
}

// ImageRemoveCmd holds image remove command
type ImageRemoveCmd struct {
	CliGlobals
	Image string `kong:"arg,name='image',required,help='Image to remove.'"`
	All   bool   `kong:"name='all',default='false',help='Remove all manifests from the database.'"`
}

func (s *ImageRemoveCmd) Run(ctx *Context) error {
	defer s.conn.Close()

	removed, err := s.imageSvc.ImageRemove(context.Background(), &pb.ImageRemoveRequest{
		Name: s.Image,
	})
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Tag", "Created", "Digest", "Size"})
	var totalSize int64
	for _, image := range removed.Manifests {
		t.AppendRow(table.Row{image.Tag, image.Created.AsTime().Format(time.RFC3339), image.Digest, units.HumanSize(float64(image.Size))})
		totalSize += image.Size
	}
	t.AppendFooter(table.Row{"Total", fmt.Sprintf("%d (%s)", len(removed.Manifests), units.HumanSize(float64(totalSize)))})
	t.Render()

	return nil
}
