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

package centralconsumer

import (
	"golang.org/x/net/context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dataphos/schema-registry-validator/internal/janitor"
	"github.com/dataphos/schema-registry-validator/internal/publisher"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/schema-registry-validator/internal/validator"
	localjson "github.com/dataphos/schema-registry-validator/internal/validator/json"
)

func TestTopicsIntoRoutingFunc(t *testing.T) {
	topics := Topics{
		Valid:       "valid-topic",
		InvalidCSV:  "deadletter",
		InvalidJSON: "deadletter",
		Deadletter:  "deadletter",
	}

	tt := []struct {
		name        string
		isValid     janitor.Result
		format      string
		destination string
	}{
		{"valid avro", janitor.Valid, avroFormat, topics.Valid},
		{"invalid avro", janitor.Invalid, avroFormat, topics.Deadletter},
		{"deadletter avro", janitor.Deadletter, avroFormat, topics.Deadletter},
		{"missing schema avro", janitor.MissingSchema, avroFormat, topics.Deadletter},
		{"valid protobuf", janitor.Valid, protobufFormat, topics.Valid},
		{"invalid protobuf", janitor.Invalid, protobufFormat, topics.Deadletter},
		{"deadletter protobuf", janitor.Deadletter, protobufFormat, topics.Deadletter},
		{"missing schema protobuf", janitor.MissingSchema, protobufFormat, topics.Deadletter},
		{"valid xml", janitor.Valid, xmlFormat, topics.Valid},
		{"invalid xml", janitor.Invalid, xmlFormat, topics.Deadletter},
		{"deadletter xml", janitor.Deadletter, xmlFormat, topics.Deadletter},
		{"missing schema xml", janitor.MissingSchema, xmlFormat, topics.Deadletter},
		{"valid json", janitor.Valid, jsonFormat, topics.Valid},
		{"invalid json", janitor.Invalid, jsonFormat, topics.InvalidJSON},
		{"deadletter json", janitor.Deadletter, jsonFormat, topics.Deadletter},
		{"missing schema json", janitor.MissingSchema, jsonFormat, topics.InvalidJSON},
		{"valid csv", janitor.Valid, csvFormat, topics.Valid},
		{"invalid csv", janitor.Invalid, csvFormat, topics.InvalidCSV},
		{"deadletter csv", janitor.Deadletter, csvFormat, topics.Deadletter},
		{"missing schema csv", janitor.MissingSchema, csvFormat, topics.InvalidCSV},
	}

	routingFunc := intoRouter(topics)

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			destination := routingFunc.Route(tc.isValid, janitor.Message{Format: tc.format})
			if destination != tc.destination {
				t.Errorf("expected and actual destination not the same (%s != %s)", tc.destination, destination)
			}
		})
	}
}

func TestOneCCPerTopic(t *testing.T) {
	ctx := context.Background()
	topics := Topics{
		Valid:       "valid",
		InvalidCSV:  "deadletter",
		InvalidJSON: "deadletter",
		Deadletter:  "deadletter",
	}

	_, b, _, _ := runtime.Caller(0)
	dir := filepath.Dir(b)
	testdataDir := filepath.Join(dir, "testdata")

	data1, err := os.ReadFile(filepath.Join(testdataDir, "data-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	data2, err := os.ReadFile(filepath.Join(testdataDir, "data-2.json"))
	if err != nil {
		t.Fatal(err)
	}
	data3, err := os.ReadFile(filepath.Join(testdataDir, "data-3.json"))
	if err != nil {
		t.Fatal(err)
	}
	schemaSpec1, err := os.ReadFile(filepath.Join(testdataDir, "schema-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	schemaSpec2, err := os.ReadFile(filepath.Join(testdataDir, "schema-2.json"))
	if err != nil {
		t.Fatal(err)
	}
	schemaSpec3, err := os.ReadFile(filepath.Join(testdataDir, "schema-3.json"))
	if err != nil {
		t.Fatal(err)
	}

	schemaRegistry := registry.NewMock()
	schemaRegistry.SetGetResponse("1", "1", schemaSpec1, nil)
	schemaRegistry.SetGetResponse("1", "2", schemaSpec2, nil)
	schemaRegistry.SetGetResponse("1", "3", schemaSpec3, nil)

	validators := make(map[string]validator.Validator)
	validators["json"] = localjson.New()
	encryptionKey := ""

	cc, err := New(schemaRegistry, &publisher.MockPublisher{}, validators, topics, Settings{}, nil, RouterFlags{}, Mode(1),
		SchemaMetadata{
			ID:      "1",
			Version: "1",
			Format:  "json",
		},
		encryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	message1 := janitor.Message{
		ID:            "",
		Key:           "",
		RawAttributes: map[string]interface{}{},
		Payload:       data1,
		IngestionTime: time.Time{},
		SchemaID:      "1",
		Version:       "1",
		Format:        "json",
	}
	message2 := janitor.Message{
		ID:            "",
		Key:           "",
		RawAttributes: map[string]interface{}{},
		Payload:       data2,
		IngestionTime: time.Time{},
		SchemaID:      "1",
		Version:       "2",
		Format:        "json",
	}
	message3 := janitor.Message{
		ID:            "",
		Key:           "",
		RawAttributes: map[string]interface{}{},
		Payload:       data3,
		IngestionTime: time.Time{},
		SchemaID:      "1",
		Version:       "3",
		Format:        "json",
	}

	tt := []struct {
		name          string
		expectedTopic string
		version       string
		message       janitor.Message
	}{
		{"valid data1 unspecified", "valid", "", message1},
		{"valid data2 specified", "valid", "2", message2},
		//{"invalid data1 unspecified", "deadletter", "", message1},
		{"valid data3 specified", "valid", "3", message3},
		//{"invalid data2 unspecified", "deadletter", "", message2},
		{"invalid data 1 against v2", "deadletter", "2", message1},
		{"invalid data 1 against v3", "deadletter", "3", message1},
		{"invalid data 2 against v3", "deadletter", "3", message2},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.message.Version = tc.version
			messageTopicPair, err := cc.Handle(ctx, tc.message)
			if err != nil {
				t.Fatal(err)
			}
			if messageTopicPair.Topic != tc.expectedTopic {
				t.Errorf("expected and actual destination not the same (%s != %s)", tc.expectedTopic, messageTopicPair.Topic)
			}
		})
	}
}
