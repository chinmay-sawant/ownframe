# perf(repaint): damage list, tiled replay, idle pacing, and experimental patching

Filled copy of `skills/PR/PR_TEMPLATE.md`, saved before opening the PR.

- Base: `origin/master` (`11f309f`)
- Head: `perf/damage-list-experimental` (`6c22a43`, pushed, working tree clean at time of writing)
- Commits: 25
- Diff: 646 files, +25986 / -1604

---

## Summary

This branch cuts per-frame repaint work to the boxes that actually changed, parks idle windows instead of spinning, and scaffolds the experimental retained-DOM patching path. Damage-list repaint, tiled replay, raster and text caches, compositor fast paths, idle pacing, devtools dock clipping, CSS masking, Spotify partial updates, the Digital Wellbeing companion, and grouped docs navigation ship together. The branch stays additive on the public API; the only behavior change to know about is idle windows presenting less often.

## Motivation / context

- The departure-board probe left one union rect per frame: the refresh bar plus the cascade band unioned to the full board, and rotation still called `Redraw` at 5.7–13.5 ms.
- Large pages were display-list bound with no per-operation damage tracking, no tile culling, and no baked raster reuse; ticking pages repainted in full.
- Idle windows spun at full vsync with nothing to draw.
- Plans: `plans/v0.0.2/performance/`, `plans/v0.0.2/damage-and-invalidation/` (audit-2026-10-12 plus phase-01 through phase-06), `plans/v0.0.2/experimental/` (01–05, all checklists closed), `plans/v0.0.2/filter-replay.md`, `plans/v0.0.2/subpixel-text.md`, `plans/v0.0.2/multi-process-isolation.md`.
- Issues: none. This repo has no tracker entries for this work.

---

## Changes

### Damage list, partial swap, per-tick opt-in (`1316914`)

- `internal/page/page_dirty_list.go` (new): capped damage list (`maxDirtyRects=8`), rect coalescing, dirty padding, bounding-box collapse past one third of the frame.
- `internal/page/page_dirty_take.go` (new): `TakeDirtyRects() ([]image.Rectangle, bool)`; `TakeDirty()` kept as bbox wrapper.
- `internal/page/page_boxindex.go` (new): id-to-box index for lookups.
- `internal/page/page_frame.go`: `UseFrameDirty()` / `FrameDirty()` consume-once opt-in; the window captures it once per `Update` (Ebiten can `Draw` without `Update`).
- `internal/window`: replay plan / partial / taker / frame-state / blit files; one buffer repaint per rect, blit only planned rects, blit nothing when clean; Ebiten per-frame clear off on desktop; chrome (dock, menu, scrollbar, outlines) reports rects and restores from buffer.
- Departure-board Phase 2: rotation retargets instead of `Redraw`; fixed-width tile grid; dump gains rotate scenario plus rect-count and area report.

### Raster cache, perf ring, dock rows (`063a59b`)

- `internal/replay/raster_*.go` (new): vector raster cache (64 entries, 4 MP total, LRU dispose); bake once, blit per copy; integer positions snap to image origin.
- `internal/window/perf_ring_quantile.go` (new): 120-frame window summarized in one pass with scratch buffer, zero alloc.
- DevTools rows memoize label widths and spans; no per-row padding strings.

### Idle pacing, tiled replay, compositor ops, text rasters (`de786e6`)

- Idle pacing (`internal/window/idle_*.go`, new): `Options.NoIdlePace`; desktop parks in event-wait after 5 clean frames plus 250 ms without input; 125 ms keep-alive lets ticks run; input, deadlines, and a 2-frame dirty burst return to vsync. Phone and wasm unchanged.
- `Page.WantFrameAt` / `WantFrameIn` / `NextFrameAt` plus `host.FrameWanter` schedule idle wakeups.
- Tiled replay (`internal/window/replay_tile*.go`, new): 256 px grid clipped to content; dirty rects repaint touched tiles only and blit damaged tiles only.
- Compositor fast paths (`internal/replay/group_*.go`, `group_blend.kage`, `xform_*.go`): finite 2D transforms via matrix; separable blend and isolation groups buffer into one Kage composite; non-separable modes, transformed grids, rotated glyphs, and oversize scratch keep the bitmap path.
- Text rasters (`internal/replay/text_raster*.go`, new): run baked once per font, content, size, spacing, color, and subpixel origin; blit while the key matches.
- `-perf` flags on the Spotify, Wispr Flow, and departure-board demos wire `SetPerf(true)` plus pprof in `-web -perf`; `scripts/perf-apps.sh` (new) measures the three demos.

### Frame deadlines, dock clip, element chunks, animations, diff rule (`29079c6`)

- Deadlines: page-declared frame times feed the idle waiter; board asks per frame while the flap spins, else the next clock second.
- Dock clip: tile planner and blit clip to the visible strip; remembered clip repaints the hidden strip on close; dead single-buffer path deleted.
- Element chunks (`internal/replay/chunk_*.go`, new): consecutive ops in one element box bake to a chunk image and blit on later damage; re-bake on content-signature change.
- Compositor animations (`internal/frame/anim*.go`, new): translate, fade, scale, and rotate patch the retained op and report old plus new boxes as damage; finished animations go inactive so the window parks.
- Diff rule: replaces the old >8-changed-ops full-frame fallback; merges within an 8 px gap, unions the closest pair past 8 rects, and goes full only when the bounding box covers a third of the frame.

### DevTools dock viewport and picker (`54b33ce`)

- Dock shrinks the viewport: `viewSize`/`viewOf` equals window minus dock and feeds scroll clamp, scrollbars, bitmap viewport, IME, menu, replay scale, and pointer mapping.
- Picker becomes a mode with a title-bar toggle; off means clicks reach controls, on means a click pins the box, opens the Ops tab, and outlines the op.
- Overlay keys gated on `Config.DevTools` / `SetDevTools`; all examples set `DevTools: true`.

### CSS masking and engine pin (`7d3a55d`, `f8f3116`)

- `internal/masking/` (new): mask image build, layer geometry, alpha and luminance, SVG fragments, repeats, and composite ops; GPU replay and the software PNG painter both honor them; `examples/css-masking/` covers the cases with pixel asserts.
- Blinkless pin moves to `c085b3423260` in both modules; the root local `replace` is gone; checksums refreshed; Oct-13 demo baselines recorded in `documentation/performance-baseline.md`.

### Experimental retained DOM and runtime scaffolding (`0950019` plus `5f0f3df`, `af707a5`, `a959d8d`, `260bf9d`, `883fcea`, `e118161`, `554bd2c`)

- `internal/state/`: stable IDs, projected patches, interceptors, change metadata, undo-by-identity, input-vs-background origin.
- `internal/element/`: retained node tree, reconcile and diff, create/move/update/delete, events, mount/update/unmount lifecycle, disposal.
- `internal/displaypatch/`: op diffs, dominate-bounds patching, precomputed bounds, pool recycle, tile merge at op boundaries.
- `internal/styleinvalid/`: per-box dependencies, path invalidation, frame-dirty ranges, text stability.
- `internal/runtimebehavior/`: loading markers, suspense placeholders, transitions, clipping and scroll containers, focus restore, promise errors, visibility and resize observers, patched-vs-repainted counters, watchdog sampling.
- `Page` patch surface: `PatchDisplay`, `PatchOp`, `DisplayPatch`, `LastPatch`, `PatchReport`, op rental pool.
- `a7cb318`: bitmap fallback records a full patch (`Changed=len(ops)`, `Added=len(ops)`, `Regions:1`, `Full:true`) so patch accounting stays consistent.
- Docs: `documentation/migrate-experimental.md`, `runtime-behavior.md`, `styleinvalid.md`.

### Spotify partial updates and logging (`6a874fa`, `0bc75aa`, `b87498e`)

- Seek, elapsed, remaining, credit, and EQ mutate the retained op in place and invalidate only the changed box; `Tick` goes through `UseFrameDirty()`.
- Row select mutates the active flag in place; pagination reuses capacity and avoids `fmt.Sprintf`; progress skips reformatting when unchanged.
- `player/logger.go` (new): timestamped NDJSON action logs with patch report, view state, memory stats, and runtime loading and watchdog fields.
- Tests pin zero-redraw ticks with dirty rects and single-redraw seeks.

### Digital Wellbeing companion and showcase (`0da8dc1`, `aacb127`, `f2404f4`)

- New `examples/digital-wellbeing/`: dashboard, week and day screens, insights, app picker, phone bridge, Android project with UsageStats query, and release APK (~30 MB).
- `showcase.md`: per-app source-code links, APK download links for Telegram, Dino, and Digital Wellbeing, and the new Digital Wellbeing section with five screenshots.
- No Telegram example code changes in this range; only showcase links.

### Docs navigation and plans (`ec245fc`, `6c22a43`)

- `frontend/src/content.js`: sidebar topics grouped into Start here, Build a screen, Run your app, Reference, Advanced topics, and Examples and comparisons; ungrouped pages fall into More topics.
- `frontend/src/main.jsx`: collapsible groups that auto-open on the active page, mobile docs menu, mobile scroll path; `DocPagination.jsx` deleted.
- `documentation/README.md`: flat 40-link list replaced by the Start-here plus three-group index; `features.md` lede rewritten around Blinkless plus Ebitengine.
- `docs/` frontend assets rebuilt; Digital Wellbeing and Teams screenshots added.
- `plans/v0.0.2/damage-and-invalidation/`: README plus audit plus phase-01 through phase-06.

---

## Impact

| Area | Impact |
|------|--------|
| Performance | Static 48-row repaint about 1.0 ms down to about 35 us in the commit probe; board idle CPU 27.4% down to 7.0%, Wispr 7.8% down to 3.3%; complex layout redraw quoted at 306.66 ms down to 34.74 ms in the earlier baseline. Per-frame tickers such as Spotify and Flappy never throttle. |
| Memory | Stress alloc per redraw quoted at 21 MB down to 5.7 MB; complex fixture 86.25 MB down to 9.27 MB; raster cache bounded at 64 entries and 4 MP; viewport path drops content-sized buffers. |
| Behavior / correctness | Oversized and ticking pages render instead of panicking; dirty repaints fall back to full frame past the rect cap or area threshold; bitmap fallback and ticking pages keep full repaint with patch accounting. |
| API / CLI | Additive only: `TakeDirtyRects`, `WantFrameAt`/`WantFrameIn`/`NextFrameAt`, `NoIdlePace`, chunk and animation patchers, patch and retained-node surface, `-perf` demo flags. No released signature removed. |
| Dependencies | Blinkless pin moves to `c085b3423260`; no new direct dependency. Examples carry the audio indirects. |
| Binary size / build time | One 30 MB Digital Wellbeing APK plus Gradle wrapper jar committed; `make build` compiles every package including examples. |

---

## Performance results

All numbers below are quoted from the branch commit bodies, measured on the authors' machines. Ticking pages (Spotify player, Flappy) never throttle under idle pacing.

### Damage list and partial swap (`1316914`, software-renderer probe, same machine)

| Metric | Before | After |
|---|---|---|
| Frames per 10 s | 120 | 760 |
| Alloc per frame | 23–54 MB | 4 KB |
| Live heap | 850 MB | flat 230 MB |
| RSS | 1.06 GB | 401–427 MB |
| Full replays after startup | - | zero |

### Raster cache and perf ring (`063a59b`, departure board 1908x999, Frame tab open)

| Metric | Before | After |
|---|---|---|
| Go heap | 345.1 MB | 20.1 MB |
| RSS | 678.3 MB | 190.7 MB |
| Alloc per sample | 6.1 MB | 2.3 MB |
| Frame p99 | 22.9 ms | 19.2 ms |
| Long frames | 1 | none |

Dock closed: idle alloc 620 KB/s down to 380 KB/s; graphics queue vertex high-water 86.6 MB down to 9.4 MB.

### Idle pacing and text rasters (`de786e6`)

| Metric | Before | After |
|---|---|---|
| Board idle CPU | 27.4% | 7.0% |
| Wispr dashboard idle CPU | 7.8% | 3.3% |
| 48-row static repaint | 1.0 ms, 873 KB, 4512 allocs | 35 us, 27 KB, 144 allocs |

### Frame deadlines, dock clip, chunks, animations (`29079c6`)

| Metric | Before | After |
|---|---|---|
| Idle presents | 8/s | 1–2/s |
| Idle CPU median / mean | 5.0% / 5.1% | 2.0% / 3.6% (20.4% `NoIdlePace` control) |
| Dock-occluded damage, 1908x1080 with 340 px dock | 152,400 px, 6 blits | 50,400 px, 4 blits |
| Damaged card repaint | 8,589 ns, 5 ops, 18 allocs | 4,432 ns, 1 blit, 3 allocs |
| Animated card frame | 170,854 ns, 203 allocs (`Redraw`) | 176 ns, 0 alloc (patch) |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| Idle windows present less often | Pass `WindowOptions{NoIdlePace: true}` to keep the old always-vsync behavior while debugging frame pacing. |
| None other | All other new surface is additive. |

---

## Test plan

- [ ] `make build` (runs `go vet -p 1 ./... ./examples/...`)
- [ ] `make test` (runs `go test -p 1 ./... ./examples/...`)
- [ ] `GOOS=js GOARCH=wasm` compile check (when replay or window surface changed)
- [ ] Departure-board dump and idle probes from the commit bodies
- [ ] Headless screenshot check for the devtools dock and rebuilt docs site

### Commands

```sh
make build
make test
bash scripts/pr-diff-stat.sh origin/master
```

Branch commits report green `make build` and `make test` on their final sources plus board idle and repaint probes; rerun here before merge.

### Local run

```
(TBD - fill after running make build and make test on the merge HEAD.)
```

---

## Screenshots / sample output

- Commit `1316914` probe: 120 frames per 10 s up to 760; alloc per frame 23–54 MB down to 4 KB; live heap 850 MB flat to 230 MB; RSS 1.06 GB down to 401–427 MB.
- Commit `063a59b` probe: heap 345.1 MB down to 20.1 MB; RSS 678.3 MB down to 190.7 MB; frame p99 22.9 ms down to 19.2 ms.
- Commit `de786e6` probe: 48-row static repaint 1.0 ms and 873 KB and 4512 allocs down to 35 us and 27 KB and 144 allocs.
- `showcase.md` and `docs/assets` carry the Digital Wellbeing screenshots and the rebuilt docs site.

---

## Related issues

- None. No tracker entries exist for this work; the plans folders above are the reference.

---

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied (`enhancement`; add `documentation` for the docs and plans)
- [ ] Filled body saved under `plans/v0.0.2/performance/pr-perf-damage-list-experimental.md`
- [ ] Related issues: none to link

Suggested command once the author decides to open the PR:

```sh
gh pr create \
  --base master \
  --head perf/damage-list-experimental \
  --title "perf(repaint): damage list, tiled replay, idle pacing, and experimental patching" \
  --body-file plans/v0.0.2/performance/pr-perf-damage-list-experimental.md \
  --assignee "@me" \
  --label enhancement
```

PR #27 is open for this branch; the body below is kept in sync with `gh pr edit --body-file`.

---

## Follow-ups (out of scope)

1. Per-shell perf hooks instead of package globals.
2. Frame percentiles in the sampler instead of per-frame sorts on the draw goroutine.
3. Wire `RuntimeSnapshot` to the window sampler or drop the empty heap columns.
4. Filter-as-layer needs the engine payload (`plans/v0.0.2/filter-replay.md`).
5. LCD subpixel text blocked on Ebiten grayscale AA (`plans/v0.0.2/subpixel-text.md`).
6. Window-host and page split with shared-memory frame (`plans/v0.0.2/multi-process-isolation.md`).
7. Rerun `perf-baseline.sh` and `perf-dump.sh` to refresh `performance-baseline.md` on one reference machine.

---

## Reviewer checklist

- [ ] Behavior matches the summary and the matrix.
- [ ] No unrelated changes in the diff.
- [ ] Public API additions documented (`TakeDirtyRects`, frame deadlines, `NoIdlePace`, patch surface).
- [ ] New fallback paths have tests (oversize, viewport clip, perf wiring).
- [ ] PR has assignee and labels.
- [ ] No secrets or generated artifacts committed.
- [ ] Diff-stat-by-extension table below matches `bash scripts/pr-diff-stat.sh origin/master`.

---

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh origin/master` on 2026-10-15.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.apk` | 1 | Binary | Binary |
| `.bat` | 1 | 92 | 0 |
| `.css` | 3 | 71 | 42 |
| `.go` | 559 | 20738 | 1143 |
| `.gradle` | 4 | 46 | 0 |
| `.html` | 4 | 927 | 3 |
| `.jar` | 1 | Binary | Binary |
| `.java` | 1 | 336 | 0 |
| `.jpg` | 2 | Binary | Binary |
| `.js` | 2 | 437 | 160 |
| `.jsx` | 2 | 56 | 50 |
| `.kage` | 1 | 86 | 0 |
| `.md` | 36 | 2561 | 200 |
| `.mod` | 2 | 2 | 2 |
| `.png` | 3 | Binary | Binary |
| `.properties` | 2 | 8 | 0 |
| `.sh` | 2 | 329 | 0 |
| `.sum` | 2 | 4 | 4 |
| `.svg` | 1 | 9 | 0 |
| `.webp` | 13 | Binary | Binary |
| `.xml` | 1 | 30 | 0 |
| No extension | 3 | 254 | 0 |
| **Total** | **646** | **25986** | **1604** |

---

## Commit list

| Commit | Subject | Files | Insertions | Deletions |
|---|---|---|---|---|
| `ec245fc` | docs(plans): add the damage-and-invalidation workstream | 9 | +924 | -0 |
| `1316914` | perf(repaint): damage list, partial swap, and the per-tick frame opt-in | 105 | +3129 | -531 |
| `063a59b` | perf(repaint): raster cache, allocation-free perf ring, lighter dock rows | 34 | +801 | -170 |
| `679b7b9` | examples: add the -perf flag to the spotify and wispr demos | 4 | +15 | -3 |
| `de786e6` | perf: idle frame pacing, tiled replay, compositor ops, text rasters | 120 | +5227 | -321 |
| `7d3a55d` | feat(css-masking): replay masked display groups | 39 | +1019 | -50 |
| `f8f3116` | chore(ownframe): update Blinkless pin and record demo baselines | 5 | +26 | -8 |
| `29079c6` | experimental: frame deadlines, dock clip, element chunks, animations, diff rule | 121 | +4411 | -353 |
| `54b33ce` | feat(devtools): dock-aware page viewport, picker mode, and op selection | 86 | +531 | -143 |
| `e118161` | Add experimental-pro phase checklist plans | 6 | +126 | -0 |
| `554bd2c` | Updating the paths | 6 | +0 | -0 |
| `0950019` | Add experimental reactive retained DOM and runtime behavior scaffolding | 72 | +3215 | -50 |
| `6a874fa` | Apply Spotify playback and progress partial-update changes | 8 | +218 | -55 |
| `0bc75aa` | Add local Spotify action log dumps | 5 | +75 | -0 |
| `5f0f3df` | Integrate remaining experimental and Spotify improvement work | 36 | +1353 | -25 |
| `af707a5` | Complete remaining experimental checklist items | 27 | +924 | -57 |
| `a959d8d` | Mark remaining experimental checklist items complete | 2 | +10 | -10 |
| `260bf9d` | Apply experimental verification fixes and checklist accuracy updates | 6 | +34 | -8 |
| `883fcea` | Implement final Phase1 metadata and state transition coverage | 14 | +525 | -16 |
| `b87498e` | Add Spotify performance optimizations and logging instrumentation | 27 | +931 | -125 |
| `a7cb318` | Record a full patch when a page falls back to a bitmap | 1 | +6 | -0 |
| `0da8dc1` | feat: add Digital Wellbeing Android companion | 29 | +1345 | -0 |
| `aacb127` | feat: expand Digital Wellbeing insights and showcase | 33 | +1804 | -362 |
| `f2404f4` | showcase: add Digital Wellbeing APK | 10 | +20 | -17 |
| `6c22a43` | Group docs navigation and simplify index pages | 10 | +257 | -240 |

Counts come from `git show <sha> --shortstat`.
