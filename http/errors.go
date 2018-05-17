package http

import (
	"fmt"
	"strings"
)

type validationError struct {
	fields  []string
	message string
}

func (e validationError) Error() string {
	return fmt.Sprintf("http validation error: Fields[%s] are not valid. %s", strings.Join(e.fields, ", "), e.message)
}
