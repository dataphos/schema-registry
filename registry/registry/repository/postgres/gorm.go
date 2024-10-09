package postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitializeGormFromEnv() (*gorm.DB, error) {
	config, err := LoadDatabaseConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return InitializeGorm(config)
}

func InitializeGorm(config DatabaseConfig) (*gorm.DB, error) {
	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		config.Host, config.User, config.Password, config.DatabaseName,
	)
	dialector := postgres.Open(connectionString)
	gcfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   config.TablePrefix,
			SingularTable: true,
		},
	}

	db, err := gorm.Open(dialector, gcfg)
	if err != nil {
		return nil, err
	}

	return db, nil
}
