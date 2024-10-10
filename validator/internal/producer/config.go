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

package producer

import (
	"reflect"
	"strings"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errtemplates"
	"github.com/go-playground/validator/v10"
	"github.com/kkyr/fig"
	"go.uber.org/multierr"
)

// Config represents all required configuration to run a standalone producer.
type Config struct {
	BaseDir          string           `toml:"base_dir" default:""`
	FileName         string           `toml:"file_name" val:"file"`
	NumberOfMessages int              `toml:"number_of_messages" default:"100"`
	RateLimit        int              `toml:"rate_limit" default:"100"`
	TopicId          string           `toml:"topic_id" val:"required"`
	Type             string           `toml:"type" val:"oneof=kafka eventhubs pubsub servicebus jetstream"`
	EncryptionKey    string           `toml:"encryption_key"`
	Kafka            KafkaConfig      `toml:"kafka"`
	Eventhubs        EventhubsConfig  `toml:"eventhubs"`
	Pubsub           PubsubConfig     `toml:"pubsub"`
	Servicebus       ServicebusConfig `toml:"servicebus"`
	Jetstream        JetstreamConfig  `toml:"jetstream"`
	RegistryConfig   RegistryConfig   `toml:"registry_config"`
	Mode             int              `toml:"mode"`
}

type KafkaConfig struct {
	Address    string                 `toml:"address"`
	TlsConfig  TlsConfig              `toml:"tls_config"`
	KrbConfig  KrbConfig              `toml:"krb_config"`
	SaslConfig SaslConfig             `toml:"sasl_config"`
	Settings   KafkaPublisherSettings `toml:"settings"`
}

type EventhubsConfig struct {
	Address    string                     `toml:"address"`
	TlsConfig  TlsConfig                  `toml:"tls_config"`
	SaslConfig SaslConfig                 `toml:"sasl_config"`
	Settings   EventhubsPublisherSettings `toml:"settings"`
}

type TlsConfig struct {
	Enabled        bool   `toml:"enabled"`
	ClientCertFile string `toml:"client_cert_file" val:"required_if=Enabled true,omitempty,file"`
	ClientKeyFile  string `toml:"client_key_file" val:"required_if=Enabled true,omitempty,file"`
	CaCertFile     string `toml:"ca_cert_file" val:"required_if=Enabled true,omitempty,file"`
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
}

type EventhubsPublisherSettings struct {
	BatchSize  int           `toml:"batch_size" default:"40"`
	BatchBytes int64         `toml:"batch_bytes" default:"5242880"`
	Linger     time.Duration `toml:"linger" default:"10ms"`
}

type PubsubConfig struct {
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

type ServicebusConfig struct {
	ConnectionString string `toml:"connection_string"`
}

type JetstreamConfig struct {
	Url      string                     `toml:"url"`
	Settings JetstreamPublisherSettings `toml:"settings"`
}

type JetstreamPublisherSettings struct {
	MaxInflightPending int `toml:"max_inflight_pending" default:"512"`
}

type RegistryConfig struct {
	URL             string        `toml:"url" val:"url"`
	GetTimeout      time.Duration `toml:"get_timeout" default:"4s"`
	RegisterTimeout time.Duration `toml:"register_timeout" default:"10s"`
	UpdateTimeout   time.Duration `toml:"update_timeout" default:"10s"`
	Type            string        `toml:"type" default:"janitor" val:"oneof=janitor apicurio"`
	GroupID         string        `toml:"groupID"`
}

// Read loads parameters from configuration file into Config struct.
func (cfg *Config) Read(fileName string) error {
	return fig.Load(cfg, fig.File(fileName), fig.Tag("toml"), fig.UseEnv(""))
}

// Validate validates Config struct.
func (cfg *Config) Validate() error {
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
		publisherStructLevelValidation,
		KafkaConfig{},
		EventhubsConfig{},
		PubsubConfig{},
		ServicebusConfig{},
		JetstreamConfig{},
	)

	if err := validate.Struct(cfg); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		var errCombined error
		for _, err := range err.(validator.ValidationErrors) {
			// Trims prefix in order to correspond to TOML key path.
			fieldName := strings.TrimPrefix(err.Namespace(), "Config.")

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

// publisherStructLevelValidation is a custom validator which validates broker
// structure depending on which type of producer is required.
func publisherStructLevelValidation(sl validator.StructLevel) {
	source := sl.Parent().Interface().(Config)

	validate := validator.New()

	switch producer := sl.Current().Interface().(type) {
	case KafkaConfig:
		if source.Type == "kafka" {
			if err := validate.Var(producer.Address, "hostname_port"); err != nil {
				sl.ReportValidationErrors("address", "", err.(validator.ValidationErrors))
			}
		}
	case EventhubsConfig:
		if source.Type == "eventhubs" {
			if err := validate.Var(producer.Address, "hostname_port"); err != nil {
				sl.ReportValidationErrors("address", "", err.(validator.ValidationErrors))
			}
		}
	case PubsubConfig:
		if source.Type == "pubsub" {
			if err := validate.Var(producer.ProjectId, "required"); err != nil {
				sl.ReportValidationErrors("project_id", "", err.(validator.ValidationErrors))
			}
		}
	case ServicebusConfig:
		if source.Type == "servicebus" {
			if err := validate.Var(producer.ConnectionString, "required"); err != nil {
				sl.ReportValidationErrors("connection_string", "", err.(validator.ValidationErrors))
			}
		}
	case JetstreamConfig:
		if source.Type == "jetstream" {
			if err := validate.Var(producer.Url, "url"); err != nil {
				sl.ReportValidationErrors("url", "", err.(validator.ValidationErrors))
			}
		}
	}
}
