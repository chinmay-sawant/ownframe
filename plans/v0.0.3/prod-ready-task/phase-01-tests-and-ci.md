# v0.0.3 - Phase 1 field: Tests and CI

> Superseded planning record. `[s]` marks an old proposal or assertion, not proof of implementation. `[-]` marks work explicitly deferred in the original plan. Current decisions and evidence live in [plans/README.md](../../README.md).

> **Parent:** [README.md](README.md)
> **Estimated effort:** 3-7 days. Race on the full library may surface real bugs; budget fix time.

---

## Overview

Go's production bar is `go test`, `go test -race`, fuzz at trust boundaries, and `govulncheck`. Visual/dirty-rect tests already protect paint. They stay. The gap is library-wide race in CI, public-API tests that cannot see `internal/`, and fuzz on crash paths and template data.

Go practice: tests live with the package, external `_test` packages for public API, `UseMemory` for clipboard, no OS side effects, fuzz `FuzzXxx` at the process edge.

## Executive summary

- CI job: `go test -race -p 1` on the library module.
- External `ownframe_test` that imports only `ownframe`.
- Fuzz: `SetCrashDir`, Fetch (already), template/HTML size, image src (already).
- Keep pixel compare and dirty-rect tests. Merge `*_more_test.go` only after the size rule for tests is decided in [phase-01-package-shape.md](phase-01-package-shape.md).

## Current source baseline

| Gate | Where | Gap |
|---|---|---|
| `make test` | `go test -p 1 ./... ./examples/...` | no `-race` |
| Sec tests | `ci.yml` job `sec`: `-run Sec` | keep |
| Fuzz smoke | emoji `FuzzReplace`, web `FuzzSanitizeLine`, fetch `FuzzBlockedIP` + `FuzzMethodHeader`, page `FuzzDOMLimits` + `FuzzImageSrc` | add crash dir, maybe template |
| Race | recorded ad hoc in old plans on state/ipc/page/replay | not a CI job |
| govulncheck | `make vuln`, `vuln.yml`, `dependency-review.yml` | keep |
| Visual | `repaint_*_test.go` at repo root, page dirty tests | keep |
| Clipboard tests | `clipboard.UseMemory` | keep; public after API-15 |

## Checklist

### 1.1 Race

- [s] `TEST-01` add `make test-race` that runs `$(GO) test -race -p 1 ./...` on the **library** module (not every example at first). Proof: `make test-race` documented in Makefile and CONTRIBUTING.
- [s] `TEST-02` CI job in `.github/workflows/ci.yml` (new job or extra step) runs that race target on ubuntu-latest. Proof: a PR without the job cannot merge once this row is `[x]`; quote the YAML in this row.
- [s] `TEST-03` fix races the first CI run finds. Each fix is its own commit if possible. Do not skip with `sync` spray. Proof: job green; list packages that failed first in this row.
- [-] `TEST-04`  examples race. Next gate: after library race is green, optionally `go test -race` on `./examples/login/...` only. Full examples race is out of 1.0 if it is too slow.

### 1.2 Public API tests

- [s] `TEST-05` `api_test.go` in package `ownframe_test` (see also `API-08`): `New`, `SetData`, `Handle`, `Redraw`, form value if public. Imports: `ownframe` only. Proof: `go list -f '{{.Imports}}' ./api_test.go` equivalent grep shows no `internal`.
- [s] `TEST-06` gate test: any test file under `./` package `ownframe_test` that imports `ownframe/internal/...` fails. Proof: trial import fails.
- [s] `TEST-07` `pagetest` package tests (after `API-18`) click a box on a tiny HTML fixture without importing `internal/page`. Proof: `go test ./pagetest` green.

### 1.3 Fuzz

Existing CI fuzz stays. Add:

- [s] `TEST-08` `FuzzSetCrashDir` in `internal/crash` (or public crash package): empty, control chars, `..`, absolute vs relative, long strings. Expected: rejected paths never become the dir; accepted paths match current validation. Proof: `go test -run=NONE -fuzz=FuzzSetCrashDir -fuzztime=15s ./internal/crash` and a CI line.
- [s] `TEST-09` keep `FuzzBlockedIP`, `FuzzMethodHeader`, `FuzzDOMLimits`, `FuzzImageSrc`, `FuzzSanitizeLine`, `FuzzReplace`. Proof: `ci.yml` sec job still lists them.
- [s] `TEST-10` optional `FuzzTemplateHTML` on `New`/`Redraw` with hostile templates under `MaxSourceSize`. Skip if `FuzzDOMLimits` already covers it. Record the decision here.

### 1.4 Paint tests (keep)

- [s] `TEST-11` do not delete `repaint_visual_test.go`, `repaint_login_test.go`, `repaint_forms_test.go`, `repaint_scroll_test.go`, `repaint_theme_test.go`, `page_dirty_*_test.go`. Proof: files still exist at 1.0.
- [-] `TEST-12`  merge `*_more_test.go` files. Reason: 2000-char cap. Next gate: [phase-01-package-shape.md](phase-01-package-shape.md) `SHAPE-06`. Until then, new test cases may add a sibling file rather than break the gate.

### 1.5 Clipboard and filepick tests

- [s] `TEST-13` clipboard tests call `UseMemory` only. Proof: grep `UseMemory` in clipboard tests; no test writes the OS clipboard.
- [s] `TEST-14` filepick tests use `page.InstallPicker` (or the public equivalent after wrap). Proof: tests do not invoke zenity.

### 1.6 Commands recorded

When closing this phase, paste:

```
make test
make lint
make test-race   # or the CI log URL
go test -run=NONE -fuzz=FuzzSetCrashDir -fuzztime=15s ./internal/crash
```

- [s] `TEST-15` those commands recorded with date and exit code in this file.

## Dependencies

- `API-01` wrapper will break tests that type `*page.Page` as `ownframe.Page`. Fix in the wrap PR, not here, if this phase lands first. Prefer landing TEST-05 after the wrap.
- Race CI can land before the wrap.

## Risks

- Full `go test -race ./...` on the library may be slow. Keep `-p 1`. Do not hide failures with `-count=1` only; flakes need a root cause.
- Fuzz in CI is already time-capped at 15s. Do not add unbounded fuzz to the PR gate.
- `make test` already runs examples. Race on examples can OOM the runner. That is why TEST-01 is library-only.

## Closure gates

- [s] `make lint` exit 0
- [s] `make test` exit 0
- [s] race job green
- [s] new crash fuzz in CI
- [s] public `ownframe_test` has zero internal imports
