package errors

type NotFoundError struct {
	Err      string
	RawError error
}

func (m *NotFoundError) Error() string {
	return "Not found error :: " + m.Err
}
