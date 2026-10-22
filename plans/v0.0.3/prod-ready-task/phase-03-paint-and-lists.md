# v0.0.3 - Phase 3 field: Paint, lists, and blinkless

> **Parent:** [README.md](README.md)
> **Status:** open. Does not block the API wrap. 1.0 needs a list **recipe** and either tile+pins or documented `SetScrollWindow`. Sticky in the engine is `[~]`.
> **Estimated effort:** dirty-from-tick is days. Tile+pins is a week-scale window change. Engine sticky is a blinkless project.

---

## Overview

Dirty tiles and idle park are real on forms. They drop out on pinned chrome (`hasViewportOps`) and on ticks that forget `UseFrameDirty`. Long lists pay full template+layout (Teams 500 posts ~7.3 s in `documentation/tricks-tips.md`).

1.0 production for ops tools is: windowed slices in the app, ticks that dirty automatically, and honest CSS. It is not a compositor rewrite ([plans/v0.0.3/optimizations/](../optimizations/README.md) stays planning).

Go practice: measure with a named command and dataset; do not treat `make test` as a benchmark; keep `RetainState` off by default; `UseFrameDirty` becoming default is a behavior change, test it.

## Executive summary

- Op patches from `SetTick` mark dirty rects inside the library.
- Pinned layers paint after the tile blit, **or** `SetScrollWindow` is the documented 1.0 contract.
- List windowing example, 200 rows, layout time recorded.
- blinkless: one pin, prefer a tagged release.

## Current source baseline

| Item | Path | Behavior |
|---|---|---|
| Partial replay bail | `internal/window/replay_partial.go` | `hasViewportOps()` forces full replay |
| Pin detect | `internal/window/replay_order.go` `hasViewportOps` | fixed or viewport-pin ops |
| Tick dirty opt-in | `internal/page/page_frame.go` `UseFrameDirty` | ticks without it replay the whole list (`replay_frame.go`) |
| Telegram pins | `examples/telegram/telegram/pin.go` | `SetScrollWindow` / `SetViewportPinZ` |
| Spotify tick | `examples/spotify-player/player/playback.go` | calls `UseFrameDirty` |
| Patch cost | `examples/css-preview/PATCH_COST.md` | ~6 ms vs ~34 ms full redraw, still copies full op list |
| Engine lock | `internal/render/engine.go` | process-wide mutex |
| blinkless | `go.mod` pseudo-version `v0.0.3-0.20261021155643-c1054ff46202` | same author |

## Checklist

### 3.1 Tick dirty by default

- [ ] `PAINT-01` when `SetTick` patches ops (or `UseFrameDirty` would have been required), the library sets the frame-dirty flag itself. Callers should not remember `UseFrameDirty`. Proof: a test ticks a fill color without calling `UseFrameDirty` and `TakeDirty` / tile path is used. `documentation/frames.md` updated.
- [ ] `PAINT-02` keep `UseFrameDirty` as a no-op or documented alias for one tag if examples already call it. Proof: Spotify/telegram/departure-board still compile.
- [ ] `PAINT-03` ticks that call `Redraw` stay full layout. Proof: existing redraw tests.

### 3.2 Pinned chrome vs tiles

Pick one and record it in this row:

- [ ] `PAINT-04a` implement paint of fixed/pin-z layers after the unfixed tile blit so `hasViewportOps` does not disable partial replay. Proof: telegram-like fixture: scroll dirties message tiles; composer/header still on screen; `hasViewportOps` no longer forces `fullReplay` in `replay_partial.go`. Visual/dirty tests updated.
- [ ] `PAINT-04b` **or** document 1.0: pinned chrome uses `SetScrollWindow` / `SetViewportPinZ` and **leaves the tile path**. `documentation/repaint.md` and `screen.md` say incremental repaint is a forms feature. Proof: those sentences exist; Telegram comments point at the docs.

1.0 may ship 04b. 04a is the product-quality follow-up. Do not leave both `[ ]`. Close one `[x]` and the other `[~]` with pointer.

- [ ] `PAINT-05` `documentation/repaint.md` names `drawReplayPartial` as the window path (see `SHAPE-16`). Proof: file agrees with `internal/window/replay_partial.go`.

### 3.3 Lists

- [ ] `PAINT-06` example recipe: a windowed list (slice window + spacers or `SetScrollWindow`) under `examples/` (new small example **or** a section in `documentation/tricks-tips.md` plus a test in an existing example). Not a virtual DOM. The app owns the slice. Proof: 200 rows of simple rows, layout/redraw time recorded with command, dataset, cold/warm, metric. Target: under 16 ms layout on the same class of machine as Teams notes, or a recorded number with hardware.
- [ ] `PAINT-07` `documentation/tricks-tips.md` Teams 50/200/500 numbers stay as the warning. Link the recipe. Proof: link exists.
- [ ] `PAINT-08` `[~]` engine `content-visibility`. Next gate: blinkless. Do not fake it in 1.0 beyond spacers.

### 3.4 Element patch (behavior already in ERR-*)

- [ ] `PAINT-09` features.md already describes refuse + cache drop. After `ERR-05`, confirm emoji-in-patched-text still documented (no emoji pass on patch). Proof: features.md element patch section.
- [ ] `PAINT-10` `RetainState` remains off by default. Proof: `Config` zero value; a test that patch without retain fails with `error`.

### 3.5 Fallback honesty

- [ ] `PAINT-11` bitmap fallback badge stays. Docs list elliptical radius, missing image, rotated glyph, non-separable blend, transformed grid as fallback triggers (`replay_reject_test.go`). Proof: features.md or screen.md bullet list cites the test file.
- [ ] `PAINT-12` `[~]` half-replay of mixed pages. Reason: all-or-nothing Replayable is the correct gate. Next gate: none for 1.0.

### 3.6 blinkless

- [ ] `PAINT-13` one pin both modules (`MOD-05`). Proof: same version string.
- [ ] `PAINT-14` `[~]` tagged blinkless `v0.0.3` (or later) instead of a pseudo-version. Next gate: blinkless repo release. 1.0 may ship with a pseudo-version if MOD-05 holds, but release notes must name the commit.
- [ ] `PAINT-15` CSS matrix (`DOC-08`) is the support contract. Engine holes are blinkless issues, not silent ownframe workarounds, except the documented `SetScrollWindow` contract.
- [ ] `PAINT-16` `[~]` `position: sticky` / flex+absolute in blinkless. Next gate: engine PR then pin bump. ownframe 1.0 documents the hole.

### 3.7 Process lock

- [ ] `PAINT-17` document `internal/render/engine.go` process-wide mutex: two pages in one process queue. Proof: `documentation/performance.md` or features.md sentence. 1.0 does not remove the lock (blinkless readers).

## Dependencies

- Do not couple this to `plans/v0.0.3/optimizations/` Chromium-phase files. Those are planning only. If a tile+pin change overlaps phase-02 tiling enhancements, pointer from this row to that file and keep **one** active `[ ]`.
- Telegram is the pin fixture; do not require a new clone.

## Risks

- Defaulting frame dirty can keep idle park from sleeping if ticks always run. Idle policy must still park when the tick did not dirty. Test: a no-op tick vs a patching tick.
- 16 ms at 200 rows is hardware-dependent. Record CPU, GOMAXPROCS, and binary (vet vs run).
- Painting pins after blit can clip wrong with devtools dock. Window already clips tiles to the visible strip; include that in PAINT-04a tests.

## Closure gates

Record the exact release command for PAINT-06. Successful execution of `make test` is not proof.

- [ ] `make lint` exit 0
- [ ] `make test` exit 0
- [ ] PAINT-04a or PAINT-04b closed
- [ ] PAINT-01 test without `UseFrameDirty`
- [ ] list recipe numbers pasted here
