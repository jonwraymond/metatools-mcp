# Changelog

## Unreleased

### Added
- Cursor-based pagination for `search_tools` and `list_namespaces` with opaque cursors.
- `notifications/tools/list_changed` support with debounce and env toggle.
- Tests for list-changed notifications and cancellation propagation.
- Progress notifications for `run_tool`, `run_chain`, and `execute_code` when a progress token is provided.
- Cancellation and timeout error codes (`cancelled`, `timeout`) for tool failures.

### Changed
- `list_namespaces` input/output schemas now include `limit`, `cursor`, and `nextCursor`.
- Cursor helpers marked deprecated in favor of toolindex tokens.
