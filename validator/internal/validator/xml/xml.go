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

package xml

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"
	"github.com/dataphos/lib-httputil/pkg/httputil"
	"github.com/dataphos/lib-retry/pkg/retry"

	"github.com/pkg/errors"
)

type Validator struct {
	Url         string
	TimeoutBase time.Duration
}

const DefaultTimeoutBase = 3 * time.Second

// New returns a new validator which validates XML messages against a schema.
//
// Performs a health check to see if the validator is available, retrying periodically until the context is cancelled
// or the health check succeeds.
func New(ctx context.Context, url string, timeoutBase time.Duration) (validator.Validator, error) {
	if err := retry.Do(ctx, retry.WithJitter(retry.Constant(2*time.Second)), func(ctx context.Context) error {
		return httputil.HealthCheck(ctx, url+"/health")
	}); err != nil {
		return nil, errors.Wrapf(err, "attempting to reach xml validator at %s failed", url)
	}

	return &Validator{
		Url:         url,
		TimeoutBase: timeoutBase,
	}, nil
}

func (v *Validator) Validate(message, schema []byte, _, _ string) (bool, error) {
	if !IsXML(message) || !IsXML(schema) {
		return false, validator.ErrDeadletter
	}

	ctx, cancel := context.WithTimeout(context.Background(), validator.EstimateHTTPTimeout(len(message), v.TimeoutBase))
	defer cancel()

	return validator.ValidateOverHTTP(ctx, message, schema, v.Url)
}

// IsXML checks if given data is valid XML.
func IsXML(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := decoder.Token()
		if err != nil {
			return err == io.EOF
		}
	}
}
