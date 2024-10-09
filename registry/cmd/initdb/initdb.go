package main

import (
	"runtime/debug"

	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/config"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errcodes"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/repository/postgres"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
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
