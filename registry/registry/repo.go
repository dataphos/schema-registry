package registry

import (
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")
var ErrUnknownComp = errors.New("unknown value for compatibility_mode")
var ErrUnknownVal = errors.New("unknown value for validity mode")
var ErrNotValid = errors.New("schema is not valid")
var ErrNotComp = errors.New("schemas are not compatible")

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
