package grpc

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/crazy-max/diun/v4/pb"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/pkg/errors"
	"go.podman.io/image/v5/docker/reference"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ImageList(_ context.Context, _ *pb.ImageListRequest) (*pb.ImageListResponse, error) {
	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	var ilr []*pb.ImageListResponse_Image
	for name, mfsts := range images {
		var latest *registry.Manifest
		for _, entry := range mfsts {
			mfst := entry.Manifest
			if latest == nil || mfst.Created.After(*latest.Created) {
				latest = new(mfst)
			}
		}
		ilr = append(ilr, &pb.ImageListResponse_Image{
			Name:           name,
			ManifestsCount: int64(len(mfsts)),
			Latest: &pb.Manifest{
				Tag:      latest.Tag,
				MimeType: latest.MIMEType,
				Digest:   latest.Digest.String(),
				Created:  timestamppb.New(*latest.Created),
				Labels:   latest.Labels,
				Platform: latest.Platform,
			},
		})
	}

	return &pb.ImageListResponse{
		Images: ilr,
	}, nil
}

func (c *Client) ImageInspect(_ context.Context, request *pb.ImageInspectRequest) (*pb.ImageInspectResponse, error) {
	ref, err := reference.ParseNormalizedNamed(request.Name)
	if err != nil {
		return nil, err
	}

	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	if _, ok := images[ref.Name()]; !ok {
		return nil, errors.Errorf("%s not found in database", ref.Name())
	}

	iir := &pb.ImageInspectResponse_Image{
		Name:      ref.Name(),
		Manifests: []*pb.Manifest{},
	}
	for _, entry := range images[ref.Name()] {
		manifest := entry.Manifest
		iir.Manifests = append(iir.Manifests, &pb.Manifest{
			Tag:      manifest.Tag,
			MimeType: manifest.MIMEType,
			Digest:   manifest.Digest.String(),
			Created:  timestamppb.New(*manifest.Created),
			Labels:   manifest.Labels,
			Platform: manifest.Platform,
		})
	}

	return &pb.ImageInspectResponse{
		Image: iir,
	}, nil
}

func (c *Client) ImageRemove(_ context.Context, request *pb.ImageRemoveRequest) (*pb.ImageRemoveResponse, error) {
	ref, err := reference.ParseNormalizedNamed(request.Name)
	if err != nil {
		return nil, err
	}

	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	if _, ok := images[ref.Name()]; !ok {
		return nil, errors.Errorf("%s not found in database", ref.Name())
	}

	_, tagged := ref.(reference.Tagged)
	_, digested := ref.(reference.Digested)

	var removed []*pb.Manifest
	for _, entry := range images[ref.Name()] {
		switch {
		case digested:
			if entry.Reference != ref.String() {
				continue
			}
		case tagged:
			if entry.Reference != ref.String() && !strings.HasPrefix(entry.Reference, ref.String()+"@") {
				continue
			}
		}
		manifest := entry.Manifest
		if err = c.db.DeleteManifest(entry.Reference); err != nil {
			return nil, err
		}
		b, _ := json.Marshal(manifest)
		removed = append(removed, &pb.Manifest{
			Tag:      manifest.Tag,
			MimeType: manifest.MIMEType,
			Digest:   manifest.Digest.String(),
			Created:  timestamppb.New(*manifest.Created),
			Labels:   manifest.Labels,
			Platform: manifest.Platform,
			Size:     int64(len(b)),
		})
	}

	return &pb.ImageRemoveResponse{
		Manifests: removed,
	}, nil
}

func (c *Client) ImagePrune(_ context.Context, _ *pb.ImagePruneRequest) (*pb.ImagePruneResponse, error) {
	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	var removed []*pb.ImagePruneResponse_Image
	for n, m := range images {
		var manifests []*pb.Manifest
		for _, entry := range m {
			manifest := entry.Manifest
			if err = c.db.DeleteManifest(entry.Reference); err != nil {
				return nil, err
			}
			b, _ := json.Marshal(manifest)
			manifests = append(manifests, &pb.Manifest{
				Tag:      manifest.Tag,
				MimeType: manifest.MIMEType,
				Digest:   manifest.Digest.String(),
				Created:  timestamppb.New(*manifest.Created),
				Labels:   manifest.Labels,
				Platform: manifest.Platform,
				Size:     int64(len(b)),
			})
		}
		removed = append(removed, &pb.ImagePruneResponse_Image{
			Name:      n,
			Manifests: manifests,
		})
	}

	return &pb.ImagePruneResponse{
		Images: removed,
	}, nil
}
