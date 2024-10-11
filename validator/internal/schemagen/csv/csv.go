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

package csv

import (
	"bytes"
	"encoding/csv"
	"github.com/dataphos/schema-registry-validator/internal/schemagen"
	"io"
	"strconv"
	"strings"
)

func New() schemagen.Generator {
	return schemagen.Func(func(data []byte) ([]byte, error) {
		reader := csv.NewReader(bytes.NewReader(data))

		reader.ReuseRecord = true
		reader.LazyQuotes = true
		reader.Comma = ','

		header, err := reader.Read()
		if err != nil {
			return nil, schemagen.ErrDeadletter
		}

		schema := parseHeaderIntoSchema(header)

		for {
			_, err = reader.Read()
			if err != nil {
				if err == io.EOF {
					return schema, nil
				}
				return nil, schemagen.ErrDeadletter
			}
		}
	})
}

func parseHeaderIntoSchema(header []string) []byte {
	var schemaBuilder bytes.Buffer

	schemaBuilder.Write([]byte("version 1.0\n"))
	schemaBuilder.Write([]byte("@totalColumns " + strconv.Itoa(len(header)) + "\n"))
	for _, key := range header {
		schemaBuilder.Write([]byte(strings.Trim(key, " ") + ":\n"))
	}

	return schemaBuilder.Bytes()
}
