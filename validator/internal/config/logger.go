package config

import (
	"fmt"
	"github.com/dataphos/lib-logger/logger"
	"os"
)

const (
	LogLevelEnvKey = "LOG_LEVEL_MINIMUM"
)

const (
	InfoLevel    = "info"
	WarnLevel    = "warn"
	ErrorLevel   = "error"
	DefaultLevel = InfoLevel
)

var levels = map[string]logger.Level{InfoLevel: logger.LevelInfo, WarnLevel: logger.LevelWarn, ErrorLevel: logger.LevelError}

// GetLogLevel returns minimum log level based on environment variable.
// Possible levels are info, warn, and error. Defaults to info.
func GetLogLevel() (logger.Level, []string) {
	warnings := make([]string, 0, 2) // warnings about log config to be logged after the logger is configured

	levelString := os.Getenv(LogLevelEnvKey)
	if levelString == "" {
		warnings = append(warnings, fmt.Sprintf("Value for '%s' not set! Using level %s.", LogLevelEnvKey, DefaultLevel))
		return levels[DefaultLevel], warnings
	}

	level, supported := levels[levelString]
	if supported {
		return level, warnings
	} else {
		warnings = append(warnings, fmt.Sprintf("Value %v for %v is not supported, using level %v.", levelString, LogLevelEnvKey, DefaultLevel))
		return levels[DefaultLevel], warnings
	}
}
