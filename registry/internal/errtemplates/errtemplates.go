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
