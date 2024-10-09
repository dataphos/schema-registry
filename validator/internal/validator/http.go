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

package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/dataphos/lib-httputil/pkg/httputil"

	"github.com/pkg/errors"
)

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

// ValidateOverHTTP requests a message validation over HTTP.
// Function returns the validation boolean result.
func ValidateOverHTTP(ctx context.Context, message, schema []byte, url string) (bool, error) {
	response, err := sendValidationRequest(ctx, message, schema, url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	var parsedBody validationResponse
	if err = json.Unmarshal(body, &parsedBody); err != nil {
		return false, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return parsedBody.Validation, nil
	case http.StatusBadRequest:
		return false, ErrDeadletter
	default:
		return false, errors.Errorf("error: status code [%v]", response.StatusCode)
	}
}

func sendValidationRequest(ctx context.Context, message, schema []byte, url string) (*http.Response, error) {
	// this can't generate an error, so it's safe to ignore
	data, _ := json.Marshal(validationRequest{Data: string(message), Schema: string(schema)})

	request, err := httputil.Post(ctx, url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(request)
}

// validationRequest contains the message and schema which are used by the validator. The structure represents an HTTP
// request body.
type validationRequest struct {
	Data   string `json:"data"`
	Schema string `json:"schema"`
}

// validationResponse contains the validation result and an info message. The structure represents an HTTP response body.
type validationResponse struct {
	Validation bool   `json:"validation"`
	Info       string `json:"info"`
}
