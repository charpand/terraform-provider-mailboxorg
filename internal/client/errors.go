package client

import (
	"errors"
	"fmt"
)

// Documented mailbox.org JSON-RPC error codes.
const (
	codeAccessDenied  = 2
	codeNoPermission  = 8
	codeInvalidAuthID = 253
)

var (
	// ErrAuthentication is returned when credentials or the session are rejected.
	ErrAuthentication = errors.New("authentication failed")
	// ErrForbidden is returned when the API refuses access to an object (code 8).
	// The API uses it both for objects the caller may not manage and for objects
	// that do not exist, so the two cannot be told apart.
	ErrForbidden = errors.New("not permitted or does not exist")
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
	if detail, ok := e.Data.(string); ok && detail != "" {
		return fmt.Sprintf("mailbox.org %s: error %d: %s (%s)", e.Method, e.Code, e.Message, detail)
	}
	return fmt.Sprintf("mailbox.org %s: error %d: %s", e.Method, e.Code, e.Message)
}

// Is lets errors.Is match the auth and permission codes. Messages are localised
// by the API, so only the numeric code is ever inspected.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrAuthentication:
		return e.Code == codeAccessDenied || e.Code == codeInvalidAuthID
	case ErrForbidden:
		return e.Code == codeNoPermission
	}
	return false
}
