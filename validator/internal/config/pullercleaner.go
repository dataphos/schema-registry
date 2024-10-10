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

package config

import (
	"time"

	"github.com/kkyr/fig"
)

// PullerCleaner represents all required configuration to run an instance of puller cleaner.
type PullerCleaner struct {
	Producer               Producer                `toml:"producer"`
	Consumer               Consumer                `toml:"consumer"`
	Registry               Registry                `toml:"registry"`
	Topics                 PullerCleanerTopics     `toml:"topics"`
	Validators             PullerCleanerValidators `toml:"validators"`
	ShouldLog              PullerCleanerShouldLog  `toml:"should_log"`
	NumCleaners            int                     `toml:"num_cleaners" default:"10"`
	MetricsLoggingInterval time.Duration           `toml:"metrics_logging_interval" default:"5s"`
	RunOptions             RunOptions              `toml:"run_options"`
}

type PullerCleanerTopics struct {
	Valid      string `toml:"valid" val:"required"`
	DeadLetter string `toml:"dead_letter" val:"required"`
}

type PullerCleanerValidators struct {
	EnableCsv           bool          `toml:"enable_csv"`
	EnableJson          bool          `toml:"enable_json"`
	CsvUrl              string        `toml:"csv_url" val:"required_if=EnableCsv true,omitempty,url"`
	CsvTimeoutBase      time.Duration `toml:"csv_timeout_base" default:"2s"`
	JsonUseAltBackend   bool          `toml:"json_use_alt_backend"`
	JsonCacheSize       int           `toml:"json_cache_size" default:"100"`
	JsonSchemaGenScript string        `toml:"json_schema_gen_script" val:"required_if=EnableJson true,omitempty,file"`
}

type PullerCleanerShouldLog struct {
	Valid      bool `toml:"valid"`
	DeadLetter bool `toml:"dead_letter"`
}

// Read loads parameters from configuration file into PullerCleaner struct.
func (cfg *PullerCleaner) Read(filename string) error {
	return fig.Load(cfg, fig.File(filename), fig.Tag("toml"), fig.UseEnv(""))
}

// Validate validates PullerCleaner struct.
func (cfg *PullerCleaner) Validate() error {
	return validate(cfg, "PullerCleaner.")
}
