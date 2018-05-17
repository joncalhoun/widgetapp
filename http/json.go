package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	app "github.com/joncalhoun/widgetapp"
	"github.com/joncalhoun/widgetapp/context"
	"golang.org/x/oauth2"
)

func renderJSON(w http.ResponseWriter, data interface{}, status int) {
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

type jsonError struct {
	Message string `json:"error"`
	Type    string `json:"type"`
}

func (e jsonError) Error() string {
	return fmt.Sprintf("json %s error: %s", e.Type, e.Message)
}

type jsonAuthMw struct {
	userService app.UserService
}

func (mw *jsonAuthMw) SetUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if len(bearer) < len("Bearer") {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimSpace(bearer[len("Bearer"):])
		user, err := mw.userService.ByToken(token)
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
func (mw *jsonAuthMw) RequireUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			renderJSON(w, jsonError{
				Message: "Unauthorized access. Do you have a valid oauth2 token set?",
				Type:    "unauthorized",
			}, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func jsonUserHandler(us app.UserService) *UserHandler {
	uh := UserHandler{
		userService: us,

		parseEmailAndPassword: func(r *http.Request) (email, password string) {
			var req struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			dec := json.NewDecoder(r.Body)
			dec.Decode(&req)
			return req.Email, req.Password
		},
		renderProcessSigninSuccess: func(w http.ResponseWriter, r *http.Request, token string) {
			t := oauth2.Token{
				TokenType:   "Bearer",
				AccessToken: token,
			}
			renderJSON(w, t, http.StatusOK)
		},
		renderProcessSigninError: func(w http.ResponseWriter, r *http.Request, err error) {
			switch err {
			case errAuthFailed:
				renderJSON(w, jsonError{
					Message: "Invalid authentication details",
					Type:    "authentication",
				}, http.StatusBadRequest)
			default:
				renderJSON(w, jsonError{
					Message: "Something went wrong. Try again later.",
					Type:    "internal_server",
				}, http.StatusInternalServerError)
			}
		},
	}
	return &uh
}

type jsonWidget struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Price int    `json:"price"`
	// userID is intentionally not here.
}

func (jw *jsonWidget) read(w app.Widget) {
	jw.ID = w.ID
	jw.Name = w.Name
	jw.Color = w.Color
	jw.Price = w.Price
}

func jsonWidgetHandler(ws app.WidgetService) *WidgetHandler {
	wh := WidgetHandler{
		widgetService: ws,

		parseWidget: func(r *http.Request) (*app.Widget, error) {
			var req struct {
				Name  string `json:"name"`
				Color string `json:"color"`
				Price int    `json:"price"`
			}
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&req)
			if err != nil {
				// We could try to parse out the price validation error here if
				// we want, or return a more generic errBadRequest, but I'm going
				// to be lazy for now and just assume this is a validation error
				return nil, validationError{
					fields:  []string{"price"},
					message: "Price must be an integer",
				}
			}
			return &app.Widget{
				Name:  req.Name,
				Color: req.Color,
				Price: req.Price,
			}, nil
		},
		renderCreateSuccess: func(w http.ResponseWriter, r *http.Request, widget *app.Widget) {
			var res jsonWidget
			res.read(*widget)
			renderJSON(w, res, http.StatusCreated)
		},
		renderCreateError: func(w http.ResponseWriter, r *http.Request, err error) {
			switch v := err.(type) {
			case validationError:
				renderJSON(w, struct {
					Fields []string `json:"fields"`
					jsonError
				}{
					Fields: v.fields,
					jsonError: jsonError{
						Message: v.message,
						Type:    "validation",
					},
				}, http.StatusBadRequest)
			default:
				renderJSON(w, jsonError{
					Message: "Something went wrong. Try again later.",
					Type:    "internal_server",
				}, http.StatusInternalServerError)
			}
		},

		renderIndexSuccess: func(w http.ResponseWriter, r *http.Request, widgets []app.Widget) error {
			res := make([]jsonWidget, 0, len(widgets))
			for _, widget := range widgets {
				var jw jsonWidget
				jw.read(widget)
				res = append(res, jw)
			}
			enc := json.NewEncoder(w)
			return enc.Encode(res)
		},
		renderIndexError: func(w http.ResponseWriter, r *http.Request, err error) {
			renderJSON(w, jsonError{
				Message: "Something went wrong. Try again later.",
				Type:    "internal_server",
			}, http.StatusInternalServerError)
		},
	}
	return &wh
}
