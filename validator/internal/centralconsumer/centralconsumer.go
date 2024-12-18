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

package centralconsumer

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/schema-registry-validator/internal/config"
	"github.com/dataphos/schema-registry-validator/internal/errcodes"
	"github.com/dataphos/schema-registry-validator/internal/errtemplates"
	"github.com/dataphos/schema-registry-validator/internal/janitor"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	jsoninternal "github.com/dataphos/schema-registry-validator/internal/validator/json"
)

// Mode is the way Central consumer works; if the mode is Default, one CC will be deployed, and it will validate multiple
// different schemas. If it's OneCCPerTopic, there will be one CC for each topic, and it will validate only one schema.
type Mode int

const (
	Default Mode = iota
	OneCCPerTopic
)

const (
	HeaderSchema  = "Header: "
	PayloadSchema = "Payload: "
)

type SchemaMetadata struct {
	ID      string
	Version string
	Format  string
}

type Schema struct {
	SchemaMetadata SchemaMetadata
	Specification  []byte
}

type SchemaDefinition struct {
	ID          string           `json:"schema_id,omitempty"`
	Type        string           `json:"schema_type"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	LastCreated string           `json:"last_created"`
	Versions    []VersionDetails `json:"schemas"`
}

type VersionDetails struct {
	Version       string `json:"version"`
	Specification []byte `json:"specification"`
}

// CentralConsumer models the central consumer process.
type CentralConsumer struct {
	Registry            registry.SchemaRegistry
	Validators          janitor.Validators
	Router              janitor.Router
	Publisher           broker.Publisher
	topicIDs            Topics
	topics              map[string]broker.Topic
	registrySem         chan struct{}
	validatorsSem       chan struct{}
	log                 logger.Log
	mode                Mode
	schema              Schema
	encryptionKey       string
	validateHeaders     bool
	defaultHeaderSchema config.DefaultHeaderSchema
}

// Settings holds settings concerning the concurrency limits for various stages of the central consumer pipeline.
type Settings struct {
	// NumSchemaCollectors defines the maximum amount of inflight requests to the schema registry.
	NumSchemaCollectors int

	// NumInferrers defines the maximum amount of inflight destination topic inference jobs (validation and routing).
	NumInferrers int

	// ValidateHeaders defines if the messages' headers will be validated
	ValidateHeaders bool

	// DefaultHeaderSchemaId is default ID of the header schema
	DefaultHeaderSchemaId string

	// DefaultHeaderSchemaVersion is default version of the header schema
	DefaultHeaderSchemaVersion string
}

// Topics defines the standard destination topics, based on validation results.
type Topics struct {
	Valid       string
	InvalidCSV  string
	InvalidJSON string
	Deadletter  string
}

// RouterFlags defines logging levels for logging each routing decision.
type RouterFlags struct {
	MissingSchema bool
	Valid         bool
	Invalid       bool
	Deadletter    bool
}

// New is a convenience function which returns a new instance of CentralConsumer.
func New(registry registry.SchemaRegistry, publisher broker.Publisher, validators janitor.Validators, topicIds Topics, settings Settings, log logger.Log, routerFlags RouterFlags, mode Mode, schemaMetadata SchemaMetadata, encryptionKey string) (*CentralConsumer, error) {
	var (
		schemaVersion VersionDetails
		format        string
	)

	topics, err := idsIntoTopics(topicIds, publisher)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.LoadingTopicsFailed)
	}

	var registrySem chan struct{}
	if settings.NumSchemaCollectors > 0 {
		registrySem = make(chan struct{}, settings.NumSchemaCollectors)
	}
	var validatorsSem chan struct{}
	if settings.NumInferrers > 0 {
		validatorsSem = make(chan struct{}, settings.NumInferrers)
	}
	if settings.ValidateHeaders {
		_, ok := validators["json"]
		if !ok {
			// if json validation is turned off, this version of json validator is used by default for validating message header
			validators["json"] = jsoninternal.New()
		}
	}

	var schemaReturned []byte
	if mode == OneCCPerTopic {
		if schemaMetadata.ID != "" {
			if schemaMetadata.Version != "" {
				schemaSpecReturned, err := registry.Get(context.Background(), schemaMetadata.ID, schemaMetadata.Version)
				if err != nil {
					return &CentralConsumer{}, err
				}
				schemaVersion.Version = schemaMetadata.Version
				schemaVersion.Specification = schemaSpecReturned
			} else {
				schemaReturned, err = registry.GetLatest(context.Background(), schemaMetadata.ID)
				if err != nil {
					return &CentralConsumer{}, err
				}
				if err = json.Unmarshal(schemaReturned, &schemaVersion); err != nil {
					return &CentralConsumer{}, errors.Wrap(err, errtemplates.UnmarshallingJSONFailed)
				}
			}
			if schemaMetadata.Format != "" {
				format = schemaMetadata.Format
			} else {
				return &CentralConsumer{}, errors.New("schema format not specified")
			}
		} else {
			return &CentralConsumer{}, errors.New("schema ID not specified")
		}
	}

	return &CentralConsumer{
		Registry:      registry,
		Validators:    validators,
		Router:        setupRoutingFunc(topicIds, routerFlags, log),
		topicIDs:      topicIds,
		Publisher:     publisher,
		topics:        topics,
		registrySem:   registrySem,
		validatorsSem: validatorsSem,
		log:           log,
		mode:          mode,
		schema: Schema{
			SchemaMetadata: SchemaMetadata{
				ID:      schemaMetadata.ID,
				Version: schemaVersion.Version,
				Format:  format,
			},
			Specification: schemaVersion.Specification,
		},
		encryptionKey:   encryptionKey,
		validateHeaders: settings.ValidateHeaders,
		defaultHeaderSchema: config.DefaultHeaderSchema{
			DefaultHeaderSchemaId:      settings.DefaultHeaderSchemaId,
			DefaultHeaderSchemaVersion: settings.DefaultHeaderSchemaVersion,
		},
	}, nil
}

// idsIntoTopics maps Topics into instances of broker.Topic.
func idsIntoTopics(topicIds Topics, publisher broker.Publisher) (map[string]broker.Topic, error) {
	topics := make(map[string]broker.Topic)

	if topicIds.Valid != "" {
		topic, err := publisher.Topic(topicIds.Valid)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.Valid))
		}
		topics[topicIds.Valid] = topic
	}

	if topicIds.InvalidJSON != "" {
		topic, err := publisher.Topic(topicIds.InvalidJSON)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.InvalidJSON))
		}
		topics[topicIds.InvalidJSON] = topic
	}

	if topicIds.InvalidCSV != "" {
		topic, err := publisher.Topic(topicIds.InvalidCSV)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.InvalidCSV))
		}
		topics[topicIds.InvalidCSV] = topic
	}

	if topicIds.Deadletter != "" {
		topic, err := publisher.Topic(topicIds.Deadletter)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.Deadletter))
		}
		topics[topicIds.Deadletter] = topic
	}
	return topics, nil
}

// setupRoutingFunc sets up the janitor.LoggingRouter, by first checking if there's a need for logging any of the routing
// decisions (if any logging level flag is set). If none of the flags are set, standard intoRouter is used,
// wrapping it with logging middleware otherwise.
func setupRoutingFunc(topics Topics, routerFlags RouterFlags, log logger.Log) janitor.Router {
	next := intoRouter(topics)

	if routerFlags.MissingSchema || routerFlags.Valid || routerFlags.Invalid || routerFlags.Deadletter {
		return janitor.LoggingRouter(
			log,
			janitor.RouterFlags{
				MissingSchema: routerFlags.MissingSchema,
				Valid:         routerFlags.Valid,
				Invalid:       routerFlags.Invalid,
				Deadletter:    routerFlags.Deadletter,
			},
			next,
		)
	}

	return next
}

const (
	avroFormat     = "avro"
	csvFormat      = "csv"
	jsonFormat     = "json"
	protobufFormat = "protobuf"
	xmlFormat      = "xml"
)

// intoRouter maps the given Topics into a janitor.LoggingRouter.
//
// All janitor.Valid messages are routed to Topics.Valid.
//
// All janitor.Deadletter messages are routed to Topics.Deadletter.
//
// If the result is janitor.MissingSchema,
// CSV and JSON formats are routed to Topics.InvalidCSV and Topics.InvalidJSON, respectively,
// while all other formats are routed to Topics.Deadletter.
//
// If the result is janitor.Invalid,
// CSV and JSON formats are routed to Topics.InvalidCSV and Topics.InvalidJSON, respectively,
// while all other formats are routed to Topics.Deadletter.
func intoRouter(topics Topics) janitor.Router {
	return janitor.RoutingFunc(func(result janitor.Result, message janitor.Message) string {
		format := message.Format

		switch result {
		case janitor.Valid:
			return topics.Valid
		case janitor.Deadletter:
			return topics.Deadletter
		case janitor.MissingSchema, janitor.Invalid:
			switch format {
			case csvFormat:
				return topics.InvalidCSV
			case jsonFormat:
				return topics.InvalidJSON
			default:
				return topics.Deadletter
			}
		default:
			return topics.Deadletter
		}
	})
}

func (cc *CentralConsumer) AsProcessor() *janitor.Processor {
	return janitor.NewProcessor(cc, cc.topics, cc.topicIDs.Deadletter, cc.log)
}

func (cc *CentralConsumer) Handle(ctx context.Context, message janitor.Message) (janitor.MessageTopicPair, error) {
	var (
		schema                []byte
		messageSchemaPair     janitor.MessageSchemaPair
		messageTopicPair      janitor.MessageTopicPair
		specificSchemaVersion VersionDetails
		err                   error
		encryptedMessageData  []byte
	)

	// header validation is turned on if a message specifies so in the header OR if validateHeaders flag is set
	// on the Validator level
	if message.RawAttributes[janitor.HeaderValidation] == "true" ||
		(cc.validateHeaders && message.RawAttributes[janitor.HeaderValidation] != "false") {
		_, ok := cc.Validators["json"]
		// it is possible json validator isn't initialized by this point so we are checking it just in case
		if !ok {
			cc.Validators["json"] = jsoninternal.New()
		}
		var (
			headerId      string
			headerVersion string
			headerSchema  []byte
		)

		headerId, headerVersion, err = getHeaderSchemaIdAndVersion(message)
		for _, e := range multierr.Errors(err) {
			if e.Error() == errtemplates.AttributeNotDefined(janitor.AttributeHeaderID).Error() {
				if cc.defaultHeaderSchema.DefaultHeaderSchemaId != "" {
					headerId = cc.defaultHeaderSchema.DefaultHeaderSchemaId
				} else {
					setMessageRawAttributes(message, "Header error", err)
					return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, nil
				}
			}
			if e.Error() == errtemplates.AttributeNotDefined(janitor.AttributeHeaderVersion).Error() {
				if cc.defaultHeaderSchema.DefaultHeaderSchemaVersion != "" {
					headerVersion = cc.defaultHeaderSchema.DefaultHeaderSchemaVersion
				} else {
					setMessageRawAttributes(message, "Header error", err)
					return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, nil
				}
			}
		}

		acquireIfSet(cc.registrySem)
		headerSchema, err = janitor.CollectSchema(ctx, headerId, headerVersion, cc.Registry)
		if err != nil {
			return cc.determineError(message, err, HeaderSchema)
		}
		releaseIfSet(cc.registrySem)

		var isValid bool
		isValid, err = janitor.ValidateHeader(message, headerSchema, cc.Validators)
		if err != nil {
			return cc.determineError(message, err, HeaderSchema)
		}
		if !isValid {
			return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, nil
		}
	}

	if cc.mode == Default {
		acquireIfSet(cc.registrySem)
		schema, err = janitor.CollectSchema(ctx, message.SchemaID, message.Version, cc.Registry)
		if err != nil {
			return cc.determineError(message, err, PayloadSchema)
		}
		messageSchemaPair = janitor.MessageSchemaPair{Message: message, Schema: schema}
		releaseIfSet(cc.registrySem)

		messageTopicPair, err = cc.getMessageTopicPair(messageSchemaPair, encryptedMessageData)
		if err != nil {
			return messageTopicPair, err
		}
		return messageTopicPair, nil

	} else if cc.mode == OneCCPerTopic {
		if message.Version == "" { // Version not set in message
			messageTopicPair, err = cc.getMessageTopicPair(janitor.MessageSchemaPair{
				Message: message,
				Schema:  cc.schema.Specification,
			}, encryptedMessageData)
			if err != nil {
				return messageTopicPair, err
			}
			if messageTopicPair.Topic == cc.topicIDs.Deadletter {
				// if message is invalid against latest schema saved in CC, then fetch latest from SR and revalidate
				messageTopicPair, err = cc.revalidatedAgainstLatest(ctx, specificSchemaVersion, message, encryptedMessageData)
				if err != nil {
					return messageTopicPair, err
				}
			}
			return messageTopicPair, nil
		} else {
			if message.Version == cc.schema.SchemaMetadata.Version {
				messageTopicPair, err = cc.getMessageTopicPair(janitor.MessageSchemaPair{
					Message: message,
					Schema:  cc.schema.Specification,
				}, encryptedMessageData)
				if err != nil {
					return messageTopicPair, err
				}
				return messageTopicPair, nil
			} else {
				acquireIfSet(cc.registrySem)
				specificSchemaVersionSpec, err := cc.Registry.Get(ctx, cc.schema.SchemaMetadata.ID, message.Version)
				if err != nil {
					if errors.Is(err, registry.ErrNotFound) {
						setMessageRawAttributes(message, "Schema error", err)
						releaseIfSet(cc.registrySem)
						return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, nil
					} else if errors.Is(err, registry.InvalidHeader) {
						setMessageRawAttributes(message, "Header error", err)
						releaseIfSet(cc.registrySem)
						return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, nil
					} else {
						setMessageRawAttributes(message, "Wrong compile", err)
						releaseIfSet(cc.registrySem)
						return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, err
					}
				}
				releaseIfSet(cc.registrySem)

				err = cc.updateIfNewer(VersionDetails{
					Version:       message.Version,
					Specification: specificSchemaVersionSpec,
				})
				if err != nil {
					setMessageRawAttributes(message, "Non number version", err)
					return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, err
				}

				messageTopicPair, err = cc.getMessageTopicPair(janitor.MessageSchemaPair{
					Message: message,
					Schema:  specificSchemaVersionSpec,
				}, encryptedMessageData)
				if err != nil {
					return messageTopicPair, err
				}
				return messageTopicPair, nil
			}
		}
	} else {
		err = errors.New("unknown CC mode")
		setMessageRawAttributes(message, "Unknown CC mode", err)
		return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, err
	}
}

func getHeaderSchemaIdAndVersion(message janitor.Message) (string, string, error) {
	var (
		id, version               string
		okId, okVersion, isString bool
		errCombined               error
	)

	// Check if header ID and header version are present
	if _, okId = message.RawAttributes[janitor.AttributeHeaderID]; !okId {
		errCombined = multierr.Append(errCombined, errtemplates.AttributeNotDefined(janitor.AttributeHeaderID))
	}
	if _, okVersion = message.RawAttributes[janitor.AttributeHeaderVersion]; !okVersion {
		errCombined = multierr.Append(errCombined, errtemplates.AttributeNotDefined(janitor.AttributeHeaderVersion))
	}

	// For those values that are present, check if they are strings (this should always be true, but check just in case)
	if okId {
		if id, isString = message.RawAttributes[janitor.AttributeHeaderID].(string); !isString {
			errCombined = multierr.Append(errCombined, errtemplates.AttributeNotAString(janitor.AttributeHeaderID))
		}
	}
	if okVersion {
		if version, isString = message.RawAttributes[janitor.AttributeHeaderVersion].(string); !isString {
			errCombined = multierr.Append(errCombined, errtemplates.AttributeNotAString(janitor.AttributeHeaderVersion))
		}
	}

	// Values that are set are going to be modified, those that are not will not be and proper error will be returned
	return id, version, errCombined
}

func (cc *CentralConsumer) getMessageTopicPair(messageSchemaPair janitor.MessageSchemaPair, encryptedMessageData []byte) (janitor.MessageTopicPair, error) {
	acquireIfSet(cc.validatorsSem)
	var err error
	if cc.encryptionKey != "" {
		encryptedMessageData = messageSchemaPair.Message.Payload //nolint:ineffassign,staticcheck // fine for now
		messageSchemaPair.Message.Payload, err = janitor.Decrypt(messageSchemaPair.Message.Payload, cc.encryptionKey)
		if err != nil {
			messageSchemaPair.Message.RawAttributes["deadLetterErrorCategory"] = "Failure to decrypt"
			messageSchemaPair.Message.RawAttributes["deadLetterErrorReason"] = err.Error()
			return janitor.MessageTopicPair{Message: messageSchemaPair.Message, Topic: cc.Router.Route(janitor.Deadletter, messageSchemaPair.Message)}, err
		}
	}
	messageTopicPair, err := janitor.InferDestinationTopic(messageSchemaPair, cc.Validators, cc.Router)
	if err != nil {
		releaseIfSet(cc.validatorsSem)
		return messageTopicPair, err
	}
	releaseIfSet(cc.validatorsSem)
	return messageTopicPair, nil
}

func (cc *CentralConsumer) updateVersion(vd VersionDetails) {
	cc.schema.SchemaMetadata.Version = vd.Version
	cc.schema.Specification = vd.Specification
}

// checkIfNewer checks if v2 is newer than v1
func checkIfNewer(v1, v2 string) (bool, error) {
	v1Int, err := strconv.Atoi(v1)
	if err != nil {
		return false, err
	}
	v2Int, err := strconv.Atoi(v2)
	if err != nil {
		return false, err
	}
	if v2Int > v1Int {
		return true, nil
	}
	return false, nil
}

// revalidatedAgainstLatest fetches latest version of schema from Schema Registry and validates the message against it
func (cc *CentralConsumer) revalidatedAgainstLatest(ctx context.Context, specificSchemaVersion VersionDetails, message janitor.Message, encryptedMessageData []byte) (janitor.MessageTopicPair, error) {
	var messageTopicPair janitor.MessageTopicPair

	acquireIfSet(cc.registrySem)
	specificSchemaVersionBytes, err := cc.Registry.GetLatest(ctx, cc.schema.SchemaMetadata.ID)
	if err != nil {
		setMessageRawAttributes(message, "Wrong compile", err)
		releaseIfSet(cc.registrySem)
		return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, err
	}
	if err = json.Unmarshal(specificSchemaVersionBytes, &specificSchemaVersion); err != nil {
		setMessageRawAttributes(message, "Broken message", err)
		releaseIfSet(cc.registrySem)
		return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, errors.Wrap(err, errtemplates.UnmarshallingJSONFailed)
	}
	releaseIfSet(cc.registrySem)

	err = cc.updateIfNewer(specificSchemaVersion)
	if err != nil {
		setMessageRawAttributes(message, "Non number version", err)
		return janitor.MessageTopicPair{Message: message, Topic: cc.Router.Route(janitor.Deadletter, message)}, err
	}

	messageTopicPair, err = cc.getMessageTopicPair(janitor.MessageSchemaPair{
		Message: message,
		Schema:  cc.schema.Specification,
	}, encryptedMessageData)
	if err != nil {
		return messageTopicPair, err
	}
	return messageTopicPair, nil
}

func (cc *CentralConsumer) updateIfNewer(versionDetails VersionDetails) error {
	newer, err := checkIfNewer(cc.schema.SchemaMetadata.Version, versionDetails.Version)
	if err != nil {
		return err
	}
	if newer {
		cc.updateVersion(versionDetails)
	}
	return nil
}

func setMessageRawAttributes(message janitor.Message, errCategory string, err error) {
	message.RawAttributes["deadLetterErrorCategory"] = errCategory
	message.RawAttributes["deadLetterErrorReason"] = err.Error()
}

func (cc *CentralConsumer) determineError(m janitor.Message, e error, errorLocation string) (janitor.MessageTopicPair, error) {
	opError := &janitor.OpError{}
	if errors.As(e, &opError) {
		if opError.Code == errcodes.RegistryUnresponsive {
			setMessageRawAttributes(m, "Registry unresponsive", e)
			releaseIfSet(cc.registrySem)
			return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, e
		} else if opError.Code == errcodes.SchemaNotRegistered {
			setMessageRawAttributes(m, errorLocation+"schema not registered", e)
			releaseIfSet(cc.registrySem)
			return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, nil
		} else if opError.Code == errcodes.MissingDataInHeader {
			setMessageRawAttributes(m, errorLocation+"schema id/version missing", e)
			releaseIfSet(cc.registrySem)
			return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, nil
		} else if opError.Code == errcodes.InvalidDataInHeader {
			setMessageRawAttributes(m, errorLocation+"schema id/version invalid", e)
			releaseIfSet(cc.registrySem)
			return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, nil
		} else if opError.Code == errcodes.DeadletterMessage {
			setMessageRawAttributes(m, errorLocation+"validation failed", e)
			releaseIfSet(cc.registrySem)
			return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, nil
		}
	}
	setMessageRawAttributes(m, "Wrong compile", e)
	releaseIfSet(cc.registrySem)
	return janitor.MessageTopicPair{Message: m, Topic: cc.Router.Route(janitor.Deadletter, m)}, e
}

func acquireIfSet(sem chan struct{}) {
	if sem != nil {
		sem <- struct{}{}
	}
}

func releaseIfSet(sem chan struct{}) {
	if sem != nil {
		<-sem
	}
}
