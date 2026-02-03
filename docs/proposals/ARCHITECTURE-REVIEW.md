# Architecture Review: Consolidated Stack

**Status:** Revised (2026-02-02)
**Date:** 2026-02-02

## Executive Summary
The consolidated stack is structurally sound and aligns with the goals of protocol-agnostic tooling, progressive discovery, and sandboxed execution. The largest risks are **runtime backend completeness** and **protocol adapter coverage**. Auth consolidation is complete and now lives in `toolops/auth`.

## What Is Solid
- **Layer separation** is clean: `toolfoundation` → `tooldiscovery` → `toolexec` → `toolcompose` → `toolops` → `toolprotocol` → `metatools-mcp`.
- **Canonical model** embeds MCP SDK for spec fidelity.
- **Progressive discovery** has a clear pipeline and interfaces.
- **Execution/runtimes** are isolated behind stable contracts.

## Gaps to Close (Priority Order)
1. **Protocol adapters** beyond MCP/OpenAI/Anthropic (A2A + Google Gemini).
2. **Runtime parity** for Kubernetes, gVisor, Kata, Firecracker, and remote backends.
3. **Auth consolidation**: complete (single source in `toolops/auth`).
4. **Spec alignment**: MCP 2025-11-25 feature coverage and testing.

## Decisions (Reaffirmed)
- Keep `CanonicalTool` as the hub-and-spoke adapter format.
- Keep protocol primitives in `toolprotocol`; do not fork per protocol.
- Treat `metatools-mcp` as a **reference server**, not the source of truth.

## References
- `ai-tools-stack/docs/roadmap.md`
- `ai-tools-stack/docs/architecture/stack-map.md`
- `ai-tools-stack/docs/architecture/protocol-crosswalk.md`
