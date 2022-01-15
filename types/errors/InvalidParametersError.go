package errors

type InvalidParametersError struct {
	Err string
}

func (m *InvalidParametersError) Error() string {
	return "Invalid Parameter :: " + m.Err
}
