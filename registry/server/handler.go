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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/schema-registry/internal/metrics"
	"github.com/dataphos/schema-registry/registry"
)

type Handler struct {
	Service *registry.Service
	log     logger.Log
}

// report is a simple wrapper of the system's message for the user.
type report struct {
	Message string `json:"message"`
}

// insertInfo represents a schema registry/evolution response for methods other than GET.
type insertInfo struct {
	Id      string `json:"identification"`
	Version string `json:"version"`
	Message string `json:"message"`
}

// NewHandler is a convenience function which returns a new instance of Handler.
func NewHandler(Service *registry.Service, log logger.Log) *Handler {
	return &Handler{
		Service: Service,
		log:     log,
	}
}

// GetSchemaVersionByIdAndVersion is a GET method that expects parameters "id" and "version" for
// retrieving the schema version from the underlying repository.
//
// It currently writes back either:
//   - status 200 with a schema version in JSON format, if the schema is registered and active
//   - status 404 with error message, if the schema version is not registered or registered but deactivated
//   - status 500 with error message, if an internal server error occurred
//
// @Title        Get schema version by schema id and version
// @Summary      Get schema version by schema id and version
// @Produce      json
// @Param        id path string true "schema id"
// @Param        version path string true "version"
// @Success      200
// @Failure      404
// @Failure      500
// @Router       /schemas/{id}/versions/{version} [get]
func (h Handler) GetSchemaVersionByIdAndVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	version := chi.URLParam(r, "version")

	details, err := h.Service.GetSchemaVersion(id, version)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Schema with id=%v and version=%v is not registered", id, version),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		} else if errors.Is(err, registry.ErrInvalidValueHeader) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Id=%v and/or version=%v are not of supported data types", id, version),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusUnprocessableEntity,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(details)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// GetSpecificationByIdAndVersion is a GET method that expects parameters "id" and "version" for
// retrieving the specification of schema version from the underlying repository.
//
// It currently writes back either:
//   - status 200 with a schema in JSON format, if the schema version is registered and active
//   - status 404 with error message, if the schema version is not registered or registered but deactivated
//   - status 500 with error message, if an internal server error occurred
//
// @Title        Get schema specification by schema id and version
// @Summary      Get schema specification by schema id and version
// @Produce      json
// @Param        id path string true "schema id"
// @Param        version path string true "version"
// @Success 	 200
// @Failure 	 404
// @Failure 	 500
// @Router       /schemas/{id}/versions/{version}/spec [get]
func (h Handler) GetSpecificationByIdAndVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	version := chi.URLParam(r, "version")

	details, err := h.Service.GetSchemaVersion(id, version)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Schema with id=%v and version=%v is not registered", id, version),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}
	specification, err := base64.StdEncoding.DecodeString(details.Specification)
	if err != nil {
		log.Println(err)
	}

	writeResponse(w, responseBodyAndCode{
		Body: specification,
		Code: http.StatusOK,
	})
}

// GetSchemaVersionsById is a GET method that expects "id" of the wanted schema and returns all active versions of the schema
//
// It currently gives the following responses:
//   - status 200 for a successful invocation along with an instance of the schema structure
//   - status 404 if there is no registered or active schema version under the given id
//   - status 500 with error message, if an internal server error occurred
//
// @Title 	 Get all active schema versions by schema id
// @Summary 	 Get all active schema versions by schema id
// @Produce      json
// @Param        id path string true "schema id"
// @Success      200
// @Failure      404
// @Failure      500
// @Router       /schemas/{id}/versions [get]
func (h Handler) GetSchemaVersionsById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	schemas, err := h.Service.ListSchemaVersions(id)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeResponse(w, responseBodyAndCode{
				Body: serializeErrorMessage(http.StatusText(http.StatusNotFound)),
				Code: http.StatusNotFound,
			})
			return
		}
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(schemas)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// GetAllSchemaVersionsById is a GET method that expects "id" of the wanted schema and returns all versions of the schema
//
// It currently gives the following responses:
//   - status 200 for a successful invocation along with an instance of the schema structure
//   - status 404 if there is no registered schema version under the given id
//   - status 500 with error message, if an internal server error occurred
//
// @Title 	 Get schema by schema id
// @Summary 	 Get schema by schema id
// @Produce      json
// @Param        id path string true "schema id"
// @Success      200
// @Failure      404
// @Failure      500
// @Router       /schemas/{id}/versions/all [get]
func (h Handler) GetAllSchemaVersionsById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	schemas, err := h.Service.ListAllSchemaVersions(id)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Schema with id=%v is not registered", id),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(schemas)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// GetLatestSchemaVersionById  is a GET method that expects "id" of the wanted schema and returns
// the latest versions of the schema
//
// It currently gives the following responses:
//   - status 200 with the latest schema version in JSON format, if the schema is registered
//   - status 404 if there is no registered or active schema under the given id
//   - status 500 with error message, if an internal server error occurred
//
// @Title        Get the latest schema version by schema id
// @Summary      Get the latest schema version by schema id
// @Produce      json
// @Param        id path string true "schema id"
// @Success      200
// @Failure      404
// @Failure      500
// @Router       /schemas/{id}/versions/latest [get]
func (h Handler) GetLatestSchemaVersionById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	details, err := h.Service.GetLatestSchemaVersion(id)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Schema with id=%v is not registered", id),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(details)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// GetAllSchemas is a GET method that retrieves all schemas
//
// It currently writes back either:
//   - status 200 with all schemas in JSON format
//   - status 404 with error message if there are no registered schemas
//   - status 500 with error message, if an internal server error occurred
//
// @Title   Get all schemas
// @Summary Get all schemas
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router  /schemas/all [get]
func (h Handler) GetAllSchemas(w http.ResponseWriter, _ *http.Request) {
	schemas, err := h.Service.GetAllSchemas()
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: "There are no schemas in Registry",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(schemas)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// GetSchemas is a GET method that retrieves all active schemas
//
// It currently writes back either:
//   - status 200 with active schemas in JSON format
//   - status 404 with error message if there are no active schemas
//   - status 500 with error message, if an internal server error occurred
//
// @Title   Get all active schemas
// @Summary Get all active schemas
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router  /schemas [get]
func (h Handler) GetSchemas(w http.ResponseWriter, _ *http.Request) {
	schemas, err := h.Service.GetSchemas()
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeResponse(w, responseBodyAndCode{
				Body: serializeErrorMessage("No active schemas registered in the Registry"),
				Code: http.StatusNoContent,
			})
			return
		}
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	body, _ := json.Marshal(schemas)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// SearchSchemas  is a GET method that expects one of the following parameters: id, version, type, name, orderBy,
// sort, limit and gets schemas that match given filter criteria
//
// It currently writes back either:
//   - status 200 with filtered schemas in JSON format
//   - status 400 with error message, if a bad search query was given
//   - status 404 with error message, if there is no schema that matches the given search criteria
//   - status 500 with error message, if an internal server error occurred
//
// @Title        Search schemas
// @Summary      Search schemas
// @Produce      json
// @Param        id query string false "schema id"
// @Param        version query string false "schema version"
// @Param        type query string false "schema type"
// @Param        name query string false "schema name"
// @Param        orderBy query string false "order by name, type, id or version"
// @Param        sort query string false "sort schemas either asc or desc"
// @Param        limit query string false "maximum number of retrieved schemas matching the criteria"
// @Param        attributes query string false "schema attributes"
// @Success      200
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /schemas/search [get]
func (h Handler) SearchSchemas(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	version := r.URL.Query().Get("version")
	schemaType := r.URL.Query().Get("type")
	name := r.URL.Query().Get("name")
	orderBy := r.URL.Query().Get("orderBy")
	sort := r.URL.Query().Get("sort")

	if orderBy == "" && sort != "" {
		orderBy = "id"
	} else if orderBy != "" && orderBy != "name" && orderBy != "id" && orderBy != "type" && orderBy != "version" {
		body, _ := json.Marshal(report{
			Message: "Bad request: unknown value for orderBy",
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusBadRequest,
		})
		return
	}

	if sort == "" && orderBy != "" {
		sort = "asc"
	} else if sort != "" && sort != "asc" && sort != "desc" {
		body, _ := json.Marshal(report{
			Message: "Bad request: unknown value for sort",
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusBadRequest,
		})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0
	var err error
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			body, _ := json.Marshal(report{
				Message: "Bad request: limit must be integer",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}
	}

	var attributes []string
	if r.URL.Query().Get("attributes") != "" {
		attributes = strings.Split(r.URL.Query().Get("attributes"), ",")
	}

	queryParams := registry.QueryParams{
		Id:         id,
		Version:    version,
		SchemaType: schemaType,
		Name:       name,
		OrderBy:    orderBy,
		Sort:       sort,
		Limit:      limit,
		Attributes: attributes,
	}

	schemas, err := h.Service.SearchSchemas(queryParams)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeResponse(w, responseBodyAndCode{
				Body: serializeErrorMessage(http.StatusText(http.StatusNotFound)),
				Code: http.StatusNotFound,
			})
			return
		}
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	if schemas == nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusNotFound)),
			Code: http.StatusNotFound,
		})
		return
	}

	body, _ := json.Marshal(schemas)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

// PostSchema is a POST function that registers the received schema to the underlying repository.
//
// The expected input schema JSON should contain following fields:
// - Description       string
// - Specification     string
// - Name              string
// - SchemaType        string
// - LastCreated       string
// - PublisherID       string
// - CompatibilityMode string
// - ValidityMode      string
//
// It currently writes back either:
//   - status 201 with newly created version details in JSON format
//   - status 400 with error message, if the schema isn't valid or the values for validity and/or compatibility mode are missing
//   - status 409 with error message, if the schema already exists
//   - status 500 with error message, if an internal server error occurred
//
// In case of correct invocation the function writes back a JSON with fields:
// - Identification int64
// - Version        int32
// - Message        string
//
// @Title        Post new schema
// @Summary      Post new schema
// @Accept       json
// @Produce      json
// @Param        data body registry.SchemaRegistrationRequest false "schema registration request"
// @Success      201
// @Failure      400
// @Failure      409
// @Failure      500
// @Router       /schemas [post]
func (h Handler) PostSchema(w http.ResponseWriter, r *http.Request) {

	registerRequest, err := readSchemaRegisterRequest(r.Body)
	if err != nil {
		if errors.Is(err, registry.ErrUnknownFormat) {
			body, _ := json.Marshal(report{
				Message: "Bad request: unknown format value",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	details, added, err := h.Service.CreateSchema(registerRequest)
	if err != nil {
		if errors.Is(err, registry.ErrUnknownComp) {
			body, _ := json.Marshal(report{
				Message: "Bad request: unknown compatibility_mode value",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}

		if errors.Is(err, registry.ErrUnknownVal) {
			body, _ := json.Marshal(report{
				Message: "Bad request: unknown validity_mode value",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}

		if errors.Is(err, registry.ErrNotValid) {
			body, _ := json.Marshal(report{
				Message: "Schema is not valid",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	if !added {
		body, _ := json.Marshal(insertInfo{
			Id:      details.SchemaID,
			Version: details.Version,
			Message: fmt.Sprintf("Schema already exists at id=%v", details.SchemaID),
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusConflict,
		})
		return
	}

	body, _ := json.Marshal(insertInfo{
		Id:      details.SchemaID,
		Version: details.Version,
		Message: "Schema successfully created",
	})
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusCreated,
	})
	metrics.AddedSchemaMetricUpdate(details.SchemaID, details.Version)
}

// PutSchema registers a new schema version in the Schema Registry. The new version is connected to other schemas
// by schema id from the request URL.
// The expected input schema JSON should contain the following field:
// - Specification    string
// The input can also include the following field:
// - Description      string
//
// It currently writes back either:
//   - status 200 with updated version details in JSON format
//   - status 400 with error message, if the schemas aren't compatible
//   - status 404 if there is no registered or active schema version under the given id
//   - status 409 with error message, if the schema already exists
//   - status 500 with error message, if an internal server error occurred
//
// In case of correct invocation the function writes back a JSON with fields:
// - Identification int64
// - Version        int32
// - Message        string
//
// In case of a bad invocation, it only returns the message.
// @Title        Put new schema version
// @Summary      Put new schema version
// @Accept       json
// @Produce      json
// @Param        id path string true "schema id"
// @Param        data body registry.SchemaUpdateRequest true "schema update request"
// @Success      200
// @Failure      400
// @Failure      404
// @Failure      409
// @Failure      500
// @Router       /schemas/{id} [put]
func (h Handler) PutSchema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	updateRequest, err := readSchemaUpdateRequest(r.Body)
	if err != nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	details, updated, err := h.Service.UpdateSchema(id, updateRequest)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			body, _ := json.Marshal(report{
				Message: fmt.Sprintf("Schema with id=%v doesn't exist", id),
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusNotFound,
			})
			return
		} else if errors.Is(err, registry.ErrNotValid) {
			body, _ := json.Marshal(report{
				Message: "Schema is not valid",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		} else if errors.Is(err, registry.ErrNotComp) {
			body, _ := json.Marshal(report{
				Message: "Schemas are not compatible",
			})
			writeResponse(w, responseBodyAndCode{
				Body: body,
				Code: http.StatusBadRequest,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	if !updated {
		body, _ := json.Marshal(insertInfo{
			Id:      details.SchemaID,
			Version: details.Version,
			Message: fmt.Sprintf("Schema already exists at id=%s", details.SchemaID),
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusConflict,
		})
		return
	}

	body, _ := json.Marshal(insertInfo{
		Id:      details.SchemaID,
		Version: details.Version,
		Message: "Schema successfully updated",
	})
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
	metrics.UpdateSchemaMetricUpdate(details.SchemaID, details.Version)
}

// DeleteSchema is a DELETE method that deactivates a schema.
// It expects the "id" of the wanted schema
//
// It currently gives the following responses:
//   - status 200 for a successful invocation along with an instance of the schema structure
//   - status 400 if the deletion caused an error
//   - status 404 if the schema does not exist or is already deactivated
//
// @Title        Delete schema by schema id
// @Summary      Delete schema by schema id
// @Produce      json
// @Param        id path string true "schema id"
// @Success      200
// @Failure      400
// @Failure      404
// @Router       /schemas/{id} [delete]
func (h Handler) DeleteSchema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	deleted, err := h.Service.DeleteSchema(id)
	if err != nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	if !deleted {
		body, _ := json.Marshal(report{Message: fmt.Sprintf("Schema with id=%s doesn't exist", id)})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusNotFound,
		})
		return
	}

	body, _ := json.Marshal(report{Message: fmt.Sprintf("Schema with id=%s successfully deleted", id)})
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
	metrics.DeletedSchemaMetricUpdate(id)
}

// DeleteSchemaVersion is a DELETE method that deletes a schema version.
// It expects the "id" and "version" of the wanted schema
//
// It currently gives the following responses:
//   - status 200 for a successful invocation along with an instance of the schema structure
//   - status 400 if the deletion caused an error
//   - status 404 if the schema version does not exist or is already deactivated
//
// @Title        Delete schema version by schema id and version
// @Summary      Delete schema version by schema id and version
// @Accept       json
// @Produce      json
// @Param        id path string true "schema id"
// @Param        version path string  true "version"
// @Success      200
// @Failure      400
// @Failure      404
// @Router       /schemas/{id}/versions/{version} [delete]
func (h Handler) DeleteSchemaVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	version := chi.URLParam(r, "version")

	deleted, err := h.Service.DeleteSchemaVersion(id, version)
	if err != nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	if !deleted {
		body, _ := json.Marshal(report{Message: fmt.Sprintf("Schema with id=%s and version=%s doesn't exist", id, version)})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusNotFound,
		})
		return
	}

	body, _ := json.Marshal(report{Message: fmt.Sprintf("Schema with id=%s and version=%s successfully deleted", id, version)})
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
	metrics.DeleteSchemaVersionMetricUpdate(id, version)
}

// HealthCheck is a GET method that gives the response status 200 to signalize
// that the Schema Registry component is up and running.
func (h Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h Handler) SchemaCompatibility(w http.ResponseWriter, r *http.Request) {
	compRequest, err := readSchemaCompatibilityRequest(r.Body)

	if err != nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(fmt.Errorf("couldn't close request body"))
			return
		}
	}(r.Body)

	compatible, err := h.Service.CheckCompatibility(compRequest.SchemaID, compRequest.NewSchema)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeResponse(w, responseBodyAndCode{
				Body: serializeErrorMessage(http.StatusText(http.StatusNotFound)),
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	if !compatible {
		body, _ := json.Marshal(insertInfo{
			Message: "Schemas are not compatible",
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusConflict,
		})
		return
	}

	body, _ := json.Marshal(compatible)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}

func (h Handler) SchemaValidity(w http.ResponseWriter, r *http.Request) {
	valRequest, err := readSchemaValidityRequest(r.Body)
	if err != nil {
		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusBadRequest)),
			Code: http.StatusBadRequest,
		})
		return
	}

	valid, err := h.Service.CheckValidity(valRequest.Format, valRequest.NewSchema, valRequest.Mode)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeResponse(w, responseBodyAndCode{
				Body: serializeErrorMessage(http.StatusText(http.StatusNotFound)),
				Code: http.StatusNotFound,
			})
			return
		}

		writeResponse(w, responseBodyAndCode{
			Body: serializeErrorMessage(http.StatusText(http.StatusInternalServerError)),
			Code: http.StatusInternalServerError,
		})
		return
	}

	if !valid {
		body, _ := json.Marshal(insertInfo{
			Message: "Schema is not valid",
		})
		writeResponse(w, responseBodyAndCode{
			Body: body,
			Code: http.StatusConflict,
		})
		return
	}

	body, _ := json.Marshal(valid)
	writeResponse(w, responseBodyAndCode{
		Body: body,
		Code: http.StatusOK,
	})
}
