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

package csv

import (
	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	gen := New()

	tt := []struct {
		name           string
		dataFilename   string
		schemaFilename string
		deadletter     bool
	}{
		{"data-1", "data-1.csv", "schema-1.csvs", false},
		{"deadletter-1", "deadletter-1-data.csv", "", true},
		{"deadletter-2", "deadletter-2-data.csv", "", true},
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

				schemaStr := string(schema)
				schemaStr = strings.ReplaceAll(schemaStr, "\r\n", "\n")
				generatedStr := string(generated)

				if schemaStr != generatedStr {
					t.Errorf("expected and generated schema not the same")
				}
			}
		})
	}
}
