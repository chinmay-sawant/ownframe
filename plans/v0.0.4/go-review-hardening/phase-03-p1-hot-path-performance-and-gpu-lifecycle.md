# v0.0.4 - Phase 03: P1 Hot-Path Performance & GPU Resource Lifecycle

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P1
> **Status:** Open (8 / 9 complete; `PERF-UIA-01` blocked on a Windows host)

---

## Overview

This phase removes avoidable per-frame syscalls, GC write-barrier churn, per-frame string parsing, and GPU texture leaks/high-water-mark growth identified in the `window`, `replay`, `motion`, and `ui` hot paths.

---

## 3.1 Per-Frame Syscall & Allocation Elimination

- [x] `PERF-WSL-01` Cache `isWSL()` with `sync.Once` (`internal/window/chrome.go:101-120`, `internal/window/update.go:33`).
  - **Root cause:** `update.go:33` evaluates `isWSLg()` → `isWSL()` on every tick (60–240+ Hz). On non-WSL Linux (where `WSL_DISTRO_NAME` and `WSL_INTEROP` are unset), `isWSL()` executes `os.ReadFile("/proc/version")` + `strings.ToLower` on every frame.
  - **Fix:** Memoize `isWSL()` with `sync.OnceValue` (keeping a test hook if unit tests override WSL detection) and short-circuit `closeRequested.Load()` before checking `isWSLg()`.
  - **Proof:** Benchmark/test verifying `Update()` performs zero filesystem reads after initialization on Linux.
  - **Evidence (done):** Commit `17caf6c9`. `isWSL()` is now `sync.OnceValue(detectWSL)` with a `readProcVersion` test hook; the close stage short-circuits on `closeRequested.Load()` before `isWSLg()`. Test `TestIsWSLReadsProcVersionOnce` (`internal/window/wsl_cache_test.go`): 1000 calls, one `/proc/version` read. Command `go test -count=1 -run 'IsWSL|Close' ./internal/window` exit 0. Benchmark (`BenchmarkIdle*`/`Update`) before/after: pending primary-session run.

- [x] `PERF-LRU-01` Replace O(N) slice shifting in `lruTouch` for `textRasterCache` (and replay LRU caches) (`internal/replay/lru.go:5-15`, `internal/replay/text_cache.go:223-255`).
  - **Root cause:** Every text cache hit linearly scans `order` (up to 256 entries) and runs `copy(order[i:], order[i+1:])` over ~112 B pointer-bearing `textRasterKey` structs (~28 KB through GC write barriers under `c.mu.Lock()`).
  - **Fix:** Replace slice shifting with an O(1) doubly-linked intrusive index list (or monotonic access-tick stamp scanned only on eviction) so cache hits touch zero pointer slices.
  - **Proof:** Benchmark in `internal/replay/` measuring hit latency and `0 allocs/op` before/after on a full 256-entry `textRasterCache`.
  - **Evidence (done):** Commit `17caf6c9`. New `lruOrder[K]` (`internal/replay/lru.go`): per-key atomic access tick, touch is one atomic store (no slice shift, no allocs); only eviction scans for the oldest tick. All five replay LRU caches (text, image, raster, corner, chunk) use it. Tests `TestLRUOrderTouchAndEvict`, `TestLRUOrderTouchAllocsZero` (0 allocs on a 256-key `textRasterKey` order), `TestRasterCacheTouchOnHit`. Command `go test -count=1 ./internal/replay` exit 0; `-race` exit 0. Added `BenchmarkLRUOrderTouch256`; before/after ns/op: pending primary-session benchmark run.

- [x] `PERF-MOT-01` Pre-compile CSS easing functions at bind time instead of re-parsing strings every frame (`internal/motion/easing.go:17-59`, `internal/motion/step.go:201-209`).
  - **Root cause:** `cssEase` returns a closure calling `Ease(value, part)`, re-running `strings.Split` and `strconv.ParseFloat` on `cubic-bezier(...)`, `steps(...)`, and `linear(...)` on every animated frame per property.
  - **Fix:** Parse easing strings once into a compact value struct (kind + 4 float handles / step count / pre-parsed linear stops) when binding animations/transitions, keeping `Ease(value, t)` as a thin wrapper.
  - **Proof:** Benchmark showing `0 allocs/op` per frame for `cubic-bezier(...)`, `steps(...)`, and `linear(...)` sampling.
  - **Evidence (done):** Commit `17caf6c9` + follow-up on `fix/v0.0.4-review-followups`. `compileEase` parses a timing function once into a closure; `Ease(string, t)` goes through a bounded (512) `sync.Map` of compiled samplers (`cachedEase`, documented on `Ease`); `cssEase` binds the compiled sampler. Test `TestCompileEaseMatchesEaseAndAllocsZero` (cubic-bezier, steps, linear(), named, invalid: 0 allocs/op sampling; compiled == Ease). `TestSampleTransformAllocatesNothingFor2D` stays green. Command `go test -count=1 ./internal/motion` exit 0. Motion benchmark delta: pending primary-session run.

- [x] `PERF-MOT-02` Reuse scratch maps/slices across retained-layer frames (`internal/replay/layer_cache.go:122-149, 594-607`, `internal/motion/runtime.go:109-126`).
  - **Root cause:** Every frame with retained layers allocates fresh `candidates`, `keep` map, `active` slice, `ActiveLayerIDs()` slice, and `SetPromotedLayers()` map.
  - **Fix:** Clear and reuse scratch maps/slices on `layerStore` and `motion.Runtime`.
  - **Proof:** Allocation benchmark on steady-state retained-layer animation verifying drop to `0` map/slice allocations per frame.
  - **Evidence (done):** Commit `17caf6c9`. `layerStore` reuses `candScratch`, `keepScratch`, `activeScratch`, `dropScratch`, `idScratch`; `capCandidates` filters in place; `ActiveLayerIDs` returns renderer-owned scratch (valid until next call, documented); `Runtime.SetPromotedLayers` clears and refills its map. Tests `TestActiveLayerIDsReusesScratch`, `TestSetPromotedLayersReusesMap` (0 allocs). Commands `go test -count=1 ./internal/replay ./internal/motion` exit 0. A full steady-state `BeginLayers` allocation benchmark: pending primary-session run (needs the GPU test rig).

- [x] `PERF-IMG-01` Remove `reflect.ValueOf(data).Pointer()` boxing and write-lock on `imageCache.get` hits (`internal/replay/images.go:116, 148-157`).
  - **Root cause:** `reflect.ValueOf(data).Pointer()` boxes the slice header per draw, and `imageCache.get` takes `c.mu.Lock()` on every hit despite `mu` being `sync.RWMutex`.
  - **Fix:** Use `unsafe.SliceData(data)` (or slice-data pointer helper) for zero-alloc identity and use `RLock` + atomic tick (paired with `PERF-LRU-01`) on cache hits.
  - **Proof:** Benchmark on repeated image draw showing reduced allocs and contention.
  - **Evidence (done):** Commit `17caf6c9`. `reflect.ValueOf(data).Pointer()` replaced with `unsafe.SliceData`; `imageCache.get` takes `RLock` and stamps an atomic tick (`lruOrder.touch`). Test `TestImageCacheKeyForRingHitAllocsZero` plus existing image cache tests under `-race`: `go test -count=1 -race -run ImageCache ./internal/replay` exit 0.

- [~] `PERF-UIA-01` Throttle/cache `placeUIA()` Win32 geometry syscalls (`internal/window/a11y_uia_windows.go:77, 99-108`).
  - **Root cause:** `placeUIA()` runs on every `Update()` tick, issuing 3 Win32 syscalls (`ClientToScreen`, `GetClientRect`, `GetDpiForWindow`) and locking `Provider.mu` even when the window has neither moved nor resized.
  - **Fix:** Track last window position/size/device-scale/titlebar-inset and only re-query/publish geometry when changed (or on periodic fallback).
  - **Proof:** Unit test verifying unchanged window state skips `SetGeometry` updates on consecutive ticks.
  - **Evidence (partial / blocked):** Commit `17caf6c9`. Throttle logic done and tested on Linux: `geometryThrottle` (`internal/window/a11y_geometry_throttle.go`) skips `ClientGeometry()` (the 3 Win32 calls) while position, size, scale and title-bar inset are unchanged, with a 120-frame periodic refresh; `placeUIA` uses it. `TestGeometryThrottleSkipsUnchangedWindow` passes (`go test -count=1 -run Geometry ./internal/window` exit 0). `GOOS=windows GOARCH=amd64|arm64 go vet ./internal/window` exit 0 is supporting evidence only. **Blocked:** the native wiring (`placeUIA`, `ebiten.WindowPosition/Size/Monitor` signature) has not run on a Windows host. Reconsideration gate: passing `windows-latest` `native-package` (HYG-CI-01) and `windows-arm64` runs of `./internal/uia ./internal/window`.

## 3.2 GPU Texture Lifecycles & Buffer Bounds

- [x] `GPU-LIFE-01` Bound and shrink `backdropState.scratch` textures and remove per-draw uniform map allocations (`internal/replay/backdrop_filter.go:40-57, 165, 182`).
  - **Root cause:** `acquireScratch` grows 3 unmanaged textures to `(maxW, maxH)` monotonically with no `ebiten.MaxImageSize()` clamp and never shrinks until `SetIdle()`/`Close()` (e.g., 200×2000 + 2000×200 permanently holds 48 MB VRAM), and allocates two 9-key `map[string]any` uniform maps per draw.
  - **Fix:** Clamp to `ebiten.MaxImageSize()`, shrink when high-water area greatly exceeds active demand over N frames, and reuse uniform maps on `backdropState`.
  - **Proof:** Unit test covering aspect-ratio swings and `MaxImageSize` clamping + graphics parity check.
  - **Evidence (done):** Commit `17caf6c9`. `scratchSizing.plan` clamps to `ebiten.MaxImageSize()` (`scratchMaxEdge`, unclamped when no game reports a limit), grows to the union, and shrinks to the window peak after 120 acquires when area exceeds 4x peak; `blurUniforms` reuses the two 9-key uniform maps. Tests `TestScratchSizingClampsAndShrinks` (200x2000 then 2000x200 swing, refusal past MaxImageSize, shrink to 100x100, no steady-state realloc), `TestBlurUniformsReuseMap` (0 allocs). Command `go test -count=1 ./internal/replay` exit 0. `make graphics-check` parity: pending primary-session run.

- [x] `GPU-LIFE-02` Complete GPU buffer disposal on `runWindow` exit and join `paceGame` timer goroutine (`internal/window/run.go:123-135`, `internal/window/idle.go:258-260`).
  - **Root cause:** `runWindow` defers dispose of `titlebar.canvas`, `renderer`, and `dev.panelCache`, but omits `dropTiles()`, `disposeViewport()`, `dropScaleBuf()`, `disposeImage()`, and `bitmapView.img.Deallocate()`; `paceGame.close()` does not wait for the timer goroutine to exit.
  - **Fix:** Add a unified `game.disposeSurfaces()` defer in `runWindow` and wait on a `sync.WaitGroup` in `paceGame.close()`. Also evaluate pooling short-lived offscreen textures in `internal/replay/blend.go:266, 422` and capping `ui/builder.go:289-327` `rendererPool` buffer capacity (e.g., 64 KiB).
  - **Proof:** `leakcheck` + `TestWindow*` lifecycle tests + `make graphics-check`.
  - **Evidence (done):** Commit `17caf6c9` + follow-up on `fix/v0.0.4-review-followups`. `shell.disposeSurfaces()` (title bar canvas, tiles, viewport, scale buffer, image, bitmap view; idempotent) is deferred in `runWindow`; `paceGame.close()` now joins the timer goroutine via `timerDone`. `ui` `putRenderer` resets `r.output.Reset()`, clears `r.stack` and `r.keys` elements so no `Element` or string references stay pinned, and bounds retained scratch (`keys`/`scratch`/`identities` capped at 4096 slots / 64 KiB) via `TestPutRendererBoundsRetainedScratch`. `blend.go:266,422` evaluated: both offscreens are `defer Deallocate`d in the same call, so pooling would only add retained VRAM; left as is. Tests `TestPaceGameCloseJoinsTimerGoroutine`, `TestDisposeSurfacesIsIdempotentOnBareShell`; `go test -count=1 -race -run 'Pace|DisposeSurfaces' ./internal/window` and `go test -count=1 -run 'PutRenderer|Alloc' ./ui` exit 0. Real-window leak run and `make graphics-check`: pending primary-session run.

- [x] `GPU-LIFE-03` Avoid blocking Ebiten's main thread with `time.Sleep` in `paceGame.Draw` (`internal/window/idle.go:313-315`) or document/verify frame-pacing tradeoffs.
  - **Root cause:** `MaxFPS` capping sleeps inside `Draw`, delaying OS input/window event pumping for up to ~33 ms at 30 FPS.
  - **Fix:** Evaluate driving `MaxFPS` via Ebiten's `SetTPS` / skip-draw pacing or short yield without starving input events; if Ebiten requires sleep for unmanaged draw cadence, bound and document behavior with a responsiveness test.
  - **Proof:** Window pacing test confirming input latency under `MaxFPS: 30` remains within budget.
  - **Evidence (done):** Commit `17caf6c9`. Evaluated: Ebiten has no draw-skip pacing hook short of reworking `SetTPS` cadence, so the capped sleep stays, now bounded and documented at the call site (`paceGame.Draw`). `TestLoopWaitNeverExceedsOnePresentInterval` proves the sleep never exceeds one present interval (<= 34 ms at 30 FPS) for any anchor; Update runs at `LoopFPS` (hotter than the cap), so input is sampled between presents. `go test -count=1 -run 'LoopWait|Pace' ./internal/window` exit 0. Caveat: real-window input latency under `MaxFPS: 30` was not measured here (no reachable window).

---

## Closure Gates

- [ ] `P1B-GATE-01` `make test`, `make test-race`, `make lint`, and `make graphics-check` exit `0`.
- [ ] `P1B-GATE-02` `scripts/perf-baseline.sh` shows no regressions (`max-ratio <= 1.5`) and measurable improvements on `textRasterCache`, `motion`, and idle `Update()`.
- [ ] `P1B-GATE-03` Record before/after benchmark numbers beside each closed row above.
