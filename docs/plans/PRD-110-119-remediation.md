# PRD-110–119 Remediation Plan

**Date:** 2026-01-31
**Owner:** Jon W. Raymond
**Scope:** PRD-110, PRD-111, PRD-112, PRD-113

## Objective

Close the Phase 1 infrastructure gaps by stabilizing repository setup, CI/CD workflows, org/repo settings, and release automation.

## Deliverables

| Item | Location | Status |
|---|---|---|
| Workflow templates | `.github/workflow-templates/*` | Done |
| Dependency review workflow | `.github/workflows/dependency-review.yml` (each repo) | Done |
| Release-please config updated | `release-please-config.json` (each repo) | Done |
| Release-please manifest normalized | `.release-please-manifest.json` | Done |
| Workflow permissions set | Repo settings (GITHUB_TOKEN write) | Done |
| Branch protection applied | main branch | Done |
| Repo settings standardized | allow-squash/rebase, delete-branch | Done |
| Topics + descriptions set | per-repo | Done |

## Plan of Record

### Task 1 — Repo Baseline (PRD-110)
- Verify all six repos exist and are public.
- Verify standard structure and workflows exist.

### Task 2 — CI/CD Templates (PRD-111)
- Add `.github/workflow-templates/` in root with CI, lint, commitlint, release-please, dependency-review templates.
- Add dependency-review workflow to each consolidated repo.

### Task 3 — GitHub Org/Repo Config (PRD-112)
- Enable workflow permissions (GITHUB_TOKEN write).
- Apply branch protection (status checks + code owner review).
- Standardize repo settings (merge strategy, delete branch on merge).
- Set topics + descriptions.
- Enable security features (vulnerability alerts + automated fixes).

### Task 4 — Release Automation (PRD-113)
- Update `release-please-config.json` to include component naming per repo.
- Normalize `.release-please-manifest.json` to `{".": "0.0.0"}`.
- Verify release-please workflow permissions enable PR creation.

## Verification Checklist

- [x] All six repos exist and are public
- [x] Workflow templates created
- [x] Dependency review workflow present in each repo
- [x] Release-please config/manifest updated per repo
- [x] Workflow permissions set to write
- [x] Branch protection enabled
- [x] Repo settings + topics applied
- [x] Security alerts enabled

## Notes

- Org name in examples is `ApertureStack`, but active repos are under `jonwraymond`. Commands will target `jonwraymond/*`.
