# TinyColor Go — Port Mortem

An idiomatic, dependency-free Go port of the public TinyColor colour-value API
used for the Port Mortem hackathon. It parses CSS-like RGB/HSL/HSV values,
named colours and object forms, and exposes conversion, readability, mixing,
palette and modifying helpers.

## Build and test

```text
go test ./... -count=1
go vet ./...
go run ./cmd/differential ./fixtures/differential.json
```

The package builds with Go 1.26 and has no network or runtime dependency. The
original JavaScript oracle is not bundled. `original_adapter_test.go` is an
explicit 45-label Go-side parity checklist; the untouched oracle and its
kickoff hashes are retained privately in the project evidence, so no claim is
made that the JavaScript test suite was copied into this repository.

## Scope

- `color.go` — port implementation.
- `names.json` — named-colour data extracted from the upstream TinyColor oracle.
- `color_test.go` and `original_adapter_test.go` — deterministic fixtures and
  parity checklist.
- `cmd/differential` — standalone fixture runner for the port.
- `DECISIONS.md` — non-trivial implementation choices and known differences.

The implementation is intentionally CPU-only and deterministic for all
non-random helpers. `Random` follows the source API's random-colour contract.

## Attribution

The source API and named-colour data are based on
[TinyColor](https://github.com/bgrins/TinyColor), an MIT-licensed project. The
port is an independent Go implementation; see `NOTICE.md` and `LICENSE`.
