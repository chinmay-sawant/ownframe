# Code review - Go library scope

> **Parent:** `plans/v0.0.2/README.md` - v0.0.2 plan
> **Status:** review complete, fixes open
> **Estimated effort:** 3 to 5 days for Phase 1 and 2, 1 week for full roadmap

---

## Overview

This is a read only review of the Go library. Scope is root `*.go` plus `internal/...`. Scope excludes `examples/...`. The `examples/` module is a separate module and was not scanned.

Ten agents ran in parallel. Two found public surface and engine issues. Two checked claims against code. Two checked SOLID packets. Four ran deep scans for concurrency, errors, memory, and over engineering. No code was changed. No test suite was run. This file is the canonical ledger. The HTML companion is `executive-summary-critical-review.html` in the same folder.

## Executive summary

Score is 6.4 out of 10. Confidence is medium high. Coverage is 24 packages in scope. Evidence is code reads plus grep. Full `make test` was skipped because this change is docs only.

What is good. The module graph has no cycles. Direct deps are four and each one earns its place. `net/http` use is modern. Optional interfaces in `internal/host` are narrow. Error chains keep `%w`. Path traversal checks exist. No `nolint` was found. No `os.Exit` lives in library code.

What is bad. `Page` holds about 50 fields and `Redraw` runs the full pipeline in one method. Window `shell` holds the full frame state and `Update` fixes 13 stages in code. Caches bake and decode while holding global locks. One paint path indexes a mask layer with no empty check. `Screen` has 27 methods while `web` needs about 8. Alias sets repeat one behavior under many names. Three packages have no importer in shipped deps. Sentinels use three prefix styles. Window stores a context. `Serve` drops cancel after prepare.

Devil's advocate read. The 6.4 is fair, not harsh. The lock scope and paint guard items deserve work. The fat struct items deserve slow refactors, not a rewrite. Some ponytail cuts go too far. Deleting `internal/ipc` because no internal caller uses it misses the point. It is public API for app authors. Orphan package deletion needs an importer check plus a release note before it ships.

Weighted score arithmetic:

| Category | Weight | Score | Points |
| --- | --- | --- | --- |
| SRP | 10% | 5 | 0.50 |
| OCP | 10% | 6 | 0.60 |
| LSP | 10% | 8 | 0.80 |
| ISP | 10% | 6 | 0.60 |
| DIP | 10% | 6 | 0.60 |
| Package cohesion and dependency direction | 15% | 6 | 0.90 |
| API ergonomics and reusable seams | 10% | 6 | 0.60 |
| Errors, context, ownership, and resource bounds | 10% | 6 | 0.60 |
| Tests, benchmarks, and validation | 10% | 7 | 0.70 |
| Dependency simplicity and language fit | 5% | 9 | 0.45 |
| Total | 100% |  | 6.35, reported as 6.4 |

Scale used: 0 to 2 repeated or breaking problems, 3 to 4 material problems with limited safe reuse, 5 to 6 mixed design with useful seams and clear debt, 7 to 8 strong design with isolated issues, 9 to 10 clear contracts with current validation.

## Phase 1: Correctness and coupling

Fix panics, races, and lock scope first. These touch paint and shared state.

### 1.1 Paint and shared state guards

- [ ] `GO-BOUNDS-01` `internal/masking/mask_apply.go:19` - empty `Layers` indexes `len-1` and nil `desc` derefs on the paint path; expected behavior: return an error for nil or empty descriptor; proof: unit test with nil and empty descriptor plus `go test ./internal/masking ./internal/replay`.
- [ ] `GO-RACE-01` `internal/state/undo_patch_identity.go:26-38` - read modify write spans two critical sections; expected behavior: one atomic update of patch tail or documented single thread contract; proof: `go test -race ./internal/state` with concurrent patch writers.
- [ ] `GO-RACE-02` `internal/page/page_frame_deadline_state.go:11-13` - `frameDirty bool` sits next to atomic `frameAt` and uses plain read and clear; expected behavior: atomic bool handoff; proof: `go test -race ./internal/page ./internal/window -run Frame`.
- [ ] `GO-LOCK-01` `internal/state/store_patch.go:7-31` - user listener runs under the write lock; expected behavior: snapshot listeners under lock then call outside; proof: test where a listener calls back into `Store` without deadlock.
- [ ] `GO-CTX-01` `internal/window/game_state.go:21` - long lived `shell` stores `context.Context`; expected behavior: thread context per call or store cancel only with a comment; proof: grep shows no stale deadline reuse plus `go vet ./internal/window`.
- [ ] `GO-CTX-02` `serve.go:13-32` - `Serve` checks context in prepare then drops it before `web.Serve`; expected behavior: shutdown on cancel or remove context from signature; proof: cancel during serve stops the listener in an integration test.

### 1.2 Error paths that hide cause

- [ ] `GO-ERR-01` `internal/print/print.go:61-65` - helper failures collapse to bare `ErrNoPrinter` with cause only under a debug env flag; expected behavior: joined error keeps helper name and cause; proof: test with failing helper asserts joined message.
- [ ] `GO-ERR-02` `internal/filepick/pick_linux.go:60-64` plus darwin and wsl variants - subprocess exit and stderr collapse to `false`; expected behavior: debug log keeps exit detail; proof: failing picker logs path and exit under debug flag.
- [ ] `GO-ERR-03` `internal/crash/create.go:29-35` - first close error is dropped and write plus close errors are not joined; expected behavior: joined return keeps both; proof: fault injected close asserts joined error.
- [ ] `GO-ERR-04` `internal/clipboard/x11_dial.go:60` and `x11_read.go:32` - background loop and per event worker lack recover; expected behavior: recover logs and fails the client instead of crashing the process; proof: panic in loop is contained in a test.
- [x] `GO-SEC-01` path traversal - `internal/print/command_args.go:16-39`, `internal/crash/validate.go:31-74`, `internal/filepick/pick_wsl_path_safe.go:12-27` reject dot segments and controls; expected behavior already holds; proof: existing grep audit recorded in this report.

## Phase 2: API and reuse

Narrow the surface after correctness lands. Keep compat. Prefer deprecation over removal.

### 2.1 Fat types and seams

- [ ] `SOLID-SRP-01` `internal/page/page_types.go:24-89` - `Page` mixes template, raster, history, form, watch, perf, and accessibility; expected behavior: facade keeps signature while clusters move behind narrow methods; proof: package tests pass after one cluster extraction.
- [ ] `SOLID-SRP-02` `internal/page/page_draw.go:15-97` - `Redraw` runs template plus form sync plus cache plus layout plus caret plus diff plus fallback; expected behavior: pipeline steps are callable units with perf as decorator; proof: one stage has a unit test without a full template.
- [ ] `SOLID-SRP-03` `internal/window/game_state.go:12-81` plus `update.go:11-74` - `shell` and fixed `Update` order block single stage tests; expected behavior: order stays explicit but stages accept fakes; proof: one stage runs headless without Ebiten globals.
- [ ] `SOLID-ISP-01` `internal/host/screen.go:18-48` - fat `Screen` forces 27 methods on all consumers; expected behavior: `Screen` composes narrow interfaces and `web` takes narrow params; proof: `web/areas.go` takes a one method interface and tests drop the embed.
- [ ] `SOLID-DIP-01` `internal/frame/frame.go:6-14` - every non test file imports parent root facade; expected behavior: import `internal/page` or layout directly; proof: `go list` shows no child to parent edge.
- [ ] `SOLID-DIP-02` `internal/bitmap/shape_contains.go:15` - low level raster imports high level `render`; expected behavior: pure geometry moves down and both sides consume it; proof: `go list ./internal/bitmap` drops `render`.
- [ ] `GO-API-01` `internal/page/page_config.go:12` - flat `Config` mixes source, theme, size, and flags with exclusive pairs checked at runtime; expected behavior: grouped source and theme with illegal states rejected early; proof: conflicting `HTML` plus `File` fails at construction in a test.
- [ ] `GO-API-02` `internal/page/page_handlers.go:9` - 20 callback union with wholesale replace; expected behavior: focused registration without clobber; proof: setting one handler keeps prior handlers in a test.
- [x] `GO-API-03` narrow optionals - `host.Ticker`, `Reloader`, `Focuser`, `Dropper`, `Inspector` stay one purpose per file; expected behavior already holds; proof: `web/debug_http.go:51` and `web/reload.go:19` assert narrow interfaces.

### 2.2 Alias and naming debt

- [ ] `GO-API-04` `internal/ipc/request.go:13-21` - eight names for one handler panic; expected behavior: one canonical sentinel; proof: grep shows one canonical name plus any deprecated shim.
- [ ] `GO-API-05` `internal/page/page_patch.go:5-17` - four slice fields for one property set; expected behavior: one canonical field; proof: patch round trip test with one field.
- [ ] `GO-API-06` `internal/page/page_patch_state.go:5-52` - ten names for three behaviors; expected behavior: one name per behavior; proof: docs list three canonical methods.
- [ ] `GO-API-07` `internal/state/listeners.go:21-61` plus changed props and interceptor aliases - duplicate subscribe and bind names; expected behavior: one canonical name each; proof: external caller grep justifies each kept alias.
- [~] `GO-COHESION-01` orphan packages - `internal/runtimebehavior`, `internal/displaypatch`, `internal/styleinvalid` have no shipped importer in `go list -f Deps .`; deferred because public use and release impact are unconfirmed; owner boundary: repo owner; next gate: importer survey plus release note before delete or move.
- [~] `GO-API-08` `internal/ipc` removal proposal; deferred because it is public API for app authors, not dead code; owner boundary: repo owner; next gate: usage survey of external callers before any removal.

## Phase 3: Performance and cleanup

Measure before changing. Most items below are hypotheses from shape, not profiles.

### 3.1 Lock scope

- [~] `GO-PERF-01` `internal/render/cache.go:40-41` plus `display.go:36-37` and `document.go:20-21` - global `engineMu` spans parse and layout; deferred for lack of block profile; owner boundary: render owner; next gate: block profile plus two page concurrent benchmark.
- [~] `GO-PERF-02` `internal/replay/image_cache.go:43-55` - decode plus GPU upload under cache lock; deferred for lack of benchmem proof; owner boundary: replay owner; next gate: miss benchmark with lock hold split and double check insert.
- [~] `GO-PERF-03` `internal/replay/text_cache.go:37-53` and `internal/textrun/face.go:36-53` - font parse under global lock with dual caches; deferred for lack of profile; owner boundary: text owner; next gate: shared source test plus parse outside lock.
- [~] `GO-PERF-04` `internal/replay/chunk_cache.go:15-41`, `raster_cache.go:32-44`, `text_raster_ops.go:16-27` - bake runs under cache lock; deferred for lack of frame profile; owner boundary: replay owner; next gate: bake outside lock with single flight test.
- [x] `GO-PERF-05` no `nolint` - grep finds none in `internal/...`; expected behavior already holds; proof: grep audit recorded here.

### 3.2 Allocation shape

- [~] `GO-ALLOC-01` `internal/page/page_patch_set.go:48-54` - four identical string clones per patch; deferred because callers may rely on independence; owner boundary: page owner; next gate: bench showing patch alloc share before collapsing to one field.
- [~] `GO-ALLOC-02` `internal/replay/chunk_flatten.go:19-31` - per node make plus per chunk copy; deferred for lack of build benchmark; owner boundary: replay owner; next gate: `BenchmarkBuildChunks` with alloc delta.
- [~] `GO-ALLOC-03` `internal/textrun/run.go:60-86` and `caret.go:49-62` - prefix reshape per rune on click path; deferred for lack of caret benchmark; owner boundary: text owner; next gate: long run caret benchmark with incremental measure.
- [~] `GO-ALLOC-04` `internal/textrun/lines.go:38-76` - sort per insert plus `+=` concat; deferred for lack of profile; owner boundary: text owner; next gate: collect benchmark with single sort and builder.
- [~] `GO-ALLOC-05` `internal/bitmap/shape.go:13-17` - per pixel slice literals; deferred because compiler may hoist; owner boundary: bitmap owner; next gate: `-gcflags=-m` plus fallback raster benchmark.
- [~] `GO-ALLOC-06` `internal/bitmap/ordered.go:58` - full canvas alloc per nested group; deferred because pooling without a cap trades one issue for another; owner boundary: bitmap owner; next gate: heap profile before any pool with size cap.
- [ ] `GO-STYLE-01` `api.go` is 2044 chars against the 2000 char repo cap; expected behavior: split error block or alias block into a second root file; proof: `wc -m api.go` reads at or under 2000.

## Phase 4: Validation matrix

How the ten agents map to this ledger.

| Agent | Scope | Verdict |
| --- | --- | --- |
| Discovery 1 API | root plus host, page, fetch, ipc, crash | Confirmed fat types, alias sprawl, sentinel split, stored context |
| Discovery 2 engine | render, replay, frame, window, page, bitmap, emoji | Confirmed global locks, bake under lock, per frame allocs, no `nolint` |
| Validation 1 empirical | root plus internal with grep and reads | Confirmed mask guard miss and lock scope, refuted nine suspected index panics |
| Validation 2 idioms | root plus internal DAG and `go.mod` | Confirmed layering inversions and orphan set, passed modern `net/http` |
| SOLID SRP OCP | page, render, window, run, serve | Area 6, render 8, page 5, window 6, root 9 |
| SOLID ISP DIP | root facade plus host, ipc, web, state | Area 6, fat `Screen` is the top ISP cost |
| Concurrency deep | window, page, ipc, run, serve | Confirmed stored context and serve drop, rest is hypothesis |
| Error deep | fetch, filepick, clipboard, print, crash, web | Clean on panics and traversal, flags on swallowed subprocess detail |
| Memory deep | replay, bitmap, emoji, frame, render, element, masking, textrun | All alloc items theoretical, needs bench proof |
| Ponytail audit | root plus internal, no fixes | Ranked cut list, net minus 1150 lines claimed, needs usage survey |

Bounds checked and clean: frame fills index, pool free index, store undo tails, form options index, touch finger index, replay plan index, textrun caret lines, page dirty rect heads, rune slices from `rangeOf`, image cache order map pairing, lock pairing, `%w` chain, `os.Exit` absence.

## Phase 5: Roadmap to 10 out of 10

This is the order I would work in. Each step has a gate. No step starts without the gate before it.

1. Guard the paint path. Add the mask descriptor check and the crash close join. Gate is the two unit tests in Phase 1.
2. Fix context handling. Thread cancel through `Serve` or drop it from the signature. Remove context from `shell` or document the Ebiten constraint. Gate is the serve cancel test.
3. Narrow one seam. Change `web/areas.go` to take a one method interface and promote one local `web` interface to `host`. Gate is smaller fakes in `web` tests.
4. Collapse one alias set. Pick canonical IPC and patch names. Gate is grep showing one canonical path plus shims.
5. Split one lock. Take image decode or chunk bake out from under its mutex with a double check insert. Gate is a benchmark with less hold time and no race.
6. Settle the orphan question. Survey importers inside and outside the shipped module, including the separate examples module. Delete or move only with a release note. Gate is a written decision per package.
7. Enforce the file cap. Split `api.go` and any page file over 2000 chars. Gate is `wc -m` clean.
8. Profile before pooling. Run heap and block profiles on heavy text and image pages. Only then add caps or pools with bounds. Gate is numbers in the ledger, not shape arguments.

## Dependencies

- Blinkless engine for parse, cascade, and layout. Pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, and text. Single threaded draw shapes the lock analysis.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules are proposed in this review.

## Validation commands

Docs only change. No lint or test run is required for this report per the checklist rule for docs only changes.

- `go list ./...` was used for package scope. Examples were excluded by filter.
- `grep` plus targeted reads were used for every `file:line` claim above.
- Left open: `make test`, `go test -race ./internal/state ./internal/page`, benchmem runs for every `GO-PERF` and `GO-ALLOC` row marked deferred.
