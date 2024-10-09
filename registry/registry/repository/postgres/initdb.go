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
