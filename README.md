# MCP Kettel

MCP Kettel scans a local FastAPI project without importing or executing it, lets you choose endpoints in an interactive terminal list, and generates a Go MCP server that proxies the selected HTTP APIs.

## Run

```sh
go run ./cmd/mcp-kettel --input /path/to/fastapi-project --output ./generated-mcp
```

In the selector:

- `space` toggles an endpoint;
- `a` selects all and `n` selects none;
- `/` filters the list;
- `enter` generates the server;
- `esc` cancels without writing files.

Run the generated server:

```sh
cd generated-mcp
MCP_API_BASE_URL=http://127.0.0.1:8000 go run .
```

Set `MCP_API_AUTHORIZATION` to a complete Authorization header value when the upstream API requires one.

## Current slice

The scanner supports literal `@app.get(...)`, `@router.post(...)`, and other standard FastAPI method decorators with primitive `str`, `int`, `float`, and `bool` path/query parameters. Pydantic request bodies, dependency injection, computed routes, and decorator aliases are deliberately skipped until they can be represented safely.
