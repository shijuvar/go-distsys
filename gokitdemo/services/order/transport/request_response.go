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

// Failed implements Failer
func (r ChangeStatusResponse) Failed() error { return r.Err }
