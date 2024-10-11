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

package producer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dataphos/schema-registry-validator/internal/janitor"
	"github.com/dataphos/schema-registry-validator/internal/registry"
	"github.com/dataphos/lib-brokers/pkg/broker/inmem"
)

func TestProducer(t *testing.T) {
	publisher := inmem.Publisher{}
	topic, _ := publisher.Topic("some-topic")

	_, b, _, _ := runtime.Caller(0)
	dir := filepath.Dir(b)
	testdataDir := filepath.Join(dir, "testdata")

	messageToRegister, err := os.ReadFile(filepath.Join(testdataDir, "data-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	schemaToRegister, err := os.ReadFile(filepath.Join(testdataDir, "schema-1.json"))
	if err != nil {
		t.Fatal(err)
	}

	schemaRegistry := registry.NewMock()
	schemaRegistry.SetRegisterResponse(schemaToRegister, "1", "1", nil)

	encryptionKey := ""
	producer := New(schemaRegistry, topic, 100, 0, encryptionKey)

	root := filepath.Join(dir, "../..")
	dataset := filepath.Join(dir, "testdata/dataset.csv")

	if err = producer.LoadAndProduce(context.Background(), root, dataset, 5); err != nil {
		t.Fatal(err)
	}

	if len(publisher.Spawned[0].Published) != 5 {
		t.Fatal("publish count not as expected")
	}
	for _, published := range publisher.Spawned[0].Published {
		if published.Attributes[janitor.AttributeFormat] != "json" {
			t.Fatal("format not as expected")
		}

		if bytes.Equal(published.Data, messageToRegister) {
			if published.Attributes[janitor.AttributeSchemaID] != "1" ||
				published.Attributes[janitor.AttributeSchemaVersion] != "1" {
				t.Fatal("schema id and version not as expected")
			}
		} else {
			if _, ok := published.Attributes[janitor.AttributeSchemaID]; ok {
				t.Fatal("schema id not empty, empty expected")
			}
			if _, ok := published.Attributes[janitor.AttributeSchemaVersion]; ok {
				t.Fatal("schema version not empty, empty expected")
			}
		}
	}

}
