package errors

type InternalServerError struct {
	Err      string
	RawError error
}

func (m *InternalServerError) Error() string {
	return "Internal server error :: " + m.Err
}
