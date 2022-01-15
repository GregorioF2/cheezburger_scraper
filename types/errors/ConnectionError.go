package errors

type ConnectionError struct {
	Err string
}

func (m *ConnectionError) Error() string {
	return "Connection error :: " + m.Err
}
