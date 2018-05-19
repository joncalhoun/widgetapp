package context

import (
	"context"
	"log"

	app "github.com/joncalhoun/widgetapp"
)

type ctxKey string

const (
	userKey ctxKey = "user"
)

// WithUser will return a new context with the user value added to it.
func WithUser(ctx context.Context, user *app.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// User will return the user assigned to the context, or nil if there
// is any error or there isn't a user.
func User(ctx context.Context) *app.User {
	tmp := ctx.Value(userKey)
	if tmp == nil {
		return nil
	}
	user, ok := tmp.(*app.User)
	if !ok {
		// This shouldn't happen. If it does, it means we have a bug in our code.
		log.Printf("context: user value set incorrectly. type=%T, value=%#v", tmp, tmp)
		return nil
	}
	return user
}
