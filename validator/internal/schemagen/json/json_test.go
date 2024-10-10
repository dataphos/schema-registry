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
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"

	"github.com/pkg/errors"
)

func TestGenerate(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	gen := New("json_schema_gen.py")

	tt := []struct {
		name           string
		dataFilename   string
		schemaFilename string
		deadletter     bool
	}{
		{"deadletter-1", "deadletter-1-data.json", "", true},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(testdataDir, tc.dataFilename))
			if err != nil {
				t.Errorf("data read error: %s", err)
			}

			generated, err := gen.Generate(data)
			if tc.deadletter {
				if !errors.Is(err, schemagen.ErrDeadletter) {
					t.Error("deadletter expected")
				}
			} else {
				if err != nil {
					t.Errorf("validator error: %s", err)
				}

				schema, err := os.ReadFile(filepath.Join(testdataDir, tc.schemaFilename))
				if err != nil {
					t.Errorf("schema read error: %s", err)
				}
				if !bytes.Equal(schema, generated) {
					t.Errorf("expected and generated schema not the same")
				}
			}
		})
	}
}
