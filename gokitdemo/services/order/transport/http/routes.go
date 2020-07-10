package http

import (
	"errors"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	"github.com/shijuvar/go-distsys/gokitdemo/services/order/transport"
)

var (
	ErrBadRouting = errors.New("bad routing")
)

// initializeRoutes maps the Go kit endpoints
// to endpoints of HTTP request router
func initializeRoutes(svcEndpoints transport.Endpoints, options []kithttp.ServerOption) *mux.Router {
	r := mux.NewRouter()
	// HTTP Post - /orders
	r.Methods("POST").Path("/orders").Handler(kithttp.NewServer(
		svcEndpoints.Create,
		decodeCreateRequest,
		encodeResponse,
		options...,
	))

	// HTTP Post - /orders/{id}
	r.Methods("GET").Path("/orders/{id}").Handler(kithttp.NewServer(
		svcEndpoints.GetByID,
		decodeGetByIDRequest,
		encodeResponse,
		options...,
	))

	// HTTP Post - /orders/status
	r.Methods("POST").Path("/orders/status").Handler(kithttp.NewServer(
		svcEndpoints.ChangeStatus,
		decodeChangeStatusRequest,
		encodeResponse,
		options...,
	))
	return r
}
