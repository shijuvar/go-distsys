package http

import (
	"net/http"

	"github.com/go-kit/kit/log"
	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/shijuvar/go-distsys/gokitdemo/services/order/transport"
)

// NewService wires Go kit endpoints to the HTTP transport.
func NewService(
	svcEndpoints transport.Endpoints, options []kithttp.ServerOption, logger log.Logger,
) http.Handler {
	errorLogger := kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger))
	errorEncoder := kithttp.ServerErrorEncoder(encodeErrorResponse)
	options = append(options, errorLogger, errorEncoder)
	// Configure HTTP request routes with Go kit endpoints
	handler := initializeRoutes(svcEndpoints, options)
	return handler
}
