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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	schemaDeletedProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schema_registry",
		Name:      "schema_deleted",
		Help:      "Indicates whether the schema has been deleted (1 = schema deleted)",
	},
		[]string{"id"})
	schemaRegisteredProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schema_registry",
		Name:      "schema_registered",
		Help:      "Indicates whether new schema has been registered (1 = schema registered)",
	},
		[]string{"id", "version"})
	schemaUpdatedProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schema_registry",
		Name:      "schema_updated",
		Help:      "Indicates whether the schema has been updated (1 = schema updated)",
	},
		[]string{"id", "version"})
	schemaVersionDeletedProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schema_registry",
		Name:      "schema_version_deleted",
		Help:      "Indicates whether schema version has been deleted (1 = schema version deleted)",
	},
		[]string{"id", "version"})
)

func UpdateSchemaMetricUpdate(id string, ver string) {
	schemaUpdatedProm.WithLabelValues(id, ver).Set(1)
}

func AddedSchemaMetricUpdate(id string, ver string) {
	schemaRegisteredProm.WithLabelValues(id, ver).Set(1)
}

func DeletedSchemaMetricUpdate(id string) {
	schemaDeletedProm.WithLabelValues(id).Set(1)
}

func DeleteSchemaVersionMetricUpdate(id string, ver string) {
	schemaVersionDeletedProm.WithLabelValues(id, ver).Set(1)
}
