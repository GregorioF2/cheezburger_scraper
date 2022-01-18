package errors

type BadRequestError struct {
	Err      string
	RawError error
}

func (m *BadRequestError) Error() string {
	return "Bad request error :: " + m.Err
}
