package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	diungrpc "github.com/crazy-max/diun/v4/internal/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckDiunHealthReportsServiceStatuses(t *testing.T) {
	healthServer := grpchealth.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceGRPC, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceScheduler, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceMetrics, healthpb.HealthCheckResponse_SERVICE_UNKNOWN)

	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Second, errors.New("healthcheck test timed out"))
	defer cancel()

	output := checkDiunHealth(ctx, startHealthServer(t, healthServer))

	assert.Equal(t, healthcheckStatusHealthy, output.Status)
	assert.Equal(t, []healthcheckService{
		{Name: "grpc", Status: healthcheckStatusHealthy},
		{Name: "scheduler", Status: healthcheckStatusHealthy},
		{Name: "metrics", Status: healthcheckStatusDisabled, Message: "disabled by configuration"},
	}, output.Services)
}

func TestCheckDiunHealthReportsUnhealthyStatus(t *testing.T) {
	healthServer := grpchealth.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceGRPC, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceScheduler, healthpb.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus(diungrpc.HealthServiceMetrics, healthpb.HealthCheckResponse_SERVICE_UNKNOWN)

	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Second, errors.New("healthcheck test timed out"))
	defer cancel()

	output := checkDiunHealth(ctx, startHealthServer(t, healthServer))

	assert.Equal(t, healthcheckStatusUnhealthy, output.Status)
	assert.Equal(t, []healthcheckService{
		{Name: "grpc", Status: healthcheckStatusHealthy},
		{Name: "scheduler", Status: healthcheckStatusUnhealthy, Message: "not serving"},
		{Name: "metrics", Status: healthcheckStatusDisabled, Message: "disabled by configuration"},
	}, output.Services)
}

func startHealthServer(t *testing.T, healthServer *grpchealth.Server) string {
	t.Helper()

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(lis)
	}()

	t.Cleanup(func() {
		server.Stop()
		if err := <-errCh; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("health server failed: %v", err)
		}
	})

	return lis.Addr().String()
}
