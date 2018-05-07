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

// Widget represents the widget that we create with our app.
type Widget struct {
	ID     int
	UserID int
	Name   string
	Price  int
	Color  string
}

// User represents a user in our system.
type User struct {
	ID    int
	Email string
	Token string
}
