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
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")
var ErrUnknownComp = errors.New("unknown value for compatibility_mode")
var ErrUnknownVal = errors.New("unknown value for validity mode")
var ErrNotValid = errors.New("schema is not valid")
var ErrNotComp = errors.New("schemas are not compatible")
var ErrInvalidValueHeader = errors.New("invalid header value")

type Repository interface {
	CreateSchema(schemaRegisterRequest SchemaRegistrationRequest) (VersionDetails, bool, error)
	GetSchemaVersionByIdAndVersion(id string, version string) (VersionDetails, error)
	UpdateSchemaById(id string, schemaUpdateRequest SchemaUpdateRequest) (VersionDetails, bool, error)
	GetSchemaVersionsById(id string) (Schema, error)
	GetAllSchemaVersions(id string) (Schema, error)
	GetLatestSchemaVersion(id string) (VersionDetails, error)
	DeleteSchema(id string) (bool, error)
	DeleteSchemaVersion(id, version string) (bool, error)
	GetAllSchemas() ([]Schema, error)
	GetSchemas() ([]Schema, error)
}

// WithCache decorates the given Repository with an in-memory cache of the given size.
func WithCache(repository Repository, size int) (Repository, error) {
	return newCache(repository, size)
}
