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

package json

import (
	"bytes"
	"encoding/json"
	_errors "errors"
	"github.com/pkg/errors"

	"github.com/dataphos/schema-registry-validator/internal/validator"

	lru "github.com/hashicorp/golang-lru"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"github.com/xeipuuv/gojsonschema"
)

func New() validator.Validator {
	return validator.Func(func(message, schema []byte, _, _ string) (bool, error) {
		var parsedMessage interface{}
		if err := json.Unmarshal(message, &parsedMessage); err != nil {
			errBroken := errors.WithMessage(validator.ErrBrokenMessage, "Message is not in a valid format - "+err.Error())
			return false, errBroken
		}

		compiledSchema, err := compileSchema(schema)
		if err != nil {
			errCompile := errors.WithMessage(validator.ErrWrongCompile, err.Error())
			return false, errCompile
		}

		if err = compiledSchema.Validate(parsedMessage); err != nil {
			var validationError *jsonschema.ValidationError
			ok := _errors.As(err, &validationError)
			if !ok {
				return false, errors.New("impossible to cast error to *jsonschema.ValidationError")
			}
			errMessage, err := createErrorMessage(validationError)
			if err != nil {
				return false, errors.New("couldn't create error message")
			}
			return false, errors.WithMessage(validator.ErrFailedValidation, errMessage)
		}
		return true, nil
	})
}

func NewCached(size int) validator.Validator {
	cache, _ := lru.NewARC(size)

	return validator.Func(func(message, schema []byte, id, version string) (bool, error) {
		var parsedMessage interface{}
		if err := json.Unmarshal(message, &parsedMessage); err != nil {
			errBroken := errors.WithMessage(validator.ErrBrokenMessage, "Message is not in a valid format - "+err.Error())
			return false, errBroken
		}

		var compiledSchema *jsonschema.Schema
		key := id + "_" + version
		v, ok := cache.Get(key)
		if !ok {
			var err error
			compiledSchema, err = compileSchema(schema)
			if err != nil {
				errCompile := errors.WithMessage(validator.ErrWrongCompile, err.Error())
				return false, errCompile
			}
			cache.Add(key, compiledSchema)
		} else {
			compiledSchema = v.(*jsonschema.Schema)
		}

		if err := compiledSchema.Validate(parsedMessage); err != nil {
			var validationError *jsonschema.ValidationError
			ok := _errors.As(err, &validationError)
			if !ok {
				return false, errors.New("impossible to cast error to *jsonschema.ValidationError")
			}
			errMessage, err := createErrorMessage(validationError)
			if err != nil {
				return false, errors.New("couldn't create error message")
			}
			return false, errors.WithMessage(validator.ErrFailedValidation, errMessage)
		}
		return true, nil
	})
}

func createErrorMessage(validationError *jsonschema.ValidationError) (string, error) {
	errorMap := make(map[string]string)
	for _, cause := range validationError.Causes {
		errorMap[cause.KeywordLocation] = cause.Message
	}
	errMessage, err := json.Marshal(errorMap)
	if err != nil {
		return "", err
	}
	return string(errMessage), nil
}

func compileSchema(schema []byte) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schema)); err != nil {
		return nil, err
	}
	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

func NewGoJsonSchemaValidator() validator.Validator {
	return validator.Func(func(message, schema []byte, _, _ string) (bool, error) {
		if !json.Valid(message) {
			return false, validator.ErrDeadletter
		}

		schemaValidator, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schema))
		if err != nil {
			return false, validator.ErrDeadletter
		}

		result, err := schemaValidator.Validate(gojsonschema.NewBytesLoader(message))
		if err != nil {
			return false, err
		}
		if !result.Valid() {
			errMessage, err := createErrorMessageAlt(result.Errors())
			if err != nil {
				return false, err
			}
			return false, errors.WithMessage(validator.ErrFailedValidation, errMessage)
		}

		return result.Valid(), nil
	})
}

func NewCachedGoJsonSchemaValidator(size int) validator.Validator {
	cache, _ := lru.NewARC(size)

	return validator.Func(func(message, schema []byte, id, version string) (bool, error) {
		if !json.Valid(message) {
			return false, validator.ErrDeadletter
		}

		var compiledSchema *gojsonschema.Schema
		key := id + "_" + version
		v, ok := cache.Get(key)
		if !ok {
			var err error
			compiledSchema, err = gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schema))
			if err != nil {
				return false, validator.ErrDeadletter
			}
			cache.Add(key, compiledSchema)
		} else {
			compiledSchema = v.(*gojsonschema.Schema)
		}

		result, err := compiledSchema.Validate(gojsonschema.NewBytesLoader(message))
		if err != nil {
			return false, err
		}
		if !result.Valid() {
			errMessage, err := createErrorMessageAlt(result.Errors())
			if err != nil {
				return false, err
			}
			return false, errors.WithMessage(validator.ErrFailedValidation, errMessage)
		}

		return result.Valid(), nil
	})
}

func createErrorMessageAlt(validationError []gojsonschema.ResultError) (string, error) {
	errorMap := make(map[string]string)
	for _, e := range validationError {
		key := e.Details()["context"].(string)
		var expected, given, value string
		var ok1, ok2 bool
		if expected, ok1 = e.Details()["expected"].(string); !ok1 {
			value = "invalid value"
		}
		if given, ok2 = e.Details()["given"].(string); !ok2 {
			value = "invalid value"
		}
		if ok1 && ok2 {
			value = "expected " + expected + ", given " + given
		}
		errorMap[key] = value
	}
	errMessage, err := json.Marshal(errorMap)
	if err != nil {
		return "", err
	}
	return string(errMessage), nil
}
