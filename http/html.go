package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	app "github.com/joncalhoun/widgetapp"
	"github.com/joncalhoun/widgetapp/context"
)

type htmlAuthMw struct {
	userService app.UserService
}

// SetUser will retrieve the current user set by the session cookie
// and set it in the request context. UserViaSession will NOT redirect
// to the sign in page if the user is not found. That is left for the
// RequireUser method to handle so that some pages can optionally have
// access to the current user.
func (mw *htmlAuthMw) SetUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")
		if err != nil {
			// No user found - time to move on
			next.ServeHTTP(w, r)
			return
		}
		user, err := mw.userService.ByToken(session.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r = r.WithContext(context.WithUser(r.Context(), user))
		next.ServeHTTP(w, r)
	}
}

// RequireUser will verify that a user is set in the request context. It if is
// set correctly, the next handler will be called, otherwise it will redirect
// the user to the sign in page and the next handler will not be called.
func (mw *htmlAuthMw) RequireUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func htmlUserHandler(us app.UserService) *UserHandler {
	uh := UserHandler{
		userService: us,

		renderSignin: func(w http.ResponseWriter) {
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
		},

		parseEmailAndPassword: func(r *http.Request) (email, password string) {
			email = r.PostFormValue("email")
			password = r.PostFormValue("password")
			return email, password
		},
		renderProcessSigninSuccess: func(w http.ResponseWriter, r *http.Request, token string) {
			cookie := http.Cookie{
				Name:  "session",
				Value: token,
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/widgets", http.StatusFound)
		},
		renderProcessSigninError: func(w http.ResponseWriter, r *http.Request, err error) {
			switch err {
			case errAuthFailed:
				http.Redirect(w, r, "/signin", http.StatusFound)
			default:
				http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
			}
		},
	}

	return &uh
}

func htmlWidgetHandler(ws app.WidgetService) *WidgetHandler {
	wh := WidgetHandler{
		widgetService: ws,

		renderNew: func(w http.ResponseWriter) {
			html := `
				<!DOCTYPE html>
				<html lang="en">
				<form action="/widgets" method="POST">
					<label for="name">Name</label>
					<input type="text" id="name" name="name" placeholder="Stop Widget">

					<label for="color">Color</label>
					<input type="text" id="color" name="color" placeholder="Red">

					<label for="price">Price</label>
					<input type="number" id="price" name="price" placeholder="18">

					<button type="submit">Create it!</button>
				</form>
				</html>`
			fmt.Fprint(w, html)
		},

		parseWidget: func(r *http.Request) (*app.Widget, error) {
			// Parse form values and validate data (pretend w/ me here)
			widget := app.Widget{
				Name:  r.PostFormValue("name"),
				Color: r.PostFormValue("color"),
			}
			var err error
			widget.Price, err = strconv.Atoi(r.PostFormValue("price"))
			if err != nil {
				return nil, validationError{
					fields:  []string{"price"},
					message: "Price must be an integer",
				}
			}
			return &widget, nil
		},
		renderCreateSuccess: func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/widgets", http.StatusFound)
		},
		renderCreateError: func(w http.ResponseWriter, r *http.Request, err error) {
			switch v := err.(type) {
			case validationError:
				http.Error(w, v.message, http.StatusBadRequest)
			default:
				http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
			}
		},

		renderIndexSuccess: func(w http.ResponseWriter, r *http.Request, widgets []app.Widget) error {
			// Render the widgets
			tplStr := `
				<!DOCTYPE html>
				<html lang="en">
				<h1>Widgets</h1>
				<ul>
				{{range .}}
					<li>{{.Name}} - {{.Color}}: <b>${{.Price}}</b></li>
				{{end}}
				</ul>
				<p>
					<a href="/widgets/new">Create a new widget</a>
				</p>
				</html>`
			tpl := template.Must(template.New("").Parse(tplStr))
			return tpl.Execute(w, widgets)
		},
		renderIndexError: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		},
	}
	return &wh
}
