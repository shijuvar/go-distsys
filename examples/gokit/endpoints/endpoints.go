package endpoints

import (
	"context"
	service "github.com/shijuvar/go-distsys/examples/gokit"

	"github.com/go-kit/kit/endpoint"
)

type PostUserRequest struct {
	U service.User
}
type PostUserResponse struct {
	Err error
}

func MakePostUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PostUserRequest)
		err := s.PostUser(ctx, req.U)
		return PostUserResponse{Err: err}, nil
	}
}

type Endpoints struct {
	PostUser endpoint.Endpoint
}

func (r PostUserResponse) Failed() error { return r.Err }
