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

package apicuriosr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dataphos/lib-retry/pkg/retry"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dataphos/schema-registry-validator/internal/errtemplates"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/lib-httputil/pkg/httputil"

	"github.com/pkg/errors"
)

// SchemaRegistry is a proxy for communicating with the janitor schema registry server.
type SchemaRegistry struct {
	Url      string
	Timeouts TimeoutSettings
	GroupID  string
}

// TimeoutSettings defines the maximum amount of time for each get, register or update request.
type TimeoutSettings struct {
	GetTimeout      time.Duration
	RegisterTimeout time.Duration
	UpdateTimeout   time.Duration
}

var DefaultTimeoutSettings = TimeoutSettings{
	GetTimeout:      4 * time.Second,
	RegisterTimeout: 10 * time.Second,
	UpdateTimeout:   10 * time.Second,
}

// New returns an instance of SchemaRegistry.
//
// Performs a health check to see if the schema registry is available, retrying periodically until the context is cancelled
// or the health check succeeds.
func New(ctx context.Context, url string, timeouts TimeoutSettings, groupID string) (*SchemaRegistry, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := retry.Do(ctx, retry.WithJitter(retry.Constant(2*time.Second)), func(ctx context.Context) error {
		return httputil.HealthCheck(ctx, url+"/health")
	}); err != nil {
		return nil, errors.Wrapf(err, "attempting to reach schema registry at %s failed", url)
	}

	return &SchemaRegistry{
		Url:      url,
		Timeouts: timeouts,
		GroupID:  groupID,
	}, nil
}

func (sr *SchemaRegistry) Get(ctx context.Context, id, version string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, sr.Timeouts.GetTimeout)
	defer cancel()

	response, err := sr.sendGetRequest(ctx, id, version)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.ReadingResponseBodyFailed)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, errors.Wrapf(registry.ErrNotFound, "fetching schema %s/%s failed", id, version)
		}
		return nil, errors.Wrapf(errtemplates.BadHttpStatusCode(response.StatusCode), "fetching schema %s/%s resulted in a bad status code", id, version)
	}

	return body, nil
}

func (sr *SchemaRegistry) sendGetRequest(ctx context.Context, id string, version string) (*http.Response, error) {
	url := fmt.Sprintf("%s/apis/registry/v2/groups/%s/artifacts/%s/versions/%s", sr.Url, sr.GroupID, id, version)

	request, err := httputil.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodGet, url))
	}

	return response, nil
}

func (sr *SchemaRegistry) GetLatest(ctx context.Context, id string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, sr.Timeouts.GetTimeout)
	defer cancel()

	response, err := sr.sendGetLatestRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.ReadingResponseBodyFailed)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, errors.Wrapf(registry.ErrNotFound, "fetching schema %s/latest failed", id)
		}
		return nil, errors.Wrapf(errtemplates.BadHttpStatusCode(response.StatusCode), "fetching schema %s/latest resulted in a bad status code", id)
	}

	return body, nil
}

func (sr *SchemaRegistry) sendGetLatestRequest(ctx context.Context, id string) (*http.Response, error) {
	url := fmt.Sprintf("%s/apis/registry/v2/groups/%s/artifacts/%s/versions/latest", sr.Url, sr.GroupID, id)

	request, err := httputil.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodGet, url))
	}

	return response, nil
}

func (sr *SchemaRegistry) Register(ctx context.Context, schema []byte, schemaType, compMode, valMode string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, sr.Timeouts.RegisterTimeout)
	defer cancel()

	response, err := sr.sendRegisterRequest(ctx, schema, schemaType)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	// this needs to be before checking the status code because the response body always needs to be read
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", errors.Wrap(err, errtemplates.ReadingResponseBodyFailed)
	}

	// the schema registry returns either 200, if the new schema version is successfully inserted, or 409 if
	// the given schema already exists
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusConflict {
		return "", "", errtemplates.BadHttpStatusCode(response.StatusCode)
	}

	var info insertInfo
	if err = json.Unmarshal(body, &info); err != nil {
		return "", "", errors.Wrap(err, errtemplates.UnmarshallingJSONFailed)
	}

	compResponse, valResponse, err := sr.sendRulesRequest(ctx, info.Id, compMode, valMode)
	if err != nil {
		return "", "", err
	}
	defer compResponse.Body.Close()
	defer valResponse.Body.Close()

	// the schema registry returns 204 if the new rule is successfully added
	if compResponse.StatusCode != http.StatusNoContent {
		return "", "", errtemplates.BadHttpStatusCode(compResponse.StatusCode)
	}
	if valResponse.StatusCode != http.StatusNoContent {
		return "", "", errtemplates.BadHttpStatusCode(valResponse.StatusCode)
	}

	return info.Id, info.Version, nil
}

func (sr *SchemaRegistry) sendRegisterRequest(ctx context.Context, schema []byte, schemaType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/apis/registry/v2/groups/%s/artifacts", sr.Url, sr.GroupID)

	request, err := httputil.Post(ctx, url, fmt.Sprintf("application/json; artifactType=%s", strings.ToUpper(schemaType)), bytes.NewBuffer(schema))
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodPost, url))
	}

	return response, nil
}

type ruleRequest struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

func (sr *SchemaRegistry) sendRulesRequest(ctx context.Context, id, compMode, valMode string) (*http.Response, *http.Response, error) {
	url := fmt.Sprintf("%s/apis/registry/v2/groups/%s/artifacts/%s/rules", sr.Url, sr.GroupID, id)

	// this can't generate an error, so it's safe to ignore
	compRule, _ := json.Marshal(ruleRequest{Type: "COMPATIBILITY", Config: strings.ToUpper(compMode)})
	request, err := httputil.Post(ctx, url, "application/json", bytes.NewBuffer(compRule))
	if err != nil {
		return nil, nil, err
	}

	compResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodPost, url))
	}

	// this can't generate an error, so it's safe to ignore
	valRule, _ := json.Marshal(ruleRequest{Type: "VALIDITY", Config: strings.ToUpper(valMode)})
	request, err = httputil.Post(ctx, url, "application/json", bytes.NewBuffer(valRule))
	if err != nil {
		return nil, nil, err
	}

	valResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodPost, url))
	}

	return compResponse, valResponse, nil
}

func (sr *SchemaRegistry) Update(ctx context.Context, id string, schema []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, sr.Timeouts.UpdateTimeout)
	defer cancel()

	response, err := sr.sendUpdateRequest(ctx, id, schema)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// this needs to be before checking the status code because the response body always needs to be read
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, errtemplates.ReadingResponseBodyFailed)
	}

	// the schema registry returns 200, if the new schema version is successfully inserted
	if response.StatusCode != http.StatusOK {
		return "", errtemplates.BadHttpStatusCode(response.StatusCode)
	}

	var info insertInfo
	if err = json.Unmarshal(body, &info); err != nil {
		return "", errors.Wrap(err, errtemplates.UnmarshallingJSONFailed)
	}

	return info.Version, nil
}

func (sr *SchemaRegistry) sendUpdateRequest(ctx context.Context, id string, schema []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/apis/registry/v2/groups/%s/artifacts/%s", sr.Url, sr.GroupID, id)

	request, err := httputil.Put(ctx, url, "application/json", bytes.NewBuffer(schema))
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.HttpRequestToUrlFailed(http.MethodPut, url))
	}

	return response, nil
}
