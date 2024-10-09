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

//nolint:staticcheck,unused

package registry

import (
	"time"
)

type mockRepository struct {
	createSchemaResponse            map[string]mockCreateSchema
	getSchemaByIdAndVersionResponse map[string]mockGetSchemaVersionByIdAndVersion
	updateSchemaByIdResponse        map[string]mockUpdateSchemaById
	getSchemaVersionsResponse       map[string]mockGetSchemaVersionsById
	getAllSchemaVersionsResponse    map[string]mockGetAllSchemaVersions
	getLatestSchemaVersionResponse  map[string]mockGetLatestSchemaVersion
	deleteSchemaResponse            map[string]mockDeleteSchema
	deleteVersionResponse           map[string]mockDeleteSchemaVersion
	getSchemasResponse              map[string]mockGetSchemas
	getAllSchemasResponse           map[string]mockGetAllSchemas
}

type mockCompChecker struct {
	// checkCompResponse map[string]mockCheckComp
}

type mockValChecker struct {
	// checkValResponse map[string]mockValComp
}

//type mockCheckComp struct {
//	ok  bool
//	err error
//}

//type mockValComp struct {
//	ok  bool
//	err error
//}

type mockCreateSchema struct {
	VersionDetails VersionDetails
}

type mockGetSchemaVersionByIdAndVersion struct {
	VersionDetails VersionDetails
}

type mockUpdateSchemaById struct {
	VersionDetails VersionDetails
}

type mockGetSchemaVersionsById struct {
	schema Schema
	err    error
}

type mockGetAllSchemaVersions struct {
}

type mockGetLatestSchemaVersion struct {
}

type mockDeleteSchema struct {
}

type mockDeleteSchemaVersion struct {
}

type mockGetSchemas struct {
}

type mockGetAllSchemas struct {
}

func MockSchema(id string) Schema {
	return Schema{
		SchemaID:          id,
		SchemaType:        "mocking",
		Name:              "mocking",
		VersionDetails:    nil,
		Description:       "mocking",
		LastCreated:       "mocking",
		PublisherID:       "mocking",
		CompatibilityMode: "none",
		ValidityMode:      "none",
	}
}

func MockVersionDetails(id, version string) VersionDetails {
	return VersionDetails{
		VersionID:          id,
		Version:            version,
		SchemaID:           "mocking",
		Specification:      "mocking",
		Description:        "mocking",
		SchemaHash:         "mocking",
		CreatedAt:          time.Time{},
		VersionDeactivated: false,
	}
}

func NewMockRepository() *mockRepository {
	return &mockRepository{
		createSchemaResponse:            map[string]mockCreateSchema{},
		getSchemaByIdAndVersionResponse: map[string]mockGetSchemaVersionByIdAndVersion{},
		updateSchemaByIdResponse:        map[string]mockUpdateSchemaById{},
		getSchemaVersionsResponse:       map[string]mockGetSchemaVersionsById{},
		getAllSchemaVersionsResponse:    map[string]mockGetAllSchemaVersions{},
		getLatestSchemaVersionResponse:  map[string]mockGetLatestSchemaVersion{},
		deleteSchemaResponse:            map[string]mockDeleteSchema{},
		deleteVersionResponse:           map[string]mockDeleteSchemaVersion{},
		getSchemasResponse:              map[string]mockGetSchemas{},
		getAllSchemasResponse:           map[string]mockGetAllSchemas{},
	}
}

func (c *mockCompChecker) Check(_ string, _ []string, _ string) (bool, error) {
	return true, nil
}

func (c *mockValChecker) Check(_, _, _ string) (bool, error) {
	return true, nil
}

func (m *mockRepository) CheckCompatibility(_, _ string) (bool, error) {
	return true, nil
}

func (m *mockRepository) DeleteSchema(_ string) (bool, error) {
	return true, nil
}

func (m *mockRepository) DeleteSchemaVersion(_, _ string) (bool, error) {
	return true, nil
}

func (m *mockRepository) GetSchemas() ([]Schema, error) {
	return []Schema{{
		SchemaID:          "mocking",
		SchemaType:        "mocking",
		Name:              "mocking",
		VersionDetails:    nil,
		Description:       "mocking",
		LastCreated:       "mocking",
		PublisherID:       "mocking",
		CompatibilityMode: "none",
		ValidityMode:      "none",
	}}, nil
}

func (m *mockRepository) GetAllSchemas() ([]Schema, error) {
	return []Schema{{
		SchemaID:          "mocking",
		SchemaType:        "mocking",
		Name:              "mocking",
		VersionDetails:    nil,
		Description:       "mocking",
		LastCreated:       "mocking",
		PublisherID:       "mocking",
		CompatibilityMode: "none",
		ValidityMode:      "none",
	}}, nil
}

func (m *mockRepository) GetLatestSchemaVersion(_ string) (VersionDetails, error) {
	return MockVersionDetails("mocking", "mocking"), nil
}

func (m *mockRepository) CreateSchema(_ SchemaRegistrationRequest) (VersionDetails, bool, error) {
	return MockVersionDetails("mocking", "mocking"), true, nil
}

func (m *mockRepository) GetSchemaVersionByIdAndVersion(id string, version string) (VersionDetails, error) {
	return MockVersionDetails(id, version), nil
}

func (m *mockRepository) UpdateSchemaById(id string, _ SchemaUpdateRequest) (VersionDetails, bool, error) {
	return MockVersionDetails(id, "mocking"), true, nil
}

func (m *mockRepository) SetGetSchemaVersionsByIdResponse(id string, schema Schema, err error) {
	m.getSchemaVersionsResponse[id] = mockGetSchemaVersionsById{
		schema: schema,
		err:    err,
	}
}

func (m *mockRepository) GetSchemaVersionsById(id string) (Schema, error) {
	response := m.getSchemaVersionsResponse[id]
	return response.schema, response.err
}

func (m *mockRepository) GetAllSchemaVersions(id string) (Schema, error) {
	response := m.getSchemaVersionsResponse[id]
	return response.schema, response.err
}
