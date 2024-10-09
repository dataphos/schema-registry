package csv

import (
	"bytes"
	"encoding/csv"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"
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
