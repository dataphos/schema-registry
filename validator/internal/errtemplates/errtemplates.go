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

// Package errtemplates offers convenience functions to standardize error messages and simplify proper error wrapping.
package errtemplates

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	envVariableNotDefinedTemplate     = "env variable %s not defined"
	parsingEnvVariableFailedTemplate  = "parsing env variable %s failed"
	unsupportedBrokerTypeTemplate     = "unsupported broker type %s"
	unsupportedRegistryTypeTemplate   = "unsupported registry type %s"
	failedTopicInitializationTemplate = "creating reference to %s failed"
	attributeNotAStringTemplate       = "%s attribute is not a string"
	missingAttributeTemplate          = "%s attribute is not defined"
	mustNotBeEmptyTemplate            = "%s must not be empty"
	unsupportedMessageFormatTemplate  = "unsupported message format: %s"
	badHttpStatusCodeTemplate         = "bad status code: %d"
	httpRequestToUrlFailedTemplate    = "%s request to %s failed"
)

const (
	// ReadingResponseBodyFailed is an error message stating that reading the response body failed.
	ReadingResponseBodyFailed = "reading response body failed"

	// UnmarshallingJSONFailed is an error message stating that unmarshalling json failed.
	UnmarshallingJSONFailed = "unmarshalling json failed"

	// LoadingTopicsFailed is an error message stating that the target topics couldn't be loaded.
	LoadingTopicsFailed = "loading topics failed"
)

// EnvVariableNotDefined returns an error stating that the given env variable is not defined.
func EnvVariableNotDefined(name string) error {
	return errors.Errorf(envVariableNotDefinedTemplate, name)
}

// ParsingEnvVariableFailed returns a string stating that the given env variable couldn't be parsed properly.
func ParsingEnvVariableFailed(name string) string {
	return fmt.Sprintf(parsingEnvVariableFailedTemplate, name)
}

// UnsupportedBrokerType returns an error stating that the given broker type is not supported.
func UnsupportedBrokerType(name string) error {
	return errors.Errorf(unsupportedBrokerTypeTemplate, name)
}

// UnsupportedRegistryType returns an error stating that the given registry type is not supported.
func UnsupportedRegistryType(name string) error {
	return errors.Errorf(unsupportedRegistryTypeTemplate, name)
}

// CreatingTopicInstanceFailed returns an error stating that topic creation failed.
func CreatingTopicInstanceFailed(name string) string {
	return fmt.Sprintf(failedTopicInitializationTemplate, name)
}

// AttributeNotAString returns an error stating that the given attribute is not a string.
func AttributeNotAString(name string) error {
	return errors.Errorf(attributeNotAStringTemplate, name)
}

// AttributeNotDefined returns an error stating that the given attribute is not defined.
func AttributeNotDefined(name string) error {
	return errors.Errorf(missingAttributeTemplate, name)
}

// MustNotBeEmpty returns an error stating that the given variable must not be empty.
func MustNotBeEmpty(name string) error {
	return errors.Errorf(mustNotBeEmptyTemplate, name)
}

// UnsupportedMessageFormat returns an error stating that the message's format is not supported for validation.
func UnsupportedMessageFormat(format string) error {
	return errors.Errorf(unsupportedMessageFormatTemplate, format)
}

// BadHttpStatusCode returns an error stating that the given status code wasn't expected.
func BadHttpStatusCode(code int) error {
	return errors.Errorf(badHttpStatusCodeTemplate, code)
}

// HttpRequestToUrlFailed returns a string stating that a http method to the given url has failed.
func HttpRequestToUrlFailed(methodName, url string) string {
	return fmt.Sprintf(httpRequestToUrlFailedTemplate, methodName, url)
}
