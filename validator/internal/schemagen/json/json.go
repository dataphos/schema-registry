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

package json

import (
	"os/exec"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"
)

func New(filename string) schemagen.Generator {
	return schemagen.Func(func(data []byte) ([]byte, error) {
		// #nosec G204 this would usually be a security concern because of remote code execution,
		// but it's fine here since we execute a python script from a file, so the attacker would need to have
		// full access to the vm to execute the script, and in that case, they could just execute the script themselves
		return schemagen.ExternalCmdSchemaGenerator(exec.Command("python", filename), data)
	})
}
