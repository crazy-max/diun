package main

import (
	"context"
	"fmt"

	"github.com/crazy-max/diun/v4/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HealthcheckCmd holds healthcheck command
type HealthcheckerCmd struct {
	Test HealthcheckCmd `cmd:"" help:"Run a healthcheck."`
}

// HealthcheckCmd holds healthcheck test command
type HealthcheckCmd struct {
	GRPCAuthority string `name:"grpc-authority" default:"127.0.0.1:42286" help:"Link to Diun gRPC server."`
}

func (s *HealthcheckCmd) Run(_ *Context) error {
	conn, err := grpc.NewClient(s.GRPCAuthority, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	healthcheckSvc := pb.NewHealthcheckServiceClient(conn)

	nt, err := healthcheckSvc.Healthcheck(context.Background(), &pb.HealthcheckRequest{})
	if err != nil {
		return err
	}

	fmt.Println(nt.Message)
	return nil
}
