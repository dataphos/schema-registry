package json

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"

	"github.com/pkg/errors"
)

func TestJSONValidator_Validate(t *testing.T) {
	jsonV := New()

	tt := []struct {
		name           string
		dataFilename   string
		schemaFilename string
		valid          bool
		deadletter     bool
	}{
		{"valid-1", "valid-1-data.json", "valid-1-schema.json", true, false},
		{"valid-2", "valid-2-data.json", "valid-2-schema.json", true, false},
		{"valid-3", "valid-3-data.json", "valid-3-schema.json", true, false},
		{"valid-4", "valid-4-data.json", "valid-4-schema.json", true, false},
		//		{"invalid-1", "invalid-1-data.json", "invalid-1-schema.json", false, false},
		//		{"invalid-2", "invalid-2-data.json", "invalid-2-schema.json", false, false},
		//		{"invalid-3", "invalid-3-data.json", "invalid-3-schema.json", false, false},
		{"deadletter-1", "deadletter-1-data.json", "deadletter-1-schema.json", false, true},
		{"deadletter-2", "deadletter-2-data.json", "deadletter-2-schema.json", false, true},
		{"data-1", "data-1.json", "schema-1.json", true, false},
		{"data-2", "data-2.json", "schema-2.json", true, false},
		{"data-3", "data-3.json", "schema-3.json", true, false},
		{"data-4", "data-4.json", "schema-4.json", true, false},

		{"ref-1", "ref-data-1.json", "ref-1.json", true, false},
		{"ref-2", "ref-data-2.json", "ref-2.json", true, false},
		{"ref-3", "ref-data-3.json", "ref-3.json", true, false},
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
			schema, err := os.ReadFile(filepath.Join(testdataDir, tc.schemaFilename))
			if err != nil {
				t.Errorf("schema read error: %s", err)
			}

			valid, err := jsonV.Validate(data, schema, "", "")
			if tc.deadletter {
				if !(errors.Is(err, validator.ErrDeadletter) || errors.Is(err, validator.ErrFailedValidation) || errors.Is(err, validator.ErrWrongCompile) || errors.Is(err, validator.ErrMissingSchema) || errors.Is(err, validator.ErrBrokenMessage)) {
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

func BenchmarkValidateStandardImplementation(b *testing.B) {
	v := New()

	tt := []struct {
		dataFilename   string
		schemaFilename string
		data           []byte
		schema         []byte
	}{
		{dataFilename: "valid-1-data.json", schemaFilename: "valid-1-schema.json"},
		{dataFilename: "valid-2-data.json", schemaFilename: "valid-2-schema.json"},
		{dataFilename: "valid-3-data.json", schemaFilename: "valid-3-schema.json"},
		{dataFilename: "valid-4-data.json", schemaFilename: "valid-4-schema.json"},
		{dataFilename: "data-1.json", schemaFilename: "schema-1.json"},
		{dataFilename: "data-2.json", schemaFilename: "schema-2.json"},
		{dataFilename: "data-3.json", schemaFilename: "schema-3.json"},
		{dataFilename: "data-4.json", schemaFilename: "schema-4.json"},
	}

	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)
	testdataDir := filepath.Join(basepath, "testdata")
	for i := range tt {
		data, err := os.ReadFile(filepath.Join(testdataDir, tt[i].dataFilename))
		if err != nil {
			b.Errorf("data read error: %s", err)
		}
		schema, err := os.ReadFile(filepath.Join(testdataDir, tt[i].schemaFilename))
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}

		tt[i].data = data
		tt[i].schema = schema
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tt {
			valid, err := v.Validate(tc.data, tc.schema, "", "")
			if err != nil {
				b.Errorf("schema read error: %s", err)
			}
			if !valid {
				b.Errorf("expected valid")
			}
		}
	}
}

func BenchmarkValidateCachedImplementation(b *testing.B) {
	v := NewCached(100)

	tt := []struct {
		dataFilename   string
		schemaFilename string
		id             string
		version        string
		data           []byte
		schema         []byte
	}{
		{dataFilename: "valid-1-data.json", schemaFilename: "valid-1-schema.json", id: "1", version: "1"},
		{dataFilename: "valid-2-data.json", schemaFilename: "valid-2-schema.json", id: "2", version: "1"},
		{dataFilename: "valid-3-data.json", schemaFilename: "valid-3-schema.json", id: "3", version: "1"},
		{dataFilename: "valid-4-data.json", schemaFilename: "valid-4-schema.json", id: "4", version: "1"},
		{dataFilename: "data-1.json", schemaFilename: "schema-1.json", id: "5", version: "1"},
		{dataFilename: "data-2.json", schemaFilename: "schema-2.json", id: "6", version: "1"},
		{dataFilename: "data-3.json", schemaFilename: "schema-3.json", id: "7", version: "1"},
		{dataFilename: "data-4.json", schemaFilename: "schema-4.json", id: "8", version: "1"},
	}

	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)
	testdataDir := filepath.Join(basepath, "testdata")
	for i := range tt {
		data, err := os.ReadFile(filepath.Join(testdataDir, tt[i].dataFilename))
		if err != nil {
			b.Errorf("data read error: %s", err)
		}
		schema, err := os.ReadFile(filepath.Join(testdataDir, tt[i].schemaFilename))
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}

		tt[i].data = data
		tt[i].schema = schema
	}

	for _, tc := range tt {
		valid, err := v.Validate(tc.data, tc.schema, tc.id, tc.version)
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}
		if !valid {
			b.Errorf("expected valid")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tt {
			valid, err := v.Validate(tc.data, tc.schema, tc.id, tc.version)
			if err != nil {
				b.Errorf("schema read error: %s", err)
			}
			if !valid {
				b.Errorf("expected valid")
			}
		}
	}
}

func BenchmarkValidateGoJsonSchema(b *testing.B) {
	v := NewGoJsonSchemaValidator()

	tt := []struct {
		dataFilename   string
		schemaFilename string
		data           []byte
		schema         []byte
	}{
		{dataFilename: "valid-1-data.json", schemaFilename: "valid-1-schema.json"},
		{dataFilename: "valid-2-data.json", schemaFilename: "valid-2-schema.json"},
		{dataFilename: "valid-3-data.json", schemaFilename: "valid-3-schema.json"},
		{dataFilename: "valid-4-data.json", schemaFilename: "valid-4-schema.json"},
		{dataFilename: "data-1.json", schemaFilename: "schema-1.json"},
		{dataFilename: "data-2.json", schemaFilename: "schema-2.json"},
		{dataFilename: "data-3.json", schemaFilename: "schema-3.json"},
		{dataFilename: "data-4.json", schemaFilename: "schema-4.json"},
	}

	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)
	testdataDir := filepath.Join(basepath, "testdata")
	for i := range tt {
		data, err := os.ReadFile(filepath.Join(testdataDir, tt[i].dataFilename))
		if err != nil {
			b.Errorf("data read error: %s", err)
		}
		schema, err := os.ReadFile(filepath.Join(testdataDir, tt[i].schemaFilename))
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}

		tt[i].data = data
		tt[i].schema = schema
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tt {
			valid, err := v.Validate(tc.data, tc.schema, "", "")
			if err != nil {
				b.Errorf("schema read error: %s", err)
			}
			if !valid {
				b.Errorf("expected valid")
			}
		}
	}
}

func BenchmarkValidateCachedGoJsonSchema(b *testing.B) {
	v := NewCachedGoJsonSchemaValidator(100)

	tt := []struct {
		dataFilename   string
		schemaFilename string
		id             string
		version        string
		data           []byte
		schema         []byte
	}{
		{dataFilename: "valid-1-data.json", schemaFilename: "valid-1-schema.json", id: "1", version: "1"},
		{dataFilename: "valid-2-data.json", schemaFilename: "valid-2-schema.json", id: "2", version: "1"},
		{dataFilename: "valid-3-data.json", schemaFilename: "valid-3-schema.json", id: "3", version: "1"},
		{dataFilename: "valid-4-data.json", schemaFilename: "valid-4-schema.json", id: "4", version: "1"},
		{dataFilename: "data-1.json", schemaFilename: "schema-1.json", id: "5", version: "1"},
		{dataFilename: "data-2.json", schemaFilename: "schema-2.json", id: "6", version: "1"},
		{dataFilename: "data-3.json", schemaFilename: "schema-3.json", id: "7", version: "1"},
		{dataFilename: "data-4.json", schemaFilename: "schema-4.json", id: "8", version: "1"},
	}

	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)
	testdataDir := filepath.Join(basepath, "testdata")
	for i := range tt {
		data, err := os.ReadFile(filepath.Join(testdataDir, tt[i].dataFilename))
		if err != nil {
			b.Errorf("data read error: %s", err)
		}
		schema, err := os.ReadFile(filepath.Join(testdataDir, tt[i].schemaFilename))
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}

		tt[i].data = data
		tt[i].schema = schema
	}

	for _, tc := range tt {
		valid, err := v.Validate(tc.data, tc.schema, tc.id, tc.version)
		if err != nil {
			b.Errorf("schema read error: %s", err)
		}
		if !valid {
			b.Errorf("expected valid")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tt {
			valid, err := v.Validate(tc.data, tc.schema, tc.id, tc.version)
			if err != nil {
				b.Errorf("schema read error: %s", err)
			}
			if !valid {
				b.Errorf("expected valid")
			}
		}
	}
}
