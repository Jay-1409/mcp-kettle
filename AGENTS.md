# Agent instructions

## Project

MCP Kettel is a Go CLI that statically scans local FastAPI source, lets users select discovered routes, and generates a Go MCP server.

## Before changing code

- Read the relevant package and its tests.
- Preserve unrelated working-tree changes.
- Keep the change focused. Avoid new abstractions or dependencies unless the existing code cannot support the requirement.

## Structure

- `cmd/mcp-kettel` — CLI entry point and workflow
- `internal/scan` — source-file discovery
- `internal/scan/fastapi` — FastAPI static scanning
- `internal/scan/express` — Express static scanning
- `internal/model` — normalized candidate types
- `internal/tui` — interactive selection
- `internal/generate` — generated Go MCP server

## Validation

Run the full test suite after code changes:

```sh
go test ./...
```

Add or update a focused test for non-trivial behavior. Keep generated output directories, credentials, and local tool artifacts out of commits.

## Scanner boundaries

The scanner must not import or execute host-project code. Unsupported FastAPI constructs should be skipped safely rather than guessed. Preserve the normalized candidate model as the boundary between scanning, selection, and generation.

## Documentation

Update `README.md` and the relevant file under `docs/` when user-visible behavior changes. Keep README links relative and verify linked files exist.
