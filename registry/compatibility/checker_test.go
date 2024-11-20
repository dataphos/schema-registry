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

package compatibility

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := New(ctx, "http://localhost:8088", 2*time.Second)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCompatibilityChecker_Check(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := NewFromEnv(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type Data struct {
		Id     string `json:"id"`
		Format string `json:"format"`
		Schema string `json:"schema"`
	}

	tt := []struct {
		name       string
		schema     string
		history    string
		mode       string
		compatible bool
	}{
		{"compatible-1", "backward_json_true/schema2.json", "backward_json_true/schema1.json", "BACKWARD", true},
		{"compatible-2", "backward_json_false/schema2.json", "backward_json_false/schema1.json", "BACKWARD", false},
		{"compatible-3", "backward_avro_true/schema2.json", "backward_avro_true/schema1.json", "BACKWARD", true},
		{"compatible-4", "backward_avro_false/schema2.json", "backward_avro_false/schema1.json", "BACKWARD", false},

		{"compatible-5", "forward_json_true/schema2.json", "forward_json_true/schema1.json", "FORWARD", true},
		{"compatible-6", "forward_json_false/schema2.json", "forward_json_false/schema1.json", "FORWARD", false},
		{"compatible-7", "forward_avro_true/schema2.json", "forward_avro_true/schema1.json", "FORWARD", true},
		{"compatible-8", "forward_avro_false/schema2.json", "forward_avro_false/schema1.json", "FORWARD", false},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			newSchema, err := os.ReadFile(filepath.Join(testdataDir, tc.schema))
			if err != nil {
				t.Fatalf("newSchema read error: %s", err)
			}

			var schemaHistory []string
			previousSchema, err := os.ReadFile(filepath.Join(testdataDir, tc.history))
			if err != nil {
				t.Fatalf("schemaHistory read error: %s", err)
			}
			var previousSchemaJson Data
			if err := json.Unmarshal(previousSchema, &previousSchemaJson); err != nil {
				t.Fatalf("couldn't unmarshall schema history")
			}

			schemaHistory = append(schemaHistory, base64.StdEncoding.EncodeToString([]byte(previousSchemaJson.Schema)))

			compatible, err := checker.Check(string(newSchema), schemaHistory, tc.mode)
			if err != nil {
				t.Errorf("validator error: %s", err)
			}
			if compatible != tc.compatible {
				if compatible {
					t.Errorf("message compatible, incompatible expected")
				} else {
					t.Errorf("message incompatible, compatible expected")
				}
			}
		})
	}
}
