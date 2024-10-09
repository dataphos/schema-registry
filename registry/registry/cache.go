package registry

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/cast"
	"golang.org/x/sync/singleflight"
)

// cached decorates Service with a lru cache.
type cached struct {
	Repository
	cache *lru.TwoQueueCache
	group singleflight.Group
}

// newCache returns a new cached.
func newCache(repository Repository, size int) (*cached, error) {
	cache, err := lru.New2Q(size)
	if err != nil {
		return nil, err
	}

	return &cached{
		Repository: repository,
		cache:      cache,
		group:      singleflight.Group{},
	}, nil
}

// GetSchemaVersionByIdAndVersion overrides the Repository.GetSchemaVersionByIdAndVersion method, caching each call to the underlying Repository, while also
// making sure there's only one inflight request for the same key (if multiple goroutines request the same schema,
// only one request is actually sent down, the rest wait for the first one to share its result).
func (c *cached) GetSchemaVersionByIdAndVersion(id, version string) (VersionDetails, error) {
	// this should be faster than string concatenation
	arrKey := [2]string{id, version}
	var err error
	v, ok := c.cache.Get(arrKey)
	if !ok {
		// cache miss, we need a string version of the key to satisfy the singleflight.Group method signature
		key := id + "_" + version
		v, err, _ = c.group.Do(key, func() (interface{}, error) {
			if v, err = c.Repository.GetSchemaVersionByIdAndVersion(id, version); err != nil {
				return VersionDetails{}, err
			}
			c.cache.Add(arrKey, v)
			return v, err
		})
	}
	return v.(VersionDetails), err
}

// DeleteSchemaVersion overrides the Repository.DeleteSchemaVersion method, caching each call to the underlying Repository, while also
// making sure there's only one inflight request for the same key (if multiple goroutines request the deletion of the same
// Schema version, only one request is actually sent down, the rest wait for the first one to share its result).
func (c *cached) DeleteSchemaVersion(id, version string) (bool, error) {
	key := id + "_" + version
	bool, err, _ := c.group.Do(key, func() (interface{}, error) {
		bool, err := c.Repository.DeleteSchemaVersion(id, version)
		if err != nil {
			return false, err
		}

		arrKey := [2]string{id, version}
		c.cache.Remove(arrKey)
		return bool, err
	})
	return cast.ToBool(bool), err
}

// DeleteSchema overrides the Repository.DeleteSchema method, caching each call to the underlying Repository, while also
// making sure there's only one inflight request for the same key (if multiple goroutines request the deletion of the same
// Schema, only one request is actually sent down, the rest wait for the first one to share its result).
func (c *cached) DeleteSchema(id string) (bool, error) {
	v, err := c.Repository.GetSchemaVersionsById(id)
	if err == nil {
		//Schema with the given ID exists, and it's not already deactivated
		bool, err, _ := c.group.Do(id, func() (interface{}, error) {
			bool, err := c.Repository.DeleteSchema(id)
			if err != nil {
				return false, err
			}
			// remove schemas that are present in cache
			for _, v := range v.VersionDetails {
				arrKey := [2]string{v.SchemaID, v.Version}
				c.cache.Remove(arrKey)
			}
			return bool, nil
		})
		return cast.ToBool(bool), err
	}
	return false, nil
}
