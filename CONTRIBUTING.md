# Contributing

ownframe keeps one Go process and one HTML template. Bug reports, docs fixes, examples, and library changes are welcome. The map of the codebase is in [AGENTS.md](AGENTS.md), and the feature index is [documentation/features.md](documentation/features.md).

## Setup

The repo holds two Go modules joined by `go.work`: the library at the root and `examples/`. Use the Go version in `go.mod` (1.26.4 or newer).

- `make test` runs the test suite for both modules.
- `make build` is the compile check. It runs `go vet` over every package. Do not run `go build ./...`; it links an executable for each example, takes about a minute, and leaves the binaries in this directory.
- Run `gofmt` on every Go file you edit.
- Every Go file stays at most 2000 characters, counted with `wc -m`. Split an overlong file inside the same package and keep the behavior.

## Features and demos

A new feature, or a new demo of one (something like `examples/controls` or `examples/bind`), carries three things:

1. A runnable example under `examples/` that exercises the feature. Make it detailed enough that a reader can see the feature work without reading the library. Most examples accept `-web`; keep that path working when the feature allows it.
2. A guide under `documentation/`. Add a page for a new feature, or extend the existing page.
3. Index updates: a row in [documentation/README.md](documentation/README.md), a section in [documentation/features.md](documentation/features.md), and a row in [examples/readme.md](examples/readme.md).

When the feature changes behavior that an existing guide describes, edit that guide in the same change.

## Playground examples

An example you add just to try something can go in directly. It needs no guide and no index row, as long as `make test` passes and it does not change library behavior. When it grows into a feature demo, follow the section above.

## Pull requests

Open a PR with the template in [skills/PR/PR_TEMPLATE.md](skills/PR/PR_TEMPLATE.md): summary, changes, impact, test plan, and related issues.

- Title: `<type>(<scope>): <short imperative description>`, with a type from `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `chore`, `ci`.
- Labels: at least one. `feat` takes `enhancement`, `fix` takes `bug` or `enhancement`, `docs` takes `documentation`, and `perf`, `refactor`, and `chore` take `enhancement`.
- Self-assign the PR.
- Link issues with `Closes #N` when the PR completes one, and `Relates to #N` for partial work.
- Run `make test` before you open it. Add `make build` when the change touches more than docs.

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
