package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-sr/compatibility"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errcodes"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry"
	"github.com/dataphos/aquarium-janitor-standalone-sr/registry/repository/postgres"
	"github.com/dataphos/aquarium-janitor-standalone-sr/validity"
)

func main() {
	registerCommand := flag.NewFlagSet("register", flag.ExitOnError)
	updateCommand := flag.NewFlagSet("update", flag.ExitOnError)

	if len(os.Args) < 2 {
		log.Fatal("register or update command must be provided")
	}

	switch os.Args[1] {
	case "register":
		registerSchema(registerCommand)
	case "update":
		updateSchema(updateCommand)
	default:
		log.Fatal("command not supported")
	}
}

func registerSchema(registerCommand *flag.FlagSet) {
	filename := registerCommand.String("f", "", "the json file containing schema specification")
	schemaType := registerCommand.String("t", "", "schema type")
	name := registerCommand.String("n", "schema-janitor", "schema name")
	description := registerCommand.String("d", "description of the schema", "schema description")
	publisherId := registerCommand.String("p", "publisherId", "publisher id")
	compMode := registerCommand.String("c", "", "compatibility mode")
	valMode := registerCommand.String("v", "", "validity mode")

	err := registerCommand.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}

	if *filename == "" {
		log.Fatal("filename must be provided")
	}

	if *schemaType == "" {
		log.Fatal("type must be provided")
	}

	if *valMode == "" {
		log.Fatal("validity mode must be provided")
	}

	if *compMode == "" {
		log.Fatal("compatibility mode must be provided")
	}

	file, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}

	schemaRegistrationRequest := registry.SchemaRegistrationRequest{
		Description:       *description,
		Specification:     string(file),
		Name:              *name,
		SchemaType:        *schemaType,
		PublisherID:       *publisherId,
		CompatibilityMode: *compMode,
		ValidityMode:      *valMode,
	}

	service := createService()
	details, added, err := service.CreateSchema(schemaRegistrationRequest)

	if err != nil {
		log.Fatal(err)
	}
	if !added {
		log.Print("schema already exists")
	} else {
		log.Print("created schema under the id ", details.VersionID)
	}
}

func updateSchema(updateCommand *flag.FlagSet) {
	filename := updateCommand.String("f", "", "the json file containing updated schema specification")
	description := updateCommand.String("d", "", "updated schema description")
	id := updateCommand.String("id", "", "id of the schema")

	err := updateCommand.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}

	if *filename == "" {
		log.Fatal("filename must be provided")
	}

	if *id == "" {
		log.Fatal("id must be provided")
	}

	file, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}

	schemaUpdateRequest := registry.SchemaUpdateRequest{
		Specification: string(file),
	}
	if *description != "" {
		schemaUpdateRequest.Description = *description
	}

	service := createService()
	details, updated, err := service.UpdateSchema(*id, schemaUpdateRequest)
	if err != nil {
		log.Fatal(err)
	}
	if !updated {
		log.Print("schema already exists")
	} else {
		log.Print("schema successfully updated, added version ", details.Version)
	}
}

func createService() *registry.Service {
	db, err := postgres.InitializeGormFromEnv()
	if err != nil {
		log.Fatal(err, errcodes.DatabaseConnectionInitialization)
	}
	if !postgres.HealthCheck(db) {
		log.Fatal("database state invalid", errcodes.InvalidDatabaseState)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	compChecker, globalCompMode, err := compatibility.InitCompatibilityChecker(ctx)
	if err != nil {
		log.Fatal(err, errcodes.ExternalCheckerInitialization)
	}

	valChecker, globalValMode, err := validity.InitExternalValidityChecker(ctx)
	if err != nil {
		log.Fatal(err, errcodes.ExternalCheckerInitialization)
	}

	service := registry.New(postgres.New(db), compChecker, valChecker, globalCompMode, globalValMode)
	return service
}
