# Pluggable Architecture Proposal (Consolidated Stack)

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Summary
The stack is intentionally pluggable at every layer: tool schemas, discovery, execution, composition, operations, and protocol bindings.

## Extension Points
- **Schemas & adapters:** `toolfoundation/adapter`
- **Discovery:** `tooldiscovery/index`, `search`, `semantic`, `tooldoc`
- **Execution:** `toolexec/backend`, `toolexec/runtime`
- **Composition:** `toolcompose/set`, `toolcompose/skill`
- **Ops:** `toolops/auth`, `cache`, `observe`, `health`, `resilience`
- **Protocol:** `toolprotocol/transport`, `wire`, `stream`, `task`, `resource`, `prompt`

## Guardrails
- Keep adapters side-effect free.
- Preserve schema fidelity; emit feature-loss warnings on conversion.
- Keep runtime execution isolated behind explicit backends.

## References
- `ai-tools-stack/docs/architecture/stack-map.md`
