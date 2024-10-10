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
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dataphos/schema-registry-validator/internal/validator"

	"github.com/pkg/errors"
)

func TestNew(t *testing.T) {
	healthChecked := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/health" {
			healthChecked = true
			w.WriteHeader(http.StatusOK)
		} else {
			t.Fatal("wrong endpoint hit")
		}
	}))

	_, err := New(context.Background(), srv.URL, DefaultTimeoutBase)
	if err != nil {
		t.Fatal(err)
	}
	if !healthChecked {
		t.Error("health check not called")
	}
}

func TestNewTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/health" {
			time.Sleep(2 * time.Minute)
			w.WriteHeader(http.StatusOK)
		} else {
			t.Fatal("wrong endpoint hit")
		}
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := New(ctx, srv.URL, DefaultTimeoutBase)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expected timeout")
	}
}

func TestCSVValidator_Validate(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	csvV, err := New(context.Background(), "http://localhost:8088", DefaultTimeoutBase)
	if err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name           string
		dataFilename   string
		schemaFilename string
		valid          bool
		deadletter     bool
	}{
		{"valid-1", "valid-1-data.csv", "valid-1-schema.csvs", true, false},
		{"valid-2", "valid-2-data.csv", "valid-2-schema.csvs", true, false},
		{"valid-3", "valid-3-data.csv", "valid-3-schema.csvs", true, false},
		{"valid-4", "valid-4-data.csv", "valid-4-schema.csvs", true, false},
		{"invalid-1", "invalid-1-data.csv", "invalid-1-schema.csvs", false, false},
		{"invalid-2", "invalid-2-data.csv", "invalid-2-schema.csvs", false, false},
		{"invalid-3", "invalid-3-data.csv", "invalid-3-schema.csvs", false, false},
		{"deadletter-1", "deadletter-1-data.csv", "deadletter-1-schema.csvs", false, true},
		{"deadletter-2", "deadletter-2-data.csv", "deadletter-2-schema.csvs", false, true},
		{"deadletter-3", "deadletter-3-data.csv", "deadletter-3-schema.csvs", false, true},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(testdataDir, tc.dataFilename))
			if err != nil {
				t.Fatalf("data read error: %s", err)
			}
			schema, err := os.ReadFile(filepath.Join(testdataDir, tc.schemaFilename))
			if err != nil {
				t.Fatalf("schema read error: %s", err)
			}

			valid, err := csvV.Validate(data, schema, "", "")
			if tc.deadletter {
				if !errors.Is(err, validator.ErrDeadletter) {
					t.Error("deadletter expected")
				}
			} else {
				if err != nil {
					t.Errorf("validator error: %s", err)
				}
				if valid != tc.valid {
					if valid {
						t.Errorf("message valid, invalid expected")
					} else {
						t.Errorf("message invalid, valid expected")
					}
				}
			}
		})
	}
}
