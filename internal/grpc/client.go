package grpc

import (
	"net"

	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client represents an active grpc object
type Client struct {
	server    *grpc.Server
	authority string
	db        *db.Client
	pb.UnimplementedImageServiceServer
}

// New creates a new grpc instance
func New(authority string, db *db.Client) (*Client, error) {
	c := &Client{
		authority: authority,
		db:        db,
	}

	c.server = grpc.NewServer()
	pb.RegisterImageServiceServer(c.server, c)

	return c, nil
}

// Start runs the grpc server
func (c *Client) Start() error {
	var err error

	lis, err := net.Listen("tcp", c.authority)
	if err != nil {
		return errors.Wrap(err, "Cannot create gRPC listener")
	}

	return c.server.Serve(lis)
}

// Stop stops the grpc server
func (c *Client) Stop() {
	c.server.GracefulStop()
}
