package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
)

type responseBodyAndCode struct {
	Body []byte
	Code int
}

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

	return schemaRegisterRequest, nil
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
