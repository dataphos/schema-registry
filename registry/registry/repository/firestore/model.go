package firestore

import (
	"encoding/base64"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
	"strconv"
	"time"
)

// Schema is a structure that defines the parent entity in the schema registry.
type Schema struct {
	SchemaID       string           `json:"schema_id,omitempty" bson:"schema-id,omitempty" firestore:"schema-id"`
	SchemaType     string           `json:"schema_type" bson:"schema-type" firestore:"schema-type"`
	Name           string           `json:"name" bson:"name" firestore:"name"`
	Description    string           `json:"description" bson:"description" firestore:"description"`
	LastCreated    string           `json:"last_created" bson:"last-created" firestore:"last-created"`
	PublisherID    string           `json:"publisher_id" bson:"publisher-id" firestore:"publisher-id"`
	VersionDetails []VersionDetails `json:"schemas" bson:"schemas" firestore:"schemas"`
}

// VersionDetails represents the child entity in the schema registry model.
type VersionDetails struct {
	VersionID          string    `json:"id,omitempty" bson:"_id,omitempty" firestore:"ID"`
	Version            int32     `json:"version" bson:"version" firestore:"version"`
	SchemaID           string    `json:"schemaID" bson:"schemaID" firestore:"schemaID"`
	Description        string    `json:"description" bson:"description" firestore:"description"`
	Specification      string    `json:"specification" bson:"specification" firestore:"specification"`
	SchemaHash         string    `json:"schema_hash" bson:"schema-hash" firestore:"schema-hash"`
	CreatedAt          time.Time `json:"creation_date" bson:"creation-date" firestore:"creation-date"`
	VersionDeactivated bool      `json:"version_deactivated" bson:"version-deactivated" firestore:"version-deactivated"`
}

// intoRegistrySchema maps Schema from repository to service layer.
func intoRegistrySchema(schema Schema) registry.Schema {
	var registryVersionDetails []registry.VersionDetails
	for _, versionDetails := range schema.VersionDetails {
		registryVersionDetails = append(registryVersionDetails, intoRegistryVersionDetails(versionDetails))
	}

	return registry.Schema{
		SchemaID:       schema.SchemaID,
		SchemaType:     schema.SchemaType,
		Name:           schema.Name,
		VersionDetails: registryVersionDetails,
		Description:    schema.Description,
		LastCreated:    schema.LastCreated,
		PublisherID:    schema.PublisherID,
	}
}

// intoRegistryVersionDetails maps VersionDetails from repository to service layer.
func intoRegistryVersionDetails(VersionDetails VersionDetails) registry.VersionDetails {
	return registry.VersionDetails{
		VersionID:          VersionDetails.VersionID,
		Version:            strconv.Itoa(int(VersionDetails.Version)),
		SchemaID:           VersionDetails.SchemaID,
		Specification:      VersionDetails.Specification,
		Description:        VersionDetails.Description,
		SchemaHash:         VersionDetails.SchemaHash,
		CreatedAt:          VersionDetails.CreatedAt,
		VersionDeactivated: VersionDetails.VersionDeactivated,
	}
}

func fillOutSchema(schema registry.SchemaRegistrationRequest, documentID, hash string, version int32) *Schema {
	b64coder := base64.StdEncoding

	return &Schema{
		SchemaID:   documentID,
		SchemaType: schema.SchemaType,
		Name:       schema.Name,
		VersionDetails: []VersionDetails{
			{
				VersionID:     documentID,
				Version:       version,
				SchemaID:      documentID,
				Specification: b64coder.EncodeToString([]byte(schema.Specification)),
				Description:   "",
				SchemaHash:    hash,
				CreatedAt:     time.Now(),
			},
		},
		LastCreated: time.Now().String(),
		Description: schema.Description,
		PublisherID: "",
	}
}
