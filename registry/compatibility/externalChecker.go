package compatibility

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dataphos/aquarium-janitor-standalone-sr/compatibility/http"
	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errtemplates"
	"github.com/dataphos/lib-httputil/pkg/httputil"
	"github.com/dataphos/lib-retry/pkg/retry"
)

const (
	urlEnvKey               = "COMPATIBILITY_CHECKER_URL"
	timeoutEnvKey           = "COMPATIBILITY_CHECKER_TIMEOUT_BASE"
	globalCompatibilityMode = "GLOBAL_COMPATIBILITY_MODE"
)

const (
	DefaultTimeoutBase             = 2 * time.Second
	defaultGlobalCompatibilityMode = "BACKWARD"
)

type ExternalChecker struct {
	url         string
	TimeoutBase time.Duration
}

// NewFromEnv loads the needed environment variables and calls New.
func NewFromEnv(ctx context.Context) (*ExternalChecker, error) {
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

	return New(ctx, url, timeout)
}

// New returns a new instance of Repository.
func New(ctx context.Context, url string, timeoutBase time.Duration) (*ExternalChecker, error) {
	if err := retry.Do(ctx, retry.WithJitter(retry.Constant(2*time.Second)), func(ctx context.Context) error {
		return httputil.HealthCheck(ctx, url+"/health")
	}); err != nil {
		return nil, errors.Wrapf(err, "attempting to reach compatibility checker at %s failed", url)
	}

	return &ExternalChecker{
		url:         url,
		TimeoutBase: timeoutBase,
	}, nil
}

func (c *ExternalChecker) Check(schemaInfo string, history []string, mode string) (bool, error) {
	//check if compatibility mode is none, if it is, don't send HTTP request to java code
	if strings.ToLower(mode) == "none" {
		return true, nil
	}
	size := calculateSizeInBytes(schemaInfo, history, mode)
	ctx, cancel := context.WithTimeout(context.Background(), http.EstimateHTTPTimeout(size, c.TimeoutBase))
	defer cancel()

	decodedHistory, err := c.DecodeHistory(history)
	if err != nil {
		return false, err
	}
	return http.CheckOverHTTP(ctx, schemaInfo, decodedHistory, mode, c.url+"/")
}

func (c *ExternalChecker) DecodeHistory(history []string) ([]string, error) {
	var decodedHistory []string
	for i := 0; i < len(history); i++ {
		decoded, err := base64.StdEncoding.DecodeString(history[i])
		if err != nil {
			fmt.Println(fmt.Errorf("could not decode").Error())
			return nil, err
		}
		decodedHistory = append(decodedHistory, string(decoded))
	}
	return decodedHistory, nil
}

func calculateSizeInBytes(schema string, history []string, mode string) int {
	bytes := []byte(schema + mode)
	for i := 0; i < len(history); i++ {
		bytes = append(bytes, []byte(history[i])...)
	}
	return len(bytes)
}

func InitCompatibilityChecker(ctx context.Context) (*ExternalChecker, string, error) {
	compChecker, err := NewFromEnv(ctx)
	if err != nil {
		return nil, "", err
	}
	globalCompMode := os.Getenv(globalCompatibilityMode)
	if globalCompMode == "" {
		globalCompMode = defaultGlobalCompatibilityMode
	}
	if globalCompMode == "BACKWARD" || globalCompMode == "BACKWARD_TRANSITIVE" ||
		globalCompMode == "FORWARD" || globalCompMode == "FORWARD_TRANSITIVE" ||
		globalCompMode == "FULL" || globalCompMode == "FULL_TRANSITIVE" || globalCompMode == "NONE" {
		return compChecker, globalCompMode, nil
	}
	return nil, "", errors.Errorf("unsupported compatibility mode")
}

func CheckIfValidMode(mode *string) bool {
	if *mode == "" {
		*mode = defaultGlobalCompatibilityMode
	}
	lowerMode := strings.ToLower(*mode)
	if lowerMode != "none" && lowerMode != "backward" && lowerMode != "backward_transitive" && lowerMode != "forward" && lowerMode != "forward_transitive" && lowerMode != "full" && lowerMode != "full_transitive" {
		return false
	}
	return true
}
