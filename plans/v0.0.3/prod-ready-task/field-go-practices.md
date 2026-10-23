# v0.0.3 - Field: Go practices (all phases)

> Superseded planning record. `[s]` marks an old proposal or assertion, not proof of implementation. `[-]` marks work explicitly deferred in the original plan. Current decisions and evidence live in [plans/README.md](../../README.md).

> **Parent:** [README.md](README.md)
> **Estimated effort:** review time. Land the freeze sentences in Phase 0. Re-check at 1.0 closure.

---

## Overview

Production Go is a small exported API, `error` values, tests with `-race` and fuzz at the edges, no `init()` clients, `internal/` for implementation, and versions that do not fork. This file is the positive list and the hard guardrails. Phase files cite it; they do not copy it.

Aligns with repository Go style: `gofmt`, 2000-char files until [phase-01-package-shape.md](phase-01-package-shape.md) changes the gate, `make lint` / `make test` on non-docs PRs.

## Practices every 1.0 PR follows

### API

- [s] `GO-01` unexport until needed. New identifiers start unexported unless an example outside `internal/` requires them. Proof per PR: `git grep '^func [A-Z]'` on the diff is justified in the PR body.
- [s] `GO-02` defined public `Page` type after `API-01`. No new `type X = internalY` at root except `Config`/`Handlers`/`Box` if those stay aliases by written decision.
- [s] `GO-03` one exported name per idea. No `Foo` plus `FooAlias`. `BLEED-04` gate.
- [s] `GO-04` exported methods return named types in `ownframe` or a public subpackage. Never `host.Stats` or `*state.Store` unless re-exported in the same PR (`API-23`, `API-24`).
- [s] `GO-05` `internal/` is implementation. New example and `ownframe_test` code import public paths only (`BLEED-06`, `TEST-06`).

### Errors and control flow

- [s] `GO-06` operations that fail return `error`. Predicates return `bool`. `PatchElement` is `error` (`ERR-01`).
- [s] `GO-07` wrap with `%w`. Sentinels at package `ownframe` for cases apps switch on. Callers use `errors.Is` / `errors.As`.
- [s] `GO-08` panic only for broken invariants. `Run`/`BindMobile` recover stays. No `log.Fatal` in library packages (`ERR-16`).
- [s] `GO-09` early return on errors. No nested `else` after `return`. (`gofmt` does not catch this; reviewers do.)
- [s] `GO-10` `context.Context` first on I/O, cancel, and deadline. Honor the caller's ctx. No hidden `time.After` that ignores it.

### Constructors and state

- [s] `GO-11` `New` / `NewWithOptions` / `Option` / `WithPerf`. No second builder. Options that can fail error at construct (`ERR-10`, `ERR-11`).
- [s] `GO-12` no new `init()`. Test seams are explicit (`UseMemory`, `InstallPicker`).
- [s] `GO-13` enums and modes: zero value is unset/unknown where a silent zero would be wrong. Do not add `iota` that makes zero a live mode without a comment.
- [s] `GO-14` inject OS seams. Tests do not touch the desktop clipboard or real zenity.

### Packages and files

- [s] `GO-15` split by package when a type is too big, not by synonym file. `SHAPE-11+` after the wrap.
- [s] `GO-16` no comment-only `.go` files and no `mnd`-only const files (`SHAPE-03`).
- [s] `GO-17` `gofmt` on every edited Go file. `make lint` includes size-check until the script changes.
- [s] `GO-18` composite literals with field names when the struct can grow (`Config{HTML: ..., Width: ...}` already the example style).

### Tests and security

- [s] `GO-19` `go test -race` on the library in CI (`TEST-02`).
- [s] `GO-20` fuzz at trust boundaries: crash dir, fetch, DOM/image, web sanitize (`TEST-08`, `TEST-09`).
- [s] `GO-21` `govulncheck` stays (`MOD-08`). Do not replace it with a 50-item markdown audit.
- [s] `GO-22` `gosec` stays in `.golangci.yml`. New `exec.Command` uses separate args. New path writes go through existing crash/filepick validators.
- [s] `GO-23` Fetch SSRF block stays: no cookies, no cache, private addresses denied, https→http is `ErrDowngrade`.
- [s] `GO-24` `Drop.Read` in the callback; do not log `Drop.Path` (`OS-20`).
- [s] `GO-25` `html/template` stays the template engine (auto-escape). Do not switch user HTML to `text/template` for "ease."

### Modules

- [s] `GO-26` one blinkless version (`MOD-05`). No blinkless `replace` in this repo.
- [s] `GO-27` examples `replace` to parent while developing; tagged examples module rewrite documented (`MOD-01`..`MOD-03`).
- [s] `GO-28` no new library module requires unless `MOD-11` is updated. A little copying over a new dependency.

### Concurrency

- [s] `GO-29` existing locks stay documented (example: `changedPropsMu` in `page_patch_props.go`). New shared mutable state gets a mutex or channel and a `-race` test.
- [s] `GO-30` do not copy `Page` values (atomic fields already make copies unsafe). Tests and examples pass pointers.

## Guardrails (phrase the target, keep the ban as a gate)

Do this:

- Export a wrapper method when an app needs it.
- Return `error` from patch, pick, fetch, crash dir.
- Put helpers in `ownframe` or `ownframe/pagetest`.
- Run race and govulncheck.

Fail the PR if it does this:

- [s] `GO-BAD-01` new synonym `func (p *Page) X() { p.Y() }`
- [s] `GO-BAD-02` new root `type Foo = page.Foo` after the wrap, except the written Config/Handlers/Box decision
- [s] `GO-BAD-03` new `examples/**` import of `ownframe/internal/`
- [s] `GO-BAD-04` `PatchElement` `bool` reintroduced
- [s] `GO-BAD-05` `init()` that opens network, clipboard, or files
- [s] `GO-BAD-06` `log.Fatal` in `internal/` or root library files
- [s] `GO-BAD-07` `exec.Command("sh", "-c", userPath)`
- [s] `GO-BAD-08` second builder type next to `Option`
- [s] `GO-BAD-09` kitchen-sink root methods (new Fetch/IPC/display-list aliases)
- [s] `GO-BAD-10` SECURITY_AUDIT_4 markdown as a merge gate

## 1.0 re-check

Before closing the folder cutoff in [README.md](README.md), walk GO-01..GO-30 against the diff from the start of this work. Record date and commit.

- [s] `GO-31` walk completed. Failures become open rows in the owning phase file, not silent skips.

## Dependencies

Phase 0 writes the freeze. This file is the review checklist for Phases 1-4.

## Risks

A long practices file that nobody opens is theater. Point PR template at this file in `BLEED-07`.

## Closure gates

- [s] PR template or CONTRIBUTING links here
- [s] `GO-31` recorded at 1.0
