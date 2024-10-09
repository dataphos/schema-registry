package json

import (
	"os/exec"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/schemagen"
)

func New(filename string) schemagen.Generator {
	return schemagen.Func(func(data []byte) ([]byte, error) {
		// #nosec G204 this would usually be a security concern because of remote code execution,
		// but it's fine here since we execute a python script from a file, so the attacker would need to have
		// full access to the vm to execute the script, and in that case, they could just execute the script themselves
		return schemagen.ExternalCmdSchemaGenerator(exec.Command("python", filename), data)
	})
}
