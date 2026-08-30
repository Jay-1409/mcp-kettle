# Architecture

MCP Kettel follows a short pipeline:

1. `cmd/mcp-kettel` validates the input and output paths.
2. `internal/scan` finds supported Python, JavaScript, and TypeScript source files.
3. `internal/scan/fastapi` and `internal/scan/express` parse supported route declarations and produce normalized candidates.
4. `internal/scan/routegraph` resolves mounted routers and accumulated route prefixes.
5. `internal/tui` lets the user filter, group, and select candidates.
6. `internal/generate` writes a Go MCP server for the selected candidates.

The FastAPI scanner uses Tree-sitter. The scanners never import or execute the host project. The generated server exposes one MCP tool per selected route and forwards requests to `MCP_API_BASE_URL` over HTTP.

FastAPI scanning builds a graph of `FastAPI` and `APIRouter` declarations. Import and `include_router()` edges connect the nodes. Traversal starts at each FastAPI application, accumulates router and mount prefixes, and prevents recursive include cycles. Routers that cannot be reached from an application remain discoverable with their local prefix.

Express scanning uses the same route graph and depth-first traversal. Relative ES module and CommonJS imports connect `express.Router()` declarations through literal `.use()` mounts. The traversal accumulates nested mount prefixes and prevents include cycles.

The FastAPI graph uses depth-first traversal to resolve nested router mounts and prefixes. Literal route decorators remain visible even when their Pydantic request bodies, dependency injection, or parameter types are not ready for generation. The TUI marks those routes unavailable. Computed routes and decorator aliases remain unsupported. The output directory must not already exist.

The current workflow is synchronous and single-process. No background workers or shared service state are required. There are no architecture diagrams because the pipeline is linear and is fully described above.

The TUI can group routes by source file, HTTP method, or first route segment. Groups start collapsed and expand with `space`; expanded route rows retain the normal individual selection behavior.
