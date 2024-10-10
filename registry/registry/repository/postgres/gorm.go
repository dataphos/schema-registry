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
