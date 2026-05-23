package utils

import "fmt"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func NewAppError(status int, code, message string, err error) *AppError {
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}
