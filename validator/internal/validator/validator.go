// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package validator exposes common functionalities of all schema validators.
package validator

import "github.com/pkg/errors"

// ErrDeadletter is a special error type to help distinguish between invalid and broken messages.
var ErrDeadletter = errors.New("deadletter")

// ErrBrokenMessage is a special error type to help distinguish broken messages.
var ErrBrokenMessage = errors.New("Message is not in valid format")

// ErrWrongCompile is a special error type to help distinguish messages that had fault while compiling.
var ErrWrongCompile = errors.New("There is an error while compiling")

// ErrMissingSchema is a special error type to help distinguish messages that are missing schema.
var ErrMissingSchema = errors.New("Message is missing a schema")

// ErrMissingHeaderSchema is a special error type to help distinguish headers that are missing schema (header validation).
var ErrMissingHeaderSchema = errors.New("Header is missing a schema")

// ErrFailedValidation is a special error type to help distinguish messages that have failed in validation.
var ErrFailedValidation = errors.New("An error occured while validating message")

// ErrFailedHeaderValidation is a special error type to help distinguish messages' headers that have failed in validation.
var ErrFailedHeaderValidation = errors.New("An error occured while validating message's header")

// ErrUnsupportedFormat is a special error type to help distinguish messages that are in wrong format
var ErrUnsupportedFormat = errors.New("Message is not in a supported format")

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
