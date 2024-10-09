package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/internal/hashutils"
)

func TestGetSchemaVersionByIdAndVersion(t *testing.T) {
	// skip this test until it is not remodeled
	t.Skip()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create new sqlmock: %s", err)
	}

	dialector := postgres.New(postgres.Config{
		DriverName:           "postgres",
		DSN:                  "sqlmock_db_1",
		PreferSimpleProtocol: true,
		Conn:                 db,
	})
	gcfg := &gorm.Config{}
	dbInstance, err := gorm.Open(dialector, gcfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pdb := Repository{
		db: dbInstance,
	}
	resultRow := sqlmock.NewRows([]string{"version_id", "version", "schema_id", "specification", "description", "schema_hash", "created_at", "version_deactivated"}).
		AddRow(1, "1", 1, "test_spec", "a description", "9f8f1a88fdc11bf262095a82a607a61086641ad8da16ab4b6e104dd32920d20f", time.Now(), false)

	mock.ExpectQuery(`SELECT * FROM "version_details" WHERE schema_id = $1 and version = $2 and version_deactivated = $3 LIMIT 1`).
		WithArgs(1, "1", false).
		WillReturnRows(resultRow)

	sd, err := pdb.GetSchemaVersionByIdAndVersion("1", "1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "1", sd.Version)
	assert.Equal(t, "1", sd.SchemaID)
	assert.Equal(t, hashutils.SHA256([]byte("test_spec")), sd.SchemaHash)
}
