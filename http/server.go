package http

import (
	"net/http"

	"github.com/gorilla/mux"
	app "github.com/joncalhoun/widgetapp"
)

// NewServer returns an http.Handler for a server that has both a JSON API
// and the HTML server.
func NewServer(us app.UserService, ws app.WidgetService) http.Handler {
	html := HTMLServer(us, ws)
	json := JSONServer(us, ws)
	mux := http.NewServeMux()
	mux.Handle("/", html)
	mux.Handle("/api/", http.StripPrefix("/api", json))
	return mux
}

// HTMLServer will construct a server, apply all of the necessary routes,
// then return an http.Handler for the server.
func HTMLServer(us app.UserService, ws app.WidgetService) http.Handler {
	s := server{
		authMw: &htmlAuthMw{
			userService: us,
		},
		users:   htmlUserHandler(us),
		widgets: htmlWidgetHandler(ws),
		router:  mux.NewRouter(),
	}
	s.routes(true)
	return &s
}

// JSONServer will construct a server, apply all of the necessary routes,
// then return an http.Handler for the server.
func JSONServer(us app.UserService, ws app.WidgetService) http.Handler {
	s := server{
		authMw: &jsonAuthMw{
			userService: us,
		},
		users:   jsonUserHandler(us),
		widgets: jsonWidgetHandler(ws),
		router:  mux.NewRouter(),
	}
	s.routes(false)
	return &s
}

// server is our HTTP server with routes for all our endpoints.
type server struct {
	// unexported types - do not use the zero value
	authMw  AuthMw
	users   *UserHandler
	widgets *WidgetHandler
	router  *mux.Router
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) routes(webMode bool) {
	if webMode {
		s.router.Handle("/", http.RedirectHandler("/signin", http.StatusFound))
	}

	// User routes
	if webMode {
		s.router.HandleFunc("/signin", s.users.ShowSignin).Methods("GET")
	}
	s.router.HandleFunc("/signin", s.users.ProcessSignin).Methods("POST")

	// Widget routes
	s.router.Handle("/widgets", ApplyMwFn(s.widgets.Index,
		s.authMw.SetUser, s.authMw.RequireUser)).Methods("GET")
	s.router.Handle("/widgets", ApplyMwFn(s.widgets.Create,
		s.authMw.SetUser, s.authMw.RequireUser)).Methods("POST")
	if webMode {
		s.router.Handle("/widgets/new", ApplyMwFn(s.widgets.New,
			s.authMw.SetUser, s.authMw.RequireUser)).Methods("GET")
	}
}
