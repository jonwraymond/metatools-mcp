# MCP Spec Alignment Proposal (2025-11-25)

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Goal
Ensure `metatools-mcp` is fully aligned with MCP spec **2025-11-25**, including features, error semantics, and transport behavior.

## Required Capabilities
- **Core features:** resources, prompts, tools.
- **Discovery:** `search_tools`, `describe_tool` with stable pagination.
- **Notifications:** tool list change events.
- **Execution:** cancellation + progress.
- **Client features:** sampling, roots, elicitation (where supported).
- **Transports:** stdio and streamable HTTP/SSE compliance.

## Implementation Notes
- Validate the MCP version constant in `toolfoundation/model` against 2025-11-25.
- Ensure transport behavior matches the spec (headers, session IDs, event streams).
- Add conformance tests where possible.

## References
- MCP spec: https://modelcontextprotocol.io/specification/2025-11-25
- `ai-tools-stack/docs/architecture/protocol-crosswalk.md`
