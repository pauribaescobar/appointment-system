package domain

import (
	"errors"
)

// APPLICATION POSSIBLE ERRORS
var (
	ErrAppointmentAlreadyExists    = errors.New("Appointment already Exists")
	ErrAppointmentAlreadyProcessed = errors.New("Appointment already processed")
	ErrApointmentNotFound          = errors.New("Appointment not Found")
	ErrInvalidInput                = errors.New("Invalid input data")
	ErrUnauthorized                = errors.New("Unauthorized")
	ErrRateLimited                 = errors.New("Rate Limited")
	ErrTransient                   = errors.New("Transient Error")
	ErrInternal                    = errors.New("Internal server error")
)

type AppError struct {
	Op    string
	Kind  error // One of our APPLICATION POSSIBLE ERRORS
	Cause error // Original error
}

func (e *AppError) Error() string {
	if e.Cause == nil {
		return e.Op + ": " + e.Kind.Error()
	}

	return e.Op + ": " + e.Kind.Error() + e.Cause.Error()
}

func (e *AppError) Unwrap() error {
	return e.Kind
}

// App Error Builder
func E(op string, kind error, cause error) error {
	return &AppError{
		Op:    op,
		Kind:  kind,
		Cause: cause,
	}
}
