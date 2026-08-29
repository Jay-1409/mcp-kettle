# MCP Kettel — Initial Requirements

Status: first usable slice in implementation
Recorded: 2026-08-29

## Product goal

Build a CLI that turns selected APIs/functions from an existing codebase into an MCP server.

## Required workflow

1. The user runs the CLI with a host-directory path.
2. The CLI recursively scans that directory and discovers functions and API endpoints that could become MCP tools.
3. The CLI shows the candidates in a polished interactive terminal checklist.
4. Every candidate is checked by default.
5. The user can:
   - check or uncheck individual candidates;
   - check all candidates;
   - uncheck all candidates;
   - filter/search the candidate list.
6. The CLI generates MCP tooling for only the selected candidates.

## Generation

- AI must be optional; deterministic generation is acceptable and preferred where source metadata is sufficient.
- If AI is used, the user supplies their own provider/model credentials.
- This project supplies the prompts used for generation.
- No bundled AI account or mandatory hosted AI service.

## Current product boundary

- Input: a local host directory.
- Primary output: a generated MCP server exposing the selected APIs/functions as MCP tools.
- Selection happens before generation.
- CLI implementation: Go.
- Terminal UI: Bubble Tea v2 with the Bubbles v2 list component and Lip Gloss v2 styling.
- First supported source ecosystem: Python with FastAPI.
- FastAPI source is parsed statically from Go using the official Tree-sitter Go bindings and Python grammar; the host project is not imported or executed.
- First generated runtime: a Go MCP server using the official MCP Go SDK whose tools proxy the existing HTTP API.
- First release generation is deterministic; AI is deferred and remains optional.
- The normalized candidate model is the stable boundary between discovery, selection, optional AI enrichment, and generation.
- Deterministic generation must not depend on an AI provider or SDK; future AI integrations may produce reviewable metadata suggestions but cannot silently change routes, schemas, authentication, or request wiring.
- New specification and framework scanners should be added through a small shared scanner contract and explicit registration.
- Additional frameworks, direct function imports, transport choices, credential mapping details, and regeneration behavior remain later decisions.

## Acceptance criteria for the first usable slice

- A directory can be passed from the command line.
- At least one supported language/framework can be scanned without executing untrusted project code.
- Discovered endpoints/functions include enough context to identify them (name, method/type, route when applicable, and source file).
- The interactive list starts fully selected and supports individual toggle, select all, clear all, and text filtering.
- Cancelling does not write generated files.
- Confirming generates a runnable MCP server containing only the selected tools.
- AI credentials are not required for the non-AI path and are never written into generated source.

## Web research: existing products

Research date: 2026-08-29.

### Closest matches

- [github-to-mcp](https://github.com/nirholas/github-to-mcp) analyzes GitHub repositories, extracts tools from source code and API schemas, and generates MCP servers. Its documented input is a GitHub URL; its CLI does not document the requested local-directory, searchable checkbox-selection flow.
- [mcp-new](https://github.com/d1maash/mcp-new) is an interactive MCP project generator. It imports OpenAPI specs, lets users select endpoints, and optionally generates tools using a user-provided Anthropic key. It does not document arbitrary local source-directory scanning.
- [CodeSpar MCP Generator](https://codespar-docs-eight.vercel.app/docs/mcp-generator) scans API source code and generates an MCP server, including a dashboard endpoint picker. It is documented as part of CodeSpar Enterprise rather than a small local-first open-source CLI.
- [cli2mcp](https://github.com/bcmcpher/cli2mcp) scans configured Python source directories and converts Click, Typer, or argparse commands to MCP tools. It is narrower than general API/function discovery and documents list/generate commands rather than an interactive checklist.

### Adjacent tools

- [openapi-mcp-generator](https://github.com/harsha-iiiv/openapi-mcp-generator) generates TypeScript MCP servers from OpenAPI specifications.
- [FastMCP OpenAPI integration](https://gofastmcp.com/integrations/openapi) converts OpenAPI routes to MCP components and supports rule-based exclusion, but it is a library workflow rather than the requested source-scanning terminal product.
- [Code2MCP](https://arxiv.org/abs/2509.05941) describes an AI multi-agent approach for converting GitHub repositories into MCP services; it is heavier and does not target the requested human-selection CLI flow.

## Research conclusion

The broad concept already exists. The defensible initial difference is the combined workflow: **local directory -> static discovery -> searchable all-selected terminal checklist -> deterministic MCP generation, with optional bring-your-own AI assistance**.
