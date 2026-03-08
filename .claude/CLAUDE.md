# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

- `Spec(path, ...LoaderOption) (*Document, error)` — main entry point, auto-detects JSON/YAML
- `JSONSpec(path, ...LoaderOption) (*Document, error)` — explicit JSON loading
- `Analyzed(data, version, ...LoaderOption) (*Document, error)` — from raw JSON bytes
- `Embedded(orig, flat, ...LoaderOption) (*Document, error)` — from pre-parsed specs
- `Document.Expanded() (*Document, error)` — resolves all `$ref` references
- `Document.Pristine() *Document` — deep clone via gob round-trip
- `AddLoader(DocMatcher, DocLoader)` — register custom loader at package level (not thread-safe)

### Dependencies

- `github.com/go-openapi/analysis` — spec analysis
- `github.com/go-openapi/spec` — Swagger v2.0 types
- `github.com/go-openapi/swag/loading` — HTTP/file loading
- `github.com/go-openapi/swag/yamlutils` — YAML conversion
- `github.com/go-openapi/testify/v2` — test-only assertions (zero-dep testify fork)

### Notable historical design decisions

- **Loader chain pattern**: linked list of `DocLoaderWithMatch` nodes; YAML matcher checked first, JSON loader is the fallback (matches any path). Extensible via `AddLoader()` or per-call `LoaderOption`.
- **Global `spec.PathLoader` bridge**: the `spec` package's `PathLoader` function pointer is set to this package's loader, enabling cross-package `$ref` resolution.
- **Deep cloning via gob**: `Pristine()` uses `encoding/gob` round-trip to deep-copy the full `Document`, preserving all nested structures.
- **Separate `origSpec`**: `Document` keeps an untouched copy of the original spec alongside the working copy, so expansion/mutation is non-destructive.
