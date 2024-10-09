package postgres

import (
	"encoding/base64"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/internal/hashutils"
)

type Repository struct {
	db *gorm.DB
}

// New returns a new instance of Repository.
func New(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetSchemaVersionByIdAndVersion retrieves a schema version by its id and version.
// Returns registry.ErrNotFound in case there's no schema under the given id and version.
func (r *Repository) GetSchemaVersionByIdAndVersion(id, version string) (registry.VersionDetails, error) {
	var details VersionDetails
	if err := r.db.Where("schema_id = ? and version = ? and version_deactivated = ?", id, version, false).Take(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return registry.VersionDetails{}, registry.ErrNotFound
		}
		return registry.VersionDetails{}, err
	}
	return intoRegistryVersionDetails(details), nil
}

// GetSchemaVersionsById returns a Schema with all active versions.
// Returns registry.ErrNotFound in case there's no schema under the given id.
func (r *Repository) GetSchemaVersionsById(id string) (registry.Schema, error) {
	var schema Schema
	err := r.db.Preload("VersionDetails", "version_deactivated = ?", false).Take(&schema, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || len(schema.VersionDetails) == 0 {
		return registry.Schema{}, registry.ErrNotFound
	}
	if err != nil {
		return registry.Schema{}, err
	}
	return intoRegistrySchema(schema), nil
}

// GetAllSchemaVersions returns a Schema with all versions.
// Returns registry.ErrNotFound in case there's no schema under the given id.
func (r *Repository) GetAllSchemaVersions(id string) (registry.Schema, error) {
	var schema Schema
	if err := r.db.Preload("VersionDetails").Take(&schema, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return registry.Schema{}, registry.ErrNotFound
		}
		return registry.Schema{}, err
	}
	return intoRegistrySchema(schema), nil
}

// GetLatestSchemaVersion returns the latest active version of selected schema.
// Returns registry.ErrNotFound in case there's no schema under the given id.
func (r *Repository) GetLatestSchemaVersion(id string) (registry.VersionDetails, error) {
	var details VersionDetails
	if err := r.db.Where("schema_id = ? and version_deactivated = ?", id, false).Last(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return registry.VersionDetails{}, registry.ErrNotFound
		}
		return registry.VersionDetails{}, err
	}
	return intoRegistryVersionDetails(details), nil
}

// GetSchemas returns all active Schema instances.
// Returns registry.ErrNotFound in case there's no schemas.
func (r *Repository) GetSchemas() ([]registry.Schema, error) {
	var schemaList []Schema
	// This query examines if there is at least one active version of the schema and based on that, it determines whether to retrieve the schema.
	tx := r.db.Preload("VersionDetails", "version_deactivated = ?", false).Where("EXISTS (SELECT 1 FROM syntio_schema.version_details WHERE syntio_schema.version_details.schema_id = syntio_schema.schema.schema_id AND syntio_schema.version_details.version_deactivated = 'false')").Find(&schemaList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, registry.ErrNotFound
	}

	var registrySchemaList []registry.Schema
	for _, schema := range schemaList {
		registrySchemaList = append(registrySchemaList, intoRegistrySchema(schema))
	}
	return registrySchemaList, nil
}

// GetAllSchemas returns all Schema instances.
// Returns registry.ErrNotFound in case there's no schemas.
func (r *Repository) GetAllSchemas() ([]registry.Schema, error) {
	var schemaList []Schema
	tx := r.db.Preload("VersionDetails").Find(&schemaList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, registry.ErrNotFound
	}

	registrySchemaList := make([]registry.Schema, len(schemaList))
	for i, schema := range schemaList {
		registrySchemaList[i] = intoRegistrySchema(schema)
	}
	return registrySchemaList, nil
}

// CreateSchema inserts a new Schema structure.
// Returns a new VersionDetails structure and a bool flag indicating if a new version of schema was added or if it already existed.
func (r *Repository) CreateSchema(schemaRegisterRequest registry.SchemaRegistrationRequest) (registry.VersionDetails, bool, error) {
	specification := []byte(schemaRegisterRequest.Specification)
	hash := hashutils.SHA256(specification)

	// Prior to saving the schema in the database we must verify the distinctness of the schema hash and publisher ID.
	// To accomplish this, we must join the "VersionDetails" and "Schema" tables on the columns that contain the schema ID,
	// while also filtering the schemas with the specified schema hash and publisher ID. If the query does not return a schema,
	// it means that a schema with the given criteria does not exist in the database and a new one needs to be created.
	var schema Schema
	if err := r.db.Table("syntio_schema.schema").Preload("VersionDetails", "schema_hash = ? and version_deactivated = ?", hash, false).Joins("JOIN syntio_schema.version_details ON syntio_schema.version_details.schema_id = syntio_schema.schema.schema_id AND syntio_schema.version_details.schema_hash = ? and syntio_schema.version_details.version_deactivated = ?", hash, false).Where("syntio_schema.schema.publisher_id = ?", schemaRegisterRequest.PublisherID).Take(&schema).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			schema := Schema{
				SchemaType:        strings.ToLower(schemaRegisterRequest.SchemaType),
				Name:              schemaRegisterRequest.Name,
				Description:       schemaRegisterRequest.Description,
				PublisherID:       schemaRegisterRequest.PublisherID,
				LastCreated:       "1",
				CompatibilityMode: schemaRegisterRequest.CompatibilityMode,
				ValidityMode:      schemaRegisterRequest.ValidityMode,
				VersionDetails: []VersionDetails{
					{
						Version:            "1",
						Specification:      base64.StdEncoding.EncodeToString(specification),
						Description:        schemaRegisterRequest.Description,
						SchemaHash:         hash,
						CreatedAt:          time.Now(),
						VersionDeactivated: false,
						Attributes:         schemaRegisterRequest.Attributes,
					},
				},
			}
			if err := r.db.Create(&schema).Error; err != nil {
				return registry.VersionDetails{}, false, err
			}
			return intoRegistryVersionDetails(schema.VersionDetails[0]), true, nil
		}
		return registry.VersionDetails{}, false, err
	}

	return intoRegistryVersionDetails(schema.VersionDetails[0]), false, nil
}

// UpdateSchemaById updates the schema specification and description if sent.
// Returns the new VersionDetails and a flag indicating if a new version of schema was added.
func (r *Repository) UpdateSchemaById(id string, schemaUpdateRequest registry.SchemaUpdateRequest) (registry.VersionDetails, bool, error) {
	schemaId, err := strconv.Atoi(id)
	if err != nil {
		return registry.VersionDetails{}, false, errors.Wrap(err, "wrong type of schemaID")
	}

	specification := []byte(schemaUpdateRequest.Specification)
	hash := hashutils.SHA256(specification)

	var details VersionDetails
	if err = r.db.Where("schema_hash = ? and schema_id = ?", hash, id).Take(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			updated := VersionDetails{}
			err = r.db.Transaction(func(tx *gorm.DB) error {

				schema := &Schema{SchemaID: uint(schemaId)}
				if err := tx.Select("last_created").Take(&schema).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return registry.ErrNotFound
					}
					return err
				}

				lastCreated, err := strconv.Atoi(schema.LastCreated)
				if err != nil {
					return errors.Wrap(err, "wrong type of latest version")
				}
				incrementedLastCreated := strconv.Itoa(lastCreated + 1)

				updated = VersionDetails{
					Version:       incrementedLastCreated,
					Specification: base64.StdEncoding.EncodeToString(specification),
					SchemaHash:    hash,
					Description:   schemaUpdateRequest.Description,
					Attributes:    schemaUpdateRequest.Attributes,
				}

				// append the new version to the VersionDetails array
				if err = tx.Model(&schema).Association("VersionDetails").Append(&updated); err != nil {
					return errors.Wrap(err, "could not update version details")
				}

				// updating description and last_created values in schema table
				if err = tx.Model(&schema).Updates(Schema{Description: schemaUpdateRequest.Description, LastCreated: incrementedLastCreated}).Error; err != nil {
					return errors.Wrap(err, "could not update schema")
				}

				return nil
			})
			if err != nil {
				return registry.VersionDetails{}, false, err
			}
			return intoRegistryVersionDetails(updated), true, nil
		}
		return registry.VersionDetails{}, false, err
	}

	if details.VersionDeactivated {
		//Activates already existing Schema
		var schema Schema
		if err := r.db.Take(&schema, id).Error; err != nil {
			return registry.VersionDetails{}, false, err
		}
		lastCreated, err := strconv.Atoi(schema.LastCreated)
		if err != nil {
			return registry.VersionDetails{}, false, errors.Wrap(err, "wrong type of latest version")
		}
		incrementedLastCreated := strconv.Itoa(lastCreated + 1)

		// updating description and last_created values in schema table
		if err = r.db.Model(&Schema{SchemaID: uint(schemaId)}).Updates(Schema{Description: schemaUpdateRequest.Description, LastCreated: incrementedLastCreated}).Error; err != nil {
			return registry.VersionDetails{}, false, errors.Wrap(err, "could not update schema")
		}

		// activating the schema version with a new creation time and version number
		if err = r.db.Model(&details).Updates(map[string]interface{}{
			"created_at":          time.Now(),
			"version_deactivated": false,
			"version":             incrementedLastCreated,
		}).Error; err != nil {
			return registry.VersionDetails{}, false, errors.Wrap(err, "could not update version details")
		}

		details.VersionDeactivated = false
		return intoRegistryVersionDetails(details), true, nil
	}
	return intoRegistryVersionDetails(details), false, nil
}

// DeleteSchema deactivates a schema.
// Returns a boolean flag indicating if a schema with the given id existed before this call.
func (r *Repository) DeleteSchema(id string) (bool, error) {
	var schema Schema
	if err := r.db.Preload("VersionDetails", "version_deactivated = ?", false).Take(&schema, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if len(schema.VersionDetails) == 0 {
		return false, nil
	}
	// deactivation of all active versions
	tx := r.db.Model(&schema.VersionDetails).Update("version_deactivated", true)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

// DeleteSchemaVersion deactivates the specified schema version.
// Returns a boolean flag indicating if a schema with the given id and version existed before this call.
func (r *Repository) DeleteSchemaVersion(id, version string) (bool, error) {
	var details VersionDetails
	if err := r.db.Where("schema_id = ? and version = ? and version_deactivated = ?", id, version, false).Take(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	tx := r.db.Model(&details).Update("version_deactivated", true)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
