// Package registry exposes common functionalities of all schema registries.
package registry

import (
	"context"

	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("no schema registered under given id and version")

// SchemaRegistry models schema registries.
type SchemaRegistry interface {
	// Get returns the schema stored under the given id and version.
	// If no schema exists, ErrNotFound must be returned.
	Get(ctx context.Context, id, version string) ([]byte, error)
	// Get(ctx context.Context, id, version string) ([]byte, error)

	// GetLatest returns the whole schema, including the metadata and all versions
	// If no schema exists under specified id, ErrNotFound must be returned.
	GetLatest(ctx context.Context, id string) ([]byte, error)

	// Register register a new schema and returns the id and version it was registered under.
	Register(ctx context.Context, schema []byte, schemaType, compMode, valMode string) (string, string, error)
	// Register(ctx context.Context, schema []byte, schemaType, compMode, valMode string) (string, string, error)

	// Update updates the schema stored under the given id, returns the version it was registered under.
	Update(ctx context.Context, id string, schema []byte) (string, error)
	// Update(ctx context.Context, id string, schema []byte) (string, error)
}

// WithCache decorates the given SchemaRegistry with an in-memory cache of the given size.
func WithCache(registry SchemaRegistry, size int) (SchemaRegistry, error) {
	return newCache(registry, size)
}
