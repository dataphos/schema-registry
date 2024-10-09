package errtemplates

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	envVariableNotDefinedTemplate    = "env variable %s not defined"
	expectedEnvVariableAsInt         = "expected env variable %s as int, received %s instead"
	parsingEnvVariableFailedTemplate = "parsing env variable %s failed"
)

// EnvVariableNotDefined returns an error stating that the given env variable is not defined.
func EnvVariableNotDefined(name string) error {
	return errors.Errorf(envVariableNotDefinedTemplate, name)
}

// ExpectedInt returns an error stating that the given env variable was expected to be an int.
func ExpectedInt(name string, value string) error {
	return errors.Errorf(expectedEnvVariableAsInt, name, value)
}

// ParsingEnvVariableFailed returns a string stating that the given env variable couldn't be parsed properly.
func ParsingEnvVariableFailed(name string) string {
	return fmt.Sprintf(parsingEnvVariableFailedTemplate, name)
}
