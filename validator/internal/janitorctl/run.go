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
	"runtime/debug"

	"github.com/dataphos/schema-registry-validator/internal/centralconsumer"
	"github.com/dataphos/schema-registry-validator/internal/config"
	"github.com/dataphos/schema-registry-validator/internal/errcodes"
	"github.com/dataphos/schema-registry-validator/internal/janitor"
	"github.com/dataphos/schema-registry-validator/internal/pullercleaner"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
	"github.com/dataphos/lib-shutdown/pkg/graceful"
)

type ProcessorInitFunc func(context.Context, registry.SchemaRegistry, broker.Publisher) (*janitor.Processor, error)

func RunCentralConsumer(configFile string) {
	labels := logger.Labels{
		"product":   "Schema Registry",
		"component": "central_consumer",
	}
	var Commit = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}

		return ""
	}()
	if Commit != "" {
		labels["commit"] = Commit
	}

	logLevel, logConfigWarnings := config.GetLogLevel()
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logLevel))

	for _, w := range logConfigWarnings {
		log.Warn(w)
	}

	var cfg config.CentralConsumer
	if err := cfg.Read(configFile); err != nil {
		log.Fatal(err.Error(), errcodes.ReadConfigFailure)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err.Error(), errcodes.ValidateConfigFailure)
	}

	initProcessor := func(ctx context.Context, registry registry.SchemaRegistry, publisher broker.Publisher) (*janitor.Processor, error) {
		validators, err := initializeValidatorsForCentralConsumer(ctx, &cfg)
		if err != nil {
			return nil, err
		}

		cc, err := centralconsumer.New(
			registry,
			publisher,
			validators,
			centralconsumer.Topics{
				Valid:       cfg.Topics.Valid,
				InvalidCSV:  cfg.Topics.DeadLetter,
				InvalidJSON: cfg.Topics.DeadLetter,
				Deadletter:  cfg.Topics.DeadLetter,
			},
			centralconsumer.Settings{
				NumSchemaCollectors: cfg.NumSchemaCollectors,
				NumInferrers:        cfg.NumInferrers,
			},
			log,
			centralconsumer.RouterFlags{
				MissingSchema: cfg.ShouldLog.MissingSchema,
				Valid:         cfg.ShouldLog.Valid,
				Deadletter:    cfg.ShouldLog.DeadLetter,
			},
			centralconsumer.Mode(cfg.Mode),
			centralconsumer.SchemaMetadata{
				ID:      cfg.SchemaID,
				Version: cfg.SchemaVersion,
				Format:  cfg.SchemaType,
			},
			cfg.Encryption.EncryptionKey,
		)
		if err != nil {
			return nil, err
		}
		return cc.AsProcessor(), nil
	}

	run(
		graceful.WithSignalShutdown(context.Background()),
		log,
		&cfg.Registry,
		initProcessor,
		&cfg.Producer,
		&cfg.RunOptions,
		&cfg.Consumer,
	)
}

func RunPullerCleaner(configFile string) {
	labels := logger.Labels{
		"product":   "Schema Registry",
		"component": "puller_cleaner",
	}
	var Commit = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}

		return ""
	}()
	if Commit != "" {
		labels["commit"] = Commit
	}

	logLevel, logConfigWarnings := config.GetLogLevel()
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logLevel))

	for _, w := range logConfigWarnings {
		log.Warn(w)
	}

	var cfg config.PullerCleaner
	if err := cfg.Read(configFile); err != nil {
		log.Fatal(err.Error(), errcodes.ReadConfigFailure)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err.Error(), errcodes.ValidateConfigFailure)
	}

	initProcessor := func(ctx context.Context, registry registry.SchemaRegistry, publisher broker.Publisher) (*janitor.Processor, error) {
		validators, err := initializeValidatorsForPullerCleaner(ctx, &cfg)
		if err != nil {
			return nil, err
		}

		generators, err := initializeGenerators(&cfg)
		if err != nil {
			return nil, err
		}

		pc, err := pullercleaner.New(
			generators,
			registry,
			validators,
			publisher,
			pullercleaner.Topics{
				Valid:      cfg.Topics.Valid,
				Deadletter: cfg.Topics.DeadLetter,
			},
			cfg.NumCleaners,
			log,
			pullercleaner.RouterFlags{
				Valid:      cfg.ShouldLog.Valid,
				Deadletter: cfg.ShouldLog.DeadLetter,
			},
		)
		if err != nil {
			return nil, err
		}
		return pc.AsProcessor(), nil
	}

	run(
		graceful.WithSignalShutdown(context.Background()),
		log,
		&cfg.Registry,
		initProcessor,
		&cfg.Producer,
		&cfg.RunOptions,
		&cfg.Consumer,
	)
}

func run(ctx context.Context, log logger.Log, registryCfg *config.Registry, initProcessor ProcessorInitFunc, producerCfg *config.Producer, runOptions *config.RunOptions, consumerCfg *config.Consumer) {
	log.Info("initializing schema registry")
	schemaRegistry, err := initializeSchemaRegistry(ctx, log, registryCfg)
	if err != nil {
		log.Error(err.Error(), errcodes.RegistryInitialization)
		return
	}

	log.Info("initializing publisher")
	publisher, err := initializePublisher(ctx, producerCfg)
	if err != nil {
		log.Error(err.Error(), errcodes.BrokerInitialization)
		return
	}

	log.Info("initializing main component")
	processor, err := initProcessor(ctx, schemaRegistry, publisher)
	if err != nil {
		log.Error(err.Error(), errcodes.Initialization)
		return
	}

	log.Info("loading run settings")
	opts := loadRunOptions(runOptions)

	log.Info("initializing source system and running processor")
	initializeSourceSystemAndRunProcessor(ctx, processor, log, consumerCfg, opts)

	log.Info("shutting down")
}
