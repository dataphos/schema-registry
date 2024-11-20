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

package validity

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dataphos/lib-httputil/pkg/httputil"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
	"github.com/dataphos/lib-retry/pkg/retry"
	"github.com/dataphos/schema-registry/internal/config"
	"github.com/dataphos/schema-registry/internal/errtemplates"
	"github.com/dataphos/schema-registry/validity/http"
)

const (
	urlEnvKey          = "VALIDITY_CHECKER_URL"
	timeoutEnvKey      = "VALIDITY_CHECKER_TIMEOUT_BASE"
	globalValidityMode = "GLOBAL_VALIDITY_MODE"
)

const (
	DefaultTimeoutBase        = 2 * time.Second
	defaultGlobalValidityMode = "FULL"
)

type ExternalChecker struct {
	Url         string
	TimeoutBase time.Duration
	Log         logger.Log
}

// NewExternalCheckerFromEnv loads the needed environment variables and calls NewExternalChecker.
func NewExternalCheckerFromEnv(ctx context.Context) (*ExternalChecker, error) {
	url := os.Getenv(urlEnvKey)
	if url == "" {
		return nil, errtemplates.EnvVariableNotDefined(urlEnvKey)
	}
	timeout := DefaultTimeoutBase
	if timeoutStr := os.Getenv(timeoutEnvKey); timeoutStr != "" {
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.ParsingEnvVariableFailed(timeoutEnvKey))
		}
	}

	return NewExternalChecker(ctx, url, timeout)
}

// NewExternalChecker returns a new instance of ExternalChecker.
func NewExternalChecker(ctx context.Context, url string, timeoutBase time.Duration) (*ExternalChecker, error) {
	if err := retry.Do(ctx, retry.WithJitter(retry.Constant(2*time.Second)), func(ctx context.Context) error {
		return httputil.HealthCheck(ctx, url+"/health")
	}); err != nil {
		return nil, errors.Wrapf(err, "attempting to reach validity checker at %s failed", url)
	}

	labels := logger.Labels{
		"product":   "Schema Registry",
		"component": "validity_checker",
	}
	logLevel, logConfigWarnings := config.GetLogLevel()
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logLevel))
	for _, w := range logConfigWarnings {
		log.Warn(w)
	}

	return &ExternalChecker{
		Url:         url,
		TimeoutBase: timeoutBase,
		Log:         log,
	}, nil
}

func (c *ExternalChecker) Check(schema, schemaType, mode string) (bool, error) {
	//check if validity mode is none, if it is, don't send HTTP request to java code
	if strings.ToLower(mode) == "none" {
		return true, nil
	}
	if strings.ToLower(mode) == "syntax-only" || strings.ToLower(mode) == "full" {
		size := []byte(schema + schemaType + mode)
		ctx, cancel := context.WithTimeout(context.Background(), http.EstimateHTTPTimeout(len(size), c.TimeoutBase))
		defer cancel()

		valid, info, err := http.CheckOverHTTP(ctx, schemaType, schema, mode, c.Url+"/")
		c.Log.Info(info)
		return valid, err
	}

	return false, errors.Errorf("")
}

func InitExternalValidityChecker(ctx context.Context) (*ExternalChecker, string, error) {
	valChecker, err := NewExternalCheckerFromEnv(ctx)
	if err != nil {
		return nil, "", err
	}
	globalValMode := os.Getenv(globalValidityMode)
	if globalValMode == "" {
		globalValMode = defaultGlobalValidityMode
	}
	if globalValMode == "SYNTAX-ONLY" || globalValMode == "FULL" || globalValMode == "NONE" {
		return valChecker, globalValMode, nil
	}
	return nil, "", errors.Errorf("unsupported validity mode")
}

func CheckIfValidMode(mode *string) bool {
	if *mode == "" {
		*mode = defaultGlobalValidityMode
	}
	lowerMode := strings.ToLower(*mode)
	if lowerMode != "none" && lowerMode != "syntax-only" && lowerMode != "full" {
		return false
	}
	return true
}
