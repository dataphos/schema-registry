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
	//ok             bool
	//err            error
}

type mockGetSchemaVersionByIdAndVersion struct {
	VersionDetails VersionDetails
	//err            error
}

type mockUpdateSchemaById struct {
	VersionDetails VersionDetails
	//ok             bool
	//err            error
}

type mockGetSchemaVersionsById struct {
	schema Schema
	err    error
}

type mockGetAllSchemaVersions struct {
	//schema Schema
	err error
}

type mockGetLatestSchemaVersion struct {
	//schema Schema
	err error
}

type mockDeleteSchema struct {
	//ok  bool
	err error
}

type mockDeleteSchemaVersion struct {
	ok  bool
	err error
}

type mockGetSchemas struct {
	//schemas []Schema
	err error
}

type mockGetAllSchemas struct {
	//schemas []Schema
	err error
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
