# Component Library Analysis (Consolidated Stack)

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Summary
The consolidated stack is organized by capability layers. Each repo provides a focused set of packages with stable interfaces. This doc summarizes the layer map and points to the canonical stack map.

## Layer Map (Consolidated)
- **Foundation:** `toolfoundation` (canonical model + adapters)
- **Discovery:** `tooldiscovery` (registry + search + docs)
- **Execution:** `toolexec` (run + runtime + backend)
- **Composition:** `toolcompose` (tool sets + skills)
- **Operations:** `toolops` (auth + cache + observe + health + resilience)
- **Protocol:** `toolprotocol` (transport + wire + content + task + stream + session + prompt + resource + elicit)
- **Surface:** `metatools-mcp` (reference MCP server)

## References
- `ai-tools-stack/docs/architecture/stack-map.md`
