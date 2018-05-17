package http

import (
	"net/http"

	app "github.com/joncalhoun/widgetapp"
	"github.com/joncalhoun/widgetapp/context"
)

// WidgetHandler contains user-specific http.HandlerFuncs as
// methods.
type WidgetHandler struct {
	widgetService app.WidgetService

	renderNew func(http.ResponseWriter)

	parseWidget         func(*http.Request) (*app.Widget, error)
	renderCreateSuccess func(http.ResponseWriter, *http.Request, *app.Widget)
	renderCreateError   func(http.ResponseWriter, *http.Request, error)

	renderIndexSuccess func(http.ResponseWriter, *http.Request, []app.Widget) error
	renderIndexError   func(http.ResponseWriter, *http.Request, error)
}

// New renders a form for creating a new widget.
func (h *WidgetHandler) New(w http.ResponseWriter, r *http.Request) {
	h.renderNew(w)
}

// Create processes the request and creates a new widget.
func (h *WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	widget, err := h.parseWidget(r)
	if err != nil {
		h.renderCreateError(w, r, err)
		return
	}
	widget.UserID = user.ID
	if widget.Color == "Green" && widget.Price%2 != 0 {
		h.renderCreateError(w, r, validationError{
			fields:  []string{"price", "color"},
			message: "Price must be even with a color of Green",
		})
		return
	}

	// Create a new widget!
	err = h.widgetService.Create(widget)
	if err != nil {
		h.renderCreateError(w, r, err)
		return
	}
	h.renderCreateSuccess(w, r, widget)
}

// Index lists all of a users widgets.
func (h *WidgetHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())

	// Query for this user's widgets
	widgets, err := h.widgetService.ByUser(user.ID)
	if err != nil {
		h.renderIndexError(w, r, err)
		return
	}
	err = h.renderIndexSuccess(w, r, widgets)
	if err != nil {
		h.renderIndexError(w, r, err)
		return
	}
}
