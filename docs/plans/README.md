# ApertureStack Consolidation PRDs

This directory contains all Product Requirement Documents for the ApertureStack ecosystem consolidation from 15 standalone repositories into 6 consolidated repositories.

## Quick Links

- [Master Plan](./CONSOLIDATION-MASTER-PLAN.md) - Executive overview
- [Order of Operations](./PRD-ORDER-OF-OPERATIONS.md) - Execution sequence

## PRD Summary

**Total PRDs:** 41
**Total Effort:** 236 hours (~29.5 days)
**Timeline:** 6-8 weeks with parallelization

---

## Phase 0: Planning & Documentation (12h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-100](./PRD-100-master-plan.md) | Master Consolidation Plan | 4h | Done |
| [PRD-101](./PRD-101-architecture-diagrams.md) | Architecture Diagrams | 4h | Done |
| [PRD-102](./PRD-102-schema-definitions.md) | Schema Definitions | 4h | Done |

## Phase 1: Infrastructure Setup (16h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-110](./PRD-110-repo-creation.md) | Repository Creation | 8h | Done |
| [PRD-111](./PRD-111-cicd-templates.md) | CI/CD Templates | 4h | Done |
| [PRD-112](./PRD-112-github-org-config.md) | GitHub Org Config | 2h | Done |
| [PRD-113](./PRD-113-release-automation.md) | Release Automation | 2h | Done |

## Phase 2: Foundation Layer - toolfoundation (16h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-120](./PRD-120-migrate-toolmodel.md) | Migrate toolmodel | 4h | Done |
| [PRD-121](./PRD-121-migrate-tooladapter.md) | Migrate tooladapter | 4h | Done |
| [PRD-122](./PRD-122-create-toolversion.md) | Create toolversion | 8h | Done |
| [PRD-123](./PRD-123-toolfoundation-docs-alignment.md) | Docs + README alignment | 3h | Done |
| [PRD-124](./PRD-124-toolfoundation-schema-policy.md) | Schema validation policy docs | 2h | Done |
| [PRD-125](./PRD-125-toolfoundation-adapter-matrix.md) | Adapter feature matrix docs | 2h | Done |
| [PRD-126](./PRD-126-toolfoundation-version-usage.md) | Version package usage docs | 2h | Done |
| [PRD-127](./PRD-127-toolfoundation-contracts.md) | Contract verification | 1h | Done |
| [PRD-128](./PRD-128-toolfoundation-release-propagation.md) | Release + propagation | 1h | Done |
| [PRD-129](./PRD-129-toolfoundation-g2-validation.md) | Gate G2 validation | 1h | Done |

## Phase 3: Discovery Layer - tooldiscovery (18h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-130](./PRD-130-migrate-toolindex.md) | Migrate toolindex | 4h | Done |
| [PRD-131](./PRD-131-migrate-toolsearch.md) | Migrate toolsearch | 4h | Done |
| [PRD-132](./PRD-132-migrate-toolsemantic.md) | Migrate toolsemantic | 6h | Done |
| [PRD-133](./PRD-133-migrate-tooldocs.md) | Migrate tooldocs | 4h | Done |
| [PRD-134](./PRD-134-tooldiscovery-docs-alignment.md) | Docs + README alignment | 2h | Done |
| [PRD-135](./PRD-135-tooldiscovery-search-policy.md) | Search strategy policy docs | 2h | Done |
| [PRD-136](./PRD-136-tooldiscovery-semantic-contracts.md) | Semantic contracts docs | 2h | Done |
| [PRD-137](./PRD-137-tooldiscovery-progressive-docs.md) | Progressive docs details | 2h | Done |
| [PRD-138](./PRD-138-tooldiscovery-release-propagation.md) | Release + propagation | 1h | Done |
| [PRD-139](./PRD-139-tooldiscovery-validation.md) | Discovery validation | 1h | Done |

## Phase 4: Execution Layer - toolexec (18h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-140](./PRD-140-migrate-toolrun.md) | Migrate toolrun | 4h | Ready |
| [PRD-141](./PRD-141-migrate-toolruntime.md) | Migrate toolruntime | 4h | Ready |
| [PRD-142](./PRD-142-migrate-toolcode.md) | Migrate toolcode | 4h | Ready |
| [PRD-143](./PRD-143-extract-toolbackend.md) | Extract toolbackend | 6h | Ready |

## Phase 5: Composition Layer - toolcompose (12h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-150](./PRD-150-migrate-toolset.md) | Migrate toolset | 4h | Ready |
| [PRD-151](./PRD-151-complete-toolskill.md) | Complete toolskill | 8h | Ready |

## Phase 6: Operations Layer - toolops (30h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-160](./PRD-160-migrate-toolobserve.md) | Migrate toolobserve | 4h | Ready |
| [PRD-161](./PRD-161-migrate-toolcache.md) | Migrate toolcache | 4h | Ready |
| [PRD-162](./PRD-162-extract-toolauth.md) | Extract toolauth | 8h | Ready |
| [PRD-163](./PRD-163-create-toolresilience.md) | Create toolresilience | 8h | Ready |
| [PRD-164](./PRD-164-create-toolhealth.md) | Create toolhealth | 6h | Ready |

## Phase 7: Protocol Layer - toolprotocol (84h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-170](./PRD-170-create-tooltransport.md) | Create tooltransport | 8h | Ready |
| [PRD-171](./PRD-171-create-toolwire.md) | Create toolwire | 12h | Ready |
| [PRD-172](./PRD-172-create-tooldiscover.md) | Create tooldiscover | 8h | Ready |
| [PRD-173](./PRD-173-create-toolcontent.md) | Create toolcontent | 8h | Ready |
| [PRD-174](./PRD-174-create-tooltask.md) | Create tooltask | 10h | Ready |
| [PRD-175](./PRD-175-create-toolstream.md) | Create toolstream | 8h | Ready |
| [PRD-176](./PRD-176-create-toolsession.md) | Create toolsession | 6h | Ready |
| [PRD-177](./PRD-177-create-toolelicit.md) | Create toolelicit | 6h | Ready |
| [PRD-178](./PRD-178-create-toolresource.md) | Create toolresource | 10h | Ready |
| [PRD-179](./PRD-179-create-toolprompt.md) | Create toolprompt | 8h | Ready |

## Phase 8: Integration (22h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-180](./PRD-180-update-metatools-mcp.md) | Update metatools-mcp | 12h | Ready |
| [PRD-181](./PRD-181-update-ai-tools-stack.md) | Update ai-tools-stack | 4h | Ready |
| [PRD-182](./PRD-182-documentation-site.md) | Documentation Site | 6h | Ready |

## Phase 9: Cleanup (8h)

| PRD | Title | Effort | Status |
|-----|-------|--------|--------|
| [PRD-190](./PRD-190-archive-old-repos.md) | Archive Old Repos | 2h | Ready |
| [PRD-191](./PRD-191-update-submodules.md) | Update Submodules | 2h | Ready |
| [PRD-192](./PRD-192-validation.md) | Validation | 4h | Ready |

---

## Checkpoint Gates

| Gate | After PRD | Validation |
|------|-----------|------------|
| **G1** | PRD-113 | All repos created, CI working |
| **G2** | PRD-122 | Foundation layer complete |
| **G3** | PRD-143 | Discovery + Execution layers complete |
| **G4** | PRD-164 | Composition + Operations layers complete |
| **G5** | PRD-179 | Protocol layer complete |
| **G6** | PRD-182 | Integration complete |
| **G7** | PRD-192 | Full validation complete |

---

## Consolidated Repositories

| Repository | Packages |
|------------|----------|
| toolfoundation | model, adapter, version |
| tooldiscovery | index, search, semantic, docs |
| toolexec | run, runtime, code, backend |
| toolcompose | set, skill |
| toolops | observe, cache, resilience, health, auth |
| toolprotocol | transport, wire, discover, content, task, stream, session, elicit, resource, prompt |

---

## Previous Plans

| Item | File | Status |
|------|------|--------|
| PRD-016 | `2026-01-30-prd-016-interface-contracts.md` | Done |
| PRD-017 | `2026-01-30-prd-017-auth-middleware.md` | Done |
