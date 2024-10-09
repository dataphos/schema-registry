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

// Package schemagen exposes common functionalities of all schema generators.
package schemagen

import "github.com/pkg/errors"

// ErrDeadletter is a special error type to mark that schema generation was unsuccessful,
// due to the fact that the given message isn't even the right format.
var ErrDeadletter = errors.New("deadletter")

// Generator defines schema generators.
type Generator interface {
	// Generate takes data of some assumed format and returns the schema inferred from that data
	Generate([]byte) ([]byte, error)
}

// Func convenience type which is the functional equivalent of Generator.
type Func func(data []byte) ([]byte, error)

// Generate implements Generate by forwarding the call to the underlying Func.
func (f Func) Generate(data []byte) ([]byte, error) {
	return f(data)
}
