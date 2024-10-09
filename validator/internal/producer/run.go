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
	"context"
	"crypto/tls"
	"errors"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errtemplates"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry/apicuriosr"
	"log"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errcodes"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry/janitorsr"
	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-brokers/pkg/broker/jetstream"
	"github.com/dataphos/lib-brokers/pkg/broker/kafka"
	"github.com/dataphos/lib-brokers/pkg/broker/pubsub"
	"github.com/dataphos/lib-brokers/pkg/broker/servicebus"
	"github.com/dataphos/lib-httputil/pkg/httputil"
	"github.com/dataphos/lib-shutdown/pkg/graceful"
)

func Run(configFile string) {
	var cfg Config
	if err := cfg.Read(configFile); err != nil {
		log.Fatal(err.Error()+" ", errcodes.ReadConfigFailure)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err.Error()+" ", errcodes.ValidateConfigFailure)
	}

	ctx := graceful.WithSignalShutdown(context.Background())

	var sr registry.SchemaRegistry
	var err error
	switch cfg.RegistryConfig.Type {
	case "apicurio":
		sr, err = apicuriosr.New(
			ctx,
			cfg.RegistryConfig.URL,
			apicuriosr.TimeoutSettings{
				GetTimeout:      cfg.RegistryConfig.GetTimeout,
				RegisterTimeout: cfg.RegistryConfig.RegisterTimeout,
				UpdateTimeout:   cfg.RegistryConfig.UpdateTimeout,
			},
			cfg.RegistryConfig.GroupID,
		)
	case "janitor":
		sr, err = janitorsr.New(
			ctx,
			cfg.RegistryConfig.URL,
			janitorsr.TimeoutSettings{
				GetTimeout:      cfg.RegistryConfig.GetTimeout,
				RegisterTimeout: cfg.RegistryConfig.RegisterTimeout,
				UpdateTimeout:   cfg.RegistryConfig.UpdateTimeout,
			},
			cfg.RegistryConfig.GroupID,
		)
	default:
		sr, err = nil, errtemplates.UnsupportedRegistryType(cfg.Type)
	}
	if err != nil {
		log.Fatal(err)
	}

	publisher, err := selectPublisher(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	topic, err := publisher.Topic(cfg.TopicId)
	if err != nil {
		log.Fatal(err)
	}

	if err = New(sr, topic, cfg.RateLimit, Mode(cfg.Mode), cfg.EncryptionKey).LoadAndProduce(ctx, cfg.BaseDir, cfg.FileName, cfg.NumberOfMessages); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}

func selectPublisher(ctx context.Context, cfg *Config) (broker.Publisher, error) {
	switch cfg.Type {
	case "pubsub":
		return pubsub.NewPublisher(
			ctx,
			pubsub.PublisherConfig{
				ProjectID: cfg.Pubsub.ProjectId,
			},
			pubsub.PublishSettings{
				DelayThreshold:         cfg.Pubsub.Settings.DelayThreshold,
				CountThreshold:         cfg.Pubsub.Settings.CountThreshold,
				ByteThreshold:          cfg.Pubsub.Settings.ByteThreshold,
				NumGoroutines:          cfg.Pubsub.Settings.NumGoroutines,
				Timeout:                cfg.Pubsub.Settings.Timeout,
				MaxOutstandingMessages: cfg.Pubsub.Settings.MaxOutstandingMessages,
				MaxOutstandingBytes:    cfg.Pubsub.Settings.MaxOutstandingBytes,
				EnableMessageOrdering:  cfg.Pubsub.Settings.EnableMessageOrdering,
			},
		)
	case "kafka":
		var tlsConfig *tls.Config
		var krbConfig *kafka.KerberosConfig
		if cfg.Kafka.TlsConfig.Enabled {
			var err error
			tlsConfig, err = httputil.NewTLSConfig(cfg.Kafka.TlsConfig.ClientCertFile, cfg.Kafka.TlsConfig.ClientKeyFile, cfg.Kafka.TlsConfig.CaCertFile)
			if err != nil {
				return nil, err
			}
		}
		if cfg.Kafka.KrbConfig.Enabled {
			krbConfig = &kafka.KerberosConfig{
				KeyTabPath: cfg.Kafka.KrbConfig.KrbKeyTabPath,
				ConfigPath: cfg.Kafka.KrbConfig.KrbConfigPath,
				Realm:      cfg.Kafka.KrbConfig.KrbRealm,
				Service:    cfg.Kafka.KrbConfig.KrbServiceName,
				Username:   cfg.Kafka.KrbConfig.KrbUsername,
			}
		}
		return kafka.NewPublisher(
			ctx,
			kafka.ProducerConfig{
				BrokerAddr: cfg.Kafka.Address,
				TLS:        tlsConfig,
				Kerberos:   krbConfig,
			},
			kafka.ProducerSettings{
				BatchSize:  cfg.Kafka.Settings.BatchSize,
				BatchBytes: cfg.Kafka.Settings.BatchBytes,
				Linger:     cfg.Kafka.Settings.Linger,
			},
		)
	case "eventhubs":
		var tlsConfig *tls.Config
		var saslConfig *kafka.PlainSASLConfig
		if cfg.Eventhubs.TlsConfig.Enabled {
			var err error
			tlsConfig, err = httputil.NewTLSConfig(cfg.Eventhubs.TlsConfig.ClientCertFile, cfg.Eventhubs.TlsConfig.ClientKeyFile, cfg.Eventhubs.TlsConfig.CaCertFile)
			if err != nil {
				return nil, err
			}
		}
		saslConfig = &kafka.PlainSASLConfig{
			User: cfg.Eventhubs.SaslConfig.User,
			Pass: cfg.Eventhubs.SaslConfig.Password,
		}

		return kafka.NewPublisher(
			ctx,
			kafka.ProducerConfig{
				BrokerAddr:         cfg.Eventhubs.Address,
				TLS:                tlsConfig,
				PlainSASL:          saslConfig,
				DisableCompression: true,
			},
			kafka.ProducerSettings{
				BatchSize:  cfg.Eventhubs.Settings.BatchSize,
				BatchBytes: cfg.Eventhubs.Settings.BatchBytes,
				Linger:     cfg.Eventhubs.Settings.Linger,
			},
		)
	case "servicebus":
		return servicebus.NewPublisher(cfg.Servicebus.ConnectionString)
	case "jetstream":
		return jetstream.NewPublisher(
			ctx,
			cfg.Jetstream.Url,
			jetstream.PublisherSettings{
				MaxPending: cfg.Jetstream.Settings.MaxInflightPending,
			},
		)
	default:
		return nil, errors.New("unsupported broker type")
	}
}
