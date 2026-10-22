# v0.0.3 - Phase 0: Stop the bleed

> **Parent:** [README.md](README.md)
> **Status:** open. Planning only until a row has a PR and proof.
> **Estimated effort:** 1-3 days of review rules plus a small CONTRIBUTING/AGENTS edit. No `Page` wrap in this phase.

---

## Overview

Stop growing the public API in the shapes 1.0 will delete. New PRs follow these rows even before the wrapper lands. This phase is cheaper than wrapping `Page` and it prevents a week of new aliases.

Go practice: unexport until needed, one name per idea, `internal/` stays inside the library, constructors over `init()`, `error` over `bool` and panic.

## Executive summary

- Land a written freeze in AGENTS.md / CONTRIBUTING.md.
- Add a test or grep gate that fails new synonym methods and new root aliases.
- Do not wrap `Page` here. That is Phase 1.

## Current source baseline

| Area | Current | Freeze target |
|---|---|---|
| Public Page | `api.go` `type Page = page.Page` | No additional aliases at root |
| Synonyms | `form_focus_last.go`, `page_patch_state.go`, `page_patch_props.go`, `page_element.go` | No new aliases; deletions in Phase 1 |
| Examples → internal | Many, including `examples/login/login/login_logout.go` | No **new** imports. Existing stay until Phase 1 exports |
| init / globals | `SetCrashDir`, clipboard OS vs memory, `ensureAudio` in `run.go` | No new `init()`; no new package-level clients |
| Context | `Redraw`, `Run`, `Fetch` already take `ctx` | Keep first-arg rule |

## Checklist conventions

IDs `BLEED-NN`. Proof is grep, a gate test, or a doc sentence. `make lint` and `make test` before closing the phase.

### 0.1 Written freeze

- [ ] `BLEED-01` `AGENTS.md` - add a "Public API freeze until Page wrap" paragraph: no new `type X =` at module root except re-exports named in [phase-01-public-api.md](phase-01-public-api.md); no new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows; no new example `ownframe/internal/` imports. Proof: paragraph exists and names this folder.
- [ ] `BLEED-02` `CONTRIBUTING.md` - same freeze, plus "v0 may break; wrap happens before v1." Fix the Go version sentence from 1.26.4 to match `go.mod` (1.26.10 at record time). Proof: grep `1.26.4` in CONTRIBUTING is empty.
- [ ] `BLEED-03` `documentation/features.md` or this folder README - one sentence that 1.0 is desktop `Run`. Proof: sentence exists.

### 0.2 Mechanical gates

- [ ] `BLEED-04` add `internal/page/export_synonym_gate_test.go` (or a `go test` in a small `internal/apiwatch` package) that lists the **current** synonym method names as an allowlist and fails when a new exported `Page` method is added whose body is a single call to another `Page` method. Expected: existing aliases pass; a new `func (p *Page) Foo() { p.Bar() }` fails. Proof: test red on a trial alias, green on HEAD.
- [ ] `BLEED-05` add a root test or `go test` in package `ownframe` that allowlists current `type X =` aliases in root `*.go` and fails a new alias file. Expected: `api.go` `Page` alias is listed; a new `type Foo = page.Foo` fails. Proof: same red/green pair.
- [ ] `BLEED-06` add `examples/internal_import_gate_test.go` **or** a `make` script under `scripts/` that records today's `ownframe/internal/` import set and fails on a **new** import path or a new file that imports `internal/`. Do not fail the existing set in this phase. Proof: adding `import ".../internal/foo"` to login fails the gate; current login still passes.

### 0.3 Review rules (no code behavior change)

- [ ] `BLEED-07` PR template or CONTRIBUTING: reject PRs that add synonym names (`PreviousX` for `X`). Proof: sentence in `skills/PR/PR_TEMPLATE.md` or CONTRIBUTING.
- [ ] `BLEED-08` PR template: new exported methods must return named types from package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR. Proof: sentence exists.
- [ ] `BLEED-09` PR template: library code panics only for invariants. New user-facing failure paths return `error`. Proof: sentence exists.
- [ ] `BLEED-10` PR template: `context.Context` is the first parameter when the function already does I/O, cancel, or deadline. Do not add hidden `time.After` timeouts that ignore the caller's ctx. Proof: sentence exists.

### 0.4 Explicit non-goals this phase

- [~] `BLEED-11` wrapping `Page` - next gate [phase-01-public-api.md](phase-01-public-api.md) `API-01`.
- [~] `BLEED-12` deleting existing synonyms - next gate `API-10` family.
- [~] `BLEED-13` moving Fetch/IPC - next gate `API-20`. Freeze growth only.

## Dependencies

- File-size cap still applies. Gate tests must stay under 2000 characters or split.
- Do not enable a gate that fails HEAD on day one. Allowlist first, then Phase 1 shrinks the allowlist.

## Risks

- A naive "fail any alias" test breaks HEAD. Allowlist the current set.
- Examples module is a second module. A gate test in the library package cannot see `examples/` unless it walks the filesystem or lives in `scripts/`. Prefer `scripts/` plus `make lint` if `go test` in the root cannot import examples.

## Closure gates

- [ ] `make lint` exit 0
- [ ] `make test` exit 0
- [ ] Trial: a throwaway synonym method fails `BLEED-04` (revert after)
- [ ] Trial: a throwaway `internal/` import in login fails `BLEED-06` (revert after)
