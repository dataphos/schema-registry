package registry

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCacheGetSchemaVersionByIdAndVersion(t *testing.T) {
	repo := NewMockRepository()
	c, err := newCache(repo, 10)
	if err != nil {
		t.Error(err)
	}
	id, version := "1", "1"
	var storedSchema, cachedSchema VersionDetails
	// after the first call the schema should be stored in cache
	if storedSchema, err = c.GetSchemaVersionByIdAndVersion(id, version); err != nil {
		t.Error(err)
	}
	// second call returns schema from cache
	if cachedSchema, err = c.GetSchemaVersionByIdAndVersion(id, version); err != nil {
		t.Error(err)
	}

	if c.cache.Len() == 0 {
		t.Error("Schema unsuccessfully stored in cache.")
	}
	if !cmp.Equal(cachedSchema, storedSchema) {
		t.Error("Cached schema differs from the stored schema")
	}
}

func TestCacheDeleteSchemaVersion(t *testing.T) {
	repo := NewMockRepository()
	c, err := newCache(repo, 10)
	if err != nil {
		t.Error(err)
	}

	id, version := "1", "1"
	arrKey := [2]string{id, version}
	VersionDetails := MockVersionDetails(id, version)
	c.cache.Add(arrKey, VersionDetails)

	if _, err = c.DeleteSchemaVersion(id, version); err != nil {
		t.Error(err)
	}
	if _, bool := c.cache.Get(arrKey); bool {
		t.Error("Schema is still stored in cache")
	}
}

func TestDeleteSchema(t *testing.T) {
	repo := NewMockRepository()
	c, err := newCache(repo, 10)
	if err != nil {
		t.Error(err)
	}
	id := "mocking"
	schema := MockSchema(id)

	for i := 1; i <= 10; i++ {
		k := strconv.Itoa(i)
		VersionDetails := MockVersionDetails(k, k)
		arrKey := [2]string{id, k}
		c.cache.Add(arrKey, VersionDetails)
		schema.VersionDetails = append(schema.VersionDetails, VersionDetails)
	}
	repo.SetGetSchemaVersionsByIdResponse(id, schema, nil)
	if bool, err := c.DeleteSchema(id); err != nil {
		t.Error(err)
	} else {
		if c.cache.Len() != 0 {
			t.Error("Some schemas are still stored in cache")
		} else if !bool {
			t.Error("Schema does not exist")
		}
	}

}