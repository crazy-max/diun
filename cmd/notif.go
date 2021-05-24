package main

import (
	"context"
	"fmt"

	"github.com/crazy-max/diun/v4/pb"
)

// NotifCmd holds notif command
type NotifCmd struct {
	Test NotifTestCmd `kong:"cmd,help='Test notification settings.'"`
}

// NotifTestCmd holds notif test command
type NotifTestCmd struct {
	CliGlobals
}

func (s *NotifTestCmd) Run(ctx *Context) error {
	defer s.conn.Close()

	nt, err := s.notifSvc.NotifTest(context.Background(), &pb.NotifTestRequest{})
	if err != nil {
		return err
	}

	fmt.Println(nt.Message)
	return nil
}
