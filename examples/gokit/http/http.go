package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shijuvar/go-distsys/examples/gokit/endpoints"

	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(endpoints endpoints.Endpoints, options ...httptransport.ServerOption) http.Handler {
	m := http.NewServeMux()
	m.Handle("/postuser", httptransport.NewServer(endpoints.PostUser, DecodePostUserRequest, EncodePostUserResponse, options...))
	return m
}
func DecodePostUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.PostUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodePostUserResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
