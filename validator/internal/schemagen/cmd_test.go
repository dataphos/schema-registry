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
	"os"
	"os/exec"
	"testing"
)

func TestOverScriptSchemaGenerator(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	cmd := exec.Command("python", "json/json_schema_gen.py")
	_, err := ExternalCmdSchemaGenerator(cmd, []byte("{\n  \"id\": 100,\n  \"first_name\": \"syn jason\",\n  \"last_name\": \"syn oblak\",\n  \"email\": \"jsonsmail\"\n}"))
	if err != nil {
		t.Fatal(err)
	}
}
