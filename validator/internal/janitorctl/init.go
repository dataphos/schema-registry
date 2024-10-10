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

package janitorctl

import (
	"context"
	"crypto/tls"
	"net/http"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dataphos/schema-registry-validator/internal/config"
	"github.com/dataphos/schema-registry-validator/internal/errcodes"
	"github.com/dataphos/schema-registry-validator/internal/errtemplates"
	"github.com/dataphos/schema-registry-validator/internal/janitor"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/schema-registry-validator/internal/registry/apicuriosr"
	"github.com/dataphos/schema-registry-validator/internal/registry/janitorsr"
	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	csvgen "github.com/dataphos/schema-registry-validator/internal/schemagen/csv"
	jsongen "github.com/dataphos/schema-registry-validator/internal/schemagen/json"
	"github.com/dataphos/schema-registry-validator/internal/validator"
	"github.com/dataphos/schema-registry-validator/internal/validator/avro"
	"github.com/dataphos/schema-registry-validator/internal/validator/csv"
	"github.com/dataphos/schema-registry-validator/internal/validator/json"
	"github.com/dataphos/schema-registry-validator/internal/validator/protobuf"
	"github.com/dataphos/schema-registry-validator/internal/validator/xml"
	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-brokers/pkg/broker/jetstream"
	"github.com/dataphos/lib-brokers/pkg/broker/kafka"
	"github.com/dataphos/lib-brokers/pkg/broker/pubsub"
	"github.com/dataphos/lib-brokers/pkg/broker/pulsar"
	"github.com/dataphos/lib-brokers/pkg/broker/servicebus"
	"github.com/dataphos/lib-brokers/pkg/brokerutil"
	"github.com/dataphos/lib-httputil/pkg/httputil"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-streamproc/pkg/streamproc"
)

// initializeSchemaRegistry gets the janitor implementation of a schema registry, optionally decorating it with an in-memory lru cache
// if the appropriate env variable is set.
func initializeSchemaRegistry(ctx context.Context, log logger.Log, cfg *config.Registry) (registry.SchemaRegistry, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var sr registry.SchemaRegistry
	var err error
	switch cfg.Type {
	case "apicurio":
		sr, err = apicuriosr.New(
			ctx,
			cfg.URL,
			apicuriosr.TimeoutSettings{
				GetTimeout:      cfg.GetTimeout,
				RegisterTimeout: cfg.RegisterTimeout,
				UpdateTimeout:   cfg.UpdateTimeout,
			},
			cfg.GroupID,
		)
	case "janitor":
		sr, err = janitorsr.New(
			ctx,
			cfg.URL,
			janitorsr.TimeoutSettings{
				GetTimeout:      cfg.GetTimeout,
				RegisterTimeout: cfg.RegisterTimeout,
				UpdateTimeout:   cfg.UpdateTimeout,
			},
			cfg.GroupID,
		)
	default:
		sr, err = nil, errtemplates.UnsupportedRegistryType(cfg.Type)
	}
	if err != nil {
		return nil, err
	}

	cacheSize := cfg.InmemCacheSize
	if cacheSize > 0 {
		log.Infow("using in-memory cache for schema registry", logger.F{
			"cache_size": cacheSize,
		})
		return registry.WithCache(sr, cacheSize)
	}

	return sr, err
}

// initKafkaPublisher initializes an instance of Kafka Publisher
func initKafkaPublisher(ctx context.Context, cfg *config.Producer) (broker.Publisher, error) {
	var tlsConfig *tls.Config
	var krbConfig *kafka.KerberosConfig
	if cfg.Kafka.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Kafka.TlsConfig.ClientCertFile, cfg.Kafka.TlsConfig.ClientKeyFile, cfg.Kafka.TlsConfig.CaCertFile)
		tlsConfig.InsecureSkipVerify = cfg.Kafka.TlsConfig.InsecureSkipVerify
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
			Prometheus: &kafka.PrometheusConfig{
				Namespace:  "publisher",
				Registerer: prometheus.DefaultRegisterer,
				Gatherer:   prometheus.DefaultGatherer,
			},
		}, kafka.ProducerSettings{
			BatchSize:  cfg.Kafka.Settings.BatchSize,
			BatchBytes: cfg.Kafka.Settings.BatchBytes,
			Linger:     cfg.Kafka.Settings.Linger,
		},
	)
}

// initEventHubsPublisher initializes an instance of Kafka Publisher
func initEventHubsPublisher(ctx context.Context, cfg *config.Producer) (broker.Publisher, error) {
	var tlsConfig *tls.Config
	saslConfig := &kafka.PlainSASLConfig{
		User: cfg.Eventhubs.SaslConfig.User,
		Pass: cfg.Eventhubs.SaslConfig.Password,
	}

	if cfg.Eventhubs.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Eventhubs.TlsConfig.ClientCertFile, cfg.Eventhubs.TlsConfig.ClientKeyFile, cfg.Eventhubs.TlsConfig.CaCertFile)
		tlsConfig.InsecureSkipVerify = cfg.Eventhubs.TlsConfig.InsecureSkipVerify
		if err != nil {
			return nil, err
		}
	} else {
		// TLS has to be set when using eventhubs. In case it is not set, we need to skip certificate verification
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.Eventhubs.TlsConfig.InsecureSkipVerify,
		}
	}
	return kafka.NewPublisher(
		ctx,
		kafka.ProducerConfig{
			BrokerAddr: cfg.Eventhubs.Address,
			Prometheus: &kafka.PrometheusConfig{
				Namespace:  "publisher",
				Registerer: prometheus.DefaultRegisterer,
				Gatherer:   prometheus.DefaultGatherer,
			},
			TLS:                tlsConfig,
			PlainSASL:          saslConfig,
			DisableCompression: true,
		}, kafka.ProducerSettings{
			BatchSize:  cfg.Eventhubs.Settings.BatchSize,
			BatchBytes: cfg.Eventhubs.Settings.BatchBytes,
			Linger:     cfg.Eventhubs.Settings.Linger,
		},
	)
}

// initPubSubPublisher initializes an instance of PubSub Publisher
func initPubSubPublisher(ctx context.Context, cfg *config.Producer) (broker.Publisher, error) {
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
}

// initServiceBusPublisher initializes an instance of ServiceBus Publisher
func initServiceBusPublisher(cfg *config.Producer) (broker.Publisher, error) {
	return servicebus.NewPublisher(cfg.Servicebus.ConnectionString)
}

// initJetStreamPublisher initializes an instance of JetStream Publisher
func initJetStreamPublisher(ctx context.Context, cfg *config.Producer) (broker.Publisher, error) {
	return jetstream.NewPublisher(
		ctx,
		cfg.Jetstream.Url,
		jetstream.PublisherSettings{
			MaxPending: cfg.Jetstream.Settings.MaxInflightPending,
		},
	)
}

// initPulsarPublisher initializes an instance of Pulsar Publisher
func initPulsarPublisher(cfg *config.Producer) (broker.Publisher, error) {
	var tlsConfig *tls.Config
	if cfg.Kafka.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Pulsar.TlsConfig.ClientCertFile, cfg.Pulsar.TlsConfig.ClientKeyFile, cfg.Pulsar.TlsConfig.CaCertFile)
		if err != nil {
			return nil, err
		}
	}

	return pulsar.NewPublisher(
		pulsar.PublisherConfig{
			ServiceURL: cfg.Pulsar.ServiceUrl,
			TLSConfig:  tlsConfig,
		},
		pulsar.DefaultPublisherSettings,
	)
}

// initializePublisher selects and initializes an instance of broker.Publisher, depending on the value based through the appropriate
// environment variable.
func initializePublisher(ctx context.Context, cfg *config.Producer) (broker.Publisher, error) {
	switch cfg.Type {
	case "kafka":
		return initKafkaPublisher(ctx, cfg)
	case "eventhubs":
		return initEventHubsPublisher(ctx, cfg)
	case "pubsub":
		return initPubSubPublisher(ctx, cfg)
	case "servicebus":
		return initServiceBusPublisher(cfg)
	case "jetstream":
		return initJetStreamPublisher(ctx, cfg)
	case "pulsar":
		return initPulsarPublisher(cfg)

	default:
		return nil, errtemplates.UnsupportedBrokerType(cfg.Type)
	}
}

// initializeValidators initializes a map of validator.Validator,
// depending on which validators are enabled.
func initializeValidatorsForCentralConsumer(ctx context.Context, cfg *config.CentralConsumer) (map[string]validator.Validator, error) {
	validators := make(map[string]validator.Validator)

	if cfg.Validators.EnableAvro {
		validators["avro"] = avro.New()
	}

	if cfg.Validators.EnableCsv {
		csvValidator, err := csv.New(ctx, cfg.Validators.CsvUrl, cfg.Validators.CsvTimeoutBase)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't initialize csv validator")
		}
		validators["csv"] = csvValidator
	}

	if cfg.Validators.EnableJson {
		if cfg.Validators.JsonCacheSize > 0 {
			if cfg.Validators.JsonUseAltBackend {
				validators["json"] = json.NewCachedGoJsonSchemaValidator(cfg.Validators.JsonCacheSize)
			} else {
				validators["json"] = json.NewCached(cfg.Validators.JsonCacheSize)
			}
		} else {
			if cfg.Validators.JsonUseAltBackend {
				validators["json"] = json.NewGoJsonSchemaValidator()
			} else {
				validators["json"] = json.New()
			}
		}
	}

	if cfg.Validators.EnableProtobuf {
		protobufValidator, err := protobuf.New(cfg.Validators.ProtobufFilePath, cfg.Validators.ProtobufCacheSize)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't initialize protobuf validator")
		}
		validators["protobuf"] = protobufValidator
	}

	if cfg.Validators.EnableXml {
		xmlValidator, err := xml.New(ctx, cfg.Validators.XmlUrl, cfg.Validators.XmlTimeoutBase)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't initialize xml validator")
		}
		validators["xml"] = xmlValidator
	}

	return validators, nil
}

// initializeValidators initializes a map of validator.Validator,
// depending on which validators are enabled.
func initializeValidatorsForPullerCleaner(ctx context.Context, cfg *config.PullerCleaner) (map[string]validator.Validator, error) {
	validators := make(map[string]validator.Validator)

	if cfg.Validators.EnableCsv {
		csvValidator, err := csv.New(ctx, cfg.Validators.CsvUrl, cfg.Validators.CsvTimeoutBase)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't initialize csv validator")
		}
		validators["csv"] = csvValidator
	}

	if cfg.Validators.EnableJson {
		if cfg.Validators.JsonCacheSize > 0 {
			if cfg.Validators.JsonUseAltBackend {
				validators["json"] = json.NewCachedGoJsonSchemaValidator(cfg.Validators.JsonCacheSize)
			} else {
				validators["json"] = json.NewCached(cfg.Validators.JsonCacheSize)
			}
		} else {
			if cfg.Validators.JsonUseAltBackend {
				validators["json"] = json.NewGoJsonSchemaValidator()
			} else {
				validators["json"] = json.New()
			}
		}
	}

	return validators, nil
}

// initializeGenerators initializes a map of enabled schemagen.Generator for puller cleaner.
func initializeGenerators(cfg *config.PullerCleaner) (map[string]schemagen.Generator, error) {
	generators := make(map[string]schemagen.Generator)

	if cfg.Validators.EnableCsv {
		generators["csv"] = csvgen.New()
	}

	if cfg.Validators.EnableJson {
		if cfg.Validators.JsonSchemaGenScript == "" {
			return nil, errors.New("jsonSchemaGenScript not defined")
		}
		generators["json"] = jsongen.New(cfg.Validators.JsonSchemaGenScript)
	}

	return generators, nil
}

// initKafkaConsumer initializes a Kafka consumer component
func initKafkaConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	var tlsConfig *tls.Config
	var krbConfig *kafka.KerberosConfig
	if cfg.Kafka.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Kafka.TlsConfig.ClientCertFile, cfg.Kafka.TlsConfig.ClientKeyFile, cfg.Kafka.TlsConfig.CaCertFile)
		if err != nil {
			log.Error(err.Error(), errcodes.TLSInitialization)
			return
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

	iterator, err := kafka.NewBatchIterator(
		ctx,
		kafka.ConsumerConfig{
			BrokerAddr: cfg.Kafka.Address,
			GroupID:    cfg.Kafka.GroupId,
			Topic:      cfg.Kafka.Topic,
			TLS:        tlsConfig,
			Kerberos:   krbConfig,
			Prometheus: &kafka.PrometheusConfig{
				Namespace:  "consumer",
				Registerer: prometheus.DefaultRegisterer,
				Gatherer:   prometheus.DefaultGatherer,
			},
		},
		kafka.BatchConsumerSettings{
			ConsumerSettings: kafka.ConsumerSettings{
				MinBytes:             cfg.Kafka.Settings.MinBytes,
				MaxWait:              cfg.Kafka.Settings.MaxWait,
				MaxBytes:             cfg.Kafka.Settings.MaxBytes,
				MaxConcurrentFetches: cfg.Kafka.Settings.MaxConcurrentFetches,
			},
			MaxPollRecords: cfg.Kafka.Settings.MaxPollRecords,
		},
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}
	defer iterator.Close()

	srv = runMetricsServer(log)

	flowOpts := janitor.LoggingCallbacks(log, janitor.ShouldReturnFlowControl{
		OnPullErr:          streamproc.FlowControlContinue,
		OnProcessErr:       streamproc.FlowControlStop,
		OnUnrecoverable:    streamproc.FlowControlStop,
		OnThresholdReached: streamproc.FlowControlStop,
	})

	opts = append(opts, flowOpts...)

	log.Info("setup complete, running")
	if err = processor.AsBatchExecutor().Run(ctx, iterator, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initEventHubsConsumer initializes a PubSub consumer component
func initEventHubsConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	var tlsConfig *tls.Config
	saslConfig := &kafka.PlainSASLConfig{
		User: cfg.Eventhubs.SaslConfig.User,
		Pass: cfg.Eventhubs.SaslConfig.Password,
	}
	if cfg.Eventhubs.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Eventhubs.TlsConfig.ClientCertFile, cfg.Eventhubs.TlsConfig.ClientKeyFile, cfg.Eventhubs.TlsConfig.CaCertFile)
		if err != nil {
			log.Error(err.Error(), errcodes.TLSInitialization)
			return
		}
	} else {
		// TLS has to be set when using eventhubs. In case it is not set, we need to skip certificate verification
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.Eventhubs.TlsConfig.InsecureSkipVerify,
		}
	}

	iterator, err := kafka.NewBatchIterator(
		ctx,
		kafka.ConsumerConfig{
			BrokerAddr: cfg.Eventhubs.Address,
			GroupID:    cfg.Eventhubs.GroupId,
			Topic:      cfg.Eventhubs.Topic,
			TLS:        tlsConfig,
			Prometheus: &kafka.PrometheusConfig{
				Namespace:  "consumer",
				Registerer: prometheus.DefaultRegisterer,
				Gatherer:   prometheus.DefaultGatherer,
			},
			PlainSASL: saslConfig,
		},
		kafka.BatchConsumerSettings{
			ConsumerSettings: kafka.ConsumerSettings{
				MinBytes:             cfg.Eventhubs.Settings.MinBytes,
				MaxWait:              cfg.Eventhubs.Settings.MaxWait,
				MaxBytes:             cfg.Eventhubs.Settings.MaxBytes,
				MaxConcurrentFetches: cfg.Eventhubs.Settings.MaxConcurrentFetches,
			},
			MaxPollRecords: cfg.Eventhubs.Settings.MaxPollRecords,
		},
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}
	defer iterator.Close()

	srv = runMetricsServer(log)

	log.Info("setup complete, running")
	if err = processor.AsBatchExecutor().Run(ctx, iterator, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initPubSubConsumer initializes a PubSub consumer component
func initPubSubConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	receiver, err := pubsub.NewReceiver(
		ctx,
		pubsub.ReceiverConfig{
			ProjectID:      cfg.Pubsub.ProjectId,
			SubscriptionID: cfg.Pubsub.SubscriptionId,
		},
		pubsub.ReceiveSettings{
			MaxExtension:           cfg.Pubsub.Settings.MaxExtension,
			MaxExtensionPeriod:     cfg.Pubsub.Settings.MaxExtensionPeriod,
			MaxOutstandingMessages: cfg.Pubsub.Settings.MaxOutstandingMessages,
			MaxOutstandingBytes:    cfg.Pubsub.Settings.MaxOutstandingBytes,
			NumGoroutines:          cfg.Pubsub.Settings.NumGoroutines,
		},
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}
	defer func() {
		if err := receiver.Close(); err != nil {
			log.Error(err.Error(), errcodes.BrokerConnClosed)
		}
	}()

	srv = runMetricsServer(log)

	flowOpts := janitor.LoggingCallbacks(log, janitor.ShouldReturnFlowControl{
		OnPullErr:          streamproc.FlowControlStop,
		OnProcessErr:       streamproc.FlowControlContinue,
		OnUnrecoverable:    streamproc.FlowControlContinue,
		OnThresholdReached: streamproc.FlowControlStop,
	})

	opts = append(opts, flowOpts...)

	log.Info("setup complete, running")
	if err = processor.AsReceiverExecutor().Run(ctx, receiver, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initServiceBusConsumer initializes a ServiceBus consumer component
func initServiceBusConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	iterator, err := servicebus.NewBatchIterator(
		servicebus.IteratorConfig{
			ConnectionString: cfg.Servicebus.ConnectionString,
			Topic:            cfg.Servicebus.Topic,
			Subscription:     cfg.Servicebus.Subscription,
		},
		servicebus.BatchIteratorSettings{
			BatchSize: cfg.Servicebus.Settings.BatchSize,
		},
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}
	defer func() {
		if err := iterator.Close(); err != nil {
			log.Error(err.Error(), errcodes.BrokerConnClosed)
		}
	}()

	batchedReceiver := brokerutil.BatchedMessageIteratorIntoBatchedReceiver(
		iterator,
		brokerutil.IntoBatchedReceiverSettings{
			NumGoroutines: runtime.GOMAXPROCS(0),
		},
	)

	srv = runMetricsServer(log)

	flowOpts := janitor.LoggingCallbacks(log, janitor.ShouldReturnFlowControl{
		OnPullErr:          streamproc.FlowControlStop,
		OnProcessErr:       streamproc.FlowControlContinue,
		OnUnrecoverable:    streamproc.FlowControlContinue,
		OnThresholdReached: streamproc.FlowControlStop,
	})

	opts = append(opts, flowOpts...)

	log.Info("setup complete, running")
	if err = processor.AsBatchedReceiverExecutor().Run(ctx, batchedReceiver, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initJetStreamConsumer initializes a JetStream consumer component
func initJetStreamConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	iterator, err := jetstream.NewBatchIterator(
		ctx,
		jetstream.IteratorConfig{
			URL:          cfg.Jetstream.Url,
			Subject:      cfg.Jetstream.Subject,
			ConsumerName: cfg.Jetstream.ConsumerName,
		},
		jetstream.BatchIteratorSettings{
			BatchSize: cfg.Jetstream.Settings.BatchSize,
		},
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}
	defer iterator.Close()

	batchedReceiver := brokerutil.BatchedMessageIteratorIntoBatchedReceiver(
		iterator,
		brokerutil.IntoBatchedReceiverSettings{
			NumGoroutines: runtime.GOMAXPROCS(0),
		},
	)

	srv = runMetricsServer(log)

	flowOpts := janitor.LoggingCallbacks(log, janitor.ShouldReturnFlowControl{
		OnPullErr:          streamproc.FlowControlStop,
		OnProcessErr:       streamproc.FlowControlContinue,
		OnUnrecoverable:    streamproc.FlowControlContinue,
		OnThresholdReached: streamproc.FlowControlStop,
	})

	opts = append(opts, flowOpts...)

	log.Info("setup complete, running")
	if err = processor.AsBatchedReceiverExecutor().Run(ctx, batchedReceiver, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initPulsarConsumer initializes a Pulsar consumer component
func initPulsarConsumer(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	var srv *http.Server
	var tlsConfig *tls.Config
	if cfg.Kafka.TlsConfig.Enabled {
		var err error
		tlsConfig, err = httputil.NewTLSConfig(cfg.Pulsar.TlsConfig.ClientCertFile, cfg.Pulsar.TlsConfig.ClientKeyFile, cfg.Pulsar.TlsConfig.CaCertFile)
		if err != nil {
			log.Error(err.Error(), errcodes.TLSInitialization)
			return
		}
	}

	iterator, err := pulsar.NewIterator(
		pulsar.IteratorConfig{
			ServiceURL:   cfg.Pulsar.ServiceUrl,
			Topic:        cfg.Pulsar.Topic,
			Subscription: cfg.Pulsar.Subscription,
			TLSConfig:    tlsConfig,
		},
		pulsar.DefaultIteratorSettings,
	)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
	}
	defer func() {
		if err := iterator.Close(); err != nil {
			log.Error(err.Error(), errcodes.BrokerConnClosed)
		}
	}()

	batchedReceiver := brokerutil.MessageIteratorIntoBatchedReceiver(
		iterator,
		brokerutil.IntoBatchedMessageIteratorSettings{
			BatchSize: pulsar.DefaultIteratorSettings.ReceiverQueueSize,
			Timeout:   pulsar.DefaultIteratorSettings.OperationTimeout,
		},
		brokerutil.IntoBatchedReceiverSettings{
			NumGoroutines: runtime.GOMAXPROCS(0),
		},
	)

	srv = runMetricsServer(log)

	flowOpts := janitor.LoggingCallbacks(log, janitor.ShouldReturnFlowControl{
		OnPullErr:          streamproc.FlowControlStop,
		OnProcessErr:       streamproc.FlowControlContinue,
		OnUnrecoverable:    streamproc.FlowControlContinue,
		OnThresholdReached: streamproc.FlowControlStop,
	})

	opts = append(opts, flowOpts...)

	log.Info("setup complete, running")
	if err = processor.AsBatchedReceiverExecutor().Run(ctx, batchedReceiver, opts...); err != nil {
		log.Error(err.Error(), errcodes.CompletedWithErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "http server shutdown failed").Error(), errcodes.MetricsServerShutdownFailure)
	}
}

// initializeSourceSystemAndRunProcessor initializes a consumer component and runs processor.
func initializeSourceSystemAndRunProcessor(ctx context.Context, processor *janitor.Processor, log logger.Log, cfg *config.Consumer, opts []streamproc.RunOption) {
	switch cfg.Type {
	case "kafka":
		initKafkaConsumer(ctx, processor, log, cfg, opts)
	case "eventhubs":
		initEventHubsConsumer(ctx, processor, log, cfg, opts)
	case "pubsub":
		initPubSubConsumer(ctx, processor, log, cfg, opts)
	case "servicebus":
		initServiceBusConsumer(ctx, processor, log, cfg, opts)
	case "jetstream":
		initJetStreamConsumer(ctx, processor, log, cfg, opts)
	case "pulsar":
		initPulsarConsumer(ctx, processor, log, cfg, opts)
	default:
		log.Error(errtemplates.UnsupportedBrokerType(cfg.Type).Error(), errcodes.BrokerInitialization)
	}
}

func loadRunOptions(cfg *config.RunOptions) []streamproc.RunOption {
	var opts []streamproc.RunOption

	opts = append(opts, streamproc.WithErrThreshold(cfg.ErrThreshold))
	opts = append(opts, streamproc.WithErrInterval(cfg.ErrInterval))
	opts = append(opts, streamproc.WithNumRetires(cfg.NumRetries))

	return opts
}
