package errors

type ConnectionError struct {
	Err      string
	RawError error
}

func (m *ConnectionError) Error() string {
	return "Connection error :: " + m.Err
}
