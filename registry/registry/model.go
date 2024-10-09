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
	"time"
)

// Schema is a structure that defines the parent entity in the schema registry
type Schema struct {
	SchemaID          string           `json:"schema_id,omitempty"`
	SchemaType        string           `json:"schema_type"`
	Name              string           `json:"name"`
	VersionDetails    []VersionDetails `json:"schemas"`
	Description       string           `json:"description"`
	LastCreated       string           `json:"last_created"`
	PublisherID       string           `json:"publisher_id"`
	CompatibilityMode string           `json:"compatibility_mode"`
	ValidityMode      string           `json:"validity_mode"`
}

// VersionDetails represent the child entity in the schema registry model.
// The schema (specification) and version with some other details is set here.
type VersionDetails struct {
	VersionID          string    `json:"version_id,omitempty"`
	Version            string    `json:"version"`
	SchemaID           string    `json:"schema_id"`
	Specification      string    `json:"specification"`
	Description        string    `json:"description"`
	SchemaHash         string    `json:"schema_hash"`
	CreatedAt          time.Time `json:"created_at"`
	VersionDeactivated bool      `json:"version_deactivated"`
	Attributes         string    `json:"attributes"`
}

// SchemaRegistrationRequest contains information needed to register a schema.
type SchemaRegistrationRequest struct {
	Description       string `json:"description"`
	Specification     string `json:"specification"`
	Name              string `json:"name"`
	SchemaType        string `json:"schema_type"`
	LastCreated       string `json:"last_created"`
	PublisherID       string `json:"publisher_id"`
	CompatibilityMode string `json:"compatibility_mode"`
	ValidityMode      string `json:"validity_mode"`
	Attributes        string `json:"attributes"`
}

// SchemaUpdateRequest contains information needed to update a schema.
type SchemaUpdateRequest struct {
	Description   string `json:"description"`
	Specification string `json:"specification"`
	Attributes    string `json:"attributes"`
}

// SchemaCompatibilityRequest contains information needed to check compatibility of schemas
type SchemaCompatibilityRequest struct {
	SchemaID  string `json:"schema_id"`
	NewSchema string `json:"new_schema"`
}

// SchemaValidityRequest contains information needed to check validity of a schema
type SchemaValidityRequest struct {
	NewSchema string `json:"new_schema"`
	Format    string `json:"format"`
	Mode      string `json:"mode"`
}
