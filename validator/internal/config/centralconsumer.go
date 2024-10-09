package config

import (
	"time"

	"github.com/kkyr/fig"
)

// CentralConsumer represents all required configuration to run an instance of central consumer.
type CentralConsumer struct {
	Producer               Producer                  `toml:"producer"`
	Consumer               Consumer                  `toml:"consumer"`
	Registry               Registry                  `toml:"registry"`
	Topics                 CentralConsumerTopics     `toml:"topics"`
	Validators             CentralConsumerValidators `toml:"validators"`
	ShouldLog              CentralConsumerShouldLog  `toml:"should_log"`
	NumSchemaCollectors    int                       `toml:"num_schema_collectors" default:"-1"`
	NumInferrers           int                       `toml:"num_inferrers" default:"-1"`
	MetricsLoggingInterval time.Duration             `toml:"metrics_logging_interval" default:"5s"`
	RunOptions             RunOptions                `toml:"run_option"`
	Mode                   int                       `toml:"mode"`
	SchemaID               string                    `toml:"schema_id"`
	SchemaVersion          string                    `toml:"schema_version"`
	SchemaType             string                    `toml:"schema_type"`
	Encryption             Encryption                `toml:"encryption"`
}

type Encryption struct {
	EncryptionKey string `toml:"encryption_key"`
}

type CentralConsumerTopics struct {
	Valid      string `toml:"valid" val:"required"`
	DeadLetter string `toml:"dead_letter" val:"required"`
}

type CentralConsumerValidators struct {
	EnableAvro        bool          `toml:"enable_avro"`
	EnableCsv         bool          `toml:"enable_csv"`
	EnableJson        bool          `toml:"enable_json"`
	EnableProtobuf    bool          `toml:"enable_protobuf"`
	EnableXml         bool          `toml:"enable_xml"`
	CsvUrl            string        `toml:"csv_url" val:"required_if=EnableCsv true,omitempty,url"`
	CsvTimeoutBase    time.Duration `toml:"csv_timeout_base" default:"2s"`
	JsonUseAltBackend bool          `toml:"json_use_alt_backend"`
	JsonCacheSize     int           `toml:"json_cache_size" default:"100"`
	ProtobufFilePath  string        `toml:"protobuf_file_path" default:"/app/.schemas"`
	ProtobufCacheSize int           `toml:"protobuf_cache_size" default:"100"`
	XmlUrl            string        `toml:"xml_url" val:"required_if=EnableXml true,omitempty,url"`
	XmlTimeoutBase    time.Duration `toml:"xml_timeout_base" default:"3s"`
}

type CentralConsumerShouldLog struct {
	MissingSchema bool `toml:"missing_schema"`
	Valid         bool `toml:"valid"`
	DeadLetter    bool `toml:"dead_letter"`
}

// Read loads parameters from configuration file into CentralConsumer struct.
func (cfg *CentralConsumer) Read(filename string) error {
	return fig.Load(cfg, fig.File(filename), fig.Tag("toml"), fig.UseEnv(""))
}

// Validate validates CentralConsumer struct.
func (cfg *CentralConsumer) Validate() error {
	return validate(cfg, "CentralConsumer.")
}
