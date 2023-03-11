package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrRequiredParam = errors.New("missing required param")

	ErrAuth                    = &sentinelError{statusCode: http.StatusUnauthorized, message: "not authorized"}
	ErrNotFound                = &sentinelError{statusCode: http.StatusNotFound, message: "resource not found"}
	ErrFieldValidation         = &sentinelError{statusCode: http.StatusUnprocessableEntity, message: "invalid request's param"}
	ErrFieldValidationRequired = &sentinelError{statusCode: http.StatusBadRequest, message: ErrRequiredParam.Error()}
	ErrEmailRegistered         = &sentinelError{statusCode: http.StatusConflict, message: "email already registered"}
)

type APIError interface {
	APIError() (int, string)
	Causer
}

type Causer interface {
	Cause() error
}
type sentinelError struct {
	statusCode int
	message    string
}

func (s sentinelError) Cause() error {
	return s
}

func (s sentinelError) Error() string {
	return s.message
}

func (s sentinelError) APIError() (int, string) {
	return s.statusCode, s.message
}

type sentinelWrappedError struct {
	err             error
	overrideMessage string
	sentinel        *sentinelError
}

func (s sentinelWrappedError) Error() string {
	return s.err.Error()
}

func (e sentinelWrappedError) Is(err error) bool {
	return e.sentinel == err
}

func (e sentinelWrappedError) APIError() (int, string) {
	code, message := e.sentinel.APIError()
	if e.overrideMessage != "" {
		message = e.overrideMessage
	}

	return code, message
}

// Unwrap used to satisfied errors.Unwrap function which will return error wrapped from WrapError function
func (e sentinelWrappedError) Unwrap() error {
	return e.err
}

// Cause will return the first error which not implement anonymous Unwrap interface
// every error which wrapped  by sentinelWrappedError might be wrapped by another error to add context
func (e sentinelWrappedError) Cause() error {
	err := e.err
	for {
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return err
		}
		err = wrappedErr
	}
}

func WrapError(err error, sentinel *sentinelError, overrideMessage string) error {

	return sentinelWrappedError{err: err, sentinel: sentinel, overrideMessage: overrideMessage}

}

func Cause(err error) error {
	for {
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return err
		}
		err = wrappedErr
	}
}
