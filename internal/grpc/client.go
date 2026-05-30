package grpc

import (
	"context"
	"net"

	"github.com/crazy-max/diun/v4/internal/db"
	grpclogger "github.com/crazy-max/diun/v4/internal/grpc/logger"
	"github.com/crazy-max/diun/v4/internal/notif"
	"github.com/crazy-max/diun/v4/pb"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	HealthServiceGRPC      = "diun.grpc"
	HealthServiceScheduler = "diun.scheduler"
	HealthServiceMetrics   = "diun.metrics"
)

// Client represents an active grpc object
type Client struct {
	server    *grpc.Server
	health    *grpchealth.Server
	authority string
	db        *db.Client
	notif     *notif.Client
	pb.UnimplementedImageServiceServer
	pb.UnimplementedNotifServiceServer
}

// New creates a new grpc instance
func New(authority string, db *db.Client, notif *notif.Client) (*Client, error) {
	grpclogger.SetGrpcLogger(log.Level(zerolog.ErrorLevel))

	c := &Client{
		authority: authority,
		db:        db,
		notif:     notif,
	}

	c.server = grpc.NewServer()
	c.health = grpchealth.NewServer()
	pb.RegisterImageServiceServer(c.server, c)
	pb.RegisterNotifServiceServer(c.server, c)
	healthpb.RegisterHealthServer(c.server, c.health)
	c.SetHealthStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	c.SetHealthStatus(HealthServiceGRPC, healthpb.HealthCheckResponse_NOT_SERVING)
	c.SetHealthStatus(HealthServiceScheduler, healthpb.HealthCheckResponse_SERVICE_UNKNOWN)
	c.SetHealthStatus(HealthServiceMetrics, healthpb.HealthCheckResponse_SERVICE_UNKNOWN)

	return c, nil
}

// Start runs the grpc server
func (c *Client) Start() error {
	lis, err := c.Listen()
	if err != nil {
		return err
	}
	return c.Serve(lis)
}

// Listen creates the gRPC listener
func (c *Client) Listen() (net.Listener, error) {
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", c.authority)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create gRPC listener")
	}
	return lis, nil
}

// Serve serves gRPC requests on listener
func (c *Client) Serve(lis net.Listener) error {
	log.Info().Str("addr", lis.Addr().String()).Msg("gRPC server listening")

	if err := c.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
}

// Stop stops the grpc server
func (c *Client) Stop() {
	c.SetHealthStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	c.SetHealthStatus(HealthServiceGRPC, healthpb.HealthCheckResponse_NOT_SERVING)
	c.server.GracefulStop()
}

// SetHealthStatus updates the status returned by the gRPC health service.
func (c *Client) SetHealthStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	c.health.SetServingStatus(service, status)
}
