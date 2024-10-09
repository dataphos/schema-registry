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

package janitor

import (
	"context"
	"sync"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errcodes"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"

	"github.com/pkg/errors"
)

// Cleaner models a process which attempts to find the correct schema metadata for the given Message.
type Cleaner interface {
	// Clean takes a Message and tries to generate a new schema for the message, returning a copy of the original Message,
	// with the updated AttributeSchemaID and AttributeSchemaVersion properties.
	//
	// The returned errors are instances of OpError in case the caller needs additional information, like testing OpError.Deadletter.
	Clean(context.Context, Message) (Message, error)
}

// CachingCleaner combines the functionalities of SchemaGenerators, Validators, registry.SchemaRegistry in order
// to perform operations relating to the registration of newly generated schemas.
//
// This implementation of Cleaner tries to validate each new message against the last schema generated for that format.
// This is a heuristic chosen because invalid messages tend to come in bursts,
// so it's sane to assume multiple messages in sequence will have the same schema.
type CachingCleaner struct {
	Generators    SchemaGenerators
	Validators    Validators
	Registry      registry.SchemaRegistry
	LastGenerated map[string]SchemaInfo
	mu            sync.RWMutex
}

// SchemaInfo holds the schema and the schema registry related info.
type SchemaInfo struct {
	Schema  []byte
	ID      string
	Version string
}

// NewCachingCleaner returns a new CachingCleaner.
//
// Because of the caching mechanisms, CachingCleaner is NOT intended for concurrent use.
func NewCachingCleaner(generators SchemaGenerators, validators Validators, schemaRegistry registry.SchemaRegistry) *CachingCleaner {
	return &CachingCleaner{
		Generators:    generators,
		Validators:    validators,
		Registry:      schemaRegistry,
		LastGenerated: map[string]SchemaInfo{},
		mu:            sync.RWMutex{},
	}
}

// Clean implements Cleaner.
func (c *CachingCleaner) Clean(ctx context.Context, message Message) (Message, error) {
	c.mu.RLock()
	lastGeneratedOfFormat, ok := c.LastGenerated[message.Format]
	c.mu.RUnlock()
	if ok {
		isValid, err := c.Validators.Validate(message, lastGeneratedOfFormat.Schema)
		if err != nil {
			return Message{}, intoOpErr(message.ID, errcodes.ValidationFailure, err)
		}
		if isValid {
			return overwriteSchemaInfo(message, lastGeneratedOfFormat), nil
		}
	}

	schema, err := c.Generators.Generate(message)
	if err != nil {
		return Message{}, intoOpErr(message.ID, errcodes.SchemaGeneration, err)
	}
	id, version, err := c.registerOrUpdateSchema(ctx, schema, message.SchemaID, message.Format)
	if err != nil {
		return Message{}, intoOpErr(message.ID, errcodes.RegistryUnresponsive, err)
	}

	lastGeneratedOfFormat = SchemaInfo{
		Schema:  schema,
		ID:      id,
		Version: version,
	}
	c.mu.Lock()
	c.LastGenerated[message.Format] = lastGeneratedOfFormat
	c.mu.Unlock()

	return overwriteSchemaInfo(message, lastGeneratedOfFormat), nil
}

// registerOrUpdateSchema registers the schema if the given id is empty, or updates the schema
// under the given id.
func (c *CachingCleaner) registerOrUpdateSchema(ctx context.Context, schema []byte, id string, schemaType string) (string, string, error) {
	if id == "" {
		return c.Registry.Register(ctx, schema, schemaType, "none", "none")
	}

	version, err := c.Registry.Update(ctx, id, schema)
	if err != nil {
		return "", "", err
	}

	return id, version, nil
}

// overwriteSchemaInfo returns a new Message, identical to the one given, expect for the fields
// concerning the schema id and version.
func overwriteSchemaInfo(message Message, schemaInfo SchemaInfo) Message {
	message.RawAttributes[AttributeSchemaID] = schemaInfo.ID
	message.RawAttributes[AttributeSchemaVersion] = schemaInfo.Version

	message.SchemaID = schemaInfo.ID
	message.Version = schemaInfo.Version

	return message
}

// CleanerRouter wraps the Cleaner functionality with a Router.
type CleanerRouter struct {
	Cleaner Cleaner
	Router  Router
}

// CleanAndReroute attempts to clean the given Message, marking all successfully cleaned messages as Valid.
//
// In case the error returned by Cleaner.Clean evaluates OpError.Deadletter to true, the message is marked as Deadletter.
// All other errors are propagated as is.
func (cr CleanerRouter) CleanAndReroute(ctx context.Context, message Message) (MessageTopicPair, error) {
	cleaned, err := cr.Cleaner.Clean(ctx, message)
	if err != nil {
		var opError *OpError
		if errors.As(err, &opError) && opError.Deadletter() {
			return MessageTopicPair{Message: message, Topic: cr.Router.Route(Deadletter, message)}, nil
		}
		return MessageTopicPair{}, err
	}

	return MessageTopicPair{Message: cleaned, Topic: cr.Router.Route(Valid, message)}, nil
}
