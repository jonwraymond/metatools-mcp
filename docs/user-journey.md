# User Journey

This journey shows the full end-to-end agent workflow via MCP metatools.

## End-to-end flow (agent view)

![Diagram](assets/diagrams/user-journey.svg)

## Step-by-step

1. **Discover tools** with `search_tools` (summary-only results).
2. **Inspect schema** using `describe_tool` (schema or full detail).
3. **Execute** a single tool with `run_tool` or a sequence with `run_chain`.
4. **Orchestrate** complex flows using `execute_code` (optional).

## Example: full agent workflow

```text
1) search_tools("create issue", limit=5)
2) describe_tool("github:create_issue", detail_level="schema")
3) run_tool("github:create_issue", args={...})
4) run_chain([{tool_id:"github:get_issue"}, {tool_id:"github:add_label", use_previous:true}])
```

## Expected outcomes

- Stable MCP-compatible APIs for discovery, documentation, and execution.
- Consistent error objects for tool failures.
- Progressive disclosure to minimize token costs.

## Common failure modes

- Invalid input payloads (handler validation errors).
- Tool-level errors returned in `ErrorObject` with `code` and `op` fields.
- Unsupported options (e.g., `stream=true` for `run_tool`).
