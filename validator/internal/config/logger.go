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
