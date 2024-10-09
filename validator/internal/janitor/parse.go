package janitor

import (
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errtemplates"
	"github.com/dataphos/lib-streamproc/pkg/streamproc"
)

// ParseMessage parses a given broker.Message into Message, by setting Message.Payload to the value of the data field of the given
// broker.Message, and extracting the other fields from the attributes field.
//
// ParseMessage checks if the attributes field contains the following keys: AttributeSchemaID, AttributeSchemaVersion and AttributeFormat.
// If AttributeSchemaID or AttributeSchemaVersion are present, then it is assumed they are strings, returning an error otherwise.
// The AttributeFormat key must be present and must be a non-empty string.
func ParseMessage(message streamproc.Message) (Message, error) {
	parsed := Message{
		ID:            message.ID,
		Key:           message.Key,
		Payload:       message.Data,
		RawAttributes: message.Attributes,
		IngestionTime: message.IngestionTime,
	}

	attributes, err := ExtractAttributes(message.Attributes)
	if err != nil {
		return parsed, err
	}
	parsed.SchemaID = attributes.SchemaId
	parsed.Version = attributes.SchemaVersion
	parsed.Format = attributes.Format

	return parsed, nil
}

type Attributes struct {
	SchemaId      string
	SchemaVersion string
	Format        string
}

func ExtractAttributes(raw map[string]interface{}) (Attributes, error) {
	var schemaIDStr, versionStr, formatStr string

	schemaID, ok := raw[AttributeSchemaID]
	if ok {
		schemaIDStr, ok = schemaID.(string)
		if !ok {
			return Attributes{}, errtemplates.AttributeNotAString(AttributeSchemaID)
		}
	}

	version, ok := raw[AttributeSchemaVersion]
	if ok {
		versionStr, ok = version.(string)
		if !ok {
			return Attributes{}, errtemplates.AttributeNotAString(AttributeSchemaVersion)
		}
	}
	format, ok := raw[AttributeFormat]
	if !ok {
		return Attributes{}, errtemplates.AttributeNotDefined(AttributeFormat)
	}
	formatStr, ok = format.(string)
	if !ok {
		return Attributes{}, errtemplates.AttributeNotAString(AttributeFormat)
	}
	if formatStr == "" {
		return Attributes{}, errtemplates.MustNotBeEmpty(AttributeFormat)
	}

	return Attributes{
		SchemaId:      schemaIDStr,
		SchemaVersion: versionStr,
		Format:        formatStr,
	}, nil
}