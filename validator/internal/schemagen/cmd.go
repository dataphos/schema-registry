package schemagen

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

// ExternalCmdSchemaGenerator generates the schema by calling the given cmd and passing the data to its stdin.
func ExternalCmdSchemaGenerator(cmd *exec.Cmd, data []byte) ([]byte, error) {
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, string(output))
	}
	if len(output) == 0 {
		return nil, ErrDeadletter
	}

	return output, nil
}
