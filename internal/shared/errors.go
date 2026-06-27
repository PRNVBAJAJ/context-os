package shared

import "fmt"

// Code identifies the class of a domain error.
type Code string

const (
	// CodeNotFound indicates that a requested resource does not exist.
	CodeNotFound Code = "NOT_FOUND"
	// CodeInvalidInput indicates that the caller supplied invalid arguments.
	CodeInvalidInput Code = "INVALID_INPUT"
	// CodeConflict indicates that the operation conflicts with existing state.
	CodeConflict Code = "CONFLICT"
	// CodeInternal indicates an unexpected internal failure.
	CodeInternal Code = "INTERNAL"
)

// Error is the standard domain error type used throughout Context OS.
// It carries a machine-readable Code, a human-readable Message, and an
// optional Cause for error unwrapping via errors.Is / errors.As.
type Error struct {
	Code    Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause, enabling errors.Is and errors.As traversal.
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError returns a new Error with the given code and message.
func NewError(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap returns a new Error that annotates an existing error with a domain code
// and message, preserving the original as the unwrap chain.
func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}
