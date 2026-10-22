# Contributing

ownframe keeps one Go process and one HTML template. Bug reports, docs fixes, examples, and library changes are welcome. The map of the codebase is in [AGENTS.md](AGENTS.md), and the feature index is [documentation/features.md](documentation/features.md).

## Setup

The repo holds two Go modules joined by `go.work`: the library at the root and `examples/`. Use the Go version in `go.mod` (1.26.10 or newer).

- `make test` runs the test suite for both modules.
- `make test-race` runs `go test -race -p 1 ./...` on the library module only.
- `make build` is the compile check. It runs `go vet` over every package. Do not run `go build ./...`; it links an executable for each example, takes about a minute, and leaves the binaries in this directory.
- `make lint` runs golangci-lint, the documentation and plan checks, the example internal-import gate, and the website lint.
- Run `gofmt` on every Go file you edit.
- Group related methods, test cases, and fixtures by responsibility. There is no character cap.

## Review cohesion

A review answers these questions with the changed code:

- Can a reader follow an operation without opening unrelated files?
- Does each package own a clear responsibility, and are its dependencies explicit?
- Do tests keep related scenarios and fixtures together?
- Does a helper hide meaningful behavior or merely add another jump?

Use `go run scripts/merge-go.go destination.go source.go source.go` to preview a same-package consolidation, and add `-write` before the destination to apply it. The tool preserves declaration text and comments and refuses build constraints, platform suffixes, cgo, or conflicting import aliases. Keep platform-specific files separate and run the affected tests after a merge.

Split code when responsibilities diverge. Consolidate files that were split only for a size count. Do not add synonym helpers, comment-only Go files, or detached constant files solely to satisfy lint.

`make docs-check` compiles the Go snippets extracted from the guides against the public API. `make plans-check` checks the current status ledger. `make graphics-check` runs window pixel checks and replay parity in isolated Xvfb sessions. See [testing.md](documentation/testing.md) for scope and prerequisites.

The maintained teaching examples are listed in `examples/maintenance.json`. Other examples are experiments and remain in the compatibility suite; new library features need a focused example rather than another app clone.

## Public API after the Page wrap

The wrap has landed: `ownframe.Page` is a struct with an unexported inner page. v0 may still break. Keep the public API small:

- No new `type X =` alias at the module root except documented re-exports.
- No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows.
- No new example import of `github.com/chinmay-sawant/ownframe/internal/`. `make internal-import-check` records the current set.
- No synonym names (`PreviousX` for `X`).
- New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR.
- Panic only for broken invariants. New user-facing failure paths return `error`.
- `context.Context` is the first parameter when the function already does I/O, cancel, or deadline. Do not add a hidden `time.After` that ignores the caller's ctx.

The longer list is [plans/v0.0.3/prod-ready-task/field-go-practices.md](plans/v0.0.3/prod-ready-task/field-go-practices.md).

## Tagging the library and the examples module

`examples/go.mod` `replace`s `github.com/chinmay-sawant/ownframe` with `../` so a `GOWORK=off` build in this checkout uses the parent, not the published tag. Keep the blinkless version in `examples/go.mod` equal to the root `go.mod`.

Two-step tag:

1. Tag the library `vX.Y.Z`.
2. Bump the examples require to that tag, drop the relative replace, and tag `examples/vX.Y.Z`.

A tagged examples module on the proxy cannot keep `replace => ../`. Keep the replace on the development branch after the tag rewrite.

## Features and demos

A new feature, or a new demo of one (something like `examples/controls` or `examples/bind`), carries three things:

1. A runnable example under `examples/` that exercises the feature. Make it detailed enough that a reader can see the feature work without reading the library. Most examples accept `-web`; keep that path working when the feature allows it.
2. A guide under `documentation/`. Add a page for a new feature, or extend the existing page.
3. Index updates: a row in [documentation/README.md](documentation/README.md), a section in [documentation/features.md](documentation/features.md), and a row in [examples/readme.md](examples/readme.md).

When the feature changes behavior that an existing guide describes, edit that guide in the same change.

Do not add a new example whose primary purpose is resembling Telegram, Teams, Spotify, Wispr, Netflix, or another product clone. Feature demos stay small (`examples/bind`, `examples/forms`). Existing clones stay; they are not a template for new apps.

## Playground examples

An example you add just to try something can go in directly. It needs no guide and no index row, as long as `make test` passes and it does not change library behavior. When it grows into a feature demo, follow the section above.

## Pull requests

Open a PR with the template in [skills/PR/PR_TEMPLATE.md](skills/PR/PR_TEMPLATE.md): summary, changes, impact, test plan, and related issues.

- Title: `<type>(<scope>): <short imperative description>`, with a type from `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `chore`, `ci`.
- Labels: at least one. `feat` takes `enhancement`, `fix` takes `bug` or `enhancement`, `docs` takes `documentation`, and `perf`, `refactor`, and `chore` take `enhancement`.
- Self-assign the PR.
- Link issues with `Closes #N` when the PR completes one, and `Relates to #N` for partial work.
- Run `make test` before you open it. Add `make build` when the change touches more than docs. Add `make test-race` when the change touches shared mutable state.

## Bugs

Report a bug with the issue template in [skills/PR/ISSUE_TEMPLATE.md](skills/PR/ISSUE_TEMPLATE.md) and the `bug` label. Include what you did, what you expected, what happened, your Go version, and the platform (desktop, `-web`, wasm, or mobile). A small HTML template that shows the problem is the most useful attachment.

## AI-assisted code guardrails

Parts of this repo are AI-assisted, so two failure modes need manual
review on every PR: hallucinated or typosquatted module names, and
plausible-but-wrong security code.

- Every new module (Go or npm) needs manual provenance review before it
  merges: real owner, `pkg.go.dev` or registry page, stars and history,
  and no lookalike name. The dependency-review workflow and
  `go-licenses check` catch CVEs and licenses; they do not catch a
  convincing fake, so this step stays human.
- Every security-relevant change (auth, fetch, paths, crypto, IPC,
  clipboard, permissions, build flags) carries a regression test that
  fails without the fix. Name it after the finding when one exists.
- Do not paste secrets, signing keys, or customer data into prompts or
  fixtures. The gitleaks pre-commit hook and CI job scan for them; the
  `secret`/`secret` demo credentials in `examples/login` are the only
  sanctioned placeholder.

For a separate maintainer assessment, use the [review prompt and handoff exercise](documentation/maintenance-review.md). Record what the new maintainer actually ran.
