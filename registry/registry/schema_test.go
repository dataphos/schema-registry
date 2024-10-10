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

package registry

import (
	"testing"
)

func Test_DeleteSchema(t *testing.T) {
	repo := NewMockRepository()
	repo.SetGetSchemaVersionsByIdResponse("mocking", MockSchema("mocking"), nil)
	deleted, err := (*Service).DeleteSchema(New(repo, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking")
	if err != nil {
		t.Errorf("returned error")
	}
	if !deleted {
		t.Errorf("deleted returned false")
	}
}

func Test_DeleteSchemaVersion(t *testing.T) {
	deleted, err := (*Service).DeleteSchemaVersion(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking", "mocking")
	if err != nil {
		t.Errorf("returned error")
	}
	if !deleted {
		t.Errorf("deleted returned false")
	}
}

func Test_GetAllSchemas(t *testing.T) {
	schemas, _ := (*Service).GetAllSchemas(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"))
	if schemas[0].SchemaID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_GetSchemas(t *testing.T) {
	schemas, _ := (*Service).GetAllSchemas(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"))
	if schemas[0].SchemaID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_GetLatestSchemaVersion(t *testing.T) {
	VersionDetails, _ := (*Service).GetLatestSchemaVersion(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking")
	if VersionDetails.VersionID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_CreateSchema(t *testing.T) {
	sdto := SchemaRegistrationRequest{
		Description:       "mocking",
		Specification:     "mocking",
		Name:              "mocking",
		SchemaType:        "mocking",
		PublisherID:       "mocking",
		ValidityMode:      "none",
		CompatibilityMode: "none",
	}
	VersionDetails, added, err := (*Service).CreateSchema(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"), sdto)
	if err != nil {
		t.Errorf("returned error")
	}

	if !added {
		t.Errorf("could not add schema")
	}

	if VersionDetails.SchemaID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_GetSchemaVersion(t *testing.T) {
	VersionDetails, _ := (*Service).GetSchemaVersion(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking", "mocking")
	if VersionDetails.SchemaID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_UpdateSchema(t *testing.T) {
	sdto := SchemaUpdateRequest{
		Description:   "mocking",
		Specification: "mocking",
	}
	VersionDetails, added, err := (*Service).UpdateSchema(New(&mockRepository{}, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking", sdto)
	if err != nil {
		t.Errorf("returned error")
	}

	if !added {
		t.Errorf("could not add schema")
	}

	if VersionDetails.SchemaID != "mocking" {
		t.Errorf("wrong schemaId returned")
	}
}

func Test_GetSchemaVersionsById(t *testing.T) {
	repo := NewMockRepository()
	repo.SetGetSchemaVersionsByIdResponse("mocking", MockSchema("mocking"), nil)
	schema, _ := (*Service).ListSchemaVersions(New(repo, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking")
	if schema.SchemaID != "mocking" {
		t.Errorf("wrong schema ID returned")
	}
}

func Test_GetAllSchemaVersions(t *testing.T) {
	repo := NewMockRepository()
	repo.SetGetSchemaVersionsByIdResponse("mocking", MockSchema("mocking"), nil)
	schema, _ := (*Service).ListAllSchemaVersions(New(repo, &mockCompChecker{}, &mockValChecker{}, "none", "none"), "mocking")
	if schema.SchemaID != "mocking" {
		t.Errorf("wrong schema ID returned")
	}
}
