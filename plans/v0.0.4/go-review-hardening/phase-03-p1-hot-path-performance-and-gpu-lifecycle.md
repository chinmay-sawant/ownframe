# v0.0.4 - Phase 03: P1 Hot-Path Performance & GPU Resource Lifecycle

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P1
> **Status:** Open (0 / 9 complete)

---

## Overview

This phase removes avoidable per-frame syscalls, GC write-barrier churn, per-frame string parsing, and GPU texture leaks/high-water-mark growth identified in the `window`, `replay`, `motion`, and `ui` hot paths.

---

## 3.1 Per-Frame Syscall & Allocation Elimination

- [ ] `PERF-WSL-01` Cache `isWSL()` with `sync.Once` (`internal/window/chrome.go:101-120`, `internal/window/update.go:33`).
  - **Root cause:** `update.go:33` evaluates `isWSLg()` → `isWSL()` on every tick (60–240+ Hz). On non-WSL Linux (where `WSL_DISTRO_NAME` and `WSL_INTEROP` are unset), `isWSL()` executes `os.ReadFile("/proc/version")` + `strings.ToLower` on every frame.
  - **Fix:** Memoize `isWSL()` with `sync.OnceValue` (keeping a test hook if unit tests override WSL detection) and short-circuit `closeRequested.Load()` before checking `isWSLg()`.
  - **Proof:** Benchmark/test verifying `Update()` performs zero filesystem reads after initialization on Linux.

- [ ] `PERF-LRU-01` Replace O(N) slice shifting in `lruTouch` for `textRasterCache` (and replay LRU caches) (`internal/replay/lru.go:5-15`, `internal/replay/text_cache.go:223-255`).
  - **Root cause:** Every text cache hit linearly scans `order` (up to 256 entries) and runs `copy(order[i:], order[i+1:])` over ~112 B pointer-bearing `textRasterKey` structs (~28 KB through GC write barriers under `c.mu.Lock()`).
  - **Fix:** Replace slice shifting with an O(1) doubly-linked intrusive index list (or monotonic access-tick stamp scanned only on eviction) so cache hits touch zero pointer slices.
  - **Proof:** Benchmark in `internal/replay/` measuring hit latency and `0 allocs/op` before/after on a full 256-entry `textRasterCache`.

- [ ] `PERF-MOT-01` Pre-compile CSS easing functions at bind time instead of re-parsing strings every frame (`internal/motion/easing.go:17-59`, `internal/motion/step.go:201-209`).
  - **Root cause:** `cssEase` returns a closure calling `Ease(value, part)`, re-running `strings.Split` and `strconv.ParseFloat` on `cubic-bezier(...)`, `steps(...)`, and `linear(...)` on every animated frame per property.
  - **Fix:** Parse easing strings once into a compact value struct (kind + 4 float handles / step count / pre-parsed linear stops) when binding animations/transitions, keeping `Ease(value, t)` as a thin wrapper.
  - **Proof:** Benchmark showing `0 allocs/op` per frame for `cubic-bezier(...)`, `steps(...)`, and `linear(...)` sampling.

- [ ] `PERF-MOT-02` Reuse scratch maps/slices across retained-layer frames (`internal/replay/layer_cache.go:122-149, 594-607`, `internal/motion/runtime.go:109-126`).
  - **Root cause:** Every frame with retained layers allocates fresh `candidates`, `keep` map, `active` slice, `ActiveLayerIDs()` slice, and `SetPromotedLayers()` map.
  - **Fix:** Clear and reuse scratch maps/slices on `layerStore` and `motion.Runtime`.
  - **Proof:** Allocation benchmark on steady-state retained-layer animation verifying drop to `0` map/slice allocations per frame.

- [ ] `PERF-IMG-01` Remove `reflect.ValueOf(data).Pointer()` boxing and write-lock on `imageCache.get` hits (`internal/replay/images.go:116, 148-157`).
  - **Root cause:** `reflect.ValueOf(data).Pointer()` boxes the slice header per draw, and `imageCache.get` takes `c.mu.Lock()` on every hit despite `mu` being `sync.RWMutex`.
  - **Fix:** Use `unsafe.SliceData(data)` (or slice-data pointer helper) for zero-alloc identity and use `RLock` + atomic tick (paired with `PERF-LRU-01`) on cache hits.
  - **Proof:** Benchmark on repeated image draw showing reduced allocs and contention.

- [ ] `PERF-UIA-01` Throttle/cache `placeUIA()` Win32 geometry syscalls (`internal/window/a11y_uia_windows.go:77, 99-108`).
  - **Root cause:** `placeUIA()` runs on every `Update()` tick, issuing 3 Win32 syscalls (`ClientToScreen`, `GetClientRect`, `GetDpiForWindow`) and locking `Provider.mu` even when the window has neither moved nor resized.
  - **Fix:** Track last window position/size/device-scale/titlebar-inset and only re-query/publish geometry when changed (or on periodic fallback).
  - **Proof:** Unit test verifying unchanged window state skips `SetGeometry` updates on consecutive ticks.

## 3.2 GPU Texture Lifecycles & Buffer Bounds

- [ ] `GPU-LIFE-01` Bound and shrink `backdropState.scratch` textures and remove per-draw uniform map allocations (`internal/replay/backdrop_filter.go:40-57, 165, 182`).
  - **Root cause:** `acquireScratch` grows 3 unmanaged textures to `(maxW, maxH)` monotonically with no `ebiten.MaxImageSize()` clamp and never shrinks until `SetIdle()`/`Close()` (e.g., 200×2000 + 2000×200 permanently holds 48 MB VRAM), and allocates two 9-key `map[string]any` uniform maps per draw.
  - **Fix:** Clamp to `ebiten.MaxImageSize()`, shrink when high-water area greatly exceeds active demand over N frames, and reuse uniform maps on `backdropState`.
  - **Proof:** Unit test covering aspect-ratio swings and `MaxImageSize` clamping + graphics parity check.

- [ ] `GPU-LIFE-02` Complete GPU buffer disposal on `runWindow` exit and join `paceGame` timer goroutine (`internal/window/run.go:123-135`, `internal/window/idle.go:258-260`).
  - **Root cause:** `runWindow` defers dispose of `titlebar.canvas`, `renderer`, and `dev.panelCache`, but omits `dropTiles()`, `disposeViewport()`, `dropScaleBuf()`, `disposeImage()`, and `bitmapView.img.Deallocate()`; `paceGame.close()` does not wait for the timer goroutine to exit.
  - **Fix:** Add a unified `game.disposeSurfaces()` defer in `runWindow` and wait on a `sync.WaitGroup` in `paceGame.close()`. Also evaluate pooling short-lived offscreen textures in `internal/replay/blend.go:266, 422` and capping `ui/builder.go:289-327` `rendererPool` buffer capacity (e.g., 64 KiB).
  - **Proof:** `leakcheck` + `TestWindow*` lifecycle tests + `make graphics-check`.

- [ ] `GPU-LIFE-03` Avoid blocking Ebiten's main thread with `time.Sleep` in `paceGame.Draw` (`internal/window/idle.go:313-315`) or document/verify frame-pacing tradeoffs.
  - **Root cause:** `MaxFPS` capping sleeps inside `Draw`, delaying OS input/window event pumping for up to ~33 ms at 30 FPS.
  - **Fix:** Evaluate driving `MaxFPS` via Ebiten's `SetTPS` / skip-draw pacing or short yield without starving input events; if Ebiten requires sleep for unmanaged draw cadence, bound and document behavior with a responsiveness test.
  - **Proof:** Window pacing test confirming input latency under `MaxFPS: 30` remains within budget.

---

## Closure Gates

- [ ] `P1B-GATE-01` `make test`, `make test-race`, `make lint`, and `make graphics-check` exit `0`.
- [ ] `P1B-GATE-02` `scripts/perf-baseline.sh` shows no regressions (`max-ratio <= 1.5`) and measurable improvements on `textRasterCache`, `motion`, and idle `Update()`.
- [ ] `P1B-GATE-03` Record before/after benchmark numbers beside each closed row above.
