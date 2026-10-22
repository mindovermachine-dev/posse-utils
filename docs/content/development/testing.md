---
title: Testing
description: How to run unit tests, coverage and lint, plus the Detroit-school testing conventions used in this repository.
sidebar:
  order: 3
---

# Testing

## Commands

| Command | What it does |
| --- | --- |
| `make test` | Unit tests with `-race`, `-shuffle=on` and atomic coverage; writes `coverage.out` |
| `make cover` | `make test`, then prints coverage per function |
| `make cover-html` | `make test`, then opens an HTML report |
| `make cover-check` | `make test`, then fails if total coverage is below `MIN_COV` (default 80) |
| `make lint` | `golangci-lint run` |
| `make fmt` | `golangci-lint fmt` (gofmt + goimports) |
| `make vet` | `go vet ./...` |
| `make check` | `vet` + `lint` + `cover-check`; run this before every push |

Run a single package or test with plain Go:

```sh
go test ./internal/newsletter/...
go test ./internal/newsletter -run TestSubscribeThenConfirm -v
```

Change the threshold for one run: `make cover-check MIN_COV=90`.

## What is excluded from the coverage gate

`make test` removes two files from the coverage profile before the threshold is
checked (see `COVER_EXCLUDE` in the [Makefile](../Makefile)):

- `cmd/server/` is process wiring (flags, signals, listeners).
- `internal/store/firestore.go` is a thin adapter over a network client.

Both are deliberately thin. If either grows logic, move the logic into a testable
package instead of testing it through the network.

## Test style: Detroit school (classicist)

- **Test behaviour through the public API**, with real collaborators. Tests live in
  external `_test` packages (for example `newsletter_test`), so they can only use
  what callers can use.
- **Replace only I/O at the edges** with simple in-memory fakes, not mocks:
  - `store.Memory` for persistence
  - `mail.Outbox` records sent messages for assertions
  - a controllable clock (`Now func() time.Time`) for expiry tests
- **Assert on outcomes** (state in the store, messages in the outbox, HTTP status
  and `Location`), not on which methods were called.
- **Table-driven tests** for validation and rejection cases.
- The one internal test, `mail_internal_test.go`, checks message formatting that
  cannot be reached from the outside without a network.

## Writing a new test

1. Build a fixture with real objects. See `newFixture` in
   [service_test.go](../internal/newsletter/service_test.go).
2. Drive it through exported functions or HTTP handlers.
3. Assert on the store, outbox or response.
4. For anything time-dependent, advance the fixture's clock instead of sleeping.

When a new collaborator performs I/O, add an interface at its boundary plus an
in-memory fake next to the real implementation, as `store.Memory` and
`mail.Outbox` do.

## Linting

Configuration is in [.golangci.yml](../.golangci.yml) (golangci-lint v2). It
enables the `standard` set plus `bodyclose`, `errorlint`, `gocritic`, `gosec`,
`misspell`, `noctx`, `nolintlint`, `revive`, `unconvert` and `unparam`. `gosec` and
`unparam` are relaxed in tests.

Any `//nolint` needs a specific linter name and a reason.

Tests that build HTTP requests use `httptest.NewRequestWithContext`, as required
by `noctx`.

## Continuous integration

There is no CI workflow yet. A minimal one runs `make check` on every pull
request. See [Contributing](contributing.md).
