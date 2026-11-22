# fix: resolve P0-P2 review findings and bump blinkless to df3f742da29d

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `v004/review-p0-p1-p2-hardening`
- Release cycle: v0.0.4

---

## Summary

- Bumps both root `go.mod` and `examples/go.mod` to `github.com/chinmay-sawant/blinkless v0.1.6-0.20261122140852-df3f742da29d` (`chinmay-sawant/blinkless#45`).
- Resolves P0 and P1 correctness, concurrency, security, and visual-parity findings across `internal/render`, `internal/page`, `internal/web`, `internal/replay`, `internal/window`, `internal/motion`, `internal/bitmap`, `internal/masking`, `internal/atspi`, `internal/fetch`, `internal/filepick`, `internal/crash`, `ui`, and root `ownframe`.
- Updates `examples/blinkless-all` to resolve the pinned module catalog first, evaluate generated tabs lazily on demand, and clean stale `status.go` entries, while reconciling `documentation/platforms.md` and `documentation/accessibility-native.md`.

---

## Motivation / context

- Plans: `plans/v0.0.4/pr/pr-review-p0-p1-p2-hardening.md`
- Issues: none. Relates to `chinmay-sawant/blinkless#45`.

---

## Changes

### 1. Core API, Page/Render, UI, and Platform Subsystems
- **Engine lock & panic recovery (`internal/render/relayout_element.go`, `internal/page/motion_wire.go`, `internal/page/image.go`):** Added `render.RelayoutElementWithOverrides`, `render.RestampTransforms`, and `render.RestyleInlineSVG` holding `engineMu` with panic recovery.
- **`internal/web` SSE shutdown & MIME/origin hardening (`internal/web/serve_build.go`, `events_http.go`, `image_http.go`, `dom_http.go`, `state_http.go`, `audio_http.go`):** Closed subscriber channels on cancel with `BaseContext` and a bounded 2s shutdown timeout; restricted `imageContentType` to an explicit image MIME allowlist with `Content-Security-Policy: sandbox` on SVG and `X-Content-Type-Options: nosniff`; enforced `sameOrigin(r)` on `/dom`, `/state`, `/audio`, and image routes.
- **`ui` & `internal/page` fixes (`ui/mount.go`, `ui/diff.go`, `internal/page/patch.go`, `internal/page/render.go`, `internal/page/dirty.go`, `internal/page/dirty/diff.go`):** Checked `Generation() == 0` on zero-patch `View.Update` so ticks do not trigger `captureTickOps`; made `ui/diff.go` patch budgets order-independent; unified `PatchElement` and `PatchElements` commit/invalidation helpers; fixed `recordPartialBounds` empty-rect early exit, `setStickyBoxes` duplicate hit list, and `mixOpStamp` masks/tags.
- **Public API nil-receiver safety & package edges (`page.go`, `page_form.go`, `page_nav.go`, `page_scroll.go`, `internal/host/`, `internal/ipc/request.go`, `internal/atspi/`, `internal/filepick/`, `internal/fetch/`, `internal/crash/`):** Added uniform nil/zero-page guards (`TestNilAndZeroPageMethodsDoNotPanic`), UTF-8 rune-boundary truncation and atomic `appID` in `atspi`, context error preservation in `filepick`, race-free `fetch.UseSystemProxy`, pre-existing dir permission preservation in `crash.Write`, and moved `ErrNilContext` and `WebDOM` into `internal/host`.

### 2. Rendering, Bitmap, Masking, Motion, and Window Pipeline
- **Tile & layer compositing fixes (`internal/window/replay_tile.go`, `replay_partial.go`, `replay_viewport.go`, `replay.go`, `update.go`, `internal/replay/layer.go`, `layer_cache.go`):** Repainted full visible tile when an offscreen-dirtied tile receives a sub-rect damage; merged `layerDamage` even when page dirty rects are empty; called `demoteLayers()` across mode transitions; added `BeginViewportLayers` / `layerTopOrderSafe` to prevent viewport layer z-order inversion.
- **Motion, replay caches, and GPU/bitmap parity (`internal/motion/transition.go`, `internal/replay/`, `internal/bitmap/`, `internal/masking/`, `internal/window/`):** Fixed `CaptureTransitionsScoped` scratch-map shrink order; added `DisplayOpImage` to `xformPointsBox`; deallocated and pooled `backdropShader`/scratch textures on `Renderer` with `Radius <= 0` early-out and `backdropClean` reuse; enforced `layerMaxPixels` eviction on layer resize; consolidated replay LRU caches (`internal/replay/lru.go`); added ring hash cache for unchunked images; quantized `rasterKey.fx`/`fy` to 1/1024 px; fixed `color-dodge`/`color-burn` and group alpha in `group_blend.kage`, `rotate(360deg)` in `text_matrix.go`, transformed `-webkit-text-stroke`, exact integer quarter-turn bitmap math, tile-plan refresh on in-place motion, `pinClip` on viewport, tight non-masked blend group bounds, and phase-locked frame pacing.

### 3. `examples/blinkless-all` and Documentation
- Resolved `properties.json` from the pinned `blinkless` module first (`examples/blinkless-all/catalog.go`), evaluated generated tabs lazily on demand (`props.go`, `page.go`, `tabs.go`), cleaned `status.go`, and reconciled `documentation/platforms.md` and `documentation/accessibility-native.md`.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Eliminates ~1,254 eager startup layouts in `blinkless-all`, pools backdrop-filter GPU scratch textures, skips redundant `backdropClean` full-scene repaints, and caches unchunked image hashes. |
| **Memory** | Picks up `blinkless` `html.Document` leak fix and 32 MiB raster cache cap, releases oversized `motion` transition scratch maps, and deallocates `backdropShader` on `Renderer.Close()`. |
| **Behavior / correctness** | Fixes offscreen-dirty tile partial repaint corruption, dropped `layerDamage`, stale `promoted` layer IDs across mode switches, viewport layer z-order inversion, `engineMu` race hazards, `-web` SSE shutdown hang, and SVG/image route MIME handling. |
| **API / CLI** | Root `Page`, `Form`, and `Scroll` methods are now uniformly safe on nil and zero-value `&ownframe.Page{}` receivers. |
| **Dependencies** | Bumped `github.com/chinmay-sawant/blinkless` to `v0.1.6-0.20261122140852-df3f742da29d` in `go.mod` and `examples/go.mod`. |
| **Binary size / build time** | No new third-party dependencies. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing `Run`, `Serve`, and `Page` callers keep working without changes. |

---

## Test plan

- [x] `make build BUILD_P=4`
- [x] `make lint` (all 10 checks)
- [x] `make test TEST_P=4`
- [x] `make graphics-check` (all 20 isolated Xvfb GPU pixel sessions)

### Commands

```sh
make build BUILD_P=4
make lint
make test TEST_P=4
make graphics-check
```

---

## Screenshots / sample output

```text
PASS maintenance tooling
PASS linter version regression
PASS Go library and examples (v2.15.0)
PASS documentation snippets
PASS plan states
PASS tracked file size guardrail
PASS example maintenance inventory
PASS example internal imports
PASS ui generated properties drift
PASS website ESLint
```

---

## Related issues

- Relates to `chinmay-sawant/blinkless#45`.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`bug`, `enhancement`)
- [x] Related issues filled (`chinmay-sawant/blinkless#45`)
- [x] Filled body committed under `plans/v0.0.4/pr/pr-review-p0-p1-p2-hardening.md`

---

## Follow-ups (out of scope)

- Unifying `internal/window` paint branches into a single surface state machine and shipping Windows/macOS/Android native accessibility backends.

---

## Reviewer checklist

- [x] Behavior matches summary and test plan
- [x] No unrelated changes in diff
- [x] Public API / CLI changes documented
- [x] New rules have fixture coverage when applicable
- [x] PR has assignee and labels
- [x] Related issues use correct Closes/Relates keywords
- [x] No secrets or generated artifacts committed
- [x] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 115 | 3253 | 687 |
| `.kage` | 2 | 39 | 28 |
| `.md` | 3 | 46 | 29 |
| `.mod` | 2 | 2 | 2 |
| `.sum` | 2 | 4 | 0 |
| **Total** | **124** | **3344** | **746** |
