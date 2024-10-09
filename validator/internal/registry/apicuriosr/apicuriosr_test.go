package apicuriosr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

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

	_, err := New(context.Background(), srv.URL, DefaultTimeoutSettings, "default")
	if err != nil {
		t.Fatal(err)
	}
	if !healthChecked {
		t.Fatal("health check not called")
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

	_, err := New(ctx, srv.URL, DefaultTimeoutSettings, "default")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expected timeout")
	}
}

func TestGet(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
	schema := []byte("some specification")

	id, version := "1", "1"

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == fmt.Sprintf("/apis/registry/v2/groups/default/artifacts/%s/versions/%s", id, version) {
			_, _ = writer.Write(schema)
		} else {
			t.Fatal("wrong endpoint called")
		}
	}))
	defer srv.Close()

	registry := SchemaRegistry{
		Url:      srv.URL,
		Timeouts: DefaultTimeoutSettings,
	}

	spec, err := registry.Get(context.Background(), id, version)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(schema, spec) {
		t.Fatalf("expected and actual spec not the same (%s != %s)", schema, spec)
	}
}

func TestRegister(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
	schema := []byte("some specification")
	schemaType := "json"

	requestResponse := insertInfo{
		Id:      "1",
		Version: "1",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && request.URL.Path == "/apis/registry/v2/groups/default/artifacts" {
			defer request.Body.Close()
			registration, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(registration) != string(schema) {
				t.Fatal("expected and actual schema not the same")
			}

			writer.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(writer).Encode(requestResponse); err != nil {
				t.Fatal(err)
			}
		} else if request.Method == http.MethodPost && request.URL.Path == fmt.Sprintf("/apis/registry/v2/groups/default/artifacts/%s/rules", requestResponse.Id) {
			writer.WriteHeader(http.StatusNoContent)
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
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}
	schema := []byte("some specification")
	schemaId := "1"

	requestResponse := insertInfo{
		Id:      schemaId,
		Version: "2",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPut && request.URL.Path == fmt.Sprintf("/apis/registry/v2/groups/default/artifacts/%s", schemaId) {
			defer request.Body.Close()
			registration, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(registration) != string(schema) {
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
