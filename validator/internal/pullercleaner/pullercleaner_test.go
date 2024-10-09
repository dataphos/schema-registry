package pullercleaner

import (
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitor"
)

func TestTopicsIntoRoutingFunc(t *testing.T) {
	topics := Topics{
		Valid:      "valid-topic",
		Deadletter: "deadletter",
	}

	tt := []struct {
		name        string
		isValid     janitor.Result
		format      string
		destination string
	}{
		{"valid json", janitor.Valid, jsonFormat, topics.Valid},
		{"invalid json", janitor.Invalid, jsonFormat, topics.Deadletter},
		{"deadletter json", janitor.Deadletter, jsonFormat, topics.Deadletter},
		{"missing schema json", janitor.MissingSchema, jsonFormat, topics.Deadletter},
		{"valid csv", janitor.Valid, csvFormat, topics.Valid},
		{"invalid csv", janitor.Invalid, csvFormat, topics.Deadletter},
		{"deadletter csv", janitor.Deadletter, csvFormat, topics.Deadletter},
		{"missing schema csv", janitor.MissingSchema, csvFormat, topics.Deadletter},
	}

	routingFunc := IntoRoutingFunc(topics)

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
