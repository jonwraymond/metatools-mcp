# Multi-Tenancy Extension

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Overview
Introduce tenant-aware boundaries across discovery, execution, and policy enforcement.

## Goals
- Tenant-scoped tool registries and search.
- Tenant-aware execution isolation.
- Clear authz model for cross-tenant access.

## Design
- **Tenant identity** sourced from `toolops/auth` identity context.
- **Tenant registry**: namespace tool IDs by tenant and enforce visibility rules.
- **Runtime isolation**: map tenant to runtime backend or resource pool.

## Dependencies
- `toolops/auth` for identity + scopes.
- `tooldiscovery/index` for scoped search.
- `toolexec/runtime` for backend isolation.

## References
- `ai-tools-stack/docs/roadmap.md`
