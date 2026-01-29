# Changelog

## [Unreleased]

### Added
- Cursor-based pagination for `search_tools` and `list_namespaces` with opaque cursors.
- `notifications/tools/list_changed` support with debounce and env toggle.
- Tests for list-changed notifications and cancellation propagation.

### Changed
- `list_namespaces` input/output schemas now include `limit`, `cursor`, and `nextCursor`.
- Cursor helpers marked deprecated in favor of toolindex tokens.

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
