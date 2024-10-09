package postgres

import (
	"os"

	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errtemplates"
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
