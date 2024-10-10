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

package janitor

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	publishCountProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "published_messages_total",
		Help:      "The total number of published messages",
	})
	bytesProcessedProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "processed_bytes_total",
		Help:      "The total number of processed bytes",
	})
	processingTimesProm = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace:  "schema_registry",
		Name:       "processing_times_milliseconds",
		Help:       "Processing times of published messages in milliseconds",
		MaxAge:     5 * time.Minute,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

// UpdateSuccessMetrics updates Prometheus metrics: publishCountProm, bytesProcessedProm, and processingTimesProm.
func UpdateSuccessMetrics(messages ...Message) {
	publishCountProm.Add(float64(len(messages)))

	for _, message := range messages {
		messageProcessingTime := time.Since(message.IngestionTime).Milliseconds()
		processingTimesProm.Observe(float64(messageProcessingTime))

		bytesProcessedProm.Add(float64(len(message.Payload)))
	}
}

var (
	publishDLCountProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "published_dead_letter_messages_total",
		Help:      "The total number of published dead letter messages",
	})
	bytesDLProcessedProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "processed_dead_letter_bytes_total",
		Help:      "The total number of processed dead letter bytes",
	})
	processingDLTimesProm = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace:  "schema_registry",
		Name:       "processing_dead_letter_times_milliseconds",
		Help:       "Processing times of published dead letter messages in milliseconds",
		MaxAge:     5 * time.Minute,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

// UpdateSuccessDLMetrics updates Prometheus metrics: publishDLCountProm, bytesDLProcessedProm, and processingDLTimesProm.
func UpdateSuccessDLMetrics(messages ...Message) {
	publishDLCountProm.Add(float64(len(messages)))

	for _, message := range messages {
		messageProcessingTime := time.Since(message.IngestionTime).Milliseconds()
		processingDLTimesProm.Observe(float64(messageProcessingTime))

		bytesDLProcessedProm.Add(float64(len(message.Payload)))
	}
}

var (
	nackCountProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "nack_messages_total",
		Help:      "The total number of nack messages",
	})
	nackBytesProcessedProm = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "schema_registry",
		Name:      "nack_processed_bytes_total",
		Help:      "The total number of nack processed bytes",
	})
	nackProcessingTimesProm = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace:  "schema_registry",
		Name:       "nack_processing_times_milliseconds",
		Help:       "Processing times of nack messages in milliseconds",
		MaxAge:     5 * time.Minute,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

// UpdateFailureMetrics updates Prometheus metrics: nackCountProm, nackBytesProcessedProm, and nackProcessingTimesProm.
func UpdateFailureMetrics(messages ...Message) {
	nackCountProm.Add(float64(len(messages)))

	for _, message := range messages {
		msgNackProcessingTime := time.Since(message.IngestionTime).Milliseconds()
		nackProcessingTimesProm.Observe(float64(msgNackProcessingTime))

		nackBytesProcessedProm.Add(float64(len(message.Payload)))
	}
}
