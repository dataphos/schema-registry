package csv

import (
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"
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
