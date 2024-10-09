package avro

import (
	"bytes"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"

	"github.com/hamba/avro"
)

func New() validator.Validator {
	return validator.Func(func(message, schema []byte, _, _ string) (bool, error) {
		parsedSchema, err := avro.Parse(string(schema))
		if err != nil {
			return false, validator.ErrDeadletter
		}

		var data interface{}
		if err = avro.Unmarshal(parsedSchema, message, &data); err != nil {
			return false, nil
		}

		reserializedMessage, err := avro.Marshal(parsedSchema, data)
		if err != nil {
			return false, nil
		}

		return bytes.Equal(reserializedMessage, message), nil
	})
}
