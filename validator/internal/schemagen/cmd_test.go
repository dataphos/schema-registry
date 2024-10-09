package schemagen

import (
	"os"
	"os/exec"
	"testing"
)

func TestOverScriptSchemaGenerator(t *testing.T) {
	if os.Getenv("MANUAL_TEST") == "" {
		t.Skip()
	}

	cmd := exec.Command("python", "json/json_schema_gen.py")
	_, err := ExternalCmdSchemaGenerator(cmd, []byte("{\n  \"id\": 100,\n  \"first_name\": \"syn jason\",\n  \"last_name\": \"syn oblak\",\n  \"email\": \"jsonsmail\"\n}"))
	if err != nil {
		t.Fatal(err)
	}
}
