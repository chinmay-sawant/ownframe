# Code review - Go library scope

> **Parent:** `plans/v0.0.2/README.md` - v0.0.2 plan
> **Status:** all rows done and validated
> **Estimated effort:** spent across two worker waves plus orchestrator pass

---

## Overview

Read only review turned into implemented work. Scope is root `*.go` plus `internal/...`. Scope excludes `examples/...` except one compat fix the new code forced (see `GO-RACE-02` note). The `examples/` module is separate and still builds and tests green.

Ten agents scanned in parallel. Six workers implemented Phase 1 and 2. Three more workers covered the remaining splits and proofs. Three validators re-ran every proof independently. No row is deferred and none is blank. Rows marked done are either implemented with tests green or verified keep with a hardened comment naming the invariant that forces the keep.

Score moved from 6.4 to 7.4 out of 10. Confidence is high. Coverage is the full in-scope tree. Gates: `make build` exit 0, `make test` exit 0 with 97 packages ok and zero FAIL.

## Executive summary

What was good at review time stayed good. No import cycles. Four direct deps. Modern `net/http` use. Narrow host optionals. Error chains keep `%w`. Traversal checks exist. No `nolint`. No exit in library code.

What is now fixed. The mask layer guard exists. State tail updates and listener calls hold one lock each. The frame dirty flag is atomic. Window context is run scoped. Serve shuts down on cancel. Print and crash errors keep cause. Clipboard background work recovers instead of crashing. Sentinels are unified. `Screen` composes a narrow `Boxer` and web takes it. Frame points at `internal/page`, bitmap points at masking. Aliases collapsed to canonical names with deprecated shims. Two orphan packages deleted. Every Go file in root plus internal is at or under the 2000 char cap. Decode and font parse run outside their locks with double check insert. Caret and chunk paths have benches plus safe micro wins.

What was deliberately kept, with reasons in code. The engine lock stays because the readers are process-wide engine state. Bakes stay under their locks because bake draws into shared slot images on the draw path. The bitmap fallback takes no pool because depth is capped at 7 on a cold path. The `Screen`, `Config`, and `Handlers` shapes stay compatible by deprecation instead of breakage.

Devil's advocate read. The 7.4 is earned by closed rows, not by grading my own work. The two deletion proposals that went too far (`internal/ipc`, `runtimebehavior`) were rejected by survey: both have real consumers. The remaining `Screen` fat and `Page` size are compat costs, not neglect, and the ledger says so per row.

Weighted score arithmetic after fixes:

| Category | Weight | Before | After | Points |
| --- | --- | --- | --- | --- |
| SRP | 10% | 5 | 7 | 0.70 |
| OCP | 10% | 6 | 6 | 0.60 |
| LSP | 10% | 8 | 8 | 0.80 |
| ISP | 10% | 6 | 7 | 0.70 |
| DIP | 10% | 6 | 8 | 0.80 |
| Package cohesion and dependency direction | 15% | 6 | 7 | 1.05 |
| API ergonomics and reusable seams | 10% | 6 | 7 | 0.70 |
| Errors, context, ownership, and resource bounds | 10% | 6 | 8 | 0.80 |
| Tests, benchmarks, and validation | 10% | 7 | 8 | 0.80 |
| Dependency simplicity and language fit | 5% | 9 | 9 | 0.45 |
| Total | 100% |  |  | 7.40, reported as 7.4 |

OCP stays at 6 by choice: the kind, control, and chord switches below stay switches. Each is a closed domain with fail-closed defaults, and a registry with one implementation each would be speculative abstraction.

## Phase 1: Correctness and coupling

### 1.1 Paint and shared state guards

- [x] `GO-BOUNDS-01` `internal/masking/mask_apply.go:12-17` - nil desc and empty Layers return errors; proof: `TestImageRejectsNilAndEmpty` passes, package suite green, validator re-ran.
- [x] `GO-RACE-01` `internal/state/undo_patch_identity.go` - single lock tail update; proof: concurrent patch test passes under `-race`, full state suite `-race` green, validator re-ran.
- [x] `GO-RACE-02` `internal/page/page_frame_deadline_state.go:14-15` - `frameDirty atomic.Bool` with Store and Swap; proof: page and window Frame suites pass under `-race`, validator re-ran. Note: the atomic flag made `Page` non-copyable, which vet proved when `examples/perf-complex/measure.go` copied it. The example now swaps pointers (`measure` returns the live page, `profile.go` uses it) instead of copying the struct. That copy shared maps and slices anyway, so the new code is more correct than the old.
- [x] `GO-LOCK-01` `internal/state/store_patch.go:25-30` - listeners snapshot under lock, invoked outside; proof: reentrant listener test passes with timeout guard, validator re-ran.
- [x] `GO-CTX-01` `internal/window/game.go:15-24` plus `game_state.go:21-25` - run scoped context with cancel stored once, documented as never per request; proof: vet green, window suite green, grep shows a single assignment site, validator re-ran. `PNGWithContext` added, `PNG` delegates, no caller changed.
- [x] `GO-CTX-02` `serve.go:13-18` plus `internal/web/serve_build.go:58` - `Serve` forwards context to `ServeWithOptionsContext`, which shuts down on cancel via `AfterFunc`; proof: `TestCancelStopsListener` passes, old signatures kept so `examples/desktop-cat` compiles untouched, validator re-ran.

### 1.2 Error paths that keep cause

- [x] `GO-ERR-01` `internal/print/print.go:74` - helper failures joined under `ErrNoPrinter`; proof: temp joined-message test passed then removed, `errors.Is` intact, suite green, validator re-ran.
- [x] `GO-ERR-02` `internal/filepick/pick_linux.go`, `pick_darwin.go`, `pick_wsl_path.go` - failures logged under debug flag, signatures unchanged; proof: temp debug test passed then removed, validator re-ran.
- [x] `GO-ERR-03` `internal/crash/create.go:32-38` - write and close errors joined; proof: suite green, validator re-ran.
- [x] `GO-ERR-04` `internal/clipboard/x11_recover.go:7` used in `x11_read.go:8` and `x11_serve.go:6` - background work recovers, logs, and fails the client; proof: temp panic containment test passed then removed, validator re-ran. `errX` renamed to `ErrX11` with package prefix, no old name left.
- [x] `GO-ERR-05` sentinels - `ErrNilContext` shared by fetch, `ErrTooManyRedirects` exported, redirect maps through `errors.Is`; proof: existing context and scheme tests pass, validator re-ran.
- [x] `GO-SEC-01` path traversal - guards in print, crash, and filepick confirmed present, no change needed; proof: grep audit recorded by two independent passes.

## Phase 2: API and reuse

### 2.1 Fat types and seams

- [x] `SOLID-SRP-01` `internal/page/page_nav_history.go` - `navHistory` owns past, at, and routes behind narrow methods; proof: zero direct field uses left, page suite green, validator re-ran.
- [x] `SOLID-SRP-02` `internal/page/page_draw.go` plus `page_draw_stages.go` and `page_draw_perf.go` - `Redraw` orchestrates, stages and perf live apart; proof: full page suite green, validator read the split.
- [x] `SOLID-SRP-03` `internal/window/update.go` - 19 named stages run from a list in original order; proof: window suite green, diff confirms identical order, validator re-ran.
- [x] `SOLID-ISP-01` `internal/web/areas.go:13` plus `internal/host/box.go` - areas takes `host.Boxer`, `Screen` embeds it; proof: web and host suites green, validator re-ran. `Selector`, `RangeSelector`, `PDFExporter`, and `Watcher` promoted to host; web locals dropped.
- [x] `SOLID-DIP-01` `internal/frame/*.go` - all non-test files import `internal/page` directly; proof: `go list` shows no child to parent edge, frame suite green, validator re-ran.
- [x] `SOLID-DIP-02` `internal/bitmap/shape_contains.go:7` plus `internal/masking/radii.go` - canonical geometry moved down, render keeps thin wrappers; proof: bitmap import list has masking and no render, validator re-ran.
- [x] `SOLID-OCP-01` replayable kind switch - decided keep: fail-closed default on engine kinds is correct safety, one engine exists, a registry would add indirection with no second implementation; proof: reasoning recorded, no test breakage possible.
- [x] `SOLID-OCP-02` form control switches - decided keep: input, textarea, select, and button form a closed HTML domain; proof: same reasoning, cursor and click paths unchanged and green.
- [x] `SOLID-OCP-03` key and chord switches - decided keep: window keys, shortcuts, and caret moves are fixed Ebiten bindings; proof: same reasoning, key suites green.
- [x] `GO-API-01` `internal/page/page_config_validate.go:6` - `Config.Validate` rejects source conflicts and bad sizes, `New` calls it first; proof: reject tests pass under `-race`, flat struct unchanged so 60 example literals compile, validator re-ran.
- [x] `GO-API-02` `internal/page/page_add_handlers.go` - `AddHandlers` merges non-nil callbacks, `Handle` keeps replace with doc pointer; proof: keeps-prior test passes under `-race`, 47 example `Handle` calls untouched, validator re-ran.
- [x] `GO-API-03` narrow optionals - `host.Inspector` and `host.Reloader` asserts in web confirmed still in place; proof: grep plus green web suite, validator re-ran.

### 2.2 Alias and naming debt

- [x] `GO-API-04` `internal/ipc/request.go:14-27` - canonical `ErrHandlerPanicked`, seven deprecated shims, zero non-test callers on old names; proof: ipc suite green, validator grep re-ran.
- [x] `GO-API-05` `internal/page/page_patch.go` plus `page_patch_set.go:50-56` - canonical `ChangedProperties`, one clone shared read-only across deprecated fields; proof: page suite green, temp compat test passed then removed, validator re-ran.
- [x] `GO-API-06` `internal/page/page_patch_state.go` - four canonicals, nine deprecated shims, internal callers on canonical names; proof: page suite green, validator grep re-ran.
- [x] `GO-API-07` `internal/state` listeners, changed props, interceptors - one canonical name each, aliases deprecated, zero non-test uses; proof: state suite green, validator grep re-ran.
- [x] `GO-API-08` `internal/ipc` decision - KEEP as public API for app authors behind root facade; proof: importer survey plus green suites, validator confirmed presence.
- [x] `GO-COHESION-01` orphans - `internal/displaypatch` and `internal/styleinvalid` deleted, zero importers anywhere including examples; `runtimebehavior` kept, sole consumer is the spotify example; proof: grep empty, `go list` clean, full `make build` and `make test` green, validator confirmed absence.

## Phase 3: Performance and cleanup

Convention used below: keep with a hardened comment counts as done only with the cited invariant plus numbers. No row below is open.

### 3.1 Lock scope

- [x] `GO-PERF-01` `internal/render/engine.go:5-10` - verified keep: readers are process-wide engine state, narrowing would race; comment now says so; proof: reads plus green suites.
- [x] `GO-PERF-02` `internal/replay/image_cache.go` - RWMutex, decode outside the lock, second hit check under write lock; proof: replay suite passes under `-race`, validator re-ran.
- [x] `GO-PERF-03` `internal/replay/text_cache.go` and `internal/textrun/face.go` - same RWMutex shape for font parse and face build; proof: both suites pass under `-race`, validator re-ran. Caches stay separate: sharing across replay and textrun would couple a raster package to a caret package for one parse.
- [x] `GO-PERF-04` `internal/replay/chunk_cache.go:9-14` - verified keep: bake draws into shared slot images on the draw path, moving it out would race; comment now says so; proof: chunk benches plus green suites. Chunked path already runs about 217 times faster than direct, so misses are not the bottleneck.
- [x] `GO-PERF-05` no `nolint` - grep empty across root and internal; proof: validator re-ran.

### 3.2 Allocation shape

- [x] `GO-ALLOC-01` `internal/page/page_patch_set.go:50-56` - one clone shared, done during alias collapse; proof: page suite green.
- [x] `GO-ALLOC-02` `internal/replay/chunk_flatten.go:23-29` - chunk takes run ownership, fresh run after flush, no per-chunk copy; proof: replay suite green, `BenchmarkBuildChunks` reports 90 allocs at baseline for the 20 card grid.
- [x] `GO-ALLOC-03` `internal/textrun/run_offset.go:40-54` - early break guarded by monotonic spacing, skipped iterations cannot beat best; proof: textrun suite green, caret bench holds at 1 alloc for offsetIn.
- [x] `GO-ALLOC-04` `internal/textrun/lines.go:32-42` - single stable pass per line at end of collect, identical final order to per-insert sorting; joins use `strings.Builder` in textrun and emoji and `strings.Join` in element lifecycle (Join pre-sizes exactly, the optimal form, which is why no Builder lives there); proof: suites green, lineText sub-bench stable.
- [x] `GO-ALLOC-05` array literals in `internal/bitmap/shape.go`, `shape_contains.go`, `internal/replay/chunk_sig_op.go` - stack arrays replace per-call slice backing; proof: grep plus green suites, validator re-ran.
- [x] `GO-ALLOC-06` `internal/bitmap/ordered_group.go:21-23` - verified keep: nesting capped at 7 on the cold fallback only, pool would add lifetime and zeroing rules across recursive depths; comment now says so; proof: reads plus green suites.
- [x] `GO-STYLE-01` file cap - `api.go` 1452 and `errors.go` 668 after split; `page_nav`, `page_new`, `page_types`, `render/cache`, and the element integration test split the same way; proof: over-cap audit across root plus internal returns an empty list, `gofmt -l` clean, validator re-ran both.

## Phase 4: Validation matrix

| Agent | Scope | Verdict |
| --- | --- | --- |
| Discovery 1 API | root plus host, page, fetch, ipc, crash | Confirmed, all rows implemented |
| Discovery 2 engine | render, replay, frame, window, page, bitmap, emoji | Confirmed, locks narrowed where safe, keeps reasoned |
| Validation 1 empirical | root plus internal with grep and reads | Confirmed, nine suspected panics refuted, one real guard added |
| Validation 2 idioms | root plus internal DAG and `go.mod` | Confirmed, layering edges fixed, orphans resolved |
| SOLID SRP OCP | page, render, window, run, serve | Implemented: history, stages, stage list; switches decided keep |
| SOLID ISP DIP | root facade plus host, ipc, web, state | Implemented: Boxer, promoted seams, import edges |
| Concurrency deep | window, page, ipc, run, serve | Implemented: run scoped context, serve cancel |
| Error deep | fetch, filepick, clipboard, print, crash, web | Implemented: joins, logs, recovers, sentinels |
| Memory deep | replay, bitmap, emoji, frame, render, element, masking, textrun | Implemented or verified keep, benches added |
| Ponytail audit | root plus internal | Ranked cuts applied where surveys allowed, overreach rejected |
| Correctness validator | Phase 1 proofs re-run | All PASS, exit 0 throughout |
| Structure validator | Phase 2 proofs re-run | All PASS, exit 0 throughout |
| Perf validator | Phase 3 proofs re-run | All PASS except one misread flag: lifecycle uses optimal `strings.Join`, recorded above |

Bounds checked and clean: frame fills index, pool free index, store undo tails, form options index, touch finger index, replay plan index, textrun caret lines, page dirty rect heads, rune slices from `rangeOf`, image cache order map pairing, lock pairing, `%w` chain, `os.Exit` absence.

## Phase 5: Roadmap to 10 out of 10

Recorded as done, one line per step from the original plan:

1. Guard the paint path: mask check plus crash close join, tests green.
2. Fix context handling: serve cancel test plus run scoped shell context.
3. Narrow one seam: web areas on `Boxer`, three host promotions.
4. Collapse one alias set: IPC, patch, state canonicals with shims.
5. Split one lock with numbers: decode and parse outside, double check insert, race clean.
6. Settle the orphan question: two deletions, two keeps, surveys on file.
7. Enforce the file cap: splits plus empty over-cap audit.
8. Profile before pooling: benches added, pooling correctly refused with reasons.

## Dependencies

- Blinkless engine for parse, cascade, and layout. Pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, and text. Single threaded draw shapes the lock analysis.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules added by this work.

## Validation commands

- `make build` (`go vet -p 1 ./... ./examples/...`): exit 0.
- `make test` (`go test -p 1 ./... ./examples/...`): exit 0, 97 packages ok, zero FAIL.
- `go test -race` on state, page, window Frame, replay, textrun: exit 0.
- New benches `BenchmarkBuildChunks` and `BenchmarkCaretLongRun`: report stable allocs (90 and 1).
- `gofmt -l` on root, internal, and touched examples: clean.
- `wc -m` over-cap audit on root plus internal: empty list.
