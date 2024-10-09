package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/base64"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/internal/hashutils"
)

// Rewrite all functions to work differently based on whether the schema is active or not.

// Firestore implementation for database.DBExecutor interface.
type DB struct {
	Collection string
	Client     *firestore.Client
}

// CREDENTIAL FILE NEED TO BE SET UNDER ENV VAR "GOOGLE_APPLICATION_CREDENTIALS"
func New() (*DB, error) {
	ctx := context.Background()

	collection := os.Getenv("COLLECTION")
	projectId := os.Getenv("PROJECT_ID")

	client, initErr := firestore.NewClient(ctx, projectId)
	if initErr != nil {
		log.Fatalf("Firestore client initialization failed. Server can't start properly.\nError: %s", initErr)
	}
	return &DB{
		Collection: collection,
		Client:     client,
	}, nil
}

// DeleteSchema takes the id of the schema which needs to be deleted
// will return bool that indicates if the schema was deleted and error
func (db *DB) DeleteSchema(id string) (bool, error) {
	ctx := context.Background()
	_, err := db.Client.Collection(db.Collection).Doc(id).Delete(ctx)
	if err != nil {
		return false, err
	}
	return true, err
}

// DeleteSchemaVersion takes the id of the schema and version which needs to be deleted
// will return bool that indicates if the schema was deleted and error
func (db *DB) DeleteSchemaVersion(id, version string) (bool, error) {
	ctx := context.Background()
	deleted := false // flag to track if the version exists

	document, err := db.Client.Collection(db.Collection).Doc(id).Get(ctx)
	// remove the one that has the version==version
	var result *Schema
	if err = document.DataTo(&result); err != nil {
		return false, err
	}
	var newDetails []VersionDetails
	for i, v := range result.VersionDetails {
		versionInt, _ := strconv.Atoi(version)
		if v.Version != int32(versionInt) {
			newDetails = append(newDetails, result.VersionDetails[i:i+1]...)
			continue
		}
		// set deleted to true, so we know that the version existed
		deleted = true
	}
	if deleted {
		// update the new version details only if the version exists
		_, err = db.Client.Collection(db.Collection).Doc(id).Set(ctx, map[string]interface{}{
			"schemas": newDetails,
		}, firestore.MergeAll)
		if err != nil {
			return false, err
		}
	}

	return deleted, nil
}

// GetSchemaVersionByIdAndVersion retrieves a schema by its id and version. Method returns a boolean flag which defines
// if the wanted schema exists. If the schema is found, representing Payload is retrieved.
func (db *DB) GetSchemaVersionByIdAndVersion(id, version string) (registry.VersionDetails, error) {
	ctx := context.Background()
	document, err := db.Client.Collection(db.Collection).Doc(id).Get(ctx)
	if err != nil {
		log.Printf("could not get schema by that ID: %v", err)
		return registry.VersionDetails{}, err
	}
	var result *Schema
	if err = document.DataTo(&result); err != nil {
		return registry.VersionDetails{}, err
	}

	for i, v := range result.VersionDetails {
		versionInt, _ := strconv.Atoi(version)
		if v.Version == int32(versionInt) {
			newDetails := result.VersionDetails[i : i+1]
			return intoRegistryVersionDetails(newDetails[0]), nil
		}
	}
	return registry.VersionDetails{}, err
}

// GetSchemaVersionsById returns a list of VersionDetails from the database.
// The input arguments are the request context and the document/row ID.
// The return value is a list of model.VersionDetails and an error in case of a fault or failure.
// Has to be rewritten to return only active schemas
func (db *DB) GetSchemaVersionsById(id string) (registry.Schema, error) {
	ctx := context.Background()
	document, err := db.Client.Collection(db.Collection).Doc(id).Get(ctx)
	if err != nil {
		log.Printf("could not get schema by that ID: %v", err)
		return registry.Schema{}, err
	}
	var result Schema
	if err = document.DataTo(&result); err != nil {
		return registry.Schema{}, err
	}
	return intoRegistrySchema(result), nil
}

// GetAllSchemaVersions returns a list of VersionDetails from the database.
// The input arguments are the request context and the document/row ID.
// The return value is a list of model.VersionDetails and an error in case of a fault or failure.
func (db *DB) GetAllSchemaVersions(id string) (registry.Schema, error) {
	ctx := context.Background()
	document, err := db.Client.Collection(db.Collection).Doc(id).Get(ctx)
	if err != nil {
		log.Printf("could not get schema by that ID: %v", err)
		return registry.Schema{}, err
	}
	var result Schema
	if err = document.DataTo(&result); err != nil {
		return registry.Schema{}, err
	}
	return intoRegistrySchema(result), nil
}

// GetLatestSchemaVersion retrieves the latest VersionDetails of a schema by its id.
func (db *DB) GetLatestSchemaVersion(id string) (registry.VersionDetails, error) {
	document, err := db.Client.Collection(db.Collection).Doc(id).Get(context.Background())
	if err != nil {
		log.Printf("could not get schema by that ID: %v", err)
		return registry.VersionDetails{}, err
	}
	var result Schema
	if err = document.DataTo(&result); err != nil {
		return registry.VersionDetails{}, err
	}

	for i, v := range result.VersionDetails {
		// because versions start from 1
		if int(v.Version) == len(result.VersionDetails) {
			newDetails := result.VersionDetails[i : i+1]
			result.VersionDetails = newDetails
			return intoRegistryVersionDetails(result.VersionDetails[0]), nil
		}
	}
	return registry.VersionDetails{}, nil
}

// GetSchemas retrieves all schemas in the Collection
// Has to be rewritten to return only active schemas
func (db *DB) GetSchemas() ([]registry.Schema, error) {
	var allSchemas []registry.Schema
	iter := db.Client.Collection(db.Collection).Documents(context.Background())
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var result Schema
		if err = doc.DataTo(&result); err != nil {
			return nil, err
		}
		allSchemas = append(allSchemas, intoRegistrySchema(result))
	}

	return allSchemas, nil
}

// GetAllSchemas retrieves all schemas in the Collection
func (db *DB) GetAllSchemas() ([]registry.Schema, error) {
	var allSchemas []registry.Schema
	iter := db.Client.Collection(db.Collection).Documents(context.Background())
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var result Schema
		if err = doc.DataTo(&result); err != nil {
			return nil, err
		}
		allSchemas = append(allSchemas, intoRegistrySchema(result))
	}
	return allSchemas, nil
}

// UpdateSchemaById updates the schema Specification e.g. creates a new entry of VersionDetails.
// The input arguments are the request context, followed by the ID string of the document/row ID,
// specification in []byte form and a flag indicating if Schema was manually updated or dynamically evolved.
// The output is a model.InsertInfo structure, a flag indicating if new version of schema was added and an error.
func (db *DB) UpdateSchemaById(id string, schemaUpdateRequest registry.SchemaUpdateRequest) (registry.VersionDetails, bool, error) {
	ctx := context.Background()
	document, err := db.Client.Collection(db.Collection).Doc(id).Get(ctx)
	if err != nil {
		log.Printf("could not get schema by that ID: %v", err)
		return registry.VersionDetails{}, false, err
	}

	var result *Schema
	if err = document.DataTo(&result); err != nil {
		return registry.VersionDetails{}, false, err
	}

	hash := hashutils.SHA256([]byte(schemaUpdateRequest.Specification))

	if exists, info := db.existsByHash(hash, result); exists {
		log.Printf("Schema %v with version %d already exists ", result.SchemaID, info.Version)
		return intoRegistryVersionDetails(*info), false, nil
	}

	newVersion := int32(len(result.VersionDetails) + 1)
	d := &VersionDetails{
		VersionID:     id,
		SchemaID:      id,
		Version:       newVersion,
		SchemaHash:    hash,
		Specification: base64.StdEncoding.EncodeToString([]byte(schemaUpdateRequest.Specification)),
	}
	details := result.VersionDetails
	details = append(details, *d)
	result.VersionDetails = details

	if _, err := db.Client.Collection(db.Collection).Doc(id).Set(ctx, result); err != nil {
		log.Println("Could not update new schema")
		return registry.VersionDetails{}, false, err
	}

	return registry.VersionDetails{
		VersionID:     id,
		SchemaID:      id,
		Version:       strconv.Itoa(int(newVersion)),
		Specification: schemaUpdateRequest.Specification,
		Description:   schemaUpdateRequest.Description,
		SchemaHash:    hash,
		CreatedAt:     time.Now(),
	}, true, nil
}

// CreateSchema persists a new Schema structure into the document or relational database.
// Input arguments are request context, and a DTO describing the basic new schema parameters.
// The output is a model.InsertInfo structure, a flag indicating if new version of schema was added and an error.
func (db *DB) CreateSchema(schemaRegisterRequest registry.SchemaRegistrationRequest) (registry.VersionDetails, bool, error) {
	ctx := context.Background()
	byteSchema := []byte(schemaRegisterRequest.Specification)
	hash := hashutils.SHA256(byteSchema)

	it := db.Client.Collection(db.Collection).Documents(ctx)
	for sh, err := it.Next(); err != iterator.Done; sh, err = it.Next() {
		if err != nil {
			log.Println("Could not read existing data")
			return registry.VersionDetails{}, false, err
		}
		var schema *Schema
		err = sh.DataTo(&schema)
		if err != nil {
			log.Printf("err while transfering schema to struct: %v", err)
			return registry.VersionDetails{}, false, err
		}
		if exists, _ := db.existsByHash(hash, schema); exists {
			return intoRegistryVersionDetails(schema.VersionDetails[0]), false, err
		}
	}
	version := int32(1)
	doc := db.Client.Collection(db.Collection).NewDoc()

	sc := fillOutSchema(schemaRegisterRequest, doc.ID, hash, version)
	if _, err := doc.Set(ctx, sc); err != nil {
		log.Println("Could not create new schema")
		return registry.VersionDetails{}, false, err
	}
	return intoRegistryVersionDetails(sc.VersionDetails[0]), true, nil
}

// existsByHash is a helper function that checks if there is a corresponding hash in a schema.
// the input arguments are a sha-256 hash of the schema specification and a model.Schema struct.
// the output is a flag indicating if the schema exists, and were the flag true the corresponding model.InsertInfo
// struct.
func (db *DB) existsByHash(hash string, schema *Schema) (bool, *VersionDetails) {
	var info *VersionDetails
	for _, sd := range schema.VersionDetails {
		if sd.SchemaHash == hash {
			info = &VersionDetails{
				VersionID:     sd.VersionID,
				Version:       sd.Version,
				SchemaID:      sd.SchemaID,
				Specification: sd.Specification,
				SchemaHash:    hash,
				Description:   sd.Description,
			}
			return true, info
		}
	}
	return false, nil
}
