package errors

import (
	"fmt"
)

type Error struct {
	Err        string
	StatusCode int
}

func NewError(err string, statusCode int) *Error {
	return &Error{
		err,
		statusCode,
	}
}

func ChainNewError(err *Error, prefix string) *Error {
	err.Err = fmt.Sprintf("%s\n%s", prefix, err.Err)
	return err
}
