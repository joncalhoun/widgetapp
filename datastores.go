// package app defines all of our domain types. Things like our representation
// of a user, a widget, and anything else specific to our domain.
//
// This does not include anything specific to the underlying technology. For
// instance, if we wanted to define a UserService interface that described
// functions we could use to interact with a users database that is fine, but
// we wouldn't add any database specific implementations here.
//
// See https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1 for
// more info on this.
package app

import (
	"errors"
)

var (
	// ErrNotFound is an implementation agnostic error that should be returned
	// by any service implementation when a record was not located.
	ErrNotFound = errors.New("app: resource requested could not be found")
)

// UserService defines the interface used to interact with the users datastore.
// Implementations can be found in packages like the postgres package.
type UserService interface {
	ByEmail(email string) (*User, error)
	ByToken(token string) (*User, error)
	UpdateToken(userID int, token string) error
}

// WidgetService defines the interface used to interact with the widgets
// datastore. Implementations can be found in packages like the postges package.
type WidgetService interface {
	ByUser(userID int) ([]Widget, error)
	Create(widget *Widget) error
}
