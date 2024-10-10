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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dataphos/lib-httputil/pkg/httputil"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// checkRequest contains a new schema, its type and a validity mode which should be enforced. The structure represents an HTTP request body.
type checkRequest struct {
	Schema string `json:"schema"`
	Format string `json:"format"`
	Mode   string `json:"mode"`
}

// checkResponse contains the validity result and an info message. The structure represents an HTTP response body.
type checkResponse struct {
	Result bool   `json:"result"`
	Info   string `json:"info"`
}

// HTTPTimeoutBytesUnit the base amount of bytes used by EstimateHTTPTimeout.
const HTTPTimeoutBytesUnit = 1024 * 100

// EstimateHTTPTimeout calculates the expected timeout, by dividing the size given in bytes with HTTPTimeoutBytesUnit, and then
// multiplying the coefficient with the given time duration.
//
// If the given size is less than HTTPTimeoutBytesUnit, base is returned, to avoid problems due to the http overhead which isn't fully linear.
func EstimateHTTPTimeout(size int, base time.Duration) time.Duration {
	coef := int(math.Round(float64(size) / float64(HTTPTimeoutBytesUnit)))
	if coef <= 1 {
		return base
	}

	return time.Duration(coef) * base
}

// CheckOverHTTP requests a schema check over HTTP.
// Function returns false if schema isn't valid.
func CheckOverHTTP(ctx context.Context, schemaType, schema, mode, url string) (bool, error) {
	response, err := sendCheckRequest(ctx, schemaType, schema, mode, url)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(errors.New("couldn't close response body"))
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	var parsedBody checkResponse
	if err = json.Unmarshal(body, &parsedBody); err != nil {
		return false, err
	}

	valid := parsedBody.Result

	switch response.StatusCode {
	case http.StatusOK:
		return valid, nil
	case http.StatusBadRequest:
		return valid, nil
	default:
		return valid, errors.Errorf("error: status code [%v]", response.StatusCode)
	}
}

func sendCheckRequest(ctx context.Context, schemaType, schema, mode, url string) (*http.Response, error) {
	// this can't generate an error, so it's safe to ignore
	data, _ := json.Marshal(checkRequest{Schema: schema, Format: schemaType, Mode: mode})

	request, err := httputil.Post(ctx, url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(request)
}
