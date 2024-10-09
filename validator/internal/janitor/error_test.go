package janitor

import (
	"syscall"
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator"

	"github.com/pkg/errors"
)

func TestDeadletter(t *testing.T) {
	opError := OpError{
		Err: validator.ErrDeadletter,
	}
	if !opError.Deadletter() {
		t.Fatal("expected Deadletter")
	}

	opError.Err = errors.New("oops")
	if opError.Deadletter() {
		t.Fatal("shouldn't be Deadletter")
	}

	opError.Err = schemagen.ErrDeadletter
	if !opError.Deadletter() {
		t.Fatal("should be Deadletter")
	}
}

func TestTemporary(t *testing.T) {
	opError := OpError{
		Err: registry.ErrNotFound,
	}
	if !opError.Temporary() {
		t.Fatal("expected temporary")
	}

	opError.Err = syscall.ECONNREFUSED
	if opError.Temporary() {
		t.Fatal("expected not temporary")
	}
}
