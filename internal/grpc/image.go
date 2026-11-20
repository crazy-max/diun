package grpc

import (
	"context"
	"strings"

	"github.com/crazy-max/diun/v4/pb"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/distribution/reference"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ImageList(_ context.Context, _ *pb.ImageListRequest) (*pb.ImageListResponse, error) {
	images, err := c.db.ListImage()
	if err != nil {
		return nil, err
	}

	var ilr []*pb.ImageListResponse_Image
	for name, mfsts := range images {
		latest := mfsts[0].Manifest
		for _, entry := range mfsts[1:] {
			mfst := entry.Manifest
			if mfst.Created != nil && (latest.Created == nil || mfst.Created.After(*latest.Created)) {
				latest = mfst
			}
		}
		ilr = append(ilr, &pb.ImageListResponse_Image{
			Name:           name,
			ManifestsCount: int64(len(mfsts)),
			Latest:         manifestToProto(latest, 0),
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
		iir.Manifests = append(iir.Manifests, manifestToProto(entry.Manifest, 0))
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
		size, err := c.db.DeleteManifest(entry.Reference)
		if err != nil {
			return nil, err
		}
		removed = append(removed, manifestToProto(entry.Manifest, size))
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
			size, err := c.db.DeleteManifest(entry.Reference)
			if err != nil {
				return nil, err
			}
			manifests = append(manifests, manifestToProto(entry.Manifest, size))
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

func manifestToProto(manifest registry.Manifest, size int64) *pb.Manifest {
	result := &pb.Manifest{
		Tag:      manifest.Tag,
		MimeType: manifest.MIMEType,
		Digest:   manifest.Digest.String(),
		Labels:   manifest.Labels,
		Platform: manifest.Platform,
		Size:     size,
	}
	if manifest.Created != nil {
		result.Created = timestamppb.New(*manifest.Created)
	}
	return result
}
