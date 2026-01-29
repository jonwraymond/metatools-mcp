package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidateCmd(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "metatools.yaml")

		yaml := `
server:
  name: test
transport:
  type: stdio
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cmd := NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"config", "validate", "--config", configPath})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !contains(buf.String(), "valid") {
			t.Errorf("Output should indicate config is valid, got: %s", buf.String())
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "metatools.yaml")

		yaml := `
transport:
  type: invalid
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cmd := NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"config", "validate", "--config", configPath})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("Execute() should return error for invalid config")
		}
	})
}
