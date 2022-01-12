package main

import (
	"github.com/crazy-max/diun/v4/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CliHandler is a cli interface
type CliHandler interface {
	BeforeApply() error
}

// CliGlobals holds globals cli attributes
type CliGlobals struct {
	CliHandler `kong:"-"`

	conn     *grpc.ClientConn      `kong:"-"`
	imageSvc pb.ImageServiceClient `kong:"-"`
	notifSvc pb.NotifServiceClient `kong:"-"`

	GRPCAuthority string `kong:"name='grpc-authority',default='127.0.0.1:42286',help='Link to Diun gRPC API.'"`
}

// BeforeApply is a hook that run cli cmd are executed.
func (s *CliGlobals) BeforeApply() (err error) {
	s.conn, err = grpc.Dial(s.GRPCAuthority, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	s.imageSvc = pb.NewImageServiceClient(s.conn)
	s.notifSvc = pb.NewNotifServiceClient(s.conn)
	return
}
