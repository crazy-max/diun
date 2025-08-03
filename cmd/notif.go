package main

import (
	"context"
	"fmt"

	"github.com/crazy-max/diun/v4/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotifCmd holds notif command
type NotifCmd struct {
	Test NotifTestCmd `cmd:"" help:"Test notification settings."`
}

// NotifTestCmd holds notif test command
type NotifTestCmd struct {
	GRPCAuthority string `name:"grpc-authority" default:"127.0.0.1:42286" help:"Link to Diun gRPC server."`
}

func (s *NotifTestCmd) Run(_ *Context) error {
	conn, err := grpc.NewClient(s.GRPCAuthority, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	notifSvc := pb.NewNotifServiceClient(conn)

	nt, err := notifSvc.NotifTest(context.Background(), &pb.NotifTestRequest{})
	if err != nil {
		return err
	}

	fmt.Println(nt.Message)
	return nil
}
