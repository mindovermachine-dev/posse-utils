---
title: Contributing
description: Workflow, conventions and checklist for making a change to posse-utils.
---

## Workflow

1. Open the project in the [devcontainer](getting-started.md).
2. Create a branch. With the GitHub CLI extension installed by the devcontainer,
   `gh workon` starts work on an issue (it wraps `gh tt workon`). Plain
   `git switch -c` works too.
3. Make a small, focused change, with tests.
4. Run `make check`. It must pass: zero lint issues and coverage at or above 80%.
5. Open a pull request. Describe the behaviour change and how you verified it.

`gh wrapup` and `gh deliver` are the matching aliases to finish and deliver the
work; they come from the TakT extension (`devx-cafe/gh-tt`), configured in
[`.devcontainer/.gh_alias.yml`](../.devcontainer/.gh_alias.yml).

## Conventions

- **Standard Go**: `gofmt`/`goimports`, idiomatic error handling, `internal/` for
  everything that is not a command.
- **Errors**: wrap with `%w` and compare with `errors.Is`. Sentinel errors are
  exported where callers need to branch on them.
- **Boundaries are interfaces** (`store.Store`, `mail.Sender`); implementations are
  concrete types. Keep interfaces small and defined next to their consumer's needs.
- **No global state.** Pass dependencies in through constructors.
- **Time is injected** (`Now func() time.Time`), never read directly in logic that
  needs testing.
- **Comments** explain why, not what, in one short line.
- **Dependencies**: prefer the standard library. Add a module only when it earns
  its keep, then run `make tidy`.
- **Logging** uses `log/slog`. Never log tokens, passwords or full email bodies in
  production paths. (`LogSender` logs bodies, and exists only for local use.)

## Adding a feature

New capabilities live in their own package under `internal/`, next to
`newsletter`. Reuse `store`, `mail` and `token`. Planned modules:

- `events` for free event signups with capacity, duplicate protection through a
  hashed document ID, and a waitlist
- `syndicate` to post new content to Bluesky and Mastodon

For each one:

1. Write the service with its behaviour tests first, using the in-memory fakes.
2. Add an HTTP handler with `Register(mux)` as in `newsletter`.
3. Wire it in `cmd/server/main.go` and add any config to
   `internal/config`, with tests.
4. Document the endpoints in [API reference](api.md) and variables in
   [Configuration](configuration.md).

## Pull request checklist

- [ ] `make check` passes
- [ ] New behaviour has tests; fixed bugs have a regression test
- [ ] No secrets, tokens or real email addresses in code, tests or docs
- [ ] `docs/` updated if endpoints, config or workflows changed
- [ ] `go.mod`/`go.sum` tidy (`make tidy` leaves no diff)

## Security issues

Do not open a public issue for a vulnerability. Contact the maintainer privately.
