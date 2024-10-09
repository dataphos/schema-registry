package producer

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitor"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"
	"github.com/dataphos/lib-brokers/pkg/broker"

	"github.com/pkg/errors"
)

// IntoBrokerMessages converts the given ProcessEntry slice into a slice of broker.Message, by calling IntoBrokerMessage on each entry.
func IntoBrokerMessages(ctx context.Context, entries []ProcessedEntry, registry registry.SchemaRegistry) ([]broker.OutboundMessage, error) {
	brokerMessages := make([]broker.OutboundMessage, len(entries))
	for i, curr := range entries {
		message, err := IntoBrokerMessage(ctx, curr, registry)
		if err != nil {
			return nil, err
		}
		brokerMessages[i] = message
	}
	return brokerMessages, nil
}

// IntoBrokerMessage converts the given ProcessedEntry into broker.Message, while also performing schema registration, depending
// on the value of ProcessedEntry.ShouldRegister.
func IntoBrokerMessage(ctx context.Context, entry ProcessedEntry, registry registry.SchemaRegistry) (broker.OutboundMessage, error) {
	attributes := make(map[string]interface{})

	attributes[janitor.AttributeFormat] = entry.Format
	if entry.Version != "" {
		attributes[janitor.AttributeSchemaVersion] = entry.Version
	}

	if entry.ShouldRegister {
		var schemaId, versionId string
		var err error
		schemaId, versionId, err = registry.Register(ctx, entry.Schema, entry.Format, entry.CompatibilityMode, entry.ValidityMode)
		if err != nil {
			return broker.OutboundMessage{}, err
		}

		log.Printf("schema registered under %s/%s\n", schemaId, versionId)

		attributes[janitor.AttributeSchemaID] = schemaId
		attributes[janitor.AttributeSchemaVersion] = versionId

		for k, v := range entry.AdditionalAttributes {
			attributes[k] = v
		}
	}

	return broker.OutboundMessage{
		Data:       entry.Message,
		Attributes: attributes,
	}, nil
}

// ProcessedEntry is the processed version of Entry.
type ProcessedEntry struct {
	Message              []byte
	Schema               []byte
	ShouldRegister       bool
	Format               string
	CompatibilityMode    string
	ValidityMode         string
	Version              string
	AdditionalAttributes map[string]interface{}
}

// ProcessEntries processes the given Entry slice, by calling ProcessEntry on each Entry.
func ProcessEntries(baseDir string, entries []Entry) ([]ProcessedEntry, error) {
	messageSchemaPairs := make([]ProcessedEntry, len(entries))
	var loaded ProcessedEntry
	var err error
	for i, curr := range entries {
		loaded, err = ProcessEntry(baseDir, curr)
		if err != nil {
			return nil, err
		}
		messageSchemaPairs[i] = loaded
	}
	return messageSchemaPairs, nil
}

// ProcessEntry processes the given Entry, by loading ProcessedEntry.Message and ProcessedEntry.Schema from the filenames
// in the relevant Entry fields. It is assumed that the correct absolute path of the given filename can be gained by calling
// filepath.Join with the given baseDir as the first argument.
func ProcessEntry(baseDir string, entry Entry) (ProcessedEntry, error) {
	var message, schema []byte
	var err error

	messageFilename := filepath.Join(baseDir, entry.BlobFilename)
	log.Printf("loading message data from %s\n", messageFilename)
	message, err = os.ReadFile(messageFilename)
	if err != nil {
		return ProcessedEntry{}, err
	}

	if entry.SchemaFilename == "" {
		if entry.ShouldRegister {
			return ProcessedEntry{}, errors.New("invalid entry: can't register an entry since no file was given")
		}
	} else {
		schemaFilename := filepath.Join(baseDir, entry.SchemaFilename)
		log.Printf("loading schema data from %s\n", schemaFilename)
		schema, err = os.ReadFile(schemaFilename)
		if err != nil {
			return ProcessedEntry{}, err
		}
	}

	return ProcessedEntry{
		Message:              message,
		Schema:               schema,
		ShouldRegister:       entry.ShouldRegister,
		Format:               entry.Format,
		CompatibilityMode:    entry.CompatibilityMode,
		ValidityMode:         entry.ValidityMode,
		Version:              entry.Version,
		AdditionalAttributes: entry.Attributes,
	}, nil
}

type Entry struct {
	BlobFilename      string
	SchemaFilename    string
	ShouldRegister    bool
	Format            string
	CompatibilityMode string
	ValidityMode      string
	Version           string
	Attributes        map[string]interface{}
}

// LoadEntries loads the entries from the given csv file.
func LoadEntries(filename string) ([]Entry, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseEntriesFile(bytes.NewReader(file))
}

// parseEntriesFile parses the given csv file.
func parseEntriesFile(file io.Reader) ([]Entry, error) {
	reader := csv.NewReader(file)

	reader.Comma = ','

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, len(lines))
	var curr Entry
	for i, line := range lines {
		if len(line) == 4 {
			curr, err = parseLineNoSchema(line)
		} else if len(line) == 5 {
			curr, err = parseLineOneCCPerTopic(line)
		} else {
			curr, err = parseLine(line)
		}
		if err != nil {
			return nil, err
		}
		entries[i] = curr
	}

	return entries, nil
}

// parseLine parses a given line of the csv file.
func parseLine(line []string) (Entry, error) {
	shouldRegister, err := strconv.ParseBool(line[2])
	if err != nil {
		return Entry{}, err
	}

	var attributes = make(map[string]interface{})
	attrs := strings.Split(line[6], ";")
	for _, at := range attrs {
		if at == "" {
			continue
		}
		pair := strings.SplitN(at, "=", 2)
		attributes[pair[0]] = pair[1]
	}

	return Entry{
		BlobFilename:      line[0],
		SchemaFilename:    line[1],
		ShouldRegister:    shouldRegister,
		Format:            line[3],
		CompatibilityMode: line[4],
		ValidityMode:      line[5],
		Version:           "",
		Attributes:        attributes,
	}, nil
}

func parseLineNoSchema(line []string) (Entry, error) {
	return Entry{
		BlobFilename:      line[0],
		SchemaFilename:    "",
		ShouldRegister:    false,
		Format:            line[1],
		CompatibilityMode: line[2],
		ValidityMode:      line[3],
		Version:           "",
		Attributes:        nil,
	}, nil
}

func parseLineOneCCPerTopic(line []string) (Entry, error) {
	return Entry{
		BlobFilename:      line[0],
		SchemaFilename:    "",
		ShouldRegister:    false,
		Version:           line[1],
		Format:            line[2],
		CompatibilityMode: line[3],
		ValidityMode:      line[4],
		Attributes:        nil,
	}, nil
}
