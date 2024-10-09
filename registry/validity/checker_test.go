package validity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewExternalChecker(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := NewExternalChecker(ctx, "http://localhost:8089", 1*time.Second)

	if err != nil {
		t.Fatal(err)
	}
}

func TestValidityChecker_Check(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker, err := NewExternalCheckerFromEnv(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type Data struct {
		Schema string
	}

	tt := []struct {
		name           string
		schemaFilename string
		schemaType     string
		validity       string
		valid          bool
	}{
		{"valid_json_syntax1", "valid_json_syntax1.json", "json", "syntax-only", true},
		{"invalid_json_syntax1", "invalid_json_syntax1.json", "json", "syntax-only", false},
		{"invalid_json_full1", "invalid_json_full1.json", "json", "full", false},
		{"valid_avro_full1", "valid_avro_full1.json", "avro", "full", true},
		{"invalid_avro_full1", "invalid_avro_full1.json", "avro", "full", false},
		{"invalid_avro_syntax1", "invalid_avro_syntax1.json", "avro", "syntax-only", false},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(testdataDir, tc.schemaFilename))
			if err != nil {
				t.Error(err)
			}

			var payload Data
			err = json.Unmarshal(content, &payload)
			if err != nil {
				t.Error(err)
			}

			newSchema := payload.Schema

			valid, err := checker.Check(newSchema, tc.schemaType, tc.validity)
			if err != nil {
				t.Errorf("validity error: %s", err)
			}
			if valid != tc.valid {
				if valid {
					t.Errorf("message valid, invalid expected")
				} else {
					t.Errorf("message invalid, valid expected")
				}
			}
		})
	}
}
