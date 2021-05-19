package grpc

import (
	"context"

	"github.com/crazy-max/diun/v4/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ImageList(ctx context.Context, request *pb.ImageListRequest) (*pb.ImageListResponse, error) {
	manifests, err := c.db.ListImageLatest()
	if err != nil {
		return nil, err
	}

	var images []*pb.ImageListResponse_Image
	for _, image := range manifests {
		images = append(images, &pb.ImageListResponse_Image{
			Name:        image.Name,
			LastTag:     image.Tag,
			LastUpdated: timestamppb.New(*image.Created),
			LastDigest:  image.Digest.String(),
		})
	}

	return &pb.ImageListResponse{
		Image: images,
	}, nil
}

func (c *Client) ImageInspect(ctx context.Context, request *pb.ImageInspectRequest) (*pb.ImageInspectResponse, error) {
	panic("implement me")
}

func (c *Client) ImageRemove(ctx context.Context, request *pb.ImageRemoveRequest) (*pb.ImageRemoveResponse, error) {
	panic("implement me")
}
