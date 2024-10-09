package xml

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"

	"github.com/pkg/errors"
)

func TestNew(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
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
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
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

func TestXMLValidator_Validate(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	xmlV, err := New(context.Background(), "http://localhost:8089", DefaultTimeoutBase)
	if err != nil {
		t.Fatal("validator constructor error", err)
	}

	tt := []struct {
		name           string
		dataFilename   string
		schemaFilename string
		valid          bool
		deadletter     bool
	}{
		{"valid-1", "valid-1-data.xml", "valid-1-schema.xsd", true, false},
		{"valid-2", "valid-2-data.xml", "valid-2-schema.xsd", true, false},
		{"valid-3", "valid-3-data.xml", "valid-3-schema.xsd", true, false},
		{"invalid-1", "invalid-1-data.xml", "invalid-1-schema.xsd", false, false},
		{"invalid-2", "invalid-2-data.xml", "invalid-2-schema.xsd", false, false},
		{"deadletter-1", "deadletter-1-data.xml", "deadletter-1-schema.xsd", false, true},
		{"deadletter-2", "deadletter-2-data.xml", "deadletter-2-schema.xsd", false, true},
		{"deadletter-3", "deadletter-3-data.xml", "deadletter-3-schema.xsd", false, true},
		{"data-1", "data-1.xml", "schema-1.xsd", true, false},
		{"data-2", "data-2.xml", "schema-2.xsd", false, false},
		{"data-3", "data-3.xml", "schema-3.xsd", true, false},
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

			valid, err := xmlV.Validate(data, schema, "", "")
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
