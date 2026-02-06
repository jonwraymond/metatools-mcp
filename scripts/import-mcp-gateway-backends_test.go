package main

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestImportBackendsYAML_FiltersRemoteHTTPAndSSE(t *testing.T) {
	in := []byte(`
backends:
  - name: local
    command: echo
    args: ["hi"]

  - name: remote-http
    url: "https://example.com/mcp"
    headers:
      Authorization: "Bearer secretref:bws:project/dotenv/key/TOKEN"

  - name: remote-streamable
    transport: "streamable-http"
    url: "https://stream.example.com/mcp"

  - name: remote-sse
    url: "sse://sse.example.com/mcp"

  - name: should-exclude
    command: somebin
    url: "https://exclude.example.com/mcp"
`)

	out, err := importBackendsYAML(in)
	if err != nil {
		t.Fatalf("importBackendsYAML returned error: %v", err)
	}

	var parsed outputConfig
	if err := yaml.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if !parsed.Secrets.Strict {
		t.Fatalf("expected secrets.strict=true")
	}
	bws, ok := parsed.Secrets.Providers["bws"]
	if !ok || !bws.Enabled {
		t.Fatalf("expected bws provider enabled")
	}

	if len(parsed.Backends.MCP) != 3 {
		t.Fatalf("expected 3 remote backends, got %d", len(parsed.Backends.MCP))
	}
	if parsed.Backends.MCP[0].Name != "remote-http" {
		t.Fatalf("unexpected ordering or name: %q", parsed.Backends.MCP[0].Name)
	}
	if parsed.Backends.MCP[1].Name != "remote-sse" {
		t.Fatalf("unexpected ordering or name: %q", parsed.Backends.MCP[1].Name)
	}
	if parsed.Backends.MCP[2].Name != "remote-streamable" {
		t.Fatalf("unexpected ordering or name: %q", parsed.Backends.MCP[2].Name)
	}
	if got := parsed.Backends.MCP[0].Headers["Authorization"]; got != "Bearer secretref:bws:project/dotenv/key/TOKEN" {
		t.Fatalf("header secretref was not preserved: %q", got)
	}
}

