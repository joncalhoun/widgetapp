package psql

import (
	"database/sql"
	"strings"

	app "github.com/joncalhoun/widgetapp"
)

// UserService is a PostgreSQL specific implementation of the user datastore.
type UserService struct {
	DB *sql.DB
}

// ByEmail will look for a user with the same email address, or return
// app.ErrNotFound if no user is found. Any other errors raised by the sql
// package are passed through.
//
// ByEmail is NOT case sensitive.
func (s *UserService) ByEmail(email string) (*app.User, error) {
	user := app.User{
		Email: strings.ToLower(email),
	}
	row := s.DB.QueryRow(`SELECT id, token FROM users WHERE email=$1;`, user.Email)
	err := row.Scan(&user.ID, &user.Token)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, app.ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// ByToken will look for a user with the same token as provided, or return
// app.ErrNotFound if no user is found. Any other errors raised by the sql
// package are passed through.
//
// ByToken IS case sensitive.
func (s *UserService) ByToken(token string) (*app.User, error) {
	user := app.User{
		Token: token,
	}
	row := s.DB.QueryRow(`SELECT id, email FROM users WHERE token=$1;`, user.Token)
	err := row.Scan(&user.ID, &user.Email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, app.ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// UpdateToken will update the token of the user specified by userID
func (s *UserService) UpdateToken(userID int, token string) error {
	_, err := s.DB.Exec(`UPDATE users SET token=$2 WHERE id=$1`, userID, token)
	return err
}
