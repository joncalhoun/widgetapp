package postgres

import (
	"database/sql"
	"log"

	app "github.com/joncalhoun/widgetapp"
)

// WidgetService is a PostgreSQL specific implementation of the widget datastore.
type WidgetService struct {
	DB *sql.DB
}

// ByUser will retrieve all widgets with the specified userID.
//
// Long term this may need to support pagination of some sort, but short term
// this is fine.
func (s *WidgetService) ByUser(userID int) ([]app.Widget, error) {
	rows, err := s.DB.Query(`SELECT id, name, price, color FROM widgets WHERE userID=$1`, userID)
	if err != nil {
		return nil, err
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
	// We forgot to check this before
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return widgets, nil
}

// Create will create a the widget provided.
func (s *WidgetService) Create(widget *app.Widget) error {
	_, err := s.DB.Exec(`INSERT INTO widgets(userID, name, price, color) VALUES($1, $2, $3, $4)`, widget.UserID, widget.Name, widget.Price, widget.Color)
	return err
}
