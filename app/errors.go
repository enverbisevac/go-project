package app

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNameIsRequired       = ErrInvalid("name is required")
	ErrFullnameIsRequired   = ErrInvalid("full_name is required")
	ErrEmailFieldIsRequired = ErrInvalid("email is required")
	ErrEmailIsNotValid      = ErrInvalid("email is not valid")
	ErrPasswordIsRequired   = ErrInvalid("password is required")
	ErrPasswordIsTooShort   = ErrInvalid("password is too short")
	ErrPasswordIsTooLong    = ErrInvalid("password is too long")
	ErrPasswordIsCommon     = ErrInvalid("password is common")
)

type Status string

// Application status codes.
//
// NOTE: These are meant to be generic and they map well to HTTP error codes.
// Different applications can have very different status code requirements so
// these should be expanded as needed (or introduce subcodes).
const (
	StatusConflict        Status = "conflict"
	StatusInternal        Status = "internal"
	StatusInvalid         Status = "invalid"
	StatusNotFound        Status = "not_found"
	StatusNotImplemented  Status = "not_implemented"
	StatusUnauthenticated Status = "unauthenticated"
	StatusUnauthorized    Status = "unauthorized"
)

// Error represents an application-specific error. Application errors can be
// unwrapped by the caller to extract out the code & message.
//
// Any non-application error (such as a disk error) should be reported as an
// StatusInternal and the human user should only see "Internal error" as the
// message. These low-level internal error details should only be logged and
// reported to the operator of the application (not the end user).
type Error struct {
	// Machine-readable status code.
	Status Status

	// Human-readable error message.
	Message string

	// Payload
	Payload any

	// Source error
	Err error
}

// Error implements the error interface. Not used by the application otherwise.
func (e *Error) Error() string {
	return fmt.Sprintf("app error: code=%s, message=%s", e.Status, e.Message)
}

// ErrorStatus unwraps an application error and returns its code.
// Non-application errors always return StatusInternal.
func ErrorStatus(err error) Status {
	var e *Error

	if err == nil {
		return ""
	}

	if errors.As(err, &e) {
		return e.Status
	}

	return StatusInternal
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error

	if err == nil {
		return ""
	}

	if errors.As(err, &e) {
		return e.Message
	}

	return "Internal error."
}

// ErrorPayload unwraps an application error and returns its payload.
// Non-application errors always return nil.
func ErrorPayload(err error) any {
	var e *Error

	if err == nil {
		return nil
	}

	if errors.As(err, &e) {
		return e.Payload
	}

	return nil
}

// SourceError unwraps an application error and returns its source error.
// Non-application errors always return input error.
func SourceError(err error) error {
	var e *Error

	if err == nil {
		return nil
	}

	if errors.As(err, &e) {
		return e.Err
	}

	return err
}

// Errorf is a helper function to return an error with a given
// code and formatted message.
func Errorf(code Status, format string, args ...interface{}) error {
	var err error

	newargs := make([]any, 0, len(args))

	for _, arg := range args {
		switch t := arg.(type) {
		case error:
			err = t
		default:
			newargs = append(newargs, t)
		}
	}

	// TODO: needs better handling when err is Error

	if err != nil && strings.Contains(format, "%w") {
		format = strings.Replace(format, "%w", "%v", -1)
		newargs = append(newargs, ErrorMessage(err))
	}

	message := fmt.Sprintf(format, newargs...)

	return &Error{
		Status:  code,
		Message: message,
		Err:     err,
	}
}

// ErrorWithPayload is a helper function to return an error with a given
// payload.
func ErrorWithPayload(err error, payload any) error {
	var e *Error

	if err == nil {
		return nil
	}

	if errors.As(err, &e) {
		e.Payload = payload
		return e
	}
	return err
}

// ErrConflict is a helper function to return an conflict error with a
// given code and formatted message.
func ErrConflict(format string, args ...interface{}) error {
	return Errorf(StatusConflict, format, args...)
}

// ErrInternal is a helper function to return an internal error with a
// given code and formatted message.
func ErrInternal(format string, args ...interface{}) error {
	return Errorf(StatusInternal, format, args...)
}

// ErrInvalid is a helper function to return an invalid argument error
// with a given code and formatted message.
func ErrInvalid(format string, args ...interface{}) error {
	return Errorf(StatusInvalid, format, args...)
}

// ErrNotFound is a helper function to return an not_found error
// with a given code and formatted message.
func ErrNotFound(format string, args ...interface{}) error {
	return Errorf(StatusNotFound, format, args...)
}

// ErrNotImplemented is a helper function to return an not_implemented error
// with a given code and formatted message.
func ErrNotImplemented(format string, args ...interface{}) error {
	return Errorf(StatusNotImplemented, format, args...)
}

// ErrUnauthenticated is a helper function to return unauthenticated error
// with a given code and formatted message.
func ErrUnauthenticated(format string, args ...interface{}) error {
	return Errorf(StatusUnauthenticated, format, args...)
}

// ErrUnauthorized is a helper function to return unauthorized error
// with a given code and formatted message.
func ErrUnauthorized(format string, args ...interface{}) error {
	return Errorf(StatusUnauthorized, format, args...)
}

// ErrFieldIsMandatory is a helper function to return error
// with a given mandatory field.
func ErrFieldIsMandatory(field string) error {
	return ErrInvalid("%s is mandatory field", field)
}
