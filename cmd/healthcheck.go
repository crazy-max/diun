package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	diungrpc "github.com/crazy-max/diun/v4/internal/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	healthcheckStatusHealthy   = "healthy"
	healthcheckStatusUnhealthy = "unhealthy"
	healthcheckStatusDisabled  = "disabled"
	healthcheckStatusUnknown   = "unknown"
)

var (
	errHealthcheckFailed   = errors.New("healthcheck failed")
	errHealthcheckTimedOut = errors.New("healthcheck timed out")
)

// HealthcheckCmd holds healthcheck command args and flags.
type HealthcheckCmd struct {
	Raw           bool          `name:"raw" default:"false" help:"JSON output."`
	Timeout       time.Duration `name:"timeout" default:"3s" help:"Timeout for healthcheck requests."`
	GRPCAuthority string        `name:"grpc-authority" default:"127.0.0.1:42286" help:"Link to Diun gRPC server."`
}

type healthcheckOutput struct {
	Status   string               `json:"status"`
	Services []healthcheckService `json:"services"`
}

type healthcheckService struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

var healthcheckServices = []struct {
	name    string
	service string
}{
	{name: "grpc", service: diungrpc.HealthServiceGRPC},
	{name: "scheduler", service: diungrpc.HealthServiceScheduler},
	{name: "metrics", service: diungrpc.HealthServiceMetrics},
}

func (s *HealthcheckCmd) Run(_ *Context) error {
	ctx, cancel := context.WithTimeoutCause(context.Background(), s.Timeout, errHealthcheckTimedOut)
	defer cancel()

	output := checkDiunHealth(ctx, s.GRPCAuthority)
	if s.Raw {
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))
	} else {
		printHealthcheckOutput(output)
	}

	if output.Status != healthcheckStatusHealthy {
		return errHealthcheckFailed
	}
	return nil
}

func checkDiunHealth(ctx context.Context, grpcAuthority string) healthcheckOutput {
	conn, err := grpc.NewClient(grpcAuthority, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return unhealthyHealthcheckOutput(err.Error())
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)
	overall, err := checkHealthService(ctx, client, "")
	if err != nil {
		return unhealthyHealthcheckOutput(err.Error())
	}

	output := healthcheckOutput{
		Status:   healthStatus(overall),
		Services: make([]healthcheckService, 0, len(healthcheckServices)),
	}
	if output.Status != healthcheckStatusHealthy {
		output.Status = healthcheckStatusUnhealthy
	}

	for _, svc := range healthcheckServices {
		respStatus, err := checkHealthService(ctx, client, svc.service)
		if err != nil {
			output.Services = append(output.Services, healthcheckService{
				Name:    svc.name,
				Status:  healthcheckStatusUnhealthy,
				Message: err.Error(),
			})
			output.Status = healthcheckStatusUnhealthy
			continue
		}

		serviceStatus := healthStatus(respStatus)
		service := healthcheckService{
			Name:   svc.name,
			Status: serviceStatus,
		}
		switch serviceStatus {
		case healthcheckStatusDisabled:
			service.Message = "disabled by configuration"
		case healthcheckStatusUnhealthy:
			service.Message = "not serving"
			output.Status = healthcheckStatusUnhealthy
		case healthcheckStatusUnknown:
			service.Message = strings.ToLower(respStatus.String())
			output.Status = healthcheckStatusUnhealthy
		}
		output.Services = append(output.Services, service)
	}

	return output
}

func checkHealthService(ctx context.Context, client healthpb.HealthClient, service string) (healthpb.HealthCheckResponse_ServingStatus, error) {
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{
		Service: service,
	}, grpc.WaitForReady(true))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return healthpb.HealthCheckResponse_SERVICE_UNKNOWN, nil
		}
		return healthpb.HealthCheckResponse_UNKNOWN, err
	}
	return resp.Status, nil
}

func healthStatus(status healthpb.HealthCheckResponse_ServingStatus) string {
	switch status {
	case healthpb.HealthCheckResponse_SERVING:
		return healthcheckStatusHealthy
	case healthpb.HealthCheckResponse_NOT_SERVING:
		return healthcheckStatusUnhealthy
	case healthpb.HealthCheckResponse_SERVICE_UNKNOWN:
		return healthcheckStatusDisabled
	default:
		return healthcheckStatusUnknown
	}
}

func unhealthyHealthcheckOutput(message string) healthcheckOutput {
	return healthcheckOutput{
		Status: healthcheckStatusUnhealthy,
		Services: []healthcheckService{
			{
				Name:    "grpc",
				Status:  healthcheckStatusUnhealthy,
				Message: message,
			},
		},
	}
}

func printHealthcheckOutput(output healthcheckOutput) {
	fmt.Printf("Diun is %s\n", output.Status)
	for _, service := range output.Services {
		if service.Message != "" {
			fmt.Printf("%s: %s (%s)\n", service.Name, service.Status, service.Message)
			continue
		}
		fmt.Printf("%s: %s\n", service.Name, service.Status)
	}
}
