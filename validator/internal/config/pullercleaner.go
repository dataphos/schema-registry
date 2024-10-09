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
