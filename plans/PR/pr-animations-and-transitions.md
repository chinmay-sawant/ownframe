## Summary

Make CSS animations and transitions run automatically in the page loop. A new `internal/motion` runtime binds declarations after every display publish, samples them on the window tick, writes composited paint directly, routes geometry through the engine's override relayout, and reports the damage. Retained compositor layers, reusable viewport scroll buffers, and per-frame cache warmup cut scroll frame gaps in the 33-card gallery, and the example and docs let a reader check the shipped behavior against the declarations.

## Motivation / context

- Plans: `plans/v0.0.3/animations-and-transitions/` (11 animation phases, 5 transition phases, 5 example phases, plus `coverage.md` and the priority-0 engine plan).
- Engine work: `chinmay-sawant/blinkless` branch `feature/animations-and-transitions` supplies the three primitives this branch consumes: `layout.RelayoutElementWithOverrides`, `layout.RestampTransforms`, and `layout.RestyleInlineSVG`. Blinkless merged that branch as #25, the race fixes as #26, and the copyless transform accessor as #27. The pin moves to `v0.0.4-0.20261025105255-96d7b9d95338`.
- Before this branch the runtime packages existed but the page did not start motion by itself; the examples that animated drove their own clocks. `documentation/css-motion.md` describes the caller-driven path that remains available for Go-side motion through `internal/frame`.

## Changes

### Automatic motion driver

- Add `internal/motion` (31 files): a runtime that binds keyframe players per element, samples once per frame, and returns damage. It does not import page, window, or replay.
- Wire the driver into the display commit and the page tick. `commitDisplay` captures transitions against the old display, commits the new one, then binds motion. `runTick` samples motion before the `SetTick` callback, then caret blink, then the 3D restamp pass (`internal/page/page_draw.go:172-186`, `internal/page/page_tick.go:43-97`).
- Step transitions before animations so an animation writing the same field wins. An active, paused, unsampled, or filling animation owns the field and cancels the competing transition (`internal/motion/step.go:1125-1143`, `internal/page/motion_wire.go:101-137`).
- Capture transitions from the paint the old display presented, so an interruption reverses from the screen rather than the declared base. A reversal shortens the duration by the CSS reversing factor and starts from the presented value (`internal/motion/transition.go:207-213`, `:806-894`).
- Add `Page.Probe().Motion()` returning `MotionStats`: bound, active, sampled, restamp, rebind, error, gap, and transition counters. Add `Page.Window().SetAnimations`, `Animations`, `SetTransitions`, and `Transitions`; both drivers are on by default (`page_probe.go:17-28`, `page_window_door.go:25-63`).

### What animates

- Composited fields: `transform` (2D) writes the operation matrix and `opacity` writes element opacity.
- Paint: `background-color`, `accent-color`, `color`, `text-decoration-color`, `text-emphasis-color`, and the border, outline, column-rule, and row-rule colours write only the nearest owner's fill, text, or stroke operations. Keyframes never write `W`, `H`, or `Alpha` (`internal/cssanimation/keyframes.go:26-45`, `internal/cssmotion/paint.go:12-19`).
- 3D and perspective samples leave the 2D op untouched and request `layout.RestampTransforms`. The page turns `Config.RetainState` on by itself and owes one redraw before the first restamp (`internal/motion/restamp.go:2-26`, `internal/page/motion_wire.go:292-342`).
- Layout: width/height and logical extents, min/max extents, inset, margin, padding, gap, flex basis/grow/shrink, border widths, border radii, font-size, line-height, letter-spacing, and word-spacing go through `layout.RelayoutElementWithOverrides`. The budget is one element relayout per frame (`internal/cssanimation/layout_values.go:193-270`, `internal/motion/relayout.go:10-15`).
- Inline SVG `fill`, `stroke`, the two opacities, `stroke-width`, `stroke-dasharray`, and `stroke-dashoffset` go through `layout.RestyleInlineSVG`, at most one restyle per frame (`internal/motion/svg.go:1-33`).
- Report unsupported properties as gaps instead of silently freezing: `filter`, `backdrop-filter`, `clip-path`, mask position/size, `mix-blend-mode`, `isolation`, the individual `translate`/`rotate`/`scale` longhands, shadows, gradients, grid templates, and shape properties (`internal/cssanimation/keyframes.go:57-77`, `internal/cssanimation/layout_values.go:166-188`).

### Damage and retained layers

- A bind returns at most 24 damage rects, coalescing within 8 px and collapsing to one bounding box when full. A step adds only the union of the before and after bounds, skips offscreen entries, and skips elements the compositor promoted (`internal/motion/damage.go:9-48`, `internal/motion/step.go:96-105`).
- Promote a visible element with an active 2D transform or opacity track whose base and delta invert. An element that also writes colour, layout, or SVG paint stays direct, and a transition plus animation on one element merge into one layer entry (`internal/motion/layers.go:106-233`).
- The replay renderer keeps up to 16 layers and 32 Mi pixels with LRU eviction. A layer whose content signature changes on two consecutive frames is demoted and blocked for 60 frames. A held layer survives its rest pose (`internal/replay/layer_cache.go:11-25`, `internal/replay/layer.go:117-226`).
- Composite one texture blit per layer on both the tile and viewport paths, and skip the operations the layer owns in the background replay. Promotion, demotion, or a moved layer box invalidates the cached viewport background because overscan can hold a stale unpromoted pose (`internal/replay/layer_paint.go:29-159`, `internal/window/replay_viewport.go:48-57`).

### Scroll buffers and cache warmup

- A page taller than three viewports keeps one window-sized viewport buffer plus up to 256 px of overscan instead of the content-sized tile grid (`internal/window/replay_oversize.go:9-15`).
- Scrolling across an overscan edge copies the rendered overlap into a second reusable image, swaps the buffers, and replays only newly exposed strips. Dirty rects are applied to the copied pixels too. A disjoint jump rebuilds, and both buffers are disposed on resize or path change (`internal/window/replay_viewport_move.go:10-39`, `internal/window/replay_viewport.go:68-118`).
- Warm one viewport-wide band of at most 128 px per frame into a reused scratch image. A page beyond 12 Mi pixels warms only the viewport neighbourhood, and the idle pacer stays awake until the sweep finishes (`internal/window/replay_viewport_prewarm.go:11-94`).
- Reuse hit-test list storage across display commits and 3D re-projection, keep the ID index as pointers into the current box slice, and skip the DevTools box snapshot when nothing is hovered or pinned (`internal/page/hit/hit.go`, `internal/window/devtools_life.go:28-33`).

### DevTools and drawing policy

- Apply dock geometry in `devSync` before `resize`, so a pointer close cannot switch geometry halfway through an update (`internal/window/devtools_own.go:44-52`, `devtools_sync.go:7-30`).
- Forward presses outside the panel to the page when the picker is off, give the resize edge its own cursor, and clear picker, hover, and pinned state on close (`internal/window/devtools_pick.go:69-185`, `devtools_life.go:44-57`).
- Disable VSync in every window constructor and in the idle wake-up. Active drawing is uncapped; animation timing and clean-page idle pacing are unchanged. `NoVsync` was added and removed inside the branch, so the final tree has no such option; `WindowOptions.NoIdlePace` remains (`internal/window/game.go:17-18`, `idle_frame.go:45-62`).

### Examples

- Rewrite `examples/animations` as a 33-card, 9-section CSS gallery: timing, direction, delay and fill, iteration, 2D transforms, 3D transforms, perspective, paint, and path/vendor alias. Every card has one stage, a property name, a CSS snippet, an Expect line, and a Play button. A selected card plays alone; Replay restarts it; the header button pauses or resumes all cards (`examples/animations/animations.html`, `examples/animations/page.go:37-109`).
- Open the DevTools dock by default with Frame timings, add `-dev`, `-performance`, `-profile-seconds`, `-web`, `-addr`, size, and `-monitor` flags, and add an opt-in native scroll profile plus `BenchmarkAnimationTick` (`examples/animations/main.go:19-26`, `profile.go`, `benchmark_test.go`).
- Add `examples/motion-conformance`, a CSS-only page with 24 rows and three transition toggles.
- Migrate `examples/css-preview` off its hand-driven player stack. Click state now lives in template data, a click publishes one `SetData` plus `Redraw`, and restart uses duplicate `@keyframes` blocks. Nine cards keep a manual sample: six geometry transitions and three trigger cards (`examples/css-preview/demo.go:44-82`, `manual_motion.go:3-88`).
- Give the Wispr dashboard sidebar `background-color` a 120 ms transition and sample it mid-run in a test.

### Documentation, plans, and dependencies

- Add `documentation/css-motion.md` and update `features.md`, `frames.md`, `repaint.md`, `optimizations.md`, `performance.md`, `devtools.md`, `testing.md`, `window.md`, `compatibility-matrix.md`, and `AGENTS.md`.
- Record the branch status under `plans/v0.0.3/animations-and-transitions/`, including the phase plans, `coverage.md`, and the example audit.
- Move `github.com/chinmay-sawant/blinkless` to `v0.0.4-0.20261025105255-96d7b9d95338` in `go.mod` and `examples/go.mod`, matching blinkless master after #25, #26, and #27. There is still no `replace` for blinkless.
- Shrink `scripts/example-internal-allowlist.txt` from 40 rows to 16 and extend `scripts/check-graphics.sh` and `scripts/test_graphics_gate.py` with the three viewport GPU tests, for eight required invocations.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | One desktop session (1908x999, dock open, VSync off): `BenchmarkAnimationTick` median fell from 468,145 to 275,957 bytes/tick and from 0.748 to 0.603 ms/tick. Median scroll frame-gap p99 fell from 164.8 to 42.9 ms and the worst gap from 444.0 to 58.1 ms across three interleaved native runs. The first one-second heap sample fell from 135.8 to 54.6 MiB. These are single-machine measurements, not a gate; frame gaps include scheduling and driver waits, and 50-80 ms gaps remain. |
| **Memory** | The motion runtime allocates zero bytes per live step, per layer read, per timeline sample, and per effect-table step (asserted with `testing.AllocsPerRun`). Layer textures are capped at 16 entries and 32 Mi pixels, about 128 MiB. The viewport path holds two window-sized images plus overscan, and keeps warmup scratch images alive. |
| **Behavior / correctness** | CSS `@keyframes` and `transition` declarations now advance without application code. Automatic capture, reversal, interruption, timing, fill, direction, composition, ranges, and named scroll/view timelines are covered. Triggers are resolved, but three `css-preview` trigger cards keep a manual sample. Geometry transitions relayout through the engine override path. Discrete transitions cover `visibility`, `content-visibility`, `z-index`, and `position` at the midpoint. `display`, `order`, grid-line integers, and the generic `inset` shorthand still snap. |
| **API / CLI** | Additive root API: `Page.Probe().Motion()` with `MotionStats` (a documented alias), and `Page.Window().SetAnimations`/`Animations`/`SetTransitions`/`Transitions`. No existing signature changes. New example flags are listed above. |
| **Dependencies** | The Blinkless pin moves to `v0.0.4-0.20261025105255-96d7b9d95338`, the blinkless master commit after #25, #26, and #27. No other module changes and no new external dependencies. |
| **Binary size / build time** | Adds `internal/motion`, a rewritten `animations` example, and a new `motion-conformance` example. No new C or wasm artifacts. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None in Go signatures | No source migration needed. |
| VSync is now off for every desktop and mobile window constructor | Active drawing is uncapped. An application that relied on display-refresh pacing has no public switch back; `WindowOptions.NoIdlePace` only affects clean-page idle pacing. Documented in `documentation/window.md:63`, `:94-95`. |
| Some transitions still snap | `display`, `order`, grid-line integers, and the generic `inset` shorthand do not interpolate; they take the destination at the start or midpoint and report a skip reason. |

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports. `MotionStats` is the one added alias, recorded as plan item API-25 and allowed by `alias_allow_test.go:26`.
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows. `MotionStats` and the switch doors are the public API.
- [x] No new example `ownframe/internal/` imports. `make internal-import-check` shrinks the allowlist from 40 to 16 rows.
- [x] No synonym names.
- [x] New exported methods return named types in package `ownframe`: `Motion() MotionStats`, `Animations() bool`, `Transitions() bool`. No `host.Stats` or `*state.Store` leaks.
- [x] Library code panics only for invariants; the new bind, restamp, layout, and SVG paths return errors and record them in `MotionStats.LastError` / `RestampError`.
- [x] `context.Context` stays first where I/O or cancel applies; the restamp and override paths use the caller's context.
- [x] No new examples resembling Telegram, Teams, Spotify, or Wispr. `animations` and `motion-conformance` are CSS property galleries.

## Test plan

- [x] `make test`
- [x] `make lint` using the repository's pinned `golangci-lint` v2.15.0
- [x] `go test -race -p 1 ./internal/window`
- [x] `make build`
- [x] `make module-check`
- [x] `make graphics-check` with all eight required GPU tests (recorded on `c25b869e`; the gate needs a GPU host)
- [x] `make docs-check plans-check examples-check`
- [x] `make internal-import-check`
- [x] `make site-check` and `npm --prefix frontend run check:site`
- [x] `go test -v -p 1 ./examples/animations/...`
- [x] `cd examples && go test -count=1 ./motion-conformance/...`
- [x] `go test -race -count=10 -run 'TestSecondStartRestartsAManualTransition|TestSecondStartRestartsATriggerCard' ./examples/css-preview`
- [x] `go test ./internal/replay ./internal/window`
- [x] `BenchmarkDrawRectScan` within 1.5x of `origin/master` (1.26x on the development machine)

### Commands

```sh
make test
make lint
make build
make module-check
make graphics-check
make docs-check plans-check examples-check
make internal-import-check
make site-check
go test -v -p 1 ./examples/animations/...
go test -race -p 1 ./internal/window
cd examples && go test -count=1 ./motion-conformance/...
go test -race -count=10 -run 'TestSecondStartRestartsAManualTransition|TestSecondStartRestartsATriggerCard' ./examples/css-preview
go test ./internal/replay ./internal/window
go test -bench=BenchmarkDrawRectScan -benchmem -count=5 ./internal/replay
```

Focused coverage in the branch:

- Motion: `TestRuntimeStepAllocFree`, `TestPromotedLayerStopsStepDamage`, `TestLayerEntriesReadAllocFree`, `TestStepDoesNotAllocate`
- Transitions: `TestInterruptShortensTheReverseByProgress`, `TestInterruptWithEaseMatchesBrowserNumbers`, `TestAnimationPrecedenceCancelsARunningTransition`, `TestSnapshotToCurrentValueCancels`, `TestLayoutAnimationWinsOverWidthTransition`
- Engine paths: `TestStepLayoutRunsOneElementPerFrame`, `TestLayoutBudgetRunsOneElementPerFrame`, `TestRotateYAnimationRestampsTheProjectionOnEveryTick`, `TestNamedViewTimelineSamplesEntryCoverExit`
- Window: `TestPartialFrameAddsLayerDamage`, `TestViewportCompositesRetainedLayer`, `TestViewportScrollPixels`, `TestViewportAnimatedReturnPixels`, `TestViewportDirectAnimationReturnPixels`, `TestViewportExposedCoversOnlyNewPixels`, `TestViewportWarmupEventuallyCoversThePage`
- Examples: `TestNativeDelayAndFillReplay`, `TestNativeScrollProfile` (both opt-in on a GPU host), `BenchmarkAnimationTick`

## Screenshots / sample output

`go run ./examples/animations` opens the 33-card gallery with the DevTools dock on the Frame tab. Each card shows its CSS snippet and Expect line, and a bottom-right button reads Play, then Replay once selected. Play all resumes shared playback; Pause all freezes it. `go run ./examples/motion-conformance` opens the CSS-only conformance page, and `go run ./examples/css-preview` opens the 43-card property catalog.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] `enhancement` and `documentation` labels applied.
- [x] Related issues checked; none are linked.
- [x] Filled body committed under `plans/PR/pr-animations-and-transitions.md`.

## Follow-ups (out of scope)

- `display` entry/exit cross-fades, `order`, grid-line integers, and the generic `inset` shorthand still snap. Morphing children under `display` needs retention work that this branch does not ship.
- Keyframed `filter`, `backdrop-filter`, `clip-path`, mask position/size, `mix-blend-mode`, and `isolation` are reported as gaps.
- `transform-origin` keyframes, nested timeline scopes, and a named scroll-container timeline are not implemented.
- A promoted layer uses normal blending only; `mix-blend-mode` is not read.
- `Serve` still shows a picture page and does not tick, so automatic motion is a native-host behavior.
- Documentation cleanup: the parent animations plan still says "planning" while the track READMEs say landed; `compatibility-matrix.md:1083-1086` still describes caller-driven transitions; `examples/css-preview/PROFILING.md` and `PATCH_COST.md` still describe the deleted patch-cost path.
- The documentation behavior audit found more drift than the CI gates cover. `documentation/css-motion.md` and `features.md` still say geometry, discrete, layout-keyframe, and SVG-paint animation are missing or next-phase, while the runtime ships them. The `MotionStats` table lists 23 of 35 fields. The animations example is described as having no tick callback, but a Play or Replay click registers one for a frame. `plans/v0.0.3/animations-and-transitions/` indexes and checklists still read "Planned" under a landed track. These need a docs pass; no gate checks prose against code.
- No checked-in performance gate exists for tick time, scroll gaps, or FPS. The allocation assertions pin storage reuse only.

## Reviewer checklist

- [ ] Behavior matches summary and test plan.
- [ ] No unrelated changes in diff.
- [ ] Public API changes are documented.
- [ ] No new root aliases, synonym methods, or example `internal/` imports.
- [ ] New exported methods return named `ownframe` types.
- [ ] New rules have fixture coverage when applicable.
- [ ] PR has assignee and labels.
- [ ] Related issues use correct Closes/Relates keywords.
- [ ] No secrets or generated artifacts committed.
- [ ] Diff-stat-by-extension table pasted at the bottom.

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 2 | 88 | 1 |
| `.go` | 364 | 37171 | 3781 |
| `.html` | 4 | 1730 | 510 |
| `.js` | 2 | 144 | 134 |
| `.json` | 1 | 1 | 0 |
| `.md` | 44 | 5873 | 119 |
| `.mod` | 2 | 2 | 2 |
| `.py` | 1 | 3 | 0 |
| `.sh` | 1 | 3 | 0 |
| `.sum` | 2 | 4 | 4 |
| `.txt` | 1 | 5 | 29 |
| **Total** | **424** | **45024** | **4580** |
