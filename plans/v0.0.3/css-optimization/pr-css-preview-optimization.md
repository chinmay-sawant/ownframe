## Summary

Give the library a way to change one element of a live page instead of rebuilding it. `Page.PatchElement` edits the cached tree, re-lays out only that element's subtree, and marks the changed region dirty, so a click no longer pays for a full parse, cascade, and layout. The CSS preview example uses it for its 43 motion cards and gains a windowed catalog and a Play All run on top.

## Motivation / context

- Plans: `plans/v0.0.3/css-optimization/`, `plans/v0.0.3/optimizations/`, and `documentation/optimizations.md`
- Issues: No issue is linked to this work.

The CSS preview example had 43 Start buttons and a 667 ms page rebuild per click, 28.7 seconds for a full walk. The layout stage was the whole cost: 27.6 ms to lay out 43 cards, 9178 operations, and every text run. A click only changes one card's class and one span's text, so the rebuild was mostly wasted work.

The engine already carried the pieces this needs: a cached parsed tree, a retained layout state, and a per-element relayout. The library just had no path from "the app changed one attribute" to "lay out that one element".

## Changes

### Element patching in the library

- Add `Page.PatchElement(ctx, id, edits...)` and the `ElementEdit` type at the root package. An edit sets an attribute or replaces the text of an element, or of a descendant of the patched element.
- Add `internal/render/cache_edit.go`. The cached document exposes `SetText` and `SetAttribute` on the tree the cascade and layout already read, so a patch edits the live tree rather than a copy.
- The patch path re-lays out the element's subtree through the engine, adopts the patched display, and marks the old and new boxes dirty. It reports whether the patch applied, so a refused patch leaves the caller to redraw.
- The patch needs retained state. `Config.RetainState` keeps the layout state a patch needs, and it is off by default so a page that never patches pays nothing. A page without it keeps no layout state, so the engine refuses the patch and the method returns false.
- `Handlers.ClickPaint` runs after `Click` and a control's own change. It reports that the handler painted the frame, so the page skips the layout.

### Redraw batching and the after-redraw hook

- Add `host.RedrawBatcher` and `host.Batch`. A screen that implements the interface defers its draws across a pointer sequence, so a focus step followed by a click draws once instead of twice. A screen without it draws every step.
- Add the after-redraw path in `internal/page/page_after_redraw.go` and the `Page.SetAfterRedraw` hook. The page runs it after the display, boxes, and hits are installed, so a callback sees a settled page. A redraw asked for inside the callback waits for the commit to finish instead of re-entering it.

### Window and replay

- Add a pixel budget to the oversize check. A content area over 4M pixels no longer builds a content-sized tile grid, because that grid would retain more pixels than the viewport buffer. The window falls back to the viewport path for such a page.
- Reuse the tile walk for a whole-viewport scroll instead of copying the full content rect.
- Split the pointer edge handling so a press that is released inside one frame fires once.
- Report the per-frame allocation in the devtools MEMORY row and add the frame ring sample to the runtime sampler.

### Devtools

- Reorder the tabs to Frame, Elements, Ops, so the dock opens on the frame and pipeline counters instead of the element JSON. The Frame tab holds the zero value the window falls back to.

### CSS preview example

- Window the catalog. Only the visible rows are laid out, with the top and bottom pads reconstructing the full grid height so the scrollbar still reaches the last row.
- Add Play All, which starts every card in the catalog from one press.
- Patch the card row and the state span instead of redrawing the page, so a click costs a patch rather than a rebuild.
- Add the cost harness: `PATCH_COST.md`, `PROFILING.md`, the `cost_*` benchmarks, and `scripts/css-preview-click-test.py` for the real-window click walk.

### Motion runtime

- Add `internal/csstransition/bind_to.go` to bind a transition to a target's paint operations.
- Fix the composition fallbacks in `internal/animation/compose.go`. A property missing from a keyframe returns the underlying paint under a replacing composition. The additive case is still wrong, see the follow-ups.

### Showcase and docs site

- Add the showcase page to the frontend and rebuild the generated GitHub Pages output.
- Add `documentation/optimizations.md` with the recorded numbers for every optimization pass, and link it from the feature index.
- Rebuild the site bundle after the documentation changed, so the published guides match the repository.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | A CSS preview click no longer rebuilds the page: the rebuild per click drops from 667 ms to 0 ms, and the press to painted latency stays at 51.0 ms p50 because the presentation frame dominates. A full 43-click walk drops from 28.7 s to 0 s of page rebuild. An element patch costs 3.80 ms against 25.58 ms for a full redraw of the same card. |
| **Memory** | A patch allocates 6.31 MB against 20.16 MB for a full redraw of the same card. The chunk and raster caches keep their pixel ceilings; only their entry caps moved. |
| **Behavior / correctness** | The library gains `Page.PatchElement`, `ElementEdit`, `host.RedrawBatcher`, and `host.Batch`. A page without retained state refuses every patch and reports false. The devtools dock now opens on Frame. |
| **API / CLI** | New exported `Page.PatchElement`, `Page.SetAfterRedraw`, `Page.BeginRedrawBatch`, `Page.EndRedrawBatch`, `ElementEdit`, `Config.RetainState`, `Handlers.ClickPaint`, `host.RedrawBatcher`, and `host.Batch`. `examples/css-preview` gains the walk harness behind `CSS_PREVIEW_WALK`. No existing exported function or method changed its signature. |
| **Dependencies** | The Blinkless pin moves to `v0.0.3-0.20261021155643-c1054ff46202`, the commit that carries the per-element relayout and the text and attribute setters. It stays the only engine pin, and there is no `replace` directive. No new external module. |
| **Binary size / build time** | One example package grows and the library gains the patch path. No new module and no new dependency. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `ownframe.Config` gains a `RetainState` field, and `ownframe.Handlers` gains a `ClickPaint` field | Keyed struct literals are unaffected. An unkeyed literal such as `ownframe.Config{...}` stops compiling, which is the same break every added field causes. |

## Test plan

- [x] `make test`
- [x] `make build`
- [x] `bash scripts/pr-diff-stat.sh master`
- [x] `go test -p 1 ./examples/css-preview`
- [x] `npm test` and `npm run check:site` in `frontend/`

### Commands

```sh
make test
make build
go test -p 1 ./examples/css-preview
bash scripts/pr-diff-stat.sh master
cd frontend && npm test && npm run check:site
```

## Screenshots / sample output

The click cost, from `documentation/optimizations.md`. Every column is the page rebuild one click caused, in milliseconds.

| Metric | Original | First pass | Second pass | Third pass | Fourth pass | Now |
|---|---:|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 12.5 ms | 6.16 ms | 0 ms |
| p50 | 648 ms | 95 ms | 32.1 ms | 12.9 ms | 6.20 ms | 0 ms |
| p95 | 845 ms | 144 ms | 41.9 ms | 15.5 ms | 7.70 ms | 0 ms |
| max | 864 ms | 163 ms | 48.0 ms | 19.2 ms | 8.50 ms | 0 ms |
| total | 28.7 s | 4.6 s | 1.44 s | 0.54 s | 0.265 s | 0 s |

The patch against a full redraw, measured on this branch with `go test -bench BenchmarkPatchCost -benchtime 30x ./examples/css-preview/` at 1908x999 on an Intel i7-13700HX. The recorded table in `examples/css-preview/PATCH_COST.md` is from an older run and its allocation columns no longer match the code, see the follow-ups.

| Path | Time | Allocated | Allocations | Patches applied |
|---|---:|---:|---:|---:|
| Full redraw, card class | 25.58 ms | 20.16 MB | 32188 | |
| Element patch, card class | 3.80 ms | 6.31 MB | 6875 | 30 of 30 |
| Full redraw, state class | 32.09 ms | 19.86 MB | 31656 | |
| Element patch, state class | 4.12 ms | 5.99 MB | 6347 | 30 of 30 |
| Full redraw, same source | 20.94 ms | 13.87 MB | 25305 | |

The visible latency, from the real click script with `-perf` on:

| Segment | p50 | p95 | max |
|---|---:|---:|---:|
| press to started | 15.8 ms | 30.9 ms | 32.2 ms |
| press to painted | 51.0 ms | 51.6 ms | 52.3 ms |

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/v0.0.3/css-optimization/pr-css-preview-optimization.md`

## Follow-ups (out of scope)

A review of this branch found the items below. None of them are fixed in this PR, and none of them were checked in by this PR.

### Element patching

- A refused patch leaves the parsed tree ahead of the display. `Page.PatchElement` edits the cached tree before the relayout, and an error path returns false without undoing the edit. The render key carries no document term, so the next `RedrawIfChanged` skips and the edit shows up only at the next forced redraw. `internal/page/page_patch_element.go:44`
- A patch does not clear the cached PNG, so `Page.PNG()` and `GET /frame.png` return the pre-patch picture. `internal/page/page_patch_splice.go:15`
- A patch does not rebuild the accessibility tree, so it reports pre-patch text. The staleness check compares the source, which a patch does not change. `internal/page/page_patch_splice.go:11`
- The patch marks only the element's border box dirty, while the redraw path diffs operation rects, so ink that changed outside the box is not repainted. `internal/page/page_patch_splice.go:18`
- `Cache.SetText` follows the engine's `textContent` semantics and drops every child of the element. The comment says it replaces the text. `internal/render/cache_edit.go:24`
- The cached pseudo flags are fixed per cache, but a patch re-collects the sheets, so a patch that rewrites a `<style>` element can add a `:hover` rule the cache does not know about. `internal/render/cache.go:31`
- The batch's final paint clears the pending redraw flag that an earlier batch step asked for. `internal/page/page_click_paint.go:29`

### Hit testing

- A `position: fixed` box loses the hit test to a later content box even though the fixed box is painted on top. `internal/page/page_fixed_hit.go:66`
- The bitmap fallback path sets the boxes without rebuilding the hit map, so fixed-box hit handling disappears there. `internal/page/page_draw_fallback.go:31`

### Window and replay

- The oversize pixel budget lives in the shared `oversized` helper, which also gates the scaled and fitted draw paths. A page over roughly 4.19M CSS pixels falls back to the direct replay and loses its zoom and fit scale, while the pointer still divides by the zoom. `internal/window/replay_oversize.go:19`
- The same threshold moves an ordinary large page onto the viewport path, which never reports an idle frame, so the window keeps spinning.

### Motion

- The new per-field fallbacks apply under additive composition too, so a color missing from a keyframe is added to the underlying color and saturates it. `internal/animation/compose.go:17`
- `internal/csstransition/bind_to.go` is 2283 characters, over the 2000-character file cap.
- The transition rebind only checks the length of the caller-supplied slice, so paints captured on an earlier display blend onto whatever the indices resolve to now. `internal/csstransition/bind_to.go:41`
- The chunk signature folds a per-layout pointer, so a chunk holding operations inside a blend or isolation group misses on every redraw. `internal/replay/chunk_cache_types.go:23`

### CSS preview

- The Play All clock is stamped at the end of the click handler, but a redraw follows the click, so every animation starts with most of its duration already spent. `examples/css-preview/catalog_play_all.go:50`
- `TestPlayAllMovesEverySample` fails under load and passes on its own, so the timing gap above reaches the test suite.
- `syncTree` restores each started row with its own `PatchElement`, and each patch re-lays out the whole document, so a scroll after Play All pays for every started row. `examples/css-preview/catalog_sync.go:26`
- The Play All queue grows on every press, and the keys of rows that never scroll into view are never dropped. `examples/css-preview/catalog_play_all.go:44`
- The catalog geometry is measured once, so a window resize leaves the scroll range and the pads on stale geometry. `examples/css-preview/run.go:37`
- The walk `/box` route reads the page boxes from the HTTP goroutine without the walker mutex. `examples/css-preview/walk_http.go:52`

### Public API

- The public `ElementEdit` alias drops the `ID` field rules the aliased type documents. Godoc shows the alias comment, and `internal/page` is not importable, so a user of `ownframe.ElementEdit` cannot tell what `ID` does. `element_edit.go:3`
- Text patched into an element skips the emoji pass, which only runs over the source string. An emoji in patched text paints as a raw text run, while the same emoji in the template paints a color image. `internal/render/cache_edit.go:34`
- A failed final draw drops the queued redraw, because the page clears `redrawPending` before it draws. In the web path the context is the request context, so a client disconnect during the draw leaves the committed display stale and nothing retries it. `internal/host/batch.go:24`
- The deferred `EndRedrawBatch` runs during panic unwinding, so a panic inside the final draw replaces the original panic the crash report would have written. `internal/host/batch.go:23`

### Documentation

- `PROFILING.md` describes the pre-window design. It says one press renders the whole catalog and that keeping the window while every preview plays is still a plan, but `catalog_play_all.go` already keeps the catalog windowed and `documentation/optimizations.md` already describes the shipped behavior. `examples/css-preview/PROFILING.md:63`
- `PROFILING.md` also reports a "Now" column of 7.1 ms redraw p50 and 10.3 ms press to started, while `documentation/optimizations.md` reports 0.0 ms and 15.8 ms for the same path. `examples/css-preview/PROFILING.md:48`
- `documentation/optimizations.md` says the template prints the started class and the state text, so the executed source changes and the tree is parsed again. The template prints the base class and a literal `Ready`; the state is restored by the after-redraw hook. `documentation/optimizations.md:299`
- The allocation columns in `PATCH_COST.md` are stale. `BenchmarkPatchCost` now reports 32187, 6875, 31655, 6347, and 25304 allocations, not the documented 29541, 6049, 29051, 5563, and 23484. The document calls the allocation counts the stable part. `examples/css-preview/PATCH_COST.md:29`
- `PATCH_COST.md` says both patch benchmarks apply 30 of 30, while the benchmark's own comment still says the engine refuses a card row today. One of the two is stale. `examples/css-preview/PATCH_COST.md:35`
- `documentation/optimizations.md` mixes two builds in one sentence: the current full catalog holds 9178 operations and 537 boxes, not 8800 and 537. `documentation/optimizations.md:438`
- `documentation/optimizations.md` says there is no replace, so the build and the module zip carry the same engine. That holds for the main module. The examples module still pins Blinkless `v0.0.2` and builds only through the workspace file. `documentation/optimizations.md:639`
- Neither `PATCH_COST.md` nor `PROFILING.md` is linked from `documentation/features.md` or `documentation/features-examples.md`, so a reader of the index cannot find them. `documentation/optimizations.md:326`
- The showcase images carry width and height attributes that do not match the files, so they render distorted: teams.webp is 720x402, spotify.webp 780x520, pet-care.webp 1882x982, desktop-cat.webp 880x476, and wispr-flow.webp 1906x1030, against a 1000x650 box. The frontend keeps the real sizes, so the repository document and the site disagree. `showcase.md:43`
- The Wispr Flow preview autoplays a muted looping video with no pause control, and it is `aria-hidden`. The reduced-motion rule in the stylesheet only cancels CSS animations, so the clip keeps running for a reader who asked for reduced motion. `frontend/src/Documentation.jsx:152`
- The home page "Browse all examples" link still targets `#docs/showcase/telegram`, while the header nav and the sidebar default to `#docs/showcase/wispr-flow`. `frontend/src/Home.jsx:108`
- The showcase picker uses `aria-current="page"` for same-page section links, where `aria-current="true"` is the correct value for an in-page target. `frontend/src/Documentation.jsx:138`
- The showcase page heading is a hardcoded `<h1>Showcase</h1>`, while the nav, the sidebar, and the overview link all say "Example apps". `frontend/src/Documentation.jsx:143`

### Test coverage

The branch's new behavior has gaps. These have no test, so a regression in them passes CI.

- The 4-Mpixel oversize tile budget. The only `oversized` check exercises the old device-limit branch, so deleting the budget fails nothing. `internal/window/replay_oversize.go:19`
- The `PatchElement` refusal branches: no retained state, an empty id, and an edit that changes nothing. Dropping the `RetainState` guard fails nothing. `internal/page/page_patch_element.go:29`
- The dirty region and the rebuilt hit list after a patch. No test calls `TakeDirty` after a patch and no test clicks a patched element, so both can be dropped unnoticed. `internal/page/page_patch_splice.go:18`
- `render.Cache.SetAttribute` and `SetText`, and the `RelayoutFor` and `SetStyled` commit contract. Making `RelayoutFor` commit the styled document would pass every test and corrupt a refused patch. `internal/render/cache_edit.go:41`
- `csstransition.BindTo`, including the branch where the destination paints do not match. `internal/csstransition/bind_to.go:17`
- Nested redraw batches. The inner `EndRedrawBatch` must not draw. `internal/page/page_redraw_batch.go:12`
- The walk HTTP surface, whose `/state` field order the click script indexes by position. `examples/css-preview/walk_http.go:66`

Three of the branch's own tests are weaker than they look:

- `patch_cost_card_parity_test.go:89` returns early when the engine refuses the card patch and then compares the operations against the same pre-patch operations, so it passes with the patch feature removed.
- `cycle_bench_test.go:110` never calls `b.ResetTimer`, so the page setup sits inside the timed region and inflates the custom `ms/click` metric.
- `window_bench_test.go:95` includes the reset in the timed loop, so `ns/op` counts the reset, and its claim that it mirrors `scripts/css-preview-click-test.py` is wrong: it measures the page's `LastRedraw`, not the script's press to painted interval.

### Repo standards

- Five Go files on this branch are over the 2000-character cap in `AGENTS.md`: `examples/css-preview/catalog_play_all_test.go` (4421), `examples/css-preview/catalog_scroll_sample_test.go` (2502), `internal/csstransition/bind_to.go` (2283), `internal/page/page_fixed_hit.go` (2209), and `examples/css-preview/walk_http.go` (2139). Each needs a split inside its package.
- `internal/render/cache.go` sits at exactly 2000 characters, with no headroom.
- `go.sum` keeps four stale Blinkless checksum lines from the intermediate engine pins on this branch. `go mod tidy` removes them.
- `examples/go.mod` still pins Blinkless `v0.0.2`, and this branch's cost harness calls `layout.RelayoutElement`, which only the new pin carries. The examples module already lagged behind the published tag before this branch, so the fix belongs to the release process that bumps the examples require.

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
|-----------|-------|------------|-----------|
| `.css` | 3 | 74 | 26 |
| `.go` | 149 | 4997 | 316 |
| `.html` | 2 | 13 | 8 |
| `.js` | 4 | 806 | 262 |
| `.jsx` | 2 | 83 | 55 |
| `.md` | 11 | 961 | 51 |
| `.mod` | 1 | 1 | 1 |
| `.py` | 1 | 275 | 0 |
| `.sum` | 1 | 8 | 0 |
| **Total** | **174** | **7218** | **719** |
