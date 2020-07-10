package service

import (
	"context"

	svc "github.com/shijuvar/go-distsys/examples/gokit"
)

type Service struct {
}

func (s Service) PostUser(ctx context.Context, u svc.User) error {
	return nil
}
