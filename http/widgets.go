package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	app "github.com/joncalhoun/widgetapp"
	"github.com/joncalhoun/widgetapp/context"
)

// WidgetHandler contains user-specific http.HandlerFuncs as
// methods.
type WidgetHandler struct {
	widgetService app.WidgetService
}

// New renders a form for creating a new widget.
func (h *WidgetHandler) New(w http.ResponseWriter, r *http.Request) {
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

// Create processes the request and creates a new widget.
func (h *WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())

	// Parse form values and validate data (pretend w/ me here)
	widget := app.Widget{
		UserID: user.ID,
		Name:   r.PostFormValue("name"),
		Color:  r.PostFormValue("color"),
	}
	var err error
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
	err = h.widgetService.Create(&widget)
	if err != nil {
		http.Error(w, "Something went wrong. Try again later.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/widgets", http.StatusFound)
}

// Index lists all of a users widgets.
func (h *WidgetHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())

	// Query for this user's widgets
	widgets, err := h.widgetService.ByUser(user.ID)
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
