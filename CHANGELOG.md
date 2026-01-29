# Changelog

## [0.3.0](https://github.com/jonwraymond/metatools-mcp/compare/metatools-mcp-v0.2.0...metatools-mcp-v0.3.0) (2026-01-29)


### Features

* **mcp:** implement spec alignment per PRD-015 ([037326f](https://github.com/jonwraymond/metatools-mcp/commit/037326fbcc932e16dc761f773feea50fd63773fb))
* **mcp:** implement spec alignment per PRD-015 ([bae429c](https://github.com/jonwraymond/metatools-mcp/commit/bae429cb5c43ea8f785f1e67a3b870943bf9d01f))
* **mcp:** stabilize list_changed and add changelog ([7f191e8](https://github.com/jonwraymond/metatools-mcp/commit/7f191e8ceebed1a178f540912df51771293a54cc))
* progress notifications and cancellation codes ([dc1d9a2](https://github.com/jonwraymond/metatools-mcp/commit/dc1d9a2874938518c615de6064489ff0573d8447))


### Bug Fixes

* satisfy revive context parameter order ([c1fefe6](https://github.com/jonwraymond/metatools-mcp/commit/c1fefe6d9431a2cb0eff31c90e8f35a17c3764ec))

## [Unreleased]

### Added
- _TBD_

### Changed
- _TBD_

## [0.4.0](https://github.com/jonwraymond/metatools-mcp/compare/metatools-mcp-v0.3.0...metatools-mcp-v0.4.0) (2026-01-29)

### Features

- CLI foundation (Cobra) with serve/config/version commands.
- Koanf-based configuration loading with YAML/env support and defaults.
- Transport layer wiring (stdio + SSE) with config-driven selection.
- Provider registry with built-in MCP tools for search/docs/run/execute.
- Backend registry scaffolding with local backend support and adapters.
- Middleware chain scaffolding with logging/metrics middleware.

### Docs

- Added implementation notes for backend and middleware naming.

## [0.2.0](https://github.com/jonwraymond/metatools-mcp/compare/metatools-mcp-v0.1.12...metatools-mcp-v0.2.0) (2026-01-28)


### Features

* add env-configurable search strategy ([52ddf43](https://github.com/jonwraymond/metatools-mcp/commit/52ddf4369ed3c9f93ded8a3e419e9702ace80446))
* add MCP SDK wiring, CI, and examples ([6e1bd86](https://github.com/jonwraymond/metatools-mcp/commit/6e1bd86d62920785624b6fb73dabcb29424c10c6))
* add optional toolruntime-backed executor ([cab4fef](https://github.com/jonwraymond/metatools-mcp/commit/cab4fefd1d4291c34030a1ac6ff1a1db82462780))


### Bug Fixes

* correct release-please step id ([2e5453d](https://github.com/jonwraymond/metatools-mcp/commit/2e5453d98c7cd8b7cf63b56672bc8eccc41140a1))
* drop go mod download from CI ([946f80d](https://github.com/jonwraymond/metatools-mcp/commit/946f80dd8ac2ac239b825de26e02d1b33211459c))
* pin self module for tidy on CI ([4273eb7](https://github.com/jonwraymond/metatools-mcp/commit/4273eb7b2f539212f0f127618f3e65b70da25d7f))
* remove go list from CI ([8e79076](https://github.com/jonwraymond/metatools-mcp/commit/8e790767c9abcffd12fcbe4fe5d64bad6f07f922))
* remove tidy from CI ([fb75a34](https://github.com/jonwraymond/metatools-mcp/commit/fb75a34df18a815d96e382a62af0b4f414ca6182))
* run tests only in CI ([db42d74](https://github.com/jonwraymond/metatools-mcp/commit/db42d74ccebe79391a6d71316175ae63de0d55ec))
* simplify release-please token handling ([89cd72d](https://github.com/jonwraymond/metatools-mcp/commit/89cd72da2d1124116facf96cf694c76100d6fb91))
* track cmd and pkg metatools packages ([54998df](https://github.com/jonwraymond/metatools-mcp/commit/54998dffaf9ff6371b48829413a7f3133f10ebad))
* use app token for release-please ([1ea515e](https://github.com/jonwraymond/metatools-mcp/commit/1ea515e76747ee6aabe47609ecf2b24d39be5ba8))
* use PAT for release-please ([b115b61](https://github.com/jonwraymond/metatools-mcp/commit/b115b61d194c3e151c683a8ee69c364b1df9a359))
