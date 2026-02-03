# Pluggable Authentication & Authorization Middleware

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Overview
Provide a pluggable, policy-driven auth layer for tool discovery and execution. The core primitives live in `toolops/auth`, and `metatools-mcp` wires them as middleware.

## Goals
- Single auth source of truth: `toolops/auth`.
- Request-scoped identity context for all tool operations.
- Clear separation of **authentication** (who) and **authorization** (can do what).
- Configuration-driven selection of authenticators/authorizers.

## Non-Goals
- Full IAM system.
- Persisted policy storage (can be added later).

## Architecture
- **Authenticator** validates credentials and produces an identity.
- **Authorizer** evaluates permissions for tool operations.
- **Auth middleware** extracts request metadata, authenticates, authorizes, and injects identity into context.

## Interfaces (from `toolops/auth`)
- `Authenticator`, `Authorizer`, `Identity`, `AuthRequest`, `AuthResult`

## Integration Points
- `metatools-mcp/internal/middleware` for enforcement.
- `metatools-mcp/internal/transport` for request metadata (headers, method, resource).
- `toolexec/run` for optional tool-level scope checks.

## Config Surface
- Enable/disable authenticators (JWT, API key, OAuth2 introspection).
- Authorizer selection (RBAC, static policy).
- Default decision (deny by default, allow by default).

## References
- `toolops/auth`
- `metatools-mcp/docs/plans/2026-01-30-prd-017-auth-middleware.md`
