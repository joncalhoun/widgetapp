package http

import (
	"fmt"
	"net/http"

	app "github.com/joncalhoun/widgetapp"
)

// UserHandler contains user-specific http.HandlerFuncs as
// methods.
type UserHandler struct {
	userService app.UserService
}

// ShowSignin will render the sign in form
func (h *UserHandler) ShowSignin(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="en">
<form action="/signin" method="POST">
	<label for="email">Email Address</label>
	<input type="email" id="email" name="email" placeholder="you@example.com">

	<label for="password">Password</label>
	<input type="password" id="password" name="password" placeholder="something-secret">

	<button type="submit">Sign in</button>
</form>
</html>`
	fmt.Fprint(w, html)
}

// ProcessSignin will process the signin form
func (h *UserHandler) ProcessSignin(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	// Fake the password part
	if password != "demo" {
		http.Redirect(w, r, "/signin", http.StatusNotFound)
		return
	}

	// Lookup the user ID by their email in the DB
	user, err := h.userService.ByEmail(email)
	if err != nil {
		switch err {
		case app.ErrNotFound:
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
	}

	// Create a fake session token
	token := fmt.Sprintf("fake-session-id-%d", user.ID)
	err = h.userService.UpdateToken(user.ID, token)
	if err != nil {
		http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:  "session",
		Value: token,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/widgets", http.StatusFound)
}
