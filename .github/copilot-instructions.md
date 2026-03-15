# Copilot Instructions for go-openapi/loads

## Project Overview

`github.com/go-openapi/loads` loads, parses, and analyzes Swagger/OpenAPI v2.0 specifications from local files or remote URLs in JSON and YAML formats. It is part of the `go-openapi` ecosystem.

See [docs/MAINTAINERS.md](../docs/MAINTAINERS.md) for CI/CD, release process, and repo structure details.

### Package layout (single package)

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `spec.go` | `Document` type; main entry points: `Spec`, `JSONSpec`, `Analyzed`, `Embedded` |
| `loaders.go` | Loader chain (linked list of `DocLoaderWithMatch`); `JSONDoc`, `AddLoader` |
| `options.go` | `LoaderOption` functional options (`WithDocLoader`, `WithDocLoaderMatches`, `WithLoadingOptions`) |
| `errors.go` | Sentinel errors: `ErrLoads`, `ErrNoLoader` |
| `fmts/yaml.go` | Re-exports YAML utilities from `swag` (`YAMLMatcher`, `YAMLDoc`, `YAMLToJSON`, `BytesToYAMLDoc`) |

### Key API

- `Spec(path, ...LoaderOption) (*Document, error)` --- main entry point, auto-detects JSON/YAML
- `JSONSpec(path, ...LoaderOption) (*Document, error)` --- explicit JSON loading
- `Analyzed(data, version, ...LoaderOption) (*Document, error)` --- from raw JSON bytes
- `Embedded(orig, flat, ...LoaderOption) (*Document, error)` --- from pre-parsed specs
- `Document.Expanded() (*Document, error)` --- resolves all `$ref` references
- `Document.Pristine() *Document` --- deep clone via gob round-trip
- `AddLoader(DocMatcher, DocLoader)` --- register custom loader at package level (not thread-safe)

### Dependencies

- `github.com/go-openapi/analysis` --- spec analysis
- `github.com/go-openapi/spec` --- Swagger v2.0 types
- `github.com/go-openapi/swag/loading` --- HTTP/file loading
- `github.com/go-openapi/swag/yamlutils` --- YAML conversion
- `github.com/go-openapi/testify/v2` --- test-only assertions (zero-dep testify fork)

## Building and Testing

Single module --- no workspace file.

```sh
go test ./...
```

## Conventions

Coding conventions are found beneath `.github/copilot`

### Summary

- All `.go` files must have SPDX license headers (Apache-2.0).
- Commits require DCO sign-off (`git commit -s`).
- Linting: `golangci-lint run` --- config in `.golangci.yml` (posture: `default: all` with explicit disables).
- Every `//nolint` directive **must** have an inline comment explaining why.
- Tests: `go test ./...`. CI runs on `{ubuntu, macos, windows} x {stable, oldstable}` with `-race`.
- Test framework: `github.com/go-openapi/testify/v2` (not `stretchr/testify`; `testifylint` does not work).

See `.github/copilot/` (symlinked to `.claude/rules/`) for detailed rules on Go conventions, linting, testing, and contributions.
