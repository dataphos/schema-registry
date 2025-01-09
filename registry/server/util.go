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

package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/dataphos/schema-registry/registry"
)

type responseBodyAndCode struct {
	Body []byte
	Code int
}

var supportedFormats = []string{"json", "avro", "xml", "csv", "protobuf"}

func writeResponse(w http.ResponseWriter, response responseBodyAndCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	_, _ = w.Write(response.Body)
}

func serializeErrorMessage(message string) []byte {
	encoded, _ := json.Marshal(report{Message: message})
	return encoded
}

func readSchemaRegisterRequest(body io.ReadCloser) (registry.SchemaRegistrationRequest, error) {
	encoded, err := io.ReadAll(body)
	if err != nil {
		return registry.SchemaRegistrationRequest{}, err
	}

	var schemaRegisterRequest registry.SchemaRegistrationRequest
	if err = json.Unmarshal(encoded, &schemaRegisterRequest); err != nil {
		return registry.SchemaRegistrationRequest{}, err
	}

	// check if format is unknown
	format := strings.ToLower(schemaRegisterRequest.SchemaType)
	if !containsFormat(format) {
		return registry.SchemaRegistrationRequest{}, registry.ErrUnknownFormat
	}

	return schemaRegisterRequest, nil
}

func containsFormat(format string) bool {
	for _, supportedFormat := range supportedFormats {
		if format == supportedFormat {
			return true
		}
	}
	return false
}

func readSchemaUpdateRequest(body io.ReadCloser) (registry.SchemaUpdateRequest, error) {
	encoded, err := io.ReadAll(body)
	if err != nil {
		return registry.SchemaUpdateRequest{}, err
	}

	var schemaUpdateRequest registry.SchemaUpdateRequest
	if err = json.Unmarshal(encoded, &schemaUpdateRequest); err != nil {
		return registry.SchemaUpdateRequest{}, err
	}

	return schemaUpdateRequest, nil
}

func readSchemaCompatibilityRequest(body io.ReadCloser) (registry.SchemaCompatibilityRequest, error) {
	encoded, err := io.ReadAll(body)
	if err != nil {
		return registry.SchemaCompatibilityRequest{}, err
	}

	var schemaCompatibilityRequest registry.SchemaCompatibilityRequest
	if err = json.Unmarshal(encoded, &schemaCompatibilityRequest); err != nil {
		return registry.SchemaCompatibilityRequest{}, err
	}

	return schemaCompatibilityRequest, nil
}

func readSchemaValidityRequest(body io.ReadCloser) (registry.SchemaValidityRequest, error) {
	encoded, err := io.ReadAll(body)
	if err != nil {
		return registry.SchemaValidityRequest{}, err
	}

	var schemaValidityRequest registry.SchemaValidityRequest
	if err = json.Unmarshal(encoded, &schemaValidityRequest); err != nil {
		return registry.SchemaValidityRequest{}, err
	}

	return schemaValidityRequest, nil
}
