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

package avro

import (
	"bytes"
	"github.com/pkg/errors"

	"github.com/dataphos/schema-registry-validator/internal/validator"

	"github.com/hamba/avro"
)

func New() validator.Validator {
	return validator.Func(func(message, schema []byte, _, _ string) (bool, error) {
		parsedSchema, err := avro.Parse(string(schema))
		if err != nil {
			return false, errors.WithMessage(validator.ErrParsingMessage, err.Error())
		}

		var data interface{}
		if err = avro.Unmarshal(parsedSchema, message, &data); err != nil {
			return false, errors.WithMessage(validator.ErrUnmarshalAvro, err.Error())
		}

		reserializedMessage, err := avro.Marshal(parsedSchema, data)
		if err != nil {
			return false, errors.WithMessage(validator.ErrMarshalAvro, err.Error())
		}

		return bytes.Equal(reserializedMessage, message), nil
	})
}
