package janitor

import (
	"testing"

	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
)

func TestRoutingFunc(t *testing.T) {
	tt := []struct {
		name        string
		routerFlags RouterFlags
		result      Result
	}{
		{
			"valid is propagated",
			RouterFlags{
				MissingSchema: true,
				Valid:         true,
				Deadletter:    true,
			},
			Valid,
		},
		{
			"invalid is propagated",
			RouterFlags{
				MissingSchema: true,
				Valid:         true,
				Deadletter:    true,
			},
			Invalid,
		},
		{
			"invalid is propagated",
			RouterFlags{
				MissingSchema: true,
				Valid:         true,
				Deadletter:    true,
			},
			MissingSchema,
		},
		{
			"Deadletter is propagated",
			RouterFlags{
				MissingSchema: true,
				Valid:         true,
				Deadletter:    true,
			},
			Deadletter,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			destination := "some topic"
			var called int
			next := RoutingFunc(func(result Result, message Message) string {
				if result != tc.result {
					t.Fatal("wrong result propagated")
				}
				called++

				return destination
			})

			r := LoggingRouter(standardlogger.New(logger.L{}), tc.routerFlags, next)

			actual := r.Route(tc.result, Message{})
			if actual != destination {
				t.Error("expected and actual not the same")
			}

			if called != 1 {
				t.Error("not propagated correctly")
			}
		})
	}
}
