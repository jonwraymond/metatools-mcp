package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !contains(output, "metatools") {
		t.Errorf("Help should contain 'metatools', got: %s", output)
	}
	if !contains(output, "serve") {
		t.Errorf("Help should list 'serve' subcommand, got: %s", output)
	}
	if !contains(output, "version") {
		t.Errorf("Help should list 'version' subcommand, got: %s", output)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
