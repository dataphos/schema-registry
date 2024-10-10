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

package janitor

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"
	"github.com/dataphos/lib-streamproc/pkg/streamproc"

	"github.com/pkg/errors"
)

const (
	AvroFormat     = "avro"
	CSVFormat      = "csv"
	JSONFormat     = "json"
	ProtobufFormat = "protobuf"
	XMLFormat      = "xml"
)

func TestParse(t *testing.T) {
	brokerMessage := streamproc.Message{
		Data: []byte("this is supposed to be some json data"),
		Attributes: map[string]interface{}{
			AttributeSchemaID:      "1",
			AttributeSchemaVersion: "1",
			AttributeFormat:        "json",
		},
	}

	expected := Message{
		RawAttributes: brokerMessage.Attributes,
		Payload:       brokerMessage.Data,
		SchemaID:      "1",
		Version:       "1",
		Format:        "json",
	}

	actual, err := ParseMessage(brokerMessage)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatal("expected and actual message not the same")
	}
}

func TestParseError(t *testing.T) {
	brokerMessages := []streamproc.Message{
		{
			Data: []byte("this is supposed to be some json data"),
			Attributes: map[string]interface{}{
				AttributeSchemaID:      "1",
				AttributeSchemaVersion: "1",
			},
		},
		{
			Data: []byte("this is supposed to be some json data"),
			Attributes: map[string]interface{}{
				AttributeSchemaID:      "1",
				AttributeSchemaVersion: 1,
				"format":               "json",
			},
		},
		{
			Data: []byte("this is supposed to be some json data"),
			Attributes: map[string]interface{}{
				AttributeSchemaID:      1,
				AttributeSchemaVersion: "1",
				"format":               "json",
			},
		},
	}

	for i, brokerMessage := range brokerMessages {
		brokerMessage := brokerMessage
		t.Run("parsing failure number "+strconv.Itoa(i), func(t *testing.T) {
			_, err := ParseMessage(brokerMessage)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestValidatorsValidate(t *testing.T) {
	tt := []struct {
		name              string
		message           Message
		isValid           bool
		shouldReturnError bool
	}{
		{
			name: "is valid",
			message: Message{
				SchemaID: "1",
				Version:  "1",
				Format:   JSONFormat,
				Payload:  []byte("this is some data to be validates against a Schema, not important for this unit test"),
			},
			isValid:           true,
			shouldReturnError: false,
		},
		{
			name: "is not valid",
			message: Message{
				SchemaID: "1",
				Version:  "1",
				Format:   JSONFormat,
				Payload:  []byte("this is some data to be validates against a Schema, not important for this unit test"),
			},
			isValid:           false,
			shouldReturnError: false,
		},
		{
			name: "valid, but the format is not supported",
			message: Message{
				SchemaID: "1",
				Version:  "1",
				Format:   JSONFormat,
				Payload:  []byte("this is some data to be validates against a Schema, not important for this unit test"),
			},
			isValid:           true,
			shouldReturnError: false,
		},
		{
			name: "throws error",
			message: Message{
				SchemaID: "1",
				Version:  "1",
				Format:   JSONFormat,
				Payload:  []byte("this is some data to be validates against a schema, not important for this unit test"),
			},
			isValid:           false,
			shouldReturnError: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			schema := []byte("schema for message validation, not important for this unit test")
			validators := Validators(map[string]validator.Validator{
				JSONFormat: validator.Func(func(message, schema []byte, id string, version string) (bool, error) {
					if tc.shouldReturnError {
						return false, errors.New("test error")
					}
					return tc.isValid, nil
				}),
			})

			isValid, err := validators.Validate(tc.message, schema)
			if err != nil {
				if !tc.shouldReturnError {
					if tc.message.Format == JSONFormat {
						t.Error("error occurred but was not expected", err)
					}
				}
			}
			if isValid && !tc.isValid {
				t.Error("message set to valid but not valid expected")
			}
		})
	}
}
