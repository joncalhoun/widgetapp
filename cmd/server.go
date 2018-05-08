package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	app "github.com/joncalhoun/widgetapp"
	"github.com/joncalhoun/widgetapp/mw"
	"github.com/joncalhoun/widgetapp/postgres"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "your-password"
	dbname   = "widget_demo"
)

var (
	userService   app.UserService
	widgetService app.WidgetService
)

func main() {
	// setup the DB connection
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	userService = &postgres.UserService{DB: db}
	widgetService = &postgres.WidgetService{DB: db}
	authMw := mw.Auth{
		UserService: userService,
	}

	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/signin", http.StatusFound))
	r.HandleFunc("/signin", showSignin).Methods("GET")
	r.HandleFunc("/signin", processSignin).Methods("POST")
	r.Handle("/widgets", mw.ApplyFn(allWidgets,
		authMw.UserViaSession, authMw.RequireUser)).Methods("GET")
	r.Handle("/widgets", mw.ApplyFn(createWidget,
		authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	r.Handle("/widgets/new", mw.ApplyFn(newWidget,
		authMw.UserViaSession, authMw.RequireUser)).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func newWidget(w http.ResponseWriter, r *http.Request) {
	// Ignore auth for now - do it on the POST
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
}

func createWidget(w http.ResponseWriter, r *http.Request) {
	// Verify the user is signed in
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := userService.ByToken(session.Value)
	if err != nil {
		switch err {
		case app.ErrNotFound:
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
	}

	// Parse form values and validate data (pretend w/ me here)
	widget := app.Widget{
		UserID: user.ID,
		Name:   r.PostFormValue("name"),
		Color:  r.PostFormValue("color"),
	}
	widget.Price, err = strconv.Atoi(r.PostFormValue("price"))
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}
	if widget.Color == "Green" && widget.Price%2 != 0 {
		http.Error(w, "Price must be even with a color of Green", http.StatusBadRequest)
		return
	}

	// Create a new widget!
	err = widgetService.Create(&widget)
	if err != nil {
		http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/widgets", http.StatusFound)
}

func allWidgets(w http.ResponseWriter, r *http.Request) {
	// Verify the user is signed in
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := userService.ByToken(session.Value)
	if err != nil {
		switch err {
		case app.ErrNotFound:
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
	}

	// Query for this user's widgets
	widgets, err := widgetService.ByUser(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

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
	err = tpl.Execute(w, widgets)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
}

func showSignin(w http.ResponseWriter, r *http.Request) {
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

func processSignin(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	// Fake the password part
	if password != "demo" {
		http.Redirect(w, r, "/signin", http.StatusNotFound)
		return
	}

	// Lookup the user ID by their email in the DB
	user, err := userService.ByEmail(email)
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
	err = userService.UpdateToken(user.ID, token)
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
