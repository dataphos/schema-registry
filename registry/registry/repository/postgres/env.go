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

package postgres

import (
	"os"

	"github.com/dataphos/schema-registry/internal/errtemplates"
)

type DatabaseConfig struct {
	TablePrefix  string
	Host         string
	User         string
	Password     string
	DatabaseName string
}

const (
	tablePrefixEnvKey  = "SR_TABLE_PREFIX"
	hostEnvKey         = "SR_HOST"
	userEnvKey         = "SR_USER"
	passwordEnvKey     = "SR_PASSWORD"
	databaseNameEnvKey = "SR_DBNAME"
)

func LoadDatabaseConfigFromEnv() (DatabaseConfig, error) {
	tablePrefix := os.Getenv(tablePrefixEnvKey)
	if tablePrefix == "" {
		return DatabaseConfig{}, errtemplates.EnvVariableNotDefined(tablePrefixEnvKey)
	}

	host := os.Getenv(hostEnvKey)
	if host == "" {
		return DatabaseConfig{}, errtemplates.EnvVariableNotDefined(hostEnvKey)
	}

	user := os.Getenv(userEnvKey)
	if user == "" {
		return DatabaseConfig{}, errtemplates.EnvVariableNotDefined(userEnvKey)
	}

	password := os.Getenv(passwordEnvKey)
	if password == "" {
		return DatabaseConfig{}, errtemplates.EnvVariableNotDefined(passwordEnvKey)
	}

	dbName := os.Getenv(databaseNameEnvKey)
	if dbName == "" {
		return DatabaseConfig{}, errtemplates.EnvVariableNotDefined(databaseNameEnvKey)
	}

	return DatabaseConfig{
		TablePrefix:  tablePrefix,
		Host:         host,
		User:         user,
		Password:     password,
		DatabaseName: dbName,
	}, nil
}
