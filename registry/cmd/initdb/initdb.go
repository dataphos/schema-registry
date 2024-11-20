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

package main

import (
	"runtime/debug"

	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
	"github.com/dataphos/schema-registry/internal/config"
	"github.com/dataphos/schema-registry/internal/errcodes"
	"github.com/dataphos/schema-registry/registry/repository/postgres"
)

func main() {
	labels := logger.Labels{
		"product":   "Schema Registry",
		"component": "initdb",
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

	db, err := postgres.InitializeGormFromEnv()
	if err != nil {
		log.Fatal(err.Error(), errcodes.DatabaseConnectionInitialization)
		return
	}

	if postgres.HealthCheck(db) {
		log.Warn("database already initialized")
		return
	}

	if err = postgres.Initdb(db); err != nil {
		log.Fatal(err.Error(), errcodes.DatabaseInitialization)
		return
	}
	log.Info("database initialized successfully")
}
