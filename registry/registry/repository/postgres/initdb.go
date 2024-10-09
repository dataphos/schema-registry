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
	"gorm.io/gorm"
)

// Initdb initializes the schema registry database.
func Initdb(db *gorm.DB) error {
	if err := db.Exec("create schema if not exists syntio_schema authorization postgres").Error; err != nil {
		return err
	}
	return db.AutoMigrate(&Schema{}, &VersionDetails{})
}

// HealthCheck checks if the necessary tables exist.
//
// Note that this function returns false in case of network issues as well, acting like a health check of sorts.
func HealthCheck(db *gorm.DB) bool {
	migrator := db.Migrator()
	return migrator.HasTable(&Schema{}) && migrator.HasTable(&VersionDetails{})
}
