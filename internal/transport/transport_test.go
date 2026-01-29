package transport

import "testing"

func TestTransportInterface(_ *testing.T) {
	var _ Transport = (*StdioTransport)(nil)
	var _ Transport = (*SSETransport)(nil)
}

func TestTransportInfo(t *testing.T) {
	stdio := &StdioTransport{}
	if stdio.Info().Name != "stdio" {
		t.Errorf("stdio Info().Name = %q", stdio.Info().Name)
	}

	sse := &SSETransport{Config: SSEConfig{Host: "127.0.0.1", Port: 8080, Path: "/mcp"}}
	info := sse.Info()
	if info.Name != "sse" {
		t.Errorf("sse Info().Name = %q", info.Name)
	}
	if info.Path != "/mcp" {
		t.Errorf("sse Info().Path = %q", info.Path)
	}
}
