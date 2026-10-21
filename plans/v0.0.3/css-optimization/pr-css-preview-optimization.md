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

- Add the showcase page to the frontend and rebuild the generated GitHub Pages output, then rebuild it again after the guides changed, so the published site matches the repository.
- Add `documentation/optimizations.md` with the recorded numbers for every optimization pass, and link it from the feature index.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | A CSS preview click no longer rebuilds the page. The rebuild per click drops from 667 ms to 0 ms, a 100% drop in page work, and the 43-click walk drops from 28.7 s of rebuild to none. The walk still costs about 2.2 s of waiting for the painted frames, and the press to painted latency stays at 51.0 ms p50 because the presentation frame dominates it. An element patch costs 6.08 ms against 34.5 ms for a full redraw of the same card. |
| **Memory** | A patch allocates 6.31 MB against 20.18 MB for a full redraw of the same card. The chunk and raster caches keep their pixel ceilings; only their entry caps moved. |
| **Behavior / correctness** | The library gains `Page.PatchElement`, `Page.SetAfterRedraw`, `ElementEdit`, `host.RedrawBatcher`, and `host.Batch`. A page without retained state refuses every patch and reports false. The devtools dock now opens on Frame. |
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

| page rebuild | Original | First pass | Second pass | Third pass | Fourth pass | Now |
|---|---:|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 12.5 ms | 6.16 ms | 0 ms |
| p50 | 648 ms | 95 ms | 32.1 ms | 12.9 ms | 6.20 ms | 0 ms |
| p95 | 845 ms | 144 ms | 41.9 ms | 15.5 ms | 7.70 ms | 0 ms |
| max | 864 ms | 163 ms | 48.0 ms | 19.2 ms | 8.50 ms | 0 ms |
| total | 28.7 s | 4.6 s | 1.44 s | 0.54 s | 0.265 s | 0 s |

The zero in the last column is a measurement, not a missing number. A click now patches the elements it changed and paints them, so the page's own redraw counter does not move and there is no rebuild left to time. The click script reports 0.0 ms for the rebuild of every one of the 43 clicks. The walk does not become free: it still waits for each painted frame.

Against the original, the drop in the rebuild is 84.0% after the first pass, 95.0% after the second, 98.1% after the third, 99.1% after the fourth, and 100% now.

What the walk costs now is the wait for the painted frame:

| Segment | p50 | p95 | max | average | 43-click pass |
|---|---:|---:|---:|---:|---:|
| press to started | 15.8 ms | 30.9 ms | 32.2 ms | 16.5 ms | 0.68 s |
| press to painted | 51.0 ms | 51.6 ms | 52.3 ms | 51.0 ms | 2.19 s |

The 43-click pass column is the p50 times 43 clicks. So the whole catalog now costs about 2.2 s of frame waiting, against 28.7 s of page rebuild before any of this work. The press to painted time did not move, because the presentation frame of the software compositor dominates it.

The patch against a full redraw, measured with `go test -bench BenchmarkPatchCost -benchtime 30x -count 5 ./examples/css-preview/`, median of five runs, at 1908x999 on an Intel i7-13700HX with Go 1.26.4.

| Path | Time | Allocated | Allocations | Patches applied |
|---|---:|---:|---:|---:|
| Full redraw, card class | 34.5 ms | 20.18 MB | 32188 | |
| Element patch, card class | 6.08 ms | 6.31 MB | 6875 | 30 of 30 |
| Full redraw, state class | 39.2 ms | 19.86 MB | 31656 | |
| Element patch, state class | 4.72 ms | 5.99 MB | 6347 | 30 of 30 |
| Full redraw, same source | 27.2 ms | 13.87 MB | 25304 | |

A separate run of the click walk under a virtual X server, three passes of 43 clicks, gave 17.2 ms and 51.2 ms on p50 for press to started and press to painted, and 2.37 s per 43-click pass, with a redraw of 0.0 ms on every click.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/v0.0.3/css-optimization/pr-css-preview-optimization.md`

## Follow-ups (out of scope)

A review of this branch found the items below. None of them were fixed in this PR.

Two findings from the same review were fixed here, because they were about the files this PR already touches: the generated site bundle was stale, and the recorded patch cost numbers no longer matched the code.

A follow-up on `master` has since fixed most of the list below. The items that are still open are at the end of this section.

### Fixed after the merge

- A refused patch left the parsed tree ahead of the display. The patch is now atomic: a failure drops the cached tree, so the next redraw parses the source again and the page shows the source's state. `internal/page/page_patch_element.go`
- A patch did not clear the cached PNG, so `Page.PNG()` and `GET /frame.png` served the pre-patch picture. The splice now drops it. `internal/page/page_patch_splice.go`
- A patch did not rebuild the accessibility tree. The splice now puts the edits into the tree and retires the document. `internal/page/page_patch_splice.go`
- The patch marked only the element's border box dirty. It now marks the changed operations dirty the way a redraw does, so ink outside the box is repainted. `internal/page/page_patch_splice.go`
- `Cache.SetText` follows the engine's `textContent` semantics and drops every child. The docs now say so, on the cache method and on `ElementEdit`. `internal/render/cache_edit.go`
- The cached pseudo flags were fixed per cache while a patch re-collected the sheets. The flags now follow the styled document the cache holds. `internal/render/cache.go`
- The batch's final paint cleared the pending redraw flag an earlier batch step asked for. The paint claim is now honored only when the render key is unchanged, so a focus or hover state still gets its layout. `internal/page/page_click_paint.go`
- A `position: fixed` box lost the hit test to a content box painted under it. The hit test now follows the paint order, so the fixed layer wins, and the bitmap fallback rebuilds the hit map. `internal/page/page_fixed_hit.go`
- The oversize pixel budget sat in the shared helper, so the scaled and fitted paths lost their zoom scale for a large page. The budget moved to the partial-repaint call site and became a ratio against the window area. `internal/window/replay_oversize.go`
- The composition fallbacks applied under additive composition too, so a colour missing from a keyframe saturated. The fallback is now the neutral contribution for the additive modes and the underlying paint only for a replacing one. `internal/animation/compose.go`
- The transition rebind only checked the length of the captured slice. It now verifies the resolved operations still belong to the box, so a paint captured on an earlier display is never written back. `internal/csstransition/bind_to.go`
- The chunk signature folded a per-layout pointer, so a grouped operation missed on every redraw. It folds the group's creation-order identity instead. `internal/replay/chunk_sig_op.go`
- The Play All clock was stamped before the redraw, so animations started with their duration spent. Each row now takes its origin from the drain clock. `examples/css-preview/catalog_play_all.go`
- `syncTree` restored each started row with its own patch, and every patch re-laid out the whole document. One grid patch now carries the state of every row. `examples/css-preview/catalog_sync_grid.go`
- The Play All queue grew on every press. The queue holds one entry per row, drops the rows that leave the window, and the tick parks when nothing is queued. `examples/css-preview/player_queue.go`
- The catalog geometry was measured once, so a resize left the scroll range stale. The tick re-measures once the size settles. `examples/css-preview/player_measure.go`
- The walk `/box` route read the page boxes from the HTTP goroutine without the walker mutex. The game loop now publishes a snapshot. `examples/css-preview/walk_state.go`
- A failed final draw dropped the queued redraw, and the deferred end of a batch swallowed a travelling panic. Both are fixed, with tests. `internal/host/batch.go`
- Five Go files were over the 2,000-character limit. All five are split, and `make size-check` now measures characters the way `AGENTS.md` says. `scripts/check-file-size.sh`
- The missing test coverage for the refusal branches, the dirty region, the cache commit contract, the nested batch, and the chunk group is now written. Three benchmark and parity tests that passed trivially are repaired.
- The documentation findings are fixed: `PROFILING.md` describes the shipped windowed design, the showcase images carry their real dimensions, the reports are linked from the feature index, and the preview clip has a pause control and respects the reduced-motion preference.

### Still open

- Text set through a patch does not go through the emoji pass, so an emoji in patched text paints as text. The engine's tree API only writes a text node, so the image cannot be built inside the tree. The limitation is documented on `ElementEdit` and in the feature index.
- The scaled and fitted draw paths have only the device limit, so a very large page on the zoom path can allocate a large content buffer.
- The chunk key still carries a per-plan sequence, so two displays that live at once write the same slots. The signature check keeps the pixels right.
- An inline `<style>` element whose text changed keeps its old parsed sheet, because the engine keys the inline sheet by node.
- `examples/go.mod` still pins Blinkless `v0.0.2`. The examples module already lagged behind the published tag before this branch, so the fix belongs to the release process that bumps the examples require.

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
| `.js` | 4 | 906 | 289 |
| `.jsx` | 2 | 83 | 55 |
| `.md` | 12 | 1241 | 51 |
| `.mod` | 1 | 1 | 1 |
| `.py` | 1 | 275 | 0 |
| `.sum` | 1 | 8 | 0 |
| **Total** | **175** | **7598** | **746** |
