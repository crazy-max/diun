package grpc

import (
	"context"

	"github.com/crazy-max/diun/v4/pb"
)

func (c *Client) Healthcheck(_ context.Context, _ *pb.HealthcheckRequest) (*pb.HealthcheckResponse, error) {
	return &pb.HealthcheckResponse{
		Message: "Diun is running",
	}, nil
}
