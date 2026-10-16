package client

import (
	"errors"
	"fmt"
)

// Documented mailbox.org JSON-RPC error codes.
const (
	codeAccessDenied  = 2
	codeInvalidAuthID = 253
)

var (
	// ErrAuthentication is returned when credentials or the session are rejected.
	ErrAuthentication = errors.New("authentication failed")
	// ErrTransport is returned when a request could not be completed or answered.
	ErrTransport = errors.New("transport error")
	// ErrDecode is returned when a response is not a valid JSON-RPC answer.
	ErrDecode = errors.New("decode error")
	// ErrAmbiguous is returned when a mutating request was sent but its outcome is unknown.
	ErrAmbiguous = errors.New("outcome of mutating request unknown")
)

// APIError is a JSON-RPC error object returned by the mailbox.org API.
// It carries the method name but never request parameters, so it is safe to log.
type APIError struct {
	Method  string
	Code    int
	Message string
	Data    any
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mailbox.org %s: error %d: %s", e.Method, e.Code, e.Message)
}

// Is lets errors.Is(err, ErrAuthentication) match the documented auth codes.
func (e *APIError) Is(target error) bool {
	return target == ErrAuthentication && (e.Code == codeAccessDenied || e.Code == codeInvalidAuthID)
}
