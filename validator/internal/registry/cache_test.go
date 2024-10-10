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
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
)

func TestCacheGet(t *testing.T) {
	t.Run("get returns the correct result", func(t *testing.T) {
		sr := NewMock()
		c, err := newCache(sr, 10)
		if err != nil {
			t.Error(err)
		}

		id, version := "1", "1"
		schema := []byte("schema stored in the registry")
		sr.SetGetResponse(id, version, schema, nil)

		result, err := c.Get(context.Background(), id, version)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result, schema) {
			t.Error("result and actual not the same")
		}

		result1, err := c.Get(context.Background(), id, version)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result1, schema) {
			t.Error("cached result and actual not the same")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		sr := NewMock()
		c, err := newCache(sr, 10)
		if err != nil {
			t.Error(err)
		}

		id, version := "1", "1"
		sr.SetGetResponse(id, version, nil, errors.New("oops"))

		_, err = c.Get(context.Background(), id, version)
		if err == nil {
			t.Error("expected an error")
		}
	})
}
