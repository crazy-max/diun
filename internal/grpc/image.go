package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containers/image/v5/docker/reference"
	"github.com/crazy-max/diun/v4/pb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ImageList(ctx context.Context, request *pb.ImageListRequest) (*pb.ImageListResponse, error) {
	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	var ilr []*pb.ImageListResponse_Image
	for name, manifests := range images {
		latest := &manifests[0]
		for _, manifest := range manifests {
			if manifest.Created.After(*latest.Created) {
				latest = &manifest
			}
		}
		ilr = append(ilr, &pb.ImageListResponse_Image{
			Name:           name,
			ManifestsCount: int64(len(manifests)),
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

func (c *Client) ImageInspect(ctx context.Context, request *pb.ImageInspectRequest) (*pb.ImageInspectResponse, error) {
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
	for _, manifest := range images[ref.Name()] {
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

func (c *Client) ImageRemove(ctx context.Context, request *pb.ImageRemoveRequest) (*pb.ImageRemoveResponse, error) {
	ref, err := reference.ParseNormalizedNamed(request.Name)
	if err != nil {
		return nil, err
	}

	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	if _, ok := images[ref.Name()]; !ok {
		return nil, fmt.Errorf("%s not found in database", ref.Name())
	}

	var tag string
	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	var removed []*pb.Manifest
	for _, manifest := range images[ref.Name()] {
		if len(tag) == 0 || manifest.Tag == tag {
			if err = c.db.DeleteManifest(manifest); err != nil {
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
	}

	return &pb.ImageRemoveResponse{
		Manifests: removed,
	}, nil
}

func (c *Client) ImagePrune(ctx context.Context, request *pb.ImagePruneRequest) (*pb.ImagePruneResponse, error) {
	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	var removed []*pb.ImagePruneResponse_Image
	for n, m := range images {
		var manifests []*pb.Manifest
		for _, manifest := range m {
			if err = c.db.DeleteManifest(manifest); err != nil {
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
