// Package validator exposes common functionalities of all schema validators.
package validator

import "github.com/pkg/errors"

// ErrDeadletter is a special error type to help distinguish between invalid and broken messages.
var ErrDeadletter = errors.New("deadletter")

// ErrBrokenMessage is a special error type to help distinguish broken messages.
var ErrBrokenMessage = errors.New("Message is not in valid format")

// ErrWrongCompile is a special error type to help distinguish messages that had fault while compiling.
var ErrWrongCompile = errors.New("There is an error while compiling.")

// ErrMissingSchema is a special error type to help distinguish messages that are missing schema.
var ErrMissingSchema = errors.New("Message is missing a schema")

// ErrFailedValidation is a special error type to help distinguish messages that have failed in validation.
var ErrFailedValidation = errors.New("An error occured while validating message.")

// Validator is the interface used to model message validators.
type Validator interface {
	// Validate takes a message and a schema (along with schema id and version, in case they are needed for optimization purposes)
	// and returns a bool value with the validation result.
	// Returns an error in case the implementation encounters an unrecoverable issue.
	// If the unrecoverable issue is a broken message or schema (for example, the given message isn't even in the
	// valid format), ErrDeadletter MUST be returned.
	Validate(message, schema []byte, id string, version string) (bool, error)
}

// Func convenience type which is the functional equivalent of Validator.
type Func func(message, schema []byte, id string, version string) (bool, error)

// Validate implements Validate by forwarding the call to the underlying ValidationFunc.
func (f Func) Validate(message, schema []byte, id string, version string) (bool, error) {
	return f(message, schema, id, version)
}
