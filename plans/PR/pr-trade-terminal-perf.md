## Summary

Brings the Meridian trading terminal example plus three rounds of engine and window performance work to `master`: hover restyle drops from ~180ms to ~20ms on the watchlist, first renders run 6 to 20x faster than the original build, and pixel caches, frame buffers, and motion binding are bounded and shared.

---

## Motivation / context

- Plans: agent-driven performance program on `examples/trade`, no standing plan file; numbers below are measured, not estimated.
- Issues: no linked issue

---

## Changes

### Trading terminal example (`examples/trade`)

- Nine-screen desktop trading terminal (watchlist, markets, portfolio, orders, screener, news, settings, shared topbar, rail, status bar, toast) built from HTML/CSS components over a seeded synthetic dataset; no network, persistence, or live order routing.
- Candlestick chart, portfolio line, allocation ring, and rail icons are DOM markup and inline SVG; all motion is CSS, no `SetTick`.
- Tests cover every screen render, CSS motion binding, watchlist filtering and selection, chart tabs and timeframes, order submission, rail history, settings toggles, and window-size following.

### Redraw and hover path (`internal/page`, `internal/render`)

- Identical-frame `Redraw` short-circuits when the executed source, viewport, focus, hover, theme, and images all match, counting the redraw with no style, layout, or repaint; motion, retained-state debt, and caret blink force rebuilds.
- Hover restyle re-places only the boxes the style diff touches (enter/leave from nothing scopes to the hovered element or its id-bearing scope, budget-checked) instead of relaying out the whole page; full layout remains the fallback.
- The hover patch keeps its change set to the patched operations (stable op identity and a per-element damage box), so the display diff no longer escalates a single-row hover to full-content damage.
- `host.Hoverer` and `internal/window/hover_cursor.go`: the window re-resolves a still cursor on a generation change only while the page still paints the element the window pushed, so a page that moved its own hover is left alone and a hover frame pays one layout instead of two.
- Scrollable content size intersects each box with its recorded ancestor overflow clip, so clipped paint (marquee tape, sheens) no longer extends the document or summons a horizontal scrollbar.

### Motion binding (`internal/motion`, `internal/cssanimation`, `internal/cssmotion`)

- One shared per-display op-ownership index serves all animated boxes through `BindBoxWithIndex`; transition capture and keep share one index pair instead of re-scanning per element per property. Pages with no animations or transitions build no index and pay nothing. Equivalence tests pin indexed results equal to direct scans.
- Finished entrance players mark damage once and rest instead of repainting the full page every frame; hover defers until one frame is committed.

### Replay and window (`internal/replay`, `internal/window`)

- Decoded-image cache is true LRU under the existing 64-entry and 64M-pixel caps.
- Text runs bake one white mask and tint the blit, sharing runs across colors and opacities, with LRU eviction; GPU parity test proves the blit byte-identical.
- Tall-page viewport keeps one window-sized buffer instead of buf plus spare; scaled and fitted draws share one scale buffer; each frame drops retained buffers its path never reads; the warmup scratch is unmanaged.
- Parking the window deallocates every GPU-backed pixel cache, keeps the bounded font parse cache, and returns memory to the OS, then pre-sizes containers on wake.
- Tile grid paints only the visible viewport and repaints skipped tiles on demand when scrolled in.

### Spotify player (`examples/spotify-player`)

- Equalizer and hover motion move from the Go tick to CSS declarations; the tick now paints only seek fill, time labels, and credit from the audio clock, so a stopped track parks the page.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Watchlist hover 178.5ms to 19.9ms (~89% faster); orders hover 123.6ms to 10.8ms (~91%); news hover 97.3ms to 10.6-26ms. Live-loop hover frames pay one layout, Update 141ms to 1.4ms. First renders 6 to 20x faster than the original build (news 157ms to 113ms under identical conditions in the latest round). Trade suite 11.2s to 8.7-8.9s. |
| **Memory** | Headless redraws peak at 34MB heap and retain 6MB; image, text, chunk, and layer stores are capped with LRU eviction; park releases pixel caches; frame buffers cut ~17MB steady state. |
| **Behavior / correctness** | Transformed paint is clipped to ancestor overflow boxes at composite time; abspos flex children overlay instead of consuming line space; stretched abspos resolves per CSS 2.1 section 10.6.4; max-content no longer clamps to the lane. |
| **API / CLI** | `-perf` defaults on for the trade example (`-perf=false` opts out). No library API changes. |
| **Dependencies** | Blinkless pinned to `v0.0.4-0.20261026120705-9d2ba1c7b728` (blinkless PR #28 and #29 merged: first-launch layout work, then stretch-aware subtree relayout). |
| **Binary size / build time** | No new dependencies; trade test binary unchanged in kind. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`make internal-import-check` clean, 16 allowlisted pairs)
- [x] No synonym names (`PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [x] Library code panics only for invariants; new user-facing failure paths return `error`
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's ctx
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr

## Test plan

- [x] `go vet -p 1 ./internal/... ./examples/trade/ ./examples/spotify-player/`
- [x] `go test -p 1 -count=1` on `internal/page`, `internal/window`, `internal/motion`, `internal/replay`, `internal/render`, `internal/cssanimation`, `internal/cssmotion`
- [x] `go test -p 1 -count=1 ./examples/trade/ ./examples/spotify-player/...`
- [ ] `make test` (full gate incl. website lint not run in this pass)
- [ ] `make test-race` (no shared mutable state changes in this pass)
- [ ] `make run` wall time vs baseline (hard &lt; 400ms; soft ±50ms of reference)

### Commands

```sh
go vet -p 1 ./internal/... ./examples/trade/ ./examples/spotify-player/
go test -p 1 -count=1 ./internal/page/ ./internal/window/ ./internal/motion/ ./internal/replay/ ./internal/render/ ./internal/cssanimation/ ./internal/cssmotion/
go test -p 1 -count=1 ./examples/trade/ ./examples/spotify-player/...
make internal-import-check
```

---

## Screenshots / sample output

Hover averages, `Edit().Hover` only, 1600x900 headless:

| Screen | Before | After |
|---|---|---|
| watchlist | 178.5ms | **19.9ms** |
| orders | 123.6ms | **10.8ms** |
| news | 97.3ms | **10.6 to 26ms** |

Orders needed a blinkless fix: the hover patch was refused because a flex-stretched row's kept height differed from its content height, so every orders hover fell back to a full layout. Stretch-aware subtree relayout (blinkless PR #29) accepts the assigned block size and the patch applies.

The live loop also paid layout twice per hover: the window re-pushed a stale hover point on the frame after any generation change. `host.Hoverer` lets the window confirm the page still paints the element it pushed, so one hover frame pays one layout. Update dropped from 141ms to 1.4ms.

---

## Related issues

- Relates to chinmay-sawant/blinkless#28 (perf/example-trade, merged)

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Watchlist hover patch still assembles the display list after the patch; a splice-aware or batched `RelayoutElements` would cut that remaining share.
- Per-screen hover averages include a one-time full layout on first entry to the list (79 to 175ms); steady row-to-row is the number above.
- Hover highlight PNG checks and per-screen ms targets live in a temporary test, deleted before merge; re-add if hover budgets regress.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] No new root aliases, synonym methods, or example `internal/` imports
- [ ] New exported methods return named `ownframe` types (not `host.Stats` / `*state.Store` unless re-exported here)
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 19 | 10988 | 7 |
| `.go` | 102 | 7296 | 417 |
| `.html` | 15 | 1993 | 4 |
| `.json` | 1 | 1 | 0 |
| `.md` | 6 | 197 | 11 |
| `.mod` | 2 | 3 | 2 |
| `.sum` | 2 | 22 | 4 |
| **Total** | **147** | **20500** | **445** |
