package http

import (
	"net/http"
)

// AuthMw provides middleware functions for authorizing users and setting the user
// in the request context.
type AuthMw interface {
	SetUser(next http.Handler) http.HandlerFunc
	RequireUser(next http.Handler) http.HandlerFunc
}
