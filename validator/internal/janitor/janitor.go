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

// Package janitor offers a suite of functions for collecting message schemas,
// validating the messages based on the collected schemas, and publishing them to a destination topic.
package janitor

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dataphos/schema-registry-validator/internal/errcodes"
	"github.com/dataphos/schema-registry-validator/internal/errtemplates"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	"github.com/dataphos/schema-registry-validator/internal/validator"

	"github.com/pkg/errors"

	"github.com/dataphos/lib-brokers/pkg/broker"
)

// Message defines a Message used for processing broker messages.
// Essentially, Message decorates broker messages with additional, extracted information.
type Message struct {
	ID               string
	Key              string
	RawAttributes    map[string]interface{}
	Payload          []byte
	IngestionTime    time.Time
	SchemaID         string
	Version          string
	Format           string
	HeaderValidation bool
}

const (
	// AttributeSchemaID is one of the keys expected to be found in the attributes field of the message.
	// It holds the schema id information concerning the data field of the message
	AttributeSchemaID = "schemaId"

	// AttributeSchemaVersion is one of the keys expected to be found in the attributes field of the message,
	// It holds the schema version information concerning the data field of the message.
	AttributeSchemaVersion = "versionId"

	// AttributeFormat is one of the keys expected to be found in the attributes field of the message.
	// It holds the format of the data field of the message.
	AttributeFormat = "format"

	// HeaderValidation is one of the keys that can occur in raw attributes section of header.
	// It determines if the header will be validated
	HeaderValidation = "headerValidation"

	// AttributeHeaderID is one of the keys expected in raw attributes section of header, but only if HeaderValidation is true.
	// It holds the header's schema id that is used to check header validity
	AttributeHeaderID = "headerSchemaId"

	// AttributeHeaderVersion is one of the keys expected in raw attributes section of header, but only if HeaderValidation is true.
	// It holds the header's schema version that is used to check header validity
	AttributeHeaderVersion = "headerVersionId"
)

// MessageSchemaPair wraps a Message with the Schema relating to this Message.
type MessageSchemaPair struct {
	Message Message
	Schema  []byte
}

// CollectSchema retrieves the schema with the given id and version from registry.SchemaRegistry.
//
// If schema retrieval results in registry.ErrNotFound, or id or version are an empty string,
// the Message is put on the results channel with MessageSchemaPair.Schema set to nil.
//
// The returned error is an instance of OpError for improved error handling (so that the source of this error is identifiable
// even if combined with other errors).
func CollectSchema(ctx context.Context, id string, version string, schemaRegistry registry.SchemaRegistry) ([]byte, error) {
	if id == "" || version == "" {
		return nil, nil
	}

	schema, err := schemaRegistry.Get(ctx, id, version)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			return nil, intoOpErr(id, errcodes.SchemaNotRegistered, err)
		} else if errors.Is(err, registry.InvalidHeader) {
			return nil, intoOpErr(id, errcodes.InvalidHeader, err)
		}
		return nil, intoOpErr(id, errcodes.RegistryUnresponsive, err)
	}

	return schema, nil
}

// Validators is a convenience type for a map containing validator.Validator instances for available message formats.
type Validators map[string]validator.Validator

// Validate wraps the same function of validator.Validator, by first selecting the proper validator, and then using that
// validator to determine the validity of the given Message.Payload under this schema.
//
// Returns an error if Validators doesn't contain a validator instance for the message format.
func (vs Validators) Validate(message Message, schema []byte) (bool, error) {
	v, ok := vs[strings.ToLower(message.Format)]
	if !ok {
		return false, errors.WithMessage(validator.ErrUnsupportedFormat, errtemplates.UnsupportedMessageFormat(message.Format).Error())
	}
	return v.Validate(message.Payload, schema, message.SchemaID, message.Version)
}

func GetHeaderIdAndVersion(message Message) (string, string, error) {
	var id string
	var version string
	var ok bool

	if id, ok = message.RawAttributes[AttributeHeaderID].(string); !ok {
		err := errors.New("missing header ID")
		message.RawAttributes["deadLetterErrorCategory"] = "Missing header ID"
		message.RawAttributes["deadLetterErrorReason"] = err
		return "", "", intoOpErr(message.ID, errcodes.MissingHeader, err)
	}
	if version, ok = message.RawAttributes[AttributeHeaderVersion].(string); !ok {
		err := errors.New("missing header version")
		message.RawAttributes["deadLetterErrorCategory"] = "Missing header version"
		message.RawAttributes["deadLetterErrorReason"] = err
		return "", "", intoOpErr(message.ID, errcodes.MissingHeader, err)
	}
	return id, version, nil
}

func ValidateHeader(message Message, schema []byte, validators Validators) (bool, error) {
	if len(schema) == 0 {
		errMissingHeaderSchema := errors.WithMessage(validator.ErrMissingHeaderSchema, "")
		message.RawAttributes["deadLetterErrorCategory"] = "Header schema error"
		message.RawAttributes["deadLetterErrorReason"] = errMissingHeaderSchema.Error()
		return false, nil
	}
	headerData, err := generateHeaderData(message.RawAttributes)
	if err != nil {
		message.RawAttributes["deadLetterErrorCategory"] = "Header schema error"
		message.RawAttributes["deadLetterErrorReason"] = err.Error()
		return false, err
	}

	headerSchemaId, ok := message.RawAttributes[AttributeHeaderID].(string)
	if !ok {
		return false, errors.New("header ID is not in a supported format")
	}
	headerSchemaVersion, ok := message.RawAttributes[AttributeHeaderVersion].(string)
	if !ok {
		return false, errors.New("header version is not in a supported format")
	}

	isValid, err := validators["json"].Validate(headerData, schema, headerSchemaId, headerSchemaVersion)
	if err != nil {
		if errors.Is(err, validator.ErrFailedValidation) {
			message.RawAttributes["deadLetterErrorCategory"] = "Header validation error"
			message.RawAttributes["deadLetterErrorReason"] = errors.WithMessage(validator.ErrFailedHeaderValidation, err.Error())
			return false, nil
		} else if errors.Is(err, validator.ErrDeadletter) {
			return false, nil
		}
		return false, intoOpErr(message.ID, errcodes.ValidationFailure, err)
	}

	if isValid {
		return true, nil
	}
	return false, nil
}

func generateHeaderData(rawAttributes map[string]interface{}) ([]byte, error) {
	cleanAttributes := make(map[string]interface{})
	for key, value := range rawAttributes {
		if key == HeaderValidation || key == AttributeHeaderID || key == AttributeHeaderVersion ||
			key == AttributeSchemaID || key == AttributeSchemaVersion || key == AttributeFormat {
			continue
		} else {
			cleanAttributes[key] = value
		}
	}
	headerData, err := json.Marshal(cleanAttributes)
	if err != nil {
		return nil, err
	}
	return headerData, nil
}

// SchemaGenerators is a convenience type for a map containing schemagen.Generator instances for available message formats.
type SchemaGenerators map[string]schemagen.Generator

// Generate wraps the same function of schemagen.Generator, by first selecting the proper generator, and then using that
// generator to construct a schema from the given Parsed instance.
//
// Returns an error if SchemaGenerators doesn't contain a generator instance for the MessageFormat of the given Parsed instance.
func (gs SchemaGenerators) Generate(message Message) ([]byte, error) {
	generator, ok := gs[message.Format]
	if !ok {
		return nil, errtemplates.UnsupportedMessageFormat(message.Format)
	}
	return generator.Generate(message.Payload)
}

// Router determines where should the messages be sent to.
type Router interface {
	Route(Result, Message) string
}

// RoutingFunc convenience type to allow functions to implement Router directly.
type RoutingFunc func(Result, Message) string

func (f RoutingFunc) Route(result Result, message Message) string {
	return f(result, message)
}

// Result holds the four possible outcomes concerning with routing messages to some destination topic: Valid, Invalid, Deadletter and MissingSchema.
// Valid, Invalid and Deadletter are possible outcomes of message validation, while MissingSchema occurs if there is no record
// of the Schema in the Schema Registry.
type Result int

const (
	Valid Result = iota
	Invalid
	Deadletter
	MissingSchema
)

// MessageTopicPair wraps a Message with the Topic the Message is supposed to be sent to.
type MessageTopicPair struct {
	Message Message
	Topic   string
}

// InferDestinationTopic infers the destination topic for the given MessageSchemaPair.
//
// In case MessageSchemaPair.Schema is empty, MissingSchema is passed onto the given Router to
// infer the destination topic.
//
// If the schema exists, the message is validated against it, and the Result is passed onto the Router
// to infer the destination topic. In case validation returns validator.ErrDeadletter, Deadletter is passed onto the Router.
//
// The returned error is an instance of OpError for improved error handling (so that the source of this error is identifiable
// even if combined with other errors).
func InferDestinationTopic(messageSchemaPair MessageSchemaPair, validators Validators, router Router) (MessageTopicPair, error) {
	message, schema := messageSchemaPair.Message, messageSchemaPair.Schema

	if len(schema) == 0 {
		errMissingSchema := errors.WithMessage(validator.ErrMissingSchema, "")
		message.RawAttributes["deadLetterErrorCategory"] = "Schema error"
		message.RawAttributes["deadLetterErrorReason"] = errMissingSchema.Error()
		return MessageTopicPair{Message: message, Topic: router.Route(MissingSchema, message)}, nil
	}

	isValid, err := validators.Validate(message, schema)
	if err != nil {
		if errors.Is(err, validator.ErrBrokenMessage) {
			message.RawAttributes["deadLetterErrorCategory"] = "Broken message"
			message.RawAttributes["deadLetterErrorReason"] = err.Error()
			return MessageTopicPair{Message: message, Topic: router.Route(Deadletter, message)}, nil
		}
		if errors.Is(err, validator.ErrWrongCompile) {
			message.RawAttributes["deadLetterErrorCategory"] = "Wrong compile"
			message.RawAttributes["deadLetterErrorReason"] = err.Error()
			return MessageTopicPair{Message: message, Topic: router.Route(Deadletter, message)}, nil
		}
		if errors.Is(err, validator.ErrFailedValidation) {
			message.RawAttributes["deadLetterErrorCategory"] = "Validation error"
			message.RawAttributes["deadLetterErrorReason"] = err.Error()
			return MessageTopicPair{Message: message, Topic: router.Route(Deadletter, message)}, nil
		}
		if errors.Is(err, validator.ErrUnsupportedFormat) {
			message.RawAttributes["deadLetterErrorCategory"] = "Unsupported format"
			message.RawAttributes["deadLetterErrorReason"] = err.Error()
			return MessageTopicPair{Message: message, Topic: router.Route(Deadletter, message)}, nil
		}
		if errors.Is(err, validator.ErrDeadletter) {
			return MessageTopicPair{Message: message, Topic: router.Route(Deadletter, message)}, nil
		}
		return MessageTopicPair{}, intoOpErr(message.ID, errcodes.ValidationFailure, err)
	}

	var result Result
	if isValid {
		result = Valid
	} else {
		result = Invalid
	}
	return MessageTopicPair{Message: message, Topic: router.Route(result, message)}, nil
}

// PublishToTopic publishes a Message to a broker.Topic, returning the relevant OpError in case of failure.
//
// If publishing is successful, the ack func of the underlying broker.Message is called, and the global Metrics are updated.
func PublishToTopic(ctx context.Context, message Message, topic broker.Topic) error {
	if err := topic.Publish(ctx, broker.OutboundMessage{
		Key:        message.Key,
		Data:       message.Payload,
		Attributes: message.RawAttributes,
	}); err != nil {
		return intoOpErr(message.ID, errcodes.PublishingFailure, err)
	}

	return nil
}
