package errors

import "fmt"

type ResponseError struct {
	Err        string
	StatusCode int
}

func (m *ResponseError) Error() string {
	return fmt.Sprintf("Error: %s", m.Err)
}
