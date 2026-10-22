# v0.0.3 - Production-ready library (phase-wise checklist)

> **Parent:** conversation plan for a native desktop-plus-Android 1.0, recorded 2026-10-22, host policy updated 2026-10-22. Critique of the kitchen-sink list lives in [00-scope-and-critique.md](00-scope-and-critique.md).
> **Status:** planning only. Creating these files does not change code, CI, or docs. Every implementation row starts `[ ]`.
> **Recorded:** 2026-10-22, against the tree that aliases `type Page = page.Page` in `api.go`, pins examples to `ownframe v0.0.2` with a second blinkless require and no `replace`, and has login import `internal/clipboard`.
> **Estimated effort:** Phase 0 is days. Phase 1 is the 1.0 library (weeks). Phases 2 to 4 can overlap after Phase 1's public wrapper lands. Native OS a11y is out of 1.0.

---

## Overview

Ship ownframe as a Go library a stranger can `go get` and run as a **native app** on desktop (`Run`: Linux, Windows, macOS) and on Android (`BindMobile` + `examples/android-host`). Native means one compiled Go process, Ebiten window or Activity, no Chromium, no WebView. Wasm and `-web` stay previews. iOS is absent. No widget kit. No new brand clones. Play Store, TalkBack, and process restore stay in [out-of-1.0.md](out-of-1.0.md).

The product score that drove this ledger is 6/10. The engine score is 7.5/10. 1.0 raises the product score by shrinking the public package, returning `error` values, and stopping the README from selling four hosts as one app.

Follow `skills/phase-wise-checklist/SKILLS.md`. One row is one code change or one proof. `[x]` needs current evidence in the owning file. `[~]` needs a reason and a next gate. `make lint` and `make test` before any non-docs phase is marked complete.

## Checklist conventions

- `[ ]` not started or not proven.
- `[x]` implemented and validated with evidence recorded in the owning file.
- `[~]` deferred or partial, with reason, owner, and next gate.
- Stable IDs (`API-01`, `MOD-02`) stay when rows move.
- Inspection of current source is a baseline, not completion.
- Go file size stays at most 2000 characters (`wc -m`) until [phase-01-package-shape.md](phase-01-package-shape.md) changes that rule for examples. Do not silently violate AGENTS.md.
- Do not import `internal/` from new example or public-test code.
- Panic only for broken invariants. User input, patch refuse, and I/O return `error`.
- `context.Context` stays the first argument on work that already takes it.

## Executive summary

| Phase | File | 1.0? | Status |
|---|---|---|---|
| Critique, cutoff, old vs new | [00-scope-and-critique.md](00-scope-and-critique.md) | reference | planning |
| 0 Stop the bleed | [phase-00-stop-the-bleed.md](phase-00-stop-the-bleed.md) | yes | open |
| 1 Public API | [phase-01-public-api.md](phase-01-public-api.md) | yes | open |
| 1 Modules | [phase-01-modules.md](phase-01-modules.md) | yes | open |
| 1 Tests and CI | [phase-01-tests-and-ci.md](phase-01-tests-and-ci.md) | yes | open |
| 1 Package shape | [phase-01-package-shape.md](phase-01-package-shape.md) | partial | open |
| 1 Errors and options | [phase-01-errors-and-options.md](phase-01-errors-and-options.md) | yes | open |
| 2 Docs and hosts | [phase-02-docs-and-hosts.md](phase-02-docs-and-hosts.md) | yes | open |
| 2 Desktop OS | [phase-02-desktop-os.md](phase-02-desktop-os.md) | clipboard + banner yes; AT-SPI no | open |
| 3 Paint and lists | [phase-03-paint-and-lists.md](phase-03-paint-and-lists.md) | recipe yes; engine sticky no | open |
| 4 Examples and hygiene | [phase-04-examples-and-hygiene.md](phase-04-examples-and-hygiene.md) | login + freeze yes | open |
| Go practices (all phases) | [field-go-practices.md](field-go-practices.md) | yes | open |
| Out of 1.0 | [out-of-1.0.md](out-of-1.0.md) | no | deferred |

## 1.0 cutoff

Close 1.0 when every row below is `[x]` in its owning file:

- Public `Page` is a defined type, not `type Page = page.Page`.
- Login and the clipboard example compile with zero `ownframe/internal/` imports.
- `PatchElement` returns `error` with `errors.Is` sentinels.
- `go test -race` on the library module is a CI job.
- `govulncheck` stays a CI job on both modules (already present in `.github/workflows/dependency-review.yml` and `vuln.yml`; prove it still runs).
- README names desktop (Linux, Windows, macOS) and Android as native 1.0 hosts, and names wasm and `-web` as previews.
- `documentation/features.md` has a CSS support matrix.
- Clipboard docs match Wayland/wasm/Android behavior, or Wayland paste works.
- A11y README banner: Go tree exists, no AT-SPI/UIA/TalkBack in 1.0.

List windowing is a documented example recipe, not a new engine. Native a11y stacks are [out-of-1.0.md](out-of-1.0.md).

## Phase-wise checklist (master status)

### Phase 0 - Stop the bleed

File: [phase-00-stop-the-bleed.md](phase-00-stop-the-bleed.md)

Dependency: none. Do this before any new root alias.

- [ ] Unexport by default. No new root aliases. No new `Page` methods unless an example cannot compile without them.
- [ ] One identifier per idea. No synonym methods.
- [ ] New example code uses the public package only.
- [ ] Exported methods do not return unnamed `internal/host` or `internal/state` types.
- [ ] No new `init()` or package-level mutable clients.
- [ ] Panic only for invariants. `context.Context` first where the call already takes it.

Gate: `make lint` and `make test`. Grep for new `type X =` at repo root is empty except the documented leftovers this phase lists.

### Phase 1 - Go library

Files: [phase-01-public-api.md](phase-01-public-api.md), [phase-01-modules.md](phase-01-modules.md), [phase-01-tests-and-ci.md](phase-01-tests-and-ci.md), [phase-01-package-shape.md](phase-01-package-shape.md), [phase-01-errors-and-options.md](phase-01-errors-and-options.md)

Dependency: Phase 0.

- [ ] Wrap `Page`. Host methods stay on `internal/page` as `host.Screen`.
- [ ] Delete synonym methods in the same v0 tag.
- [ ] Clipboard, crash, and `pagetest` as named public packages.
- [ ] Examples `replace` while developing. One blinkless pin.
- [ ] Race CI, fuzz at trust boundaries, public-only `api_test.go`.
- [ ] Split `internal/page` by package when wrapping, not by 1200-character file.
- [ ] `PatchElement(ctx, id, edits...) error` with `%w` sentinels.

Gate: `go get` plus copy of login compiles with zero `internal/` imports. `git grep 'type Page ='` is gone.

### Phase 2 - Desktop and Android truth

Files: [phase-02-docs-and-hosts.md](phase-02-docs-and-hosts.md), [phase-02-desktop-os.md](phase-02-desktop-os.md)

Dependency: Phase 1 public names, so docs match the wrapper.

- [ ] README host matrix: native desktop + native Android; wasm and `-web` previews.
- [ ] Android 1.0 capability list (`DOC-18`..`DOC-20`): android-host, APKs, eyes and telegram exhibits, Play/TalkBack limits.
- [ ] CSS matrix in `features.md`.
- [ ] Wayland clipboard or honest docs.
- [ ] File picker errors on the public API.
- [ ] A11y banner. Native bridges deferred.

Gate: a new user can predict paste, files, a11y, `-web`, and Android APK limits from README plus `features.md` and `platforms.md`.

### Phase 3 - Paint and lists

File: [phase-03-paint-and-lists.md](phase-03-paint-and-lists.md)

Dependency: can start after Phase 0. Do not block Phase 1.

- [ ] Pinned chrome keeps tiles, or `SetScrollWindow` is the documented contract.
- [ ] List windowing example recipe under 16 ms layout at 200 rows.
- [ ] `SetTick` op patches mark dirty rects inside the library.
- [ ] blinkless tagged releases and one pin.

Gate: recorded command, dataset, and metric in the phase file. Successful `make test` is not benchmark proof.

### Phase 4 - Teaching and repo hygiene

File: [phase-04-examples-and-hygiene.md](phase-04-examples-and-hygiene.md)

Dependency: Phase 1 public clipboard/`pagetest` so login can drop `internal/`.

- [ ] 5-screen non-clone sample. Login uses `data-bind`. Few files.
- [ ] Freeze new clones. Keep desktop-cat, telegram, one form app, one list app.
- [ ] Move `plans/`, `skills/`, `PHASES.md`, `RENDERING-FINDINGS.md` out of the library module zip.
- [ ] AGENTS.md and CONTRIBUTING match `go.mod` 1.26.10 and the curated linter list.

Gate: `make lint` and `make test`. Module zip from `go mod download` does not contain `plans/` or `skills/` if those rows are `[x]`.

## Dependencies

- Blinkless pin in root `go.mod`: `github.com/chinmay-sawant/blinkless v0.0.3-0.20261021155643-c1054ff46202`. Engine work lands in that repo first, then this pin moves.
- Ebiten v2.10.4. No idle park on wasm or Android (`internal/window/idle_supported.go`).
- Two modules joined by `go.work`. Examples currently require `ownframe v0.0.2` with **no replace**.
- File-size gate: `scripts/check-file-size.sh` and `scripts/file-size-allowlist.txt`.
- CI today: `.github/workflows/ci.yml` runs `make test` and `make lint`. Sec job runs `TestSec_*` plus short fuzz. `govulncheck` is a separate workflow. Library-wide `-race` is not a CI job.

## Risks

- Wrapping `Page` is a v0 breaking change. Do it before the first v1. After v1, Go's compatibility promise keeps the aliases forever.
- Examples import `internal/` today because the import path prefix is `github.com/chinmay-sawant/ownframe/...`. A user who copies login into `github.com/alice/myapp` cannot compile. That is the 1.0 bug.
- Splitting `internal/page` while wrapping `Page` in the same PR is too much. Wrapper first, package split second.
- Raising the 2000-character cap without a written AGENTS.md change will fail `make lint`.
- Native a11y and Android Play are year-scale. Parking them in this 1.0 folder as open `[ ]` rows will stall the cutoff. They live in [out-of-1.0.md](out-of-1.0.md) as `[~]`.

## What this ledger refuses

See [field-go-practices.md](field-go-practices.md) and [out-of-1.0.md](out-of-1.0.md). Short form:

- Publishing the implementation type as the API.
- Synonym methods so each name fits a file.
- `bool` for operations that fail for reasons.
- Returning `*state.Store` / `host.Stats` without a named public type.
- Kitchen-sink root growth (new Fetch/IPC/display-list aliases).
- Widget kit (dropdown, date, table) in the engine.
- Per-screen handler framework, i18n engine, Android Play, wasm clipboard, Serve-as-product.
- A new 50-item security audit as a 1.0 gate. Fix `gosec`, `govulncheck`, and `-race` instead.

## Closure (whole 1.0)

Record command and outcome here when claiming 1.0. Leave unchecked until then.

- [ ] `make lint` exit 0
- [ ] `make test` exit 0
- [ ] `go test -race -p 1 ./...` on the library module exit 0 (examples optional, recorded)
- [ ] `make vuln` or CI `govulncheck` on both modules exit 0
- [ ] `gofmt -l` over changed trees empty
- [ ] Public compile proof: a throwaway module outside this repo that copies login compiles
- [ ] README plus `features.md` match the host and CSS matrices (desktop + Android native; wasm/`-web` preview)
