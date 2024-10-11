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
	"syscall"
	"testing"

	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	"github.com/dataphos/schema-registry-validator/internal/validator"

	"github.com/pkg/errors"
)

func TestDeadletter(t *testing.T) {
	opError := OpError{
		Err: validator.ErrDeadletter,
	}
	if !opError.Deadletter() {
		t.Fatal("expected Deadletter")
	}

	opError.Err = errors.New("oops")
	if opError.Deadletter() {
		t.Fatal("shouldn't be Deadletter")
	}

	opError.Err = schemagen.ErrDeadletter
	if !opError.Deadletter() {
		t.Fatal("should be Deadletter")
	}
}

func TestTemporary(t *testing.T) {
	opError := OpError{
		Err: registry.ErrNotFound,
	}
	if !opError.Temporary() {
		t.Fatal("expected temporary")
	}

	opError.Err = syscall.ECONNREFUSED
	if opError.Temporary() {
		t.Fatal("expected not temporary")
	}
}
