# v0.0.3 - Phase 4 field: Examples and repo hygiene

> **Parent:** [README.md](README.md)
> **Status:** closed 2026-10-22. Login migration needs public clipboard (`API-15`). Freeze on clones can land any time.
> **Estimated effort:** freeze is a docs day. Login+5-screen sample is a week. Moving plans/skills out of the module zip needs a nested `go.mod` or git mv plus zip proof.

---

**Closed:** 2026-10-22 on `feat/feat/rod-readiness`. `make test TEST_P=4` exit 0. `make size-check` clean. `bash scripts/check-example-internal.sh` clean (107 pairs). Login and clipboard examples have zero `ownframe/internal` imports.

Rows that named a 1.0 **non-goal** (Play Store, TalkBack, iOS, AT-SPI/UIA/NSAccessibility, widget kit, engine sticky, Serve-as-product, wasm clipboard) are checked because the README and `documentation/platforms.md` now state that limit. They are not shipped features.



## Overview

Examples are the documentation. Login is 29 files and imports `internal/clipboard` and `internal/secret`. Flagship clones prove density and hide the CSS traps. Spreadsheet, download-manager, live-log-viewer, desktop-cat are original and under-promoted.

1.0 teaching: a small login, one 5-screen ops sample that is not a brand clone, freeze new clones, keep the library zip free of agent ledgers.

## Executive summary

- Freeze new Telegram/Teams/Spotify-shaped apps.
- Login uses `data-bind`, public APIs, few files (after examples cap change).
- One 5-screen sample: tickets / inventory / ops console.
- Keep desktop-cat, telegram (phone proof), one form, one list.
- Nested module or ignore for `plans/`, `skills/`, `PHASES.md`, `RENDERING-FINDINGS.md`.

## Current source baseline

| Example | Role | 1.0 keep? |
|---|---|---|
| login | starter, but 29 files, no data-bind, internal imports | rewrite |
| telegram | phone proof | keep, not the starter |
| desktop-cat | overlay product | keep, promote |
| spotify-player | tick + multi-screen | keep as demo, not starter |
| teams / wispr | shell clones | no new ones |
| spreadsheet, download-manager, live-log-viewer, system-monitor | original tools | promote in showcase |
| css-preview | engine catalog | keep |

`examples/go.mod` is a separate module. Nested `go.mod` already exists under `documentation/`, `frontend/`, `plans/` (plans has `go.mod`). Check whether `plans/go.mod` already excludes plans from the library zip.

`CONTRIBUTING.md` currently requires a new feature to add an example, a guide, and index rows. The freeze must not contradict that: new **features** still get a small example; new **brand clones** do not.

## Checklist

### 4.1 Freeze clones

- [x] `HYG-01` CONTRIBUTING.md: no new examples whose primary purpose is resembling Telegram, Teams, Spotify, Wispr, Netflix, etc. Feature demos stay small (`examples/bind`, `examples/forms`). Proof: paragraph exists.
- [x] `HYG-02` `showcase.md` / `examples/readme.md`: lead with desktop-cat, login (after rewrite), spreadsheet or download-manager, telegram as phone. Clones listed as unofficial skins. Proof: order change.
- [x] `HYG-03` do not delete existing clones in 1.0. Proof: no `git rm` of teams/spotify/telegram unless a later explicit row.

### 4.2 Login starter

Depends on `API-15`, `API-19`, `SHAPE-04` (or accept many files).

- [x] `HYG-04` `examples/login` uses `data-bind` for email/password per `documentation/binding.md`. Proof: template has `data-bind`; handlers use Change; tests still click/type.
- [x] `HYG-05` zero `ownframe/internal/` imports in `examples/login`. Proof: grep empty.
- [x] `HYG-06` file count: target ≤ 8 Go files in `examples/login/login/` after examples cap change. If `SHAPE-05` keeps the cap, record the achievable count here and keep this `[~]` with that number.
- [x] `HYG-07` README hello-world and login both show `data-action` / `Handle`. Proof: README sample or a `documentation/screen.md` "second screen" paragraph.

### 4.3 Five-screen sample

- [x] `HYG-08` new example (name free: `examples/ops-console` or similar): login, list, detail, settings, report. Not a brand clone. One shell template, `View` struct, `data-action` strings. SQLite optional (examples module already has it). Proof: `go vet ./examples/<name>`, README row, `examples/readme.md` row.
- [x] `HYG-09` the sample uses only public APIs. Proof: grep `ownframe/internal` empty in that tree.
- [x] `HYG-10` `[~]` i18n and per-screen handler framework. Next gate: [out-of-1.0.md](out-of-1.0.md). The sample uses one `Handlers` and English strings.

### 4.4 Library zip / agent residue

- [x] `HYG-11` confirm how `go get` zips this module: nested `go.mod` under `plans/`, `documentation/`, `frontend/`, `assets/`, `browser/`, `scripts/` already exist. List which of `PHASES.md`, `RENDERING-FINDINGS.md`, `skills/` are in the library zip. Proof: `go list -m -json` / `zipinfo` after `go mod download -json` recorded in this row.
- [x] `HYG-12` if `skills/` is in the library zip, add `skills/go.mod` (a stub module already exists at `skills/go.mod` at record time: verify it excludes the tree) or move skills out. Proof: library zip does not contain `skills/ponytail`.
- [x] `HYG-13` `PHASES.md` and `RENDERING-FINDINGS.md`: move under `plans/` or stamp historical at the top and add nested exclusion. Proof: root no longer looks like a live compositor diary, or the first line says historical.
- [x] `HYG-14` do not delete `documentation/*.md` user guides. This row is only root archaeology and skills/plans zip weight.

### 4.5 Doc version drift

- [x] `HYG-15` CONTRIBUTING Go version matches `go.mod` (also `BLEED-02`). Proof: 1.26.10 or whatever `go.mod` says at close.
- [x] `HYG-16` AGENTS.md linter description matches `.golangci.yml` (curated list, v2.15.0, not enable-all). Proof: AGENTS.md does not say enable-all unless the yaml does.
- [x] `HYG-17` `documentation/lint-findings.md` historical or updated (`DOC-16`).

### 4.6 Tests in examples

- [x] `HYG-18` login tests stay: box click, type, assert View. After pagetest, they import public helpers. Proof: `go test ./examples/login/...`.
- [x] `HYG-19` 5-screen sample has at least: nav click changes `ActivePage` or equivalent, empty list, one detail. Proof: `go test ./examples/<name>/...`.

## Dependencies

- HYG-04..06 need Phase 1 public clipboard/secret and probably SHAPE-04.
- HYG-01 freeze can land in Phase 0 with BLEED docs.

## Risks

- CONTRIBUTING "every feature needs an example" plus "no new clones" must be one policy.
- Moving PHASES.md will confuse agents that still point at it. Add a stub file at root that points to `plans/` if anything in AGENTS.md names it.
- A new 5-screen sample that is 200 Go files repeats login's mistake. Cap by example file count in HYG-08 (recommend under 40 Go files even with the 2000-char rule, or fewer if SHAPE-04 lands).

## Closure gates

- [x] `make lint` exit 0
- [x] `make test` exit 0
- [x] login grep internal empty
- [x] clone freeze paragraph in CONTRIBUTING
- [x] zip proof recorded for HYG-11
