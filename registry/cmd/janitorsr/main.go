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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dataphos/aquarium-janitor-standalone-sr/compatibility"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/config"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errcodes"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errtemplates"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/repository/postgres"
	"github.com/dataphos/aquarium-janitor-standalone-sr/server"
	"github.com/dataphos/aquarium-janitor-standalone-sr/validity"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
)

const (
	serverPortEnvKey = "SERVER_PORT"
)

const (
	defaultServerPort = 8080
)

// @title		Schema Registry API
// @version		1.0
func main() {
	labels := logger.Labels{
		"product":   "Schema Registry",
		"component": "registry",
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
		log.Error(err.Error(), errcodes.DatabaseConnectionInitialization)
		return
	}
	if !postgres.HealthCheck(db) {
		log.Error("database state invalid", errcodes.InvalidDatabaseState)
		return
	}

	var port int
	portStr := os.Getenv(serverPortEnvKey)
	if portStr == "" {
		port = defaultServerPort
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			log.Error(errtemplates.ExpectedInt(serverPortEnvKey, portStr).Error(), errcodes.ServerInitialization)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	compChecker, globalCompMode, err := compatibility.InitCompatibilityChecker(ctx)
	if err != nil {
		log.Error(err.Error(), errcodes.ExternalCheckerInitialization)
		return
	}
	log.Info("Successfully connected compatibility checker.")

	valChecker, globalValMode, err := validity.InitExternalValidityChecker(ctx)
	if err != nil {
		log.Error(err.Error(), errcodes.ExternalCheckerInitialization)
		return
	}
	log.Info("Successfully connected validity checker.")

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.New(server.NewHandler(registry.New(postgres.New(db), compChecker, valChecker, globalCompMode, globalValMode), log)),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		<-c

		log.Info("initiating graceful shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = srv.Shutdown(ctx); err != nil {
			log.Error(errors.Wrap(err, "graceful shutdown failed").Error(), errcodes.ServerShutdown)
		}
		close(idleConnsClosed)
	}()
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		log.Infow("starting Prometheus server", logger.F{"port": 2112})
		err1 := http.ListenAndServe(":2112", nil)
		if err1 != nil {
			log.Error(errors.Wrap(err, "an error occurred starting Prometheus server").Error(), errcodes.ServerShutdown)
		}
	}()

	log.Infow("starting server", logger.F{"port": srv.Addr})
	if err = srv.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Error(errors.Wrap(err, "an error occurred starting or closing server").Error(), errcodes.ServerShutdown)
		}
	}

	<-idleConnsClosed

	log.Info("shutting down")
}
