# Contributing

## Set up

Install Go 1.24 or newer, then run the test suite:

```sh
go test ./...
```

Run the CLI against the sample project while developing:

```sh
go run ./cmd/mcp-kettel --input ./example/sample-project --output ./generated-mcp
```

## Branches

Create focused branches from the default branch. Use a descriptive name such as `codex/add-route-filter`.

## Pull requests

- Keep changes focused on one behavior.
- Add or update a small test for non-trivial logic.
- Run `go test ./...` before opening a pull request.
- Do not commit generated output directories or credentials.
