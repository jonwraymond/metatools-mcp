package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	Version = "1.2.3"
	GitCommit = "abc123"
	BuildDate = "2026-01-28"

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !contains(output, "1.2.3") {
		t.Errorf("Version output should contain version, got: %s", output)
	}
	if !contains(output, "abc123") {
		t.Errorf("Version output should contain git commit, got: %s", output)
	}
}

func TestVersionCmd_JSON(t *testing.T) {
	Version = "1.2.3"
	GitCommit = "abc123"
	BuildDate = "2026-01-28"

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !contains(output, `"version"`) {
		t.Errorf("JSON output should contain version field, got: %s", output)
	}
}
