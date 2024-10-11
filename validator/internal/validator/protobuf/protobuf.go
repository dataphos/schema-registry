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

package protobuf

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/dataphos/schema-registry-validator/internal/validator"

	lru "github.com/hashicorp/golang-lru"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

type Validator struct {
	Dir   string
	group singleflight.Group
	cache *lru.TwoQueueCache
}

// New returns a new instance of a protobuf validator.Validator.
//
// Since the validator needs to write to disk, a path to the used directory is needed, as well
// as a cache size which will be used to avoid writing to disk for each validation request.
func New(dir string, cacheSize int) (validator.Validator, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	cache, err := lru.New2Q(cacheSize)
	if err != nil {
		return nil, err
	}

	return &Validator{
		Dir:   dir,
		cache: cache,
	}, nil
}

func (v *Validator) Validate(message, schema []byte, id, version string) (bool, error) {
	filename := id + "_" + version + ".txt"
	descriptor, err := v.getMainMessageDescriptor(filename, schema)
	if err != nil {
		return false, err
	}

	if err = descriptor.Unmarshal(message); err != nil {
		return false, nil
	}

	if hasUnknownFields(descriptor) {
		return false, nil
	}
	if err = descriptor.ValidateRecursive(); err != nil {
		return false, nil
	}

	return true, nil
}

// getMainMessageDescriptor returns a fresh dynamic.Message instance for the given schema.
//
// Because the used libraries require schemas to be read from disk, a lru cache is used to avoid I/O operations
// for each validation request. This in turn means every cache miss requires checking if the schema is stored to disk,
// (storing it if necessary), then reading and caching it.
func (v *Validator) getMainMessageDescriptor(filename string, schema []byte) (*dynamic.Message, error) {
	var descriptor *desc.FileDescriptor
	var err error

	path := filepath.Join(v.Dir, filename)
	// try to retrieve the processed .proto message from the cache
	val, ok := v.cache.Get(path)
	if !ok {
		// if it isn't in the cache, check if it is already written to disk
		if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
			if err = v.writeSchemaToFile(path, schema); err != nil {
				return nil, err
			}
		}

		// now we can load the written .proto schema into a message descriptor
		descriptor, err = loadSchemaIntoDescriptor(v.Dir, filename)
		if err != nil {
			return nil, err
		}

		v.cache.Add(path, descriptor)
	} else {
		descriptor = val.(*desc.FileDescriptor)
	}

	return parseDescriptor(descriptor)
}

// writeSchemaToFile writes the given schema under the given path.
//
// A singleflight.Group is used to ensure concurrent request for the same schema only write
// the schema once (I/O is expensive).
func (v *Validator) writeSchemaToFile(path string, schema []byte) error {
	_, err, _ := v.group.Do(path, func() (interface{}, error) {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		w := bufio.NewWriter(f)
		if _, err = w.Write(schema); err != nil {
			return nil, err
		}
		if err = w.Flush(); err != nil {
			return nil, err
		}

		return nil, f.Close()
	})
	if err != nil {
		return err
	}

	return nil
}

// loadSchemaIntoDescriptor retrieves a file descriptor of a .proto file stored under filename,
// under the given import path.
func loadSchemaIntoDescriptor(importPath, filename string) (*desc.FileDescriptor, error) {
	parser := protoparse.Parser{
		ImportPaths: []string{importPath},
	}
	fileDescriptors, err := parser.ParseFiles(filename)
	if err != nil {
		return nil, err
	}

	return fileDescriptors[0], nil
}

// parseDescriptor parses the given file descriptor into a dynamic.Message instance, returning an error
// if there are multiple top-level messages defined in the given descriptor.
func parseDescriptor(descriptor *desc.FileDescriptor) (*dynamic.Message, error) {
	messageDescriptors := descriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return nil, errors.Wrap(validator.ErrDeadletter, "no message definitions were found in the .proto file")
	}
	if len(messageDescriptors) > 1 {
		return nil, errors.Wrap(validator.ErrDeadletter, ".proto file must have exactly 1 top level message")
	}
	return dynamic.NewMessage(messageDescriptors[0]), nil
}

// hasUnknownFields recursively checks for unknown fields of the given dynamic.Message.
func hasUnknownFields(message *dynamic.Message) bool {
	if len(message.GetUnknownFields()) > 0 {
		return true
	}

	for _, v := range message.GetKnownFields() {
		field := message.GetField(v)
		if fieldMessage, ok := field.(*dynamic.Message); ok {
			if hasUnknownFields(fieldMessage) {
				return true
			}
		}
	}

	return false
}
