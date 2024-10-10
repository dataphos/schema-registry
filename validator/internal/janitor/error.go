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

package janitor

import (
	"github.com/pkg/errors"

	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	"github.com/dataphos/schema-registry-validator/internal/validator"
)

type OpError struct {
	MessageID string
	Code      int
	Err       error
}

func (e *OpError) Error() string {
	return "processing " + e.MessageID + " failed:" + e.Err.Error()
}

// Unwrap implements the optional Unwrap error method, which allows for proper usage of errors.Is and errors.As.
func (e *OpError) Unwrap() error {
	return e.Err
}

// Temporary implements the optional Temporary error method, to ensure we don't hide the temporariness of the underlying
// error (in case code checking if this error is temporary doesn't use errors.As but just converts directly).
func (e *OpError) Temporary() bool {
	var temporary interface {
		Temporary() bool
	}

	// errors.As stops at the first error down the chain which implements temporary
	// this is important because an unrecoverable error could wrap a recoverable one, so we need the "latest" of the two
	if errors.As(e.Err, &temporary) {
		return temporary.Temporary()
	}
	return true
}

// Deadletter evaluates whether the instance is a Deadletter-type error.
func (e *OpError) Deadletter() bool {
	return errors.Is(e.Err, validator.ErrDeadletter) || errors.Is(e.Err, schemagen.ErrDeadletter)
}

func intoOpErr(messageId string, code int, err error) error {
	return &OpError{
		MessageID: messageId,
		Code:      code,
		Err:       err,
	}
}
