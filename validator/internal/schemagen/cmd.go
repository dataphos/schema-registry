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

package schemagen

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

// ExternalCmdSchemaGenerator generates the schema by calling the given cmd and passing the data to its stdin.
func ExternalCmdSchemaGenerator(cmd *exec.Cmd, data []byte) ([]byte, error) {
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, string(output))
	}
	if len(output) == 0 {
		return nil, ErrDeadletter
	}

	return output, nil
}
