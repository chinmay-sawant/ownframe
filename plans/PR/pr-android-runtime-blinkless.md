# feat(android): upgrade Blinkless runtime and preserve input on rotation

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../skills/PR/PR_TEMPLATE.md), adapted to OwnFrame.

- Base: `master` (`a48a0e9`)
- Head: `feature/android-runtime-blinkless-34d9386`
- Scope: Blinkless engine upgrade, rendering and Android runtime fixes, Pixel 7 rotation follow-up

---

## Summary

Upgrade OwnFrame to Blinkless `v0.0.0-20261010174808-34d93862f9ed`, correct renderer coordinate and paint handling, and strengthen mobile input, IME, and viewport behavior. The Android follow-up keeps Telegram text focus through device rotation and refreshes the signed release APKs.

---

## Motivation / context

- Plans: [Android runtime upgrade checklist](../v0.0.2/android-runtime-upgrade-checklist.md) and [runtime validation](../v0.0.2/android-runtime-validation.md).
- Issues: no related issue was found.

The Blinkless update corrected coordinate conversion factors used by OwnFrame. The runtime audit also found gaps in software bitmap painting, fixed-layer ordering, GPU resource bounds, Android touch cancellation, and IME callback ordering. Pixel 7 rotation testing then exposed a mobile viewport sizing issue and a keyboard focus race.

---

## Changes

### Rendering and engine

- Pin both Go modules to Blinkless `v0.0.0-20261010174808-34d93862f9ed` and update consumers to use the corrected points-per-pixel and pixels-per-point conversions.
- Preserve fixed-operation culling and paint order, premultiply GPU opacity correctly, dispose evicted images, and bound font caches.
- Complete bitmap fallback painting for rounded geometry, borders, grids, opacity, affine images and text, letter spacing, and CSS blend groups. Return errors for unsupported payloads and bound temporary buffers.

### Android input and viewport

- Deliver one press per tap and handle touch cancellation and Activity pause without turning cancellation into clicks.
- Map IME bounds through scrolling, zoom, and resize; prevent stale composition callbacks from editing a newly focused field.
- Prepare responsive mobile pages at the native viewport, keep Telegram pinned controls visible, and reset list scrolling when navigating or rotating.
- Preserve Telegram composer focus when the keyboard dismissal callback arrives before the rotation resize. Ordinary keyboard dismissal still clears focus.
- Refresh and checksum the five signed arm64 release APKs.

### Documentation and regression coverage

- Add regression tests for rendering, cache limits, touch cancellation, IME targeting and rotation, viewport sizing, Telegram navigation, and form state after resize.
- Record build and Pixel 7 evidence plus outstanding device checks in the Android runtime documentation and validation plan.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Adds bounded font, image, canvas, blend, and transformed-text caches. Recorded redraw timings are measurements without a controlled before/after baseline; no Android performance claim is made. |
| **Memory** | Bounds retained GPU and bitmap fallback resources. Long-run Android memory stability has not been measured. |
| **Behavior / correctness** | Corrects display-list coordinate conversion and paint order, touch cancellation, mobile viewport sizing, and IME focus across rotation. |
| **API / CLI** | No public API or CLI change. |
| **Dependencies** | Both Go modules use the Blinkless commit above. No new module dependency was introduced. |
| **Binary size / build time** | Refreshes five signed arm64 APKs. No comparative size or build-time measurement was recorded. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing API callers require no migration. |

---

## Test plan

- [x] `make test` passed on the final branch, package concurrency 1.
- [x] `xvfb-run -a make build` passed for both modules on the final branch.
- [x] Runtime, Login, and Platform WebAssembly vet passed.
- [x] Opt-in GPU replay parity passed.
- [x] Android release APKs built serially; signatures, arm64 native libraries, and non-debuggable manifests were checked.
- [x] Pixel 7: Telegram retained the keyboard and text focus through portrait and both landscape directions; text entry worked after rotation, and Android Back dismissed the keyboard and cleared focus.
- [ ] Complete the remaining device checks for IME composition, additional densities, modal input, repeated navigation and resume, bitmap fallback, and performance.

### Commands

```sh
xvfb-run -a make test
xvfb-run -a make build
GOOS=js GOARCH=wasm go vet -p 1 ./... ./examples/login ./examples/platform
OWNFRAME_REPLAY_GPU_TEST=1 xvfb-run -a go test ./internal/replay -run TestGPUVisibleParity -count=1
```

The final rotation follow-up also ran `sh scripts/android-release.sh telegram` and installed the signed APK on a Pixel 7. See the validation plan for commands, device details, evidence, and limitations. `make test` was rerun for this PR and passed.

---

## Screenshots / sample output

No screenshots are attached. Pixel 7 screenshots and logs remain in ignored local `temp/android-runtime/` paths.

---

## Related issues

- None. No tracker issue exists for this work.

---

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Label applied (`enhancement`)
- [x] Related issues recorded as none; no placeholder issue IDs.
- [x] Filled body saved under `plans/PR/pr-android-runtime-blinkless.md`.

---

## Follow-ups (out of scope)

- Complete the remaining device checks listed in the Android runtime validation plan.
- Measure Android frame pacing, latency, and CPU/GPU/RSS use on a device or emulator.
- Address the documented bitmap typography, fitted-view fixed positioning, and process restoration limits separately.

---

## Reviewer checklist

- [ ] Confirm the Blinkless coordinate conversions match the engine contract.
- [ ] Confirm display-list and bitmap fallback paint behavior against the documented limits.
- [ ] Confirm Android rotation retains focus only when the viewport orientation changes.
- [ ] Confirm APKs and checksums match the signed arm64 artifacts.
- [x] No unrelated generated artifacts are included.
- [x] PR metadata includes an assignee and label.
- [x] Diff-stat-by-extension table generated with `scripts/pr-diff-stat.sh` is included below.

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.apk` | 5 | Binary | Binary |
| `.css` | 1 | 2 | 0 |
| `.go` | 145 | 2233 | 441 |
| `.html` | 3 | 23 | 0 |
| `.java` | 6 | 107 | 5 |
| `.md` | 12 | 589 | 6 |
| `.mod` | 2 | 2 | 3 |
| `.sum` | 2 | 4 | 4 |
| No extension | 1 | 5 | 0 |
| **Total** | **177** | **2965** | **459** |
