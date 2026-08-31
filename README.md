# MCP Kettel

> *Turn routes from APIs into an MCP server.*

[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)
[![Version: 0.1.0](https://img.shields.io/badge/version-0.1.0-blue.svg)](./go.mod)
[![Language: Go](https://img.shields.io/badge/language-Go-00ADD8.svg)](https://go.dev/)

![MCP Kettel interactive route selection](./docs/images/kettel-demo.gif)

## What's in it for you

- Find API routes without running or importing the source project.
- Review discovered routes before exposing anything.
- Search, select, or clear routes from an interactive terminal checklist.
- Generate a standalone Go MCP server with only the routes you chose.
- Keep credentials out of generated source.

## Features

- Static route discovery across supported backend frameworks.
- Recursive router, mount, and prefix resolution.
- Visibility for routes whose parameter schemas are not ready for generation.
- Support for `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, and `HEAD`.
- Primitive string, integer, float, and boolean path and query parameters.
- Deterministic MCP server generation.
- Optional upstream `Authorization` header forwarding.
- Safe cancellation and refusal to overwrite an existing output directory.

## Quick Start

Install Go 1.24 or newer, then run Kettel against a supported API project:

```sh
go run ./cmd/mcp-kettel \
  --input /path/to/api-project \
  --output ./generated-mcp
```

Try the included example:

```sh
go run ./cmd/mcp-kettel \
  --input ./example/sample-project \
  --output ./generated-mcp
```

The checklist starts with every ready route selected. Routes marked with `×` were discovered but need parameter-schema support before generation. Use `g` to cycle grouping by source file, HTTP method, or route prefix. Groups start collapsed; highlight a group and press `space` to expand or collapse it. Once expanded, use `space` on a ready route to toggle one, `a` to select all ready routes, `n` to clear all, `/` to filter, `enter` to generate, or `esc` to cancel.

| Group large route sets | Filter matching APIs |
| --- | --- |
| ![Routes grouped by source file](./docs/images/kettel-grouping.png) | ![Routes filtered by the word users](./docs/images/kettel-filter.png) |

Run the generated server beside the running source application:

```sh
cd generated-mcp
MCP_API_BASE_URL=http://127.0.0.1:8000 go run .
```

For an authenticated API, set the complete header value:

```sh
MCP_API_AUTHORIZATION="Bearer $TOKEN" \
MCP_API_BASE_URL=http://127.0.0.1:8000 \
go run .
```

If `--output` is omitted, Kettel writes to `<input>/mcp-server`.

Kettel initially supports FastAPI and Express. Contributions that add more backend frameworks are welcome. See [Contributing](./docs/contributing.md).

## Documentation

- [Architecture and data flow](./docs/architecture.md)
- [Contributing](./docs/contributing.md)

## License

Apache License 2.0. See [LICENSE](./LICENSE).
