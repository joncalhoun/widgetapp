package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	app "calhoun.io/widgetapp"
	"github.com/gorilla/mux"
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
	db *sql.DB
)

func main() {
	// setup the DB connection
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/signin", http.StatusFound))
	r.HandleFunc("/signin", showSignin).Methods("GET")
	r.HandleFunc("/signin", processSignin).Methods("POST")
	r.HandleFunc("/widgets", allWidgets).Methods("GET")
	r.HandleFunc("/widgets", createWidget).Methods("POST")
	r.HandleFunc("/widgets/new", newWidget).Methods("GET")
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
	row := db.QueryRow(`SELECT id FROM users WHERE token=$1;`, session.Value)
	var userID int
	err = row.Scan(&userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// Email doesn't map to a user in our DB
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
		return
	}

	// Parse form values and validate data (pretend w/ me here)
	name := r.PostFormValue("name")
	color := r.PostFormValue("color")
	price, err := strconv.Atoi(r.PostFormValue("price"))
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}
	if color == "Green" && price%2 != 0 {
		http.Error(w, "Price must be even with a color of Green", http.StatusBadRequest)
		return
	}

	// Create a new widget!
	_, err = db.Exec(`INSERT INTO widgets(userID, name, price, color) VALUES($1, $2, $3, $4)`, userID, name, price, color)
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
	row := db.QueryRow(`SELECT id FROM users WHERE token=$1;`, session.Value)
	var userID int
	err = row.Scan(&userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// Email doesn't map to a user in our DB
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
		return
	}

	// Query for this user's widgets
	rows, err := db.Query(`SELECT id, name, price, color FROM widgets WHERE userID=$1`, userID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var widgets []app.Widget
	for rows.Next() {
		var widget app.Widget
		err = rows.Scan(&widget.ID, &widget.Name, &widget.Price, &widget.Color)
		if err != nil {
			log.Printf("Failed to scan a widget: %v", err)
			continue
		}
		widgets = append(widgets, widget)
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
	email = strings.ToLower(email)
	row := db.QueryRow(`SELECT id FROM users WHERE email=$1;`, email)
	var id int
	err := row.Scan(&id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// Email doesn't map to a user in our DB
			http.Redirect(w, r, "/signin", http.StatusFound)
		default:
			http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		}
		return
	}

	// Create a fake session token
	token := fmt.Sprintf("fake-session-id-%d", id)
	_, err = db.Exec(`UPDATE users SET token=$2 WHERE id=$1`, id, token)
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
