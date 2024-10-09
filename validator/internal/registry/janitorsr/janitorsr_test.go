package janitorsr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

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

	_, err := New(context.Background(), srv.URL, DefaultTimeoutSettings, "default")
	if err != nil {
		t.Fatal(err)
	}
	if !healthChecked {
		t.Fatal("health check not called")
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

	_, err := New(ctx, srv.URL, DefaultTimeoutSettings, "default")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expected timeout")
	}
}

func TestGet(t *testing.T) {
	schema := []byte("some specification")

	details := VersionDetails{
		VersionID:          "1",
		Version:            "1",
		SchemaID:           "1",
		Specification:      base64.StdEncoding.EncodeToString(schema),
		Description:        "some description",
		SchemaHash:         "some schema hash",
		CreatedAt:          time.Now(),
		VersionDeactivated: false,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == fmt.Sprintf("/schemas/%s/versions/%s", details.SchemaID, details.Version) {
			_ = json.NewEncoder(writer).Encode(details)
		} else {
			t.Fatal("wrong endpoint called")
		}
	}))
	defer srv.Close()

	registry := SchemaRegistry{
		Url:      srv.URL,
		Timeouts: DefaultTimeoutSettings,
	}

	spec, err := registry.Get(context.Background(), details.SchemaID, details.Version)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(schema, spec) {
		t.Fatalf("expected and actual spec not the same (%s != %s)", schema, spec)
	}
}

func TestRegister(t *testing.T) {
	schema := []byte("some specification")
	schemaType := "json"

	requestResponse := insertInfo{
		Id:      "1",
		Version: "1",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && request.URL.Path == "/schemas" {
			defer func() {
				err := request.Body.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
			var registration registrationRequest
			if err := json.NewDecoder(request.Body).Decode(&registration); err != nil {
				t.Fatal(err)
			}
			if registration.Specification != string(schema) {
				t.Fatal("expected and actual schema not the same")
			}

			writer.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(writer).Encode(requestResponse); err != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal("wrong endpoint called")
		}
	}))
	defer srv.Close()

	registry := SchemaRegistry{
		Url:      srv.URL,
		Timeouts: DefaultTimeoutSettings,
	}

	id, version, err := registry.Register(context.Background(), schema, schemaType, "none", "none")
	if err != nil {
		t.Fatal(err)
	}

	if id != requestResponse.Id || version != requestResponse.Version {
		t.Fatal("response not parsed correctly")
	}
}

func TestUpdate(t *testing.T) {
	schema := []byte("some specification")
	schemaId := "1"

	requestResponse := insertInfo{
		Id:      schemaId,
		Version: "2",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPut && request.URL.Path == fmt.Sprintf("/schemas/%s", schemaId) {
			defer func() {
				err := request.Body.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
			var registration registrationRequest
			if err := json.NewDecoder(request.Body).Decode(&registration); err != nil {
				t.Fatal(err)
			}
			if registration.Specification != string(schema) {
				t.Fatal("expected and actual schema not the same")
			}

			writer.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(writer).Encode(requestResponse); err != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal("wrong endpoint called")
		}
	}))
	defer srv.Close()

	registry := SchemaRegistry{
		Url:      srv.URL,
		Timeouts: DefaultTimeoutSettings,
	}

	version, err := registry.Update(context.Background(), schemaId, schema)
	if err != nil {
		t.Fatal(err)
	}

	if version != requestResponse.Version {
		t.Fatal("response not parsed correctly")
	}
}
