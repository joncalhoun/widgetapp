package http

import (
	"errors"
	"fmt"
	"net/http"

	app "github.com/joncalhoun/widgetapp"
)

var (
	errAuthFailed = errors.New("http: authentication failed")
)

// UserHandler contains user-specific http.HandlerFuncs as
// methods.
type UserHandler struct {
	userService app.UserService

	renderSignin func(http.ResponseWriter)

	parseEmailAndPassword      func(*http.Request) (email, password string)
	renderProcessSigninSuccess func(http.ResponseWriter, *http.Request, string)
	renderProcessSigninError   func(http.ResponseWriter, *http.Request, error)
}

// ShowSignin will render the sign in form
func (h *UserHandler) ShowSignin(w http.ResponseWriter, r *http.Request) {
	h.renderSignin(w)
}

// ProcessSignin will process the signin form
func (h *UserHandler) ProcessSignin(w http.ResponseWriter, r *http.Request) {
	email, password := h.parseEmailAndPassword(r)
	// Fake the password part
	if password != "demo" {
		h.renderProcessSigninError(w, r, errAuthFailed)
		return
	}

	// Lookup the user ID by their email in the DB
	user, err := h.userService.ByEmail(email)
	if err != nil {
		switch err {
		case app.ErrNotFound:
			h.renderProcessSigninError(w, r, errAuthFailed)
		default:
			h.renderProcessSigninError(w, r, err)
		}
		return
	}

	// Create a fake session token
	token := fmt.Sprintf("fake-session-id-%d", user.ID)
	err = h.userService.UpdateToken(user.ID, token)
	if err != nil {
		h.renderProcessSigninError(w, r, err)
		return
	}
	h.renderProcessSigninSuccess(w, r, token)
}
