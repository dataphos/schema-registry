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

package config

import (
	"reflect"
	"strings"
	"time"

	"github.com/dataphos/schema-registry-validator/internal/errtemplates"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
)

type Producer struct {
	Type          string                    `toml:"type" val:"oneof=kafka eventhubs pubsub servicebus jetstream pulsar"`
	EncryptionKey string                    `toml:"encryption_key"`
	Kafka         KafkaPublisherConfig      `toml:"kafka"`
	Eventhubs     EventhubsPublisherConfig  `toml:"eventhubs"`
	Pubsub        PubsubPublisherConfig     `toml:"pubsub"`
	Servicebus    ServicebusPublisherConfig `toml:"servicebus"`
	Jetstream     JetstreamPublisherConfig  `toml:"jetstream"`
	Pulsar        PulsarPublisherConfig     `toml:"pulsar"`
}

type KafkaPublisherConfig struct {
	Address    string                 `toml:"address"`
	TlsConfig  TlsConfig              `toml:"tls_config"`
	KrbConfig  KrbConfig              `toml:"krb_config"`
	SaslConfig SaslConfig             `toml:"sasl_config"`
	Settings   KafkaPublisherSettings `toml:"settings"`
}

type EventhubsPublisherConfig struct {
	Address    string                     `toml:"address"`
	TlsConfig  TlsConfig                  `toml:"tls_config"`
	SaslConfig SaslConfig                 `toml:"sasl_config"`
	Settings   EventhubsPublisherSettings `toml:"settings"`
}

type TlsConfig struct {
	Enabled            bool   `toml:"enabled"`
	ClientCertFile     string `toml:"client_cert_file" val:"required_if=Enabled true,omitempty,file"`
	ClientKeyFile      string `toml:"client_key_file" val:"required_if=Enabled true,omitempty,file"`
	CaCertFile         string `toml:"ca_cert_file" val:"required_if=Enabled true,omitempty,file"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
}

type KrbConfig struct {
	Enabled        bool   `toml:"enabled"`
	KrbConfigPath  string `toml:"krb_config_path"`
	KrbKeyTabPath  string `toml:"krb_keytab_path"`
	KrbRealm       string `toml:"krb_realm"`
	KrbServiceName string `toml:"krb_service_name"`
	KrbUsername    string `toml:"krb_username"`
}

type SaslConfig struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type KafkaPublisherSettings struct {
	BatchSize  int           `toml:"batch_size" default:"40"`
	BatchBytes int64         `toml:"batch_bytes" default:"5242880"`
	Linger     time.Duration `toml:"linger" default:"10ms"`
	Acks       int           `toml:"kafka_acks" default:"1"`
}

type EventhubsPublisherSettings struct {
	BatchSize  int           `toml:"batch_size" default:"40"`
	BatchBytes int64         `toml:"batch_bytes" default:"5242880"`
	Linger     time.Duration `toml:"linger" default:"10ms"`
}

type PubsubPublisherConfig struct {
	ProjectId string                  `toml:"project_id"`
	Settings  PubsubPublisherSettings `toml:"settings"`
}

type PubsubPublisherSettings struct {
	DelayThreshold         time.Duration `toml:"delay_threshold" default:"50ms"`
	CountThreshold         int           `toml:"count_threshold" default:"50"`
	ByteThreshold          int           `toml:"byte_threshold" default:"52428800"`
	NumGoroutines          int           `toml:"num_goroutines" default:"5"`
	Timeout                time.Duration `toml:"timeout" default:"15s"`
	MaxOutstandingMessages int           `toml:"max_outstanding_messages" default:"800"`
	MaxOutstandingBytes    int           `toml:"max_outstanding_bytes" default:"1048576000"`
	EnableMessageOrdering  bool          `toml:"enable_message_ordering"`
}

type ServicebusPublisherConfig struct {
	ConnectionString string `toml:"connection_string"`
}

type JetstreamPublisherConfig struct {
	Url      string                     `toml:"url"`
	Settings JetstreamPublisherSettings `toml:"settings"`
}

type JetstreamPublisherSettings struct {
	MaxInflightPending int `toml:"max_inflight_pending" default:"512"`
}

type PulsarPublisherConfig struct {
	ServiceUrl string    `toml:"service_url"`
	TlsConfig  TlsConfig `toml:"tls_config"`
}

type Consumer struct {
	Type       string                   `toml:"type" val:"oneof=kafka eventhubs pubsub servicebus jetstream pulsar"`
	Kafka      KafkaConsumerConfig      `toml:"kafka"`
	Eventhubs  EventhubsConsumerConfig  `toml:"eventhubs"`
	Pubsub     PubsubConsumerConfig     `toml:"pubsub"`
	Servicebus ServicebusConsumerConfig `toml:"servicebus"`
	Jetstream  JetstreamConsumerConfig  `toml:"jetstream"`
	Pulsar     PulsarConsumerConfig     `toml:"pulsar"`
}

type KafkaConsumerConfig struct {
	Address   string                `toml:"address"`
	TlsConfig TlsConfig             `toml:"tls_config"`
	KrbConfig KrbConfig             `toml:"krb_config"`
	Topic     string                `toml:"topic"`
	GroupId   string                `toml:"group_id"`
	Settings  KafkaConsumerSettings `toml:"settings"`
}

type EventhubsConsumerConfig struct {
	Address    string                    `toml:"address"`
	TlsConfig  TlsConfig                 `toml:"tls_config"`
	SaslConfig SaslConfig                `toml:"sasl_config"`
	Topic      string                    `toml:"topic"`
	GroupId    string                    `toml:"group_id"`
	Settings   EventhubsConsumerSettings `toml:"settings"`
}

type KafkaConsumerSettings struct {
	MinBytes             int           `toml:"min_bytes" default:"100"`
	MaxWait              time.Duration `toml:"max_wait" default:"5s"`
	MaxBytes             int           `toml:"max_bytes" default:"10485760"`
	MaxConcurrentFetches int           `toml:"max_concurrent_fetches" default:"3"`
	MaxPollRecords       int           `toml:"max_poll_records" default:"100"`
}

type EventhubsConsumerSettings struct {
	MinBytes             int           `toml:"min_bytes" default:"100"`
	MaxWait              time.Duration `toml:"max_wait" default:"5s"`
	MaxBytes             int           `toml:"max_bytes" default:"10485760"`
	MaxConcurrentFetches int           `toml:"max_concurrent_fetches" default:"3"`
	MaxPollRecords       int           `toml:"max_poll_records" default:"100"`
}

type PubsubConsumerConfig struct {
	ProjectId      string                 `toml:"project_id"`
	SubscriptionId string                 `toml:"subscription_id"`
	Settings       PubsubConsumerSettings `toml:"settings"`
}

type PubsubConsumerSettings struct {
	MaxExtension           time.Duration `toml:"max_extension" default:"30m"`
	MaxExtensionPeriod     time.Duration `toml:"max_extension_period" default:"3m"`
	MaxOutstandingMessages int           `toml:"max_outstanding_messages" default:"1000"`
	MaxOutstandingBytes    int           `toml:"max_outstanding_bytes" default:"419430400"`
	NumGoroutines          int           `toml:"num_goroutines" default:"10"`
}

type ServicebusConsumerConfig struct {
	ConnectionString string                     `toml:"connection_string"`
	Topic            string                     `toml:"topic"`
	Subscription     string                     `toml:"subscription"`
	Settings         ServicebusConsumerSettings `toml:"settings"`
}

type ServicebusConsumerSettings struct {
	BatchSize int `toml:"batch_size" default:"100"`
}

type JetstreamConsumerConfig struct {
	Url          string                    `toml:"url"`
	Subject      string                    `toml:"subject"`
	ConsumerName string                    `toml:"consumer_name"`
	Settings     JetstreamConsumerSettings `toml:"settings"`
}

type JetstreamConsumerSettings struct {
	BatchSize int `toml:"batch_size" default:"100"`
}

type PulsarConsumerConfig struct {
	ServiceUrl   string    `toml:"service_url"`
	Topic        string    `toml:"topic"`
	Subscription string    `toml:"subscription"`
	TlsConfig    TlsConfig `toml:"tls_config"`
}

type Registry struct {
	URL             string        `toml:"url" val:"url"`
	Type            string        `toml:"type" default:"janitor" val:"oneof=janitor apicurio"`
	GroupID         string        `toml:"groupID"`
	GetTimeout      time.Duration `toml:"get_timeout" default:"4s"`
	RegisterTimeout time.Duration `toml:"register_timeout" default:"10s"`
	UpdateTimeout   time.Duration `toml:"update_timeout" default:"10s"`
	InmemCacheSize  int           `toml:"inmem_cache_size" default:"100"`
}

type RunOptions struct {
	ErrThreshold int64         `toml:"err_threshold" default:"50"`
	ErrInterval  time.Duration `toml:"err_interval" default:"1m"`
	NumRetries   int           `toml:"num_retries" default:"0"`
}

// validate validates CentralConsumer or PullerCleaner struct.
func validate(cfg interface{}, prefix string) error {
	validate := validator.New()
	validate.SetTagName("val")

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("toml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	validate.RegisterStructValidation(
		ProducerStructLevelValidation,
		KafkaPublisherConfig{},
		EventhubsPublisherConfig{},
		PubsubPublisherConfig{},
		ServicebusPublisherConfig{},
		JetstreamPublisherConfig{},
	)
	validate.RegisterStructValidation(
		ConsumerStructLevelValidation,
		KafkaConsumerConfig{},
		EventhubsConsumerConfig{},
		PubsubConsumerConfig{},
		ServicebusConsumerConfig{},
		JetstreamConsumerConfig{},
	)

	if err := validate.Struct(cfg); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		var errCombined error
		for _, err := range err.(validator.ValidationErrors) {
			// Trims prefix ("CentralConsumer." or "PullerCleaner.")
			// in order to correspond to TOML key path.
			fieldName := strings.TrimPrefix(err.Namespace(), prefix)

			switch err.Tag() {
			case "required":
				errCombined = multierr.Append(errCombined, errtemplates.RequiredTagFail(fieldName))
			case "required_if":
				errCombined = multierr.Append(errCombined, errtemplates.RequiredTagFail(fieldName))
			case "file":
				errCombined = multierr.Append(errCombined, errtemplates.FileTagFail(fieldName, err.Value()))
			case "url":
				errCombined = multierr.Append(errCombined, errtemplates.UrlTagFail(fieldName, err.Value()))
			case "oneof":
				errCombined = multierr.Append(errCombined, errtemplates.OneofTagFail(fieldName, err.Value()))
			case "hostname_port":
				errCombined = multierr.Append(errCombined, errtemplates.HostnamePortTagFail(fieldName, err.Value()))
			default:
				errCombined = multierr.Append(errCombined, err)
			}
		}
		return errCombined
	}
	return nil
}

// ProducerStructLevelValidation is a custom validator which validates broker
// structure depending on which type of producer is required.
func ProducerStructLevelValidation(sl validator.StructLevel) {
	source := sl.Parent().Interface().(Producer)

	validate := validator.New()

	switch producer := sl.Current().Interface().(type) {
	case KafkaPublisherConfig:
		if source.Type == "kafka" {
			validateMultipleHostnames(validate, sl, producer.Address)
		}
	case EventhubsPublisherConfig:
		if source.Type == "eventhubs" {
			validateMultipleHostnames(validate, sl, producer.Address)
			if err := validate.Var(producer.SaslConfig, "required"); err != nil {
				sl.ReportValidationErrors("sasl_config", "", err.(validator.ValidationErrors))
			}
		}
	case PubsubPublisherConfig:
		if source.Type == "pubsub" {
			if err := validate.Var(producer.ProjectId, "required"); err != nil {
				sl.ReportValidationErrors("project_id", "", err.(validator.ValidationErrors))
			}
		}
	case ServicebusPublisherConfig:
		if source.Type == "servicebus" {
			if err := validate.Var(producer.ConnectionString, "required"); err != nil {
				sl.ReportValidationErrors("connection_string", "", err.(validator.ValidationErrors))
			}
		}
	case JetstreamPublisherConfig:
		if source.Type == "jetstream" {
			if err := validate.Var(producer.Url, "url"); err != nil {
				sl.ReportValidationErrors("url", "", err.(validator.ValidationErrors))
			}
		}
	case PulsarPublisherConfig:
		if source.Type == "pulsar" {
			if err := validate.Var(producer.ServiceUrl, "required"); err != nil {
				sl.ReportValidationErrors("service_url", "", err.(validator.ValidationErrors))
			}
		}
	}
}

// ConsumerStructLevelValidation is a custom validator which validates broker
// structure depending on which type of consumer is required.
func ConsumerStructLevelValidation(sl validator.StructLevel) {
	source := sl.Parent().Interface().(Consumer)

	validate := validator.New()

	switch consumer := sl.Current().Interface().(type) {
	case KafkaConsumerConfig:
		if source.Type == "kafka" {
			validateMultipleHostnames(validate, sl, consumer.Address)
			if err := validate.Var(consumer.Topic, "required"); err != nil {
				sl.ReportValidationErrors("topic", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.GroupId, "required"); err != nil {
				sl.ReportValidationErrors("group_id", "", err.(validator.ValidationErrors))
			}
		}
	case EventhubsConsumerConfig:
		if source.Type == "eventhubs" {
			validateMultipleHostnames(validate, sl, consumer.Address)
			if err := validate.Var(consumer.Topic, "required"); err != nil {
				sl.ReportValidationErrors("topic", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.GroupId, "required"); err != nil {
				sl.ReportValidationErrors("group_id", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.SaslConfig, "required"); err != nil {
				sl.ReportValidationErrors("sasl_config", "", err.(validator.ValidationErrors))
			}
		}
	case PubsubConsumerConfig:
		if source.Type == "pubsub" {
			if err := validate.Var(consumer.ProjectId, "required"); err != nil {
				sl.ReportValidationErrors("project_id", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.SubscriptionId, "required"); err != nil {
				sl.ReportValidationErrors("subscription_id", "", err.(validator.ValidationErrors))
			}
		}
	case ServicebusConsumerConfig:
		if source.Type == "servicebus" {
			if err := validate.Var(consumer.ConnectionString, "required"); err != nil {
				sl.ReportValidationErrors("connection_string", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.Topic, "required"); err != nil {
				sl.ReportValidationErrors("topic", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.Subscription, "required"); err != nil {
				sl.ReportValidationErrors("subscription", "", err.(validator.ValidationErrors))
			}
		}
	case JetstreamConsumerConfig:
		if source.Type == "jetstream" {
			if err := validate.Var(consumer.Url, "url"); err != nil {
				sl.ReportValidationErrors("url", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.Subject, "required"); err != nil {
				sl.ReportValidationErrors("subject", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.ConsumerName, "required"); err != nil {
				sl.ReportValidationErrors("consumer_name", "", err.(validator.ValidationErrors))
			}
		}
	case PulsarConsumerConfig:
		if source.Type == "pulsar" {
			if err := validate.Var(consumer.ServiceUrl, "required"); err != nil {
				sl.ReportValidationErrors("service_url", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.Topic, "required"); err != nil {
				sl.ReportValidationErrors("topic", "", err.(validator.ValidationErrors))
			}
			if err := validate.Var(consumer.Subscription, "required"); err != nil {
				sl.ReportValidationErrors("subscription", "", err.(validator.ValidationErrors))
			}
		}
	}
}

func validateMultipleHostnames(validate *validator.Validate, sl validator.StructLevel, addresses string) {
	splitted := strings.Split(addresses, ",")

	for _, address := range splitted {
		if err := validate.Var(strings.Trim(address, " "), "hostname_port"); err != nil {
			sl.ReportValidationErrors("address", "", err.(validator.ValidationErrors))
		}
	}
}
