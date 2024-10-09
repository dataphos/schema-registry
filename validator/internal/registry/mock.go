package registry

import "context"

type Mock struct {
	getSchemaResponse       map[string]mockGetSchemaResponse
	getLatestSchemaResponse map[string]mockGetLatestSchemaResponse
	registrationResponse    map[string]mockRegisterResponse
	updateResponse          map[string]mockUpdateResponse
}

type mockGetSchemaResponse struct {
	schema []byte
	err    error
}

type mockGetLatestSchemaResponse struct {
	schema []byte
	err    error
}

type mockRegisterResponse struct {
	id      string
	version string
	err     error
}

type mockUpdateResponse struct {
	version string
	err     error
}

func NewMock() *Mock {
	return &Mock{
		getSchemaResponse:       map[string]mockGetSchemaResponse{},
		getLatestSchemaResponse: map[string]mockGetLatestSchemaResponse{},
		registrationResponse:    map[string]mockRegisterResponse{},
		updateResponse:          map[string]mockUpdateResponse{},
	}
}

func (m *Mock) SetGetResponse(id, version string, schema []byte, err error) {
	key := id + "_" + version
	m.getSchemaResponse[key] = mockGetSchemaResponse{
		schema: schema,
		err:    err,
	}
}

func (m *Mock) Get(_ context.Context, id, version string) ([]byte, error) {
	key := id + "_" + version
	response := m.getSchemaResponse[key]
	return response.schema, response.err
}

func (m *Mock) SetGetLatestResponse(id string, schema []byte, err error) {
	key := id
	m.getLatestSchemaResponse[key] = mockGetLatestSchemaResponse{
		schema: schema,
		err:    err,
	}
}

func (m *Mock) GetLatest(_ context.Context, id string) ([]byte, error) {
	key := id
	response := m.getLatestSchemaResponse[key]
	return response.schema, response.err
}

func (m *Mock) SetRegisterResponse(schema []byte, id, version string, err error) {
	m.registrationResponse[string(schema)] = mockRegisterResponse{
		id:      id,
		version: version,
		err:     err,
	}
}

func (m *Mock) Register(_ context.Context, schema []byte, _, _, _ string) (string, string, error) {
	response := m.registrationResponse[string(schema)]
	return response.id, response.version, response.err
}

func (m *Mock) SetUpdateResponse(id, version string, err error) {
	m.updateResponse[id] = mockUpdateResponse{
		version: version,
		err:     err,
	}
}

func (m *Mock) Update(_ context.Context, id string, _ []byte) (string, error) {
	response := m.updateResponse[id]
	return response.version, response.err
}
