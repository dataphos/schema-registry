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

package registry

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/hamba/avro/v2"
	"github.com/pkg/errors"

	"github.com/dataphos/aquarium-janitor-standalone-sr/compatibility"
	"github.com/dataphos/aquarium-janitor-standalone-sr/validity"
)

type Service struct {
	Repository     Repository
	CompChecker    compatibility.Checker
	ValChecker     validity.Checker
	GlobalCompMode string
	GlobalValMode  string
}

// Attribute search depth limit to prevent infinite recursion
const attSearchDepth = 10

const cacheSizeEnv = "CACHE_SIZE"
const defaultCacheSizeEnv = 0

type QueryParams struct {
	Id         string
	Version    string
	SchemaType string
	Name       string
	OrderBy    string
	Sort       string
	Limit      int
	Attributes []string
}

func New(Repository Repository, CompChecker compatibility.Checker, ValChecker validity.Checker, GlobalCompMode, GlobalValMode string) *Service {
	var size int
	var err error
	cacheSize := os.Getenv(cacheSizeEnv)
	if cacheSize == "" {
		size = defaultCacheSizeEnv
	} else {
		size, err = strconv.Atoi(cacheSize)
		if err != nil {
			log.Println("Cache size can not hold non-numeric data.")
			return &Service{}
		}
	}

	if size > 0 {
		log.Println("Using in-memory cache for repository")
		Repository, err = WithCache(Repository, size)
		if err != nil {
			log.Println("Encountered error while trying to create cache.")
			return &Service{}
		}
	}

	return &Service{
		Repository:     Repository,
		CompChecker:    CompChecker,
		ValChecker:     ValChecker,
		GlobalCompMode: GlobalCompMode,
		GlobalValMode:  GlobalValMode,
	}
}

// GetSchemaVersion gets the schema version with the specific id and version.
func (service *Service) GetSchemaVersion(id, version string) (VersionDetails, error) {
	return service.Repository.GetSchemaVersionByIdAndVersion(id, version)
}

// ListSchemaVersions lists all active schema versions of a specific schema.
func (service *Service) ListSchemaVersions(id string) (Schema, error) {
	return service.Repository.GetSchemaVersionsById(id)
}

// ListAllSchemaVersions lists all schema versions of a specific schema.
func (service *Service) ListAllSchemaVersions(id string) (Schema, error) {
	return service.Repository.GetAllSchemaVersions(id)
}

// GetLatestSchemaVersion gets the latest version of a certain schema.
func (service *Service) GetLatestSchemaVersion(id string) (VersionDetails, error) {
	return service.Repository.GetLatestSchemaVersion(id)
}

// GetSchemas gets all active schemas.
func (service *Service) GetSchemas() ([]Schema, error) {
	return service.Repository.GetSchemas()
}

// GetAllSchemas gets all schemas.
func (service *Service) GetAllSchemas() ([]Schema, error) {
	return service.Repository.GetAllSchemas()
}

// SearchSchemas gets filtered schemas.
func (service *Service) SearchSchemas(params QueryParams) ([]Schema, error) {
	schemas, err := service.Repository.GetSchemas()
	if err != nil {
		return schemas, errors.Wrap(err, "couldn't retrieve schemas")
	}

	var filteredSchemas []Schema
	for _, schema := range schemas {
		if params.Id != "" && schema.SchemaID != params.Id {
			continue
		}
		if params.Name != "" && !strings.Contains(schema.Name, params.Name) {
			continue
		}
		if params.SchemaType != "" && schema.SchemaType != params.SchemaType {
			continue
		}

		filteredVersions := Schema{
			SchemaID:          schema.SchemaID,
			SchemaType:        schema.SchemaType,
			Name:              schema.Name,
			Description:       schema.Description,
			LastCreated:       schema.LastCreated,
			PublisherID:       schema.PublisherID,
			CompatibilityMode: schema.CompatibilityMode,
			ValidityMode:      schema.ValidityMode,
		}
		for _, detail := range schema.VersionDetails {
			if params.Version != "" && detail.Version != params.Version {
				continue
			}
			if !containsAttributes(detail, params.Attributes) {
				continue
			}
			filteredVersions.VersionDetails = append(filteredVersions.VersionDetails, detail)
		}
		if len(filteredVersions.VersionDetails) > 0 {
			if params.OrderBy == "version" && len(filteredVersions.VersionDetails) > 1 {
				if params.Sort == "asc" {
					sort.Slice(filteredVersions.VersionDetails, func(i, j int) bool {
						return filteredVersions.VersionDetails[i].Version < filteredVersions.VersionDetails[j].Version
					})
				} else if params.Sort == "desc" {
					sort.Slice(filteredVersions.VersionDetails, func(i, j int) bool {
						return filteredVersions.VersionDetails[i].Version > filteredVersions.VersionDetails[j].Version
					})
				}
			}
			filteredSchemas = append(filteredSchemas, filteredVersions)
		}
	}

	switch params.OrderBy {
	case "name":
		if params.Sort == "asc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				return filteredSchemas[i].Name < filteredSchemas[j].Name
			})
		} else if params.Sort == "desc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				return filteredSchemas[i].Name > filteredSchemas[j].Name
			})
		}
	case "id":
		if params.Sort == "asc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				l1, l2 := len(filteredSchemas[i].SchemaID), len(filteredSchemas[j].SchemaID)
				if l1 != l2 {
					return l1 < l2
				}
				return filteredSchemas[i].SchemaID < filteredSchemas[j].SchemaID
			})
		} else if params.Sort == "desc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				l1, l2 := len(filteredSchemas[i].SchemaID), len(filteredSchemas[j].SchemaID)
				if l1 != l2 {
					return l1 > l2
				}
				return filteredSchemas[i].SchemaID > filteredSchemas[j].SchemaID
			})
		}
	case "type":
		if params.Sort == "asc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				return filteredSchemas[i].SchemaType < filteredSchemas[j].SchemaType
			})
		} else if params.Sort == "desc" {
			sort.Slice(filteredSchemas, func(i, j int) bool {
				return filteredSchemas[i].SchemaType > filteredSchemas[j].SchemaType
			})
		}
	}

	if params.Limit > 0 && params.Limit < len(filteredSchemas) {
		filteredSchemas = filteredSchemas[:params.Limit]
	}
	return filteredSchemas, nil
}

func containsAttributes(details VersionDetails, attributes []string) bool {
	numMatched := 0
	for i, filterAtt := range attributes {
		for _, att := range strings.FieldsFunc(details.Attributes, func(r rune) bool { return r == '/' || r == ',' }) {
			if filterAtt == att {
				numMatched += 1
				break
			}
		}
		if numMatched != i+1 {
			return false
		}
	}
	return true
}

// CreateSchema creates a new schema.
func (service *Service) CreateSchema(schemaRegisterRequest SchemaRegistrationRequest) (VersionDetails, bool, error) {
	if !compatibility.CheckIfValidMode(&schemaRegisterRequest.CompatibilityMode) {
		return VersionDetails{}, false, ErrUnknownComp
	}
	if !validity.CheckIfValidMode(&schemaRegisterRequest.ValidityMode) {
		return VersionDetails{}, false, ErrUnknownVal
	}
	valid, err := service.CheckValidity(schemaRegisterRequest.SchemaType, schemaRegisterRequest.Specification, schemaRegisterRequest.ValidityMode)
	if err != nil {
		return VersionDetails{}, false, err
	}
	if !valid {
		return VersionDetails{}, false, ErrNotValid
	}
	//cannot canonicalize schema that is invalid
	if strings.ToLower(schemaRegisterRequest.ValidityMode) == "syntax-only" || strings.ToLower(schemaRegisterRequest.ValidityMode) == "full" {
		canonicalSpec, err := canonicalizeSchema([]byte(schemaRegisterRequest.Specification), strings.ToLower(schemaRegisterRequest.SchemaType))
		if err != nil {
			return VersionDetails{}, false, err
		}
		schemaRegisterRequest.Specification = canonicalSpec
	}

	attributes, err := extractAttributes(schemaRegisterRequest.Specification, strings.ToLower(schemaRegisterRequest.SchemaType), attSearchDepth)
	if err != nil {
		return VersionDetails{}, false, errors.Wrap(err, "unable to extract attributes")
	}
	schemaRegisterRequest.Attributes = attributes

	return service.Repository.CreateSchema(schemaRegisterRequest)
}

// canonicalizeSchema converts the given schema to its canonical form
func canonicalizeSchema(specification []byte, schemaType string) (string, error) {
	switch schemaType {
	case "json":
		var canonicalSpec []byte
		var schema map[string]interface{}
		err := json.Unmarshal(specification, &schema)
		if err != nil {
			return "", err
		}

		required, ok := schema["required"].([]interface{})
		if ok {
			sort.Slice(required, func(i, j int) bool {
				return required[i].(string) < required[j].(string)
			})
			schema["required"] = required
			specification, err = json.Marshal(schema)
			if err != nil {
				return "", err
			}
		}

		canonicalSpec, err = jsoncanonicalizer.Transform(specification)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(canonicalSpec)), nil
	case "avro":
		schema, err := avro.Parse(string(specification))
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(schema.String()), nil
	default:
		return strings.TrimSpace(string(specification)), nil
	}
}

func extractAttributes(specification string, schemaType string, maxDepth int) (string, error) {
	switch schemaType {
	case "json":
		var schema map[string]interface{}
		err := json.Unmarshal([]byte(specification), &schema)
		if err != nil {
			return "", errors.Wrap(err, "couldn't unmarshal schema")
		}

		flatSchema := make(map[string]interface{})
		err = flattenJSON("", schema, flatSchema, 0, maxDepth, "/")
		if err != nil {
			return "", errors.Wrap(err, "unable to flatten json schema")
		}

		var allAttributes string
		for _, att := range reflect.ValueOf(flatSchema).MapKeys() {
			allAttributes += att.String() + ","
		}

		if len(allAttributes) > 0 {
			return allAttributes[:len(allAttributes)-1], nil
		} else {
			return "", nil
		}
	default:
		return "", nil
	}
}

func flattenJSON(prefix string, nested interface{}, flat map[string]interface{}, currentDepth int, maxDepth int, delimiter string) error {
	if currentDepth >= maxDepth {
		flat[prefix] = nested
		return nil
	}

	switch nested.(type) { //nolint:gosimple // fine here
	case map[string]interface{}:
		for k, v := range nested.(map[string]interface{}) { //nolint:gosimple // fine here
			newKey := k
			if currentDepth == 0 && strings.ToLower(newKey) != "properties" {
				continue
			}
			if currentDepth != 0 {
				newKey = prefix + delimiter + newKey
			}
			err := flattenJSON(newKey, v, flat, currentDepth+1, maxDepth, delimiter)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range nested.([]interface{}) { //nolint:gosimple // fine here
			newKey := strconv.Itoa(i)
			if currentDepth != 0 {
				newKey = prefix + delimiter + newKey
			}
			err := flattenJSON(newKey, v, flat, currentDepth+1, maxDepth, delimiter)
			if err != nil {
				return err
			}
		}
	default:
		flat[prefix] = nested
	}
	return nil
}

// UpdateSchema updates the schemas by assigning a new version to it.
func (service *Service) UpdateSchema(id string, schemaUpdateRequest SchemaUpdateRequest) (VersionDetails, bool, error) {
	schemas, err := service.ListSchemaVersions(id)
	if err != nil {
		return VersionDetails{}, false, err
	}

	valid, err := service.CheckValidity(schemas.SchemaType, schemaUpdateRequest.Specification, schemas.ValidityMode)
	if err != nil {
		return VersionDetails{}, false, err
	}
	if !valid {
		return VersionDetails{}, false, ErrNotValid
	}

	compatible, err := service.CheckCompatibility(schemaUpdateRequest.Specification, id)
	if err != nil {
		return VersionDetails{}, false, err
	}
	if !compatible {
		return VersionDetails{}, false, ErrNotComp
	}
	if strings.ToLower(schemas.ValidityMode) == "syntax-only" || strings.ToLower(schemas.ValidityMode) == "full" {
		canonicalSpec, err := canonicalizeSchema([]byte(schemaUpdateRequest.Specification), strings.ToLower(schemas.SchemaType))
		if err != nil {
			return VersionDetails{}, false, err
		}
		schemaUpdateRequest.Specification = canonicalSpec

	}

	attributes, err := extractAttributes(schemaUpdateRequest.Specification, schemas.SchemaType, attSearchDepth)
	if err != nil {
		return VersionDetails{}, false, errors.Wrap(err, "unable to extract attributes")
	}
	schemaUpdateRequest.Attributes = attributes

	return service.Repository.UpdateSchemaById(id, schemaUpdateRequest)
}

// DeleteSchema deletes the schema and its versions.
func (service *Service) DeleteSchema(id string) (bool, error) {
	return service.Repository.DeleteSchema(id)
}

// DeleteSchemaVersion deletes a specific version of a schema.
func (service *Service) DeleteSchemaVersion(id, version string) (bool, error) {
	return service.Repository.DeleteSchemaVersion(id, version)
}

// CheckCompatibility checks if schemas are compatible
func (service *Service) CheckCompatibility(newSchema, id string) (bool, error) {
	schemas, err := service.ListSchemaVersions(id)
	if err != nil {
		return false, err
	}

	jsonAttrs := make(map[string]string)
	jsonAttrs["id"] = id
	jsonAttrs["format"] = schemas.SchemaType
	jsonAttrs["schema"] = newSchema
	jsonMessage, err := json.Marshal(jsonAttrs)
	if err != nil {
		return false, err
	}

	var stringHistory []string
	for _, el := range schemas.VersionDetails {
		stringHistory = append(stringHistory, el.Specification)
	}
	mode := schemas.CompatibilityMode
	if schemas.CompatibilityMode == "" {
		mode = service.GlobalCompMode
	}

	return service.CompChecker.Check(string(jsonMessage), stringHistory, mode)
}

// CheckValidity checks if a schema is valid
func (service *Service) CheckValidity(schemaType, newSchema, mode string) (bool, error) {
	if mode == "" {
		mode = service.GlobalValMode
	}
	return service.ValChecker.Check(newSchema, schemaType, mode)
}
