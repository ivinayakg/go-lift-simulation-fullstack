package utils

type CustomError struct {
	Message string
}

// Error returns the error message for the CustomError type.
func (e *CustomError) Error() string {
	return e.Message
}
