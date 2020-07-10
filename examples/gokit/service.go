// service.go
package gokit

import "context"

type Service interface {
	PostUser(ctx context.Context, u User) error
}
type User struct {
	ID        string `json:"id"`
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
}
