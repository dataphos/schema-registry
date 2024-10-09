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
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errcodes"
	"github.com/dataphos/lib-logger/logger"
)

// runMetricsServer runs a http server on which Prometheus metrics are being exposed.
// All metrics that are registered to default Prometheus Registry are displayed at:
// "localhost:2112/metrics" endpoint.
func runMetricsServer(log logger.Log) *http.Server {
	http.Handle("/metrics", promhttp.Handler())

	port := ":2112"

	srv := &http.Server{Addr: port}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err.Error(), errcodes.MetricsServerFailure)
		}
	}()

	log.Info(fmt.Sprintf("exposed metrics at port %s", port))

	return srv
}
