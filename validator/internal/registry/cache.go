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
	"context"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/singleflight"
)

var cachedHitsCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "schema_registry",
	Name:      "schemas_cache_hits_total",
	Help:      "The total number of schemas cache hits",
})

// cached decorates SchemaRegistry with a lru cache.
type cached struct {
	SchemaRegistry
	cache *lru.TwoQueueCache
	group singleflight.Group
}

// newCache returns a new cached.
func newCache(registry SchemaRegistry, size int) (*cached, error) {
	cache, err := lru.New2Q(size)
	if err != nil {
		return nil, err
	}

	return &cached{
		SchemaRegistry: registry,
		cache:          cache,
		group:          singleflight.Group{},
	}, nil
}

// Get overrides the SchemaRegistry.Get method, caching each call to the underlying SchemaRegistry, while also
// making sure there's only one inflight request for the same key (if multiple goroutines request the same schema,
// only one request is actually sent down, the rest wait for the first one to share its result).
func (c *cached) Get(ctx context.Context, id, version string) ([]byte, error) {
	// this should be faster than string concatenation
	arrKey := [2]string{id, version}

	if v, ok := c.cache.Get(arrKey); ok {
		// cache hit
		cachedHitsCount.Inc()
		return v.([]byte), nil
	}

	// cache miss, we need a string version of the key to satisfy the singleflight.Group method signature
	key := id + "_" + version

	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		schema, err := c.SchemaRegistry.Get(ctx, id, version)
		if err != nil {
			return nil, err
		}

		c.cache.Add(arrKey, schema)

		return schema, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}
