package transport

import (
	"github.com/shijuvar/go-distsys/gokitdemo/services/order"
)

// CreateRequest holds the request parameters for the Create method.
type CreateRequest struct {
	Order order.Order
}

// CreateResponse holds the response values for the Create method.
type CreateResponse struct {
	ID  string `json:"id"`
	Err error  `json:"error,omitempty"`
}

// We have two options to return errors from the business logic.
//
// We could return the error via the endpoint itself. That makes certain things
// a little bit easier, like providing non-200 HTTP responses to the client. But
// Go kit assumes that endpoint errors are (or may be treated as)
// transport-domain errors. For example, an endpoint error will count against a
// circuit breaker error count.
//
// Therefore, it's often better to return service (business logic) errors in the
// response object. This means we have to do a bit more work in the HTTP
// response encoder to detect e.g. a not-found error and provide a proper HTTP
// status code. That work is done with the Errorer interface, in transport.go.
// Response types that may contain business-logic errors implement that
// interface.
func (r CreateResponse) Error() error { return r.Err }

/*
To include business error messages as annotations
in OpenCensus spans we need the Go kit Response
structs to implement the endpoint.Failer interface.
An issue report has been filed so this next step
might become deprecated in the (near) future.
*/

// Failed implements Failer
func (r CreateResponse) Failed() error { return r.Err }

// GetByIDRequest holds the request parameters for the GetByID method.
type GetByIDRequest struct {
	ID string
}

// GetByIDResponse holds the response values for the GetByID method.
type GetByIDResponse struct {
	Order order.Order `json:"order"`
	Err   error       `json:"error,omitempty"`
}

func (r GetByIDResponse) Error() error { return r.Err }

// Failed implements Failer
func (r GetByIDResponse) Failed() error { return r.Err }

// ChangeStatusRequest holds the request parameters for the ChangeStatus method.
type ChangeStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ChangeStatusResponse holds the response values for the ChangeStatus method.
type ChangeStatusResponse struct {
	Err error `json:"error,omitempty"`
}

func (r ChangeStatusResponse) Error() error { return r.Err }

// Failed implements Failer
func (r ChangeStatusResponse) Failed() error { return r.Err }
