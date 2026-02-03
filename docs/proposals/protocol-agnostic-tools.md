# Protocol-Agnostic Tools

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Summary
Normalize tool definitions across MCP, A2A, OpenAI, Anthropic, and Google by using `CanonicalTool` as the adapter hub.

## Goals
- Preserve round-trip fidelity via `SourceMeta`.
- Emit feature-loss warnings for narrower schemas.
- Provide adapters for MCP, OpenAI, Anthropic, A2A, and Gemini.

## Core Types
- `toolfoundation/adapter.CanonicalTool`
- `toolfoundation/adapter.Adapter`
- `toolfoundation/adapter.JSONSchema`

## Next Work
- Add A2A adapter (AgentCard + skills mapping).
- Add Gemini adapter (OpenAPI subset schemas).

## References
- `ai-tools-stack/docs/architecture/protocol-crosswalk.md`
