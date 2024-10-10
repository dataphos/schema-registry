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

// Package errcodes contains all the error codes used by janitor subcomponents.
package errcodes

const (
	// RegistryInitialization marks unsuccessful initialization of the schema registry dependency.
	RegistryInitialization = 100

	// RegistryUnresponsive marks an unsuccessful attempt at a schema registry operation due to the registry being unresponsive.
	RegistryUnresponsive = 101

	// BrokerInitialization marks unsuccessful initialization of the message broker related external dependencies.
	BrokerInitialization = 300

	// PullingFailure marks failures which occur while pulling messages from some source.
	PullingFailure = 301

	// PublishingFailure marks an unsuccessful attempt at message publishing.
	PublishingFailure = 302

	// BrokerConnClosed marks unsuccessful closing of the connection to the message broker.
	BrokerConnClosed = 303

	// TLSInitialization marks an unsuccessful initialization of a TLS configuration.
	TLSInitialization = 304

	// MetricsServerFailure marks failure of an HTTP server for metrics.
	MetricsServerFailure = 305

	// MetricsServerShutdownFailure marks an unsuccessful shutdown of an HTTP server for metrics.
	MetricsServerShutdownFailure = 306

	// ReadConfigFailure marks unsuccessful read of .yaml file into janitorctl structure.
	ReadConfigFailure = 400

	// ValidateConfigFailure marks unsuccessful validation of janitorctl's exposed fields.
	ValidateConfigFailure = 401

	// ValidationFailure marks an unsuccessful attempt at message validation.
	ValidationFailure = 500

	// InvalidMessage marks messages which were inferred to be invalid.
	InvalidMessage = 501

	// DeadletterMessage marks messages which were inferred to be deadletter.
	DeadletterMessage = 502

	// SchemaGeneration marks an unsuccessful attempt at schema generation.
	SchemaGeneration = 600

	// Initialization is used for general initialization failure of internal structures only,
	// initialization failure of external dependencies is marked through other, more descriptive error codes,
	Initialization = 900

	// ParsingMessage marks an unsuccessful attempt at mapping a broker message structure into the one used for processing.
	ParsingMessage = 901

	// UnrecoverableErrorEncountered declares that the system encountered an unrecoverable error.
	UnrecoverableErrorEncountered = 902

	// ErrorThresholdReached declares that the system encountered at least the threshold amount of errors.
	ErrorThresholdReached = 903

	// CompletedWithErrors marks that the process has completed but errors occurred.
	CompletedWithErrors = 904

	// Miscellaneous is used when no other available error code is fitting.
	Miscellaneous = 999
)
