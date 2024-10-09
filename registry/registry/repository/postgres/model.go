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
	"strconv"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
)

// Schema is a structure that defines the parent entity in the schema registry.
type Schema struct {
	SchemaID          uint             `gorm:"primaryKey;column:schema_id;autoIncrement"`
	SchemaType        string           `gorm:"column:schema_type;type:varchar(8)"`
	Name              string           `gorm:"column:name;type:varchar(256)"`
	Description       string           `gorm:"column:description;type:text"`
	LastCreated       string           `gorm:"column:last_created;type:varchar(8)"`
	PublisherID       string           `gorm:"column:publisher_id;type:varchar(256)"`
	VersionDetails    []VersionDetails `gorm:"foreignKey:schema_id"`
	CompatibilityMode string           `gorm:"column:compatibility_mode;type:varchar(256)"`
	ValidityMode      string           `gorm:"column:validity_mode;type:varchar(256)"`
}

// VersionDetails represents the child entity in the schema registry model.
type VersionDetails struct {
	VersionID          uint      `gorm:"primaryKey;column:version_id;autoIncrement"`
	Version            string    `gorm:"column:version;type:int;index:idver_idx"`
	SchemaID           uint      `gorm:"column:schema_id;index:idver_idx"`
	Description        string    `gorm:"column:description;type:text"`
	Specification      string    `gorm:"column:specification;type:text"`
	SchemaHash         string    `gorm:"column:schema_hash;type:varchar(256)"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	VersionDeactivated bool      `gorm:"column:version_deactivated;type:boolean"`
	Attributes         string    `gorm:"column:attributes;type:text"`
}

// intoRegistrySchema maps Schema from repository to service layer.
func intoRegistrySchema(schema Schema) registry.Schema {
	var registryVersionDetails []registry.VersionDetails
	for _, versionDetails := range schema.VersionDetails {
		registryVersionDetails = append(registryVersionDetails, intoRegistryVersionDetails(versionDetails))
	}

	return registry.Schema{
		SchemaID:          strconv.Itoa(int(schema.SchemaID)),
		SchemaType:        schema.SchemaType,
		Name:              schema.Name,
		VersionDetails:    registryVersionDetails,
		Description:       schema.Description,
		LastCreated:       schema.LastCreated,
		PublisherID:       schema.PublisherID,
		CompatibilityMode: schema.CompatibilityMode,
		ValidityMode:      schema.ValidityMode,
	}
}

// intoRegistryVersionDetails maps VersionDetails from repository to service layer.
func intoRegistryVersionDetails(VersionDetails VersionDetails) registry.VersionDetails {
	return registry.VersionDetails{
		VersionID:          strconv.Itoa(int(VersionDetails.VersionID)),
		Version:            VersionDetails.Version,
		SchemaID:           strconv.Itoa(int(VersionDetails.SchemaID)),
		Specification:      VersionDetails.Specification,
		Description:        VersionDetails.Description,
		SchemaHash:         VersionDetails.SchemaHash,
		CreatedAt:          VersionDetails.CreatedAt,
		VersionDeactivated: VersionDetails.VersionDeactivated,
		Attributes:         VersionDetails.Attributes,
	}
}
