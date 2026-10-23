# v0.0.3 - Phase 1 field: Package shape and file size

> Superseded planning record. `[s]` marks an old proposal or assertion, not proof of implementation. `[-]` marks work explicitly deferred in the original plan. Current decisions and evidence live in [plans/README.md](../../README.md).

> **Parent:** [README.md](README.md)
> **Estimated effort:** split is weeks if done as one move. 1.0 only needs the rule change plus no new synonym files.

---

## Overview

The 2000-character Go file cap produced a shredded god package (`internal/page` ~381 files, `internal/window` ~356, `internal/replay` ~134), not deep modules. Go layout is: related methods on one type, unexported until needed, split by **package** when the type is too big.

1.0 does not require merging 381 files. It requires: no new synonym files, examples not forced to look like 29-file login, and a written path to split `page` into form/nav/dirty/draw.

Go practice: one primary type per file when it has significant methods; group constructor and methods; unexport aggressively; a little copying over a new dependency.

## Executive summary

- 1.0: keep the cap for `internal/` until a split PR. Relax **examples** only after AGENTS.md and `scripts/check-file-size.sh` say so.
- Do not create `mnd_const.go` or comment-only `.go` files.
- Split `internal/page` by package after the public wrapper exists, so the public type does not grow during the move.

## Current source baseline

| Tree | Approx Go files (excl temp/frontend) | Notes |
|---|---|---|
| `internal/page` | 381 | god type `Page` |
| `internal/window` | 356 | `SUBSYSTEMS.md` is the map |
| `internal/replay` | 134 | |
| examples | ~2300 | login ~29 Go files |
| allowlist | `scripts/file-size-allowlist.txt` empty by design | gate fails over-limit unlisted files |

`examples/login/login/login_clip.go` was reported as a comment-sized file so the package stays under the cap. Confirm before deleting.

`internal/window/SUBSYSTEMS.md` already maps window groups. Replay and page have no equivalent index.

## Checklist

### 1.1 Rules (docs + gate)

- [s] `SHAPE-01` AGENTS.md - write that the 2000-char cap applies to library `internal/` and root until a split, and that **new examples** may exceed it only after `SHAPE-04`. Do not claim the cap is gone while `size-check` still fails. Proof: AGENTS.md and the script agree.
- [s] `SHAPE-02` CONTRIBUTING.md file-size paragraph matches AGENTS.md. Proof: both files quote the same number and the same trees.
- [s] `SHAPE-03` forbid new files whose only purpose is a linter or the cap: comment-only `.go`, `mnd` constants in a dedicated file. Proof: CONTRIBUTING sentence; `BLEED-04` already blocks synonym methods.

### 1.2 Examples cap (1.0 teaching)

- [s] `SHAPE-04` change `scripts/check-file-size.sh` **or** the allowlist policy so `examples/**` is excluded **or** the cap for examples is raised (recommend 8000 characters, one screen's Go in a few files). Proof: `make size-check` green; login can be merged toward fewer files in Phase 4.
- [s] `SHAPE-05` if examples stay capped, record why in this row and keep `SHAPE-04` `[~]` with that reason. Do not leave the starter at 29 files without a written decision.

### 1.3 Test file shredding

- [s] `SHAPE-06` after `SHAPE-04` or a test-only cap raise, merge `*_more_test.go` siblings back into the parent test file where they still fit. Proof: count of `*_more_test.go` drops; `go test` of those packages green.
- [-] `SHAPE-07`  if the cap stays 2000 for tests. Reason: a second scenario does not fit. Next gate: raise test-file cap to 4000. 1.0 can ship with `SHAPE-07` still `[~]`.

### 1.4 Index docs (cheap locality)

- [s] `SHAPE-08` `internal/page/SUBSYSTEMS.md` mapping form, nav, dirty, draw, patch, a11y, host to entry files. No code change. Proof: a reader finds `PatchElement` and `TakeDirty` from the index alone.
- [s] `SHAPE-09` `internal/replay/SUBSYSTEMS.md` mapping tile blit, group/mask, text cache, reject/fallback. Proof: same for `DrawVisibleUnfixedZ` vs documented `DrawRect`.
- [s] `SHAPE-10` keep `internal/window/SUBSYSTEMS.md` current when wrap PRs touch window. Proof: date line updated in that file if window changes.

### 1.5 Package split (after wrapper, can slip past 1.0)

Do not start this until `API-01` is `[x]`.

- [-] `SHAPE-11`  extract form/focus to `internal/page/form` or `internal/form` with `page.Page` delegating. Next gate: compile, `go test ./internal/page/...`, no public API change. 1.0 may leave this open.
- [-] `SHAPE-12`  extract nav (`Load`/`Back`/`Forward`/`Route`) similarly.
- [-] `SHAPE-13`  extract dirty/display commit.
- [-] `SHAPE-14`  extract draw/pipeline/cache talk to `internal/render`.
- [s] `SHAPE-15` after any split, `wc -m` on remaining `page_types.go` and the new packages. Proof: recorded counts.

### 1.6 Replay DrawRect honesty

- [s] `SHAPE-16` docs (`documentation/repaint.md`) name the window path `drawReplayPartial` → tiles, and name `DrawRect` as the test/helper path if that is still true. Proof: `repaint.md` does not claim the window calls `DrawRect` unless it does.

## Dependencies

- `SHAPE-04` interacts with Phase 4 login merge. Do Phase 4 after the examples cap changes, or login stays 29 files.
- Split packages (`SHAPE-11+`) after `API-01` so two large moves do not share a PR.

## Risks

- Raising the cap without changing the script fails CI the other way (allowlist stale) or not at all (policy lie).
- A big-bang merge of 381 files will never review. Split by package, one package per PR.
- `internal/page` tests use `package page`. Moving types breaks every test file. Budget that.

## Closure gates (1.0 subset)

1.0 can close with SHAPE-01, SHAPE-02, SHAPE-03, SHAPE-08, SHAPE-09, SHAPE-16, and a decision on SHAPE-04/05.

- [s] `make lint` (includes size-check) exit 0
- [s] `make test` exit 0
- [s] AGENTS.md and check-file-size.sh agree on which trees are capped
