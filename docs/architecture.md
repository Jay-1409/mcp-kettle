# Architecture

MCP Kettel follows a short pipeline:

1. `cmd/mcp-kettel` validates the input and output paths.
2. `internal/scan` finds supported Python, JavaScript, and TypeScript source files.
3. `internal/scan/fastapi` and `internal/scan/express` parse supported route declarations and produce normalized candidates.
4. `internal/scan/routegraph` resolves mounted routers and accumulated route prefixes.
5. `internal/tui` lets the user filter, group, and select candidates.
6. `internal/generate` writes a Go MCP server for the selected candidates.

The scanner uses Tree-sitter and never imports or executes the host project. The generated server exposes one MCP tool per selected route and forwards requests to `MCP_API_BASE_URL` over HTTP.

FastAPI scanning builds a graph of `FastAPI` and `APIRouter` declarations. Import and `include_router()` edges connect the nodes. Traversal starts at each FastAPI application, accumulates router and mount prefixes, and prevents recursive include cycles. Routers that cannot be reached from an application remain discoverable with their local prefix.

The current scanner skips Pydantic request bodies, dependency injection, computed routes, decorator aliases, and unsupported parameter types. The output directory must not already exist.

The current workflow is synchronous and single-process. No background workers or shared service state are required. There are no architecture diagrams because the pipeline is linear and is fully described above.

The TUI can group routes by source file, HTTP method, or first route segment. Groups start collapsed and expand with `space`; expanded route rows retain the normal individual selection behavior.
