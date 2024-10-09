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
