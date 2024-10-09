package csv

import (
	"bytes"
	"context"
	"encoding/csv"
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

const DefaultTimeoutBase = 2 * time.Second

// New returns a new validator which validates CSV messages against a schema.
//
// Performs a health check to see if the validator is available, retrying periodically until the context is cancelled
// or the health check succeeds.
func New(ctx context.Context, url string, timeoutBase time.Duration) (validator.Validator, error) {
	if err := retry.Do(ctx, retry.WithJitter(retry.Constant(2*time.Second)), func(ctx context.Context) error {
		return httputil.HealthCheck(ctx, url+"/health")
	}); err != nil {
		return nil, errors.Wrapf(err, "attempting to reach csv validator at %s failed", url)
	}

	return &Validator{
		Url:         url,
		TimeoutBase: timeoutBase,
	}, nil
}

func (v *Validator) Validate(message, schema []byte, _, _ string) (bool, error) {
	if !IsCSV(message) {
		return false, validator.ErrDeadletter
	}

	ctx, cancel := context.WithTimeout(context.Background(), validator.EstimateHTTPTimeout(len(message), v.TimeoutBase))
	defer cancel()

	return validator.ValidateOverHTTP(ctx, message, schema, v.Url)
}

// IsCSV checks if the given data is valid csv.
//
// The data is assumed to use ',' as delimiter.
func IsCSV(data []byte) bool {
	reader := csv.NewReader(bytes.NewReader(data))

	reader.ReuseRecord = true
	reader.Comma = ','
	reader.LazyQuotes = true

	for {
		_, err := reader.Read()
		if err != nil {
			return err == io.EOF
		}
	}
}
