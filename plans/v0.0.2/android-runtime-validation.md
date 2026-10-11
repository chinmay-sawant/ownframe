# Android runtime validation

## Scope and implementation

Work is on `feature/android-runtime-blinkless-34d9386`. Both modules require
`v0.0.0-20261010174808-34d93862f9ed`, resolving to the requested Blinkless
commit `34d93862f9ed2eec613b0cdea11d3d6da089b574`. Old engine checksums are removed.
The workspace configuration is unchanged, and the examples keep their local
OwnFrame replacement. No renderer, browser engine, JavaScript engine, or new
module dependency was introduced.

The audit reused existing cached parsing and stylesheets, display lists, dirty
buffers, history/routes, form bindings, text selection, IME composition, long
press, flick scrolling, pinch zoom, EbitenView lifecycle, and AndroidHost.
The engine API comparison identified corrected PointsPerPixel and PixelPerPoint
units; callers now use the factor matching their coordinate conversion.

Implemented corrections:

- Fixed operations use viewport culling. Scroll-buffer composition respects
  engine paint order, and GPU colors have valid premultiplied opacity.
- Bitmap fallback paints rounded geometry, borders, grids, affine images and
  text, opacity, and CSS blend groups. Invalid payloads and unsupported shaping
  options report errors. Canvas and blend-buffer sizes are bounded.
- A cached fallback viewport preserves fixed positioning during committed-size
  scrolling and reuses the GPU image on idle frames.
- One touch tap delivers one press. Pinch and native cancellation release active
  gestures. A shared native adapter clears Android touches because the current
  Ebiten input adapter ignores ACTION_CANCEL. Pause invokes cancellation.
- IME rectangles account for zoom and resize. Late callbacks cannot edit another
  focused field. Telegram queues keyboard blur and preserves composer drafts.
- Telegram reserves lateral system insets in its FrameLayout. GPU image eviction
  disposes resources, and font source/face caches have size limits.

## Automated and build checks

The initial engine upgrade checks ran on Linux amd64 with Go 1.26.4 and an Xvfb
display where needed. The device follow-up checks are recorded below.
Logs and profiles are local ignored artifacts under `temp/android-runtime/`.

| Check | Result and boundary |
| --- | --- |
| Both modules: `xvfb-run -a make test` | Passed on the initial upgrade tree, with package concurrency 1. Runs root and examples packages. |
| Compilation/vet: `xvfb-run -a make build` | Passed on the initial upgrade tree. Compiles root and examples without linking every executable. |
| GPU replay: `OWNFRAME_REPLAY_GPU_TEST=1 xvfb-run -a go test ./internal/replay -run TestGPUVisibleParity -count=1` | Passed, including fixed-layer culling and translucent fill regression. Desktop GPU path under Xvfb. |
| Runtime wasm: `GOOS=js GOARCH=wasm go vet -p 1 ./... ./examples/login ./examples/platform` | Passed on the initial upgrade tree. |
| All examples wasm vet | Blocked by existing SQLite-backed examples: modernc/sqlite lacks the required js build support. Runtime and supported examples are checked separately. |
| `sh scripts/android.sh` | Passed on the initial upgrade tree for all default Android ABIs. Builds Telegram AAR and debug APK. |
| Other Android hosts | Login, Platform, Dino, and LG Remote arm64 AARs and debug APKs passed. LG Remote passed its serial rerun. |
| Android installation and device runtime | Initially unavailable. The Pixel 7 follow-up below records subsequent installation and rotation checks. |

Regression coverage includes real engine unit conversion; bitmap paint, blending,
unknown operations, transformed text budgets, and resource limits; fallback viewport scrolling and GPU
reuse; touch press counts and cancellation; IME bounds and focus targeting;
font-cache bounds; PNG rejection of unsupported retained paint; and Telegram draft preservation on blur.
`TestAndroidLayoutsResizeAndModalInput` covers a responsive Flexbox/Grid dashboard,
long settings content, modal hit ordering, and form binding at 360×640, 640×360,
412×915, then 360×640 again. These tests exercise logical CSS sizes, not native
device density or the Android keyboard.

## Measurements

Measurements use Linux amd64, Intel Core i7-13700HX, 24 logical Go processors,
Xvfb, and the current tree. Other builds were running concurrently. They are
single-environment observations, with no controlled before/after comparison.
They establish neither Android frame pacing nor a performance improvement.

The existing page benchmarks ran three times with a one-second target and CPU
and allocation profiles:

| Benchmark | Time range | Bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| RedrawWarm | 0.337–0.429 ms | 328,442 | 126 |
| RedrawHeavyWarm | 0.586–0.868 ms | 369,050–369,063 | 106 |

These are complete warm Page.Redraw operations, including template execution,
engine relayout, and display-list preparation. They do not isolate layout time,
GPU presentation, or input-to-render latency. The allocation profile attributes
83.4% of allocation volume to Blinkless styleStore.append. Its 6.15 GiB aggregate
is allocation volume across benchmark iterations, not live RSS.

The uninstrumented Platform desktop example, after five seconds of warmup,
used 41.97% of one CPU core over a 30-second idle interval. RSS went from
157.88 MiB to 162.56 MiB. A separate 35-second idle CPU profile sampled 36.84%
of one core; 73.59% of samples were in runtime.cgocall and 82.57% cumulatively
in Ebiten OSThread.loop. This points to native graphics/engine calls but does
not identify a specific native function. The runtime still ticks and presents
at Ebiten's normal cadence, even when it reuses the HTML layout. Low idle CPU
and long-run memory stability have not been demonstrated.

Reproduce the benchmark profile from the repository root:

```sh
OWNFRAME_PROFILE_DIR="$PWD/temp/android-runtime"
xvfb-run -a go test ./internal/page -run '^$' \
  -bench 'BenchmarkRedraw(Warm|HeavyWarm)$' -benchmem -benchtime=1s -count=3 \
  -cpuprofile "$OWNFRAME_PROFILE_DIR/cpu.pprof" \
  -memprofile "$OWNFRAME_PROFILE_DIR/memory.pprof" \
  -o "$OWNFRAME_PROFILE_DIR/page.test"
go tool pprof -top temp/android-runtime/cpu.pprof
go tool pprof -top -alloc_space temp/android-runtime/memory.pprof
```

## Pixel 7 rotation follow-up

A USB-connected Pixel 7 is available through Windows ADB from WSL2. The device
runs Android 17, API 37, with a 1080×2400 display and a configured density of
411 dpi. Telegram reserves the lateral cutout before Ebitengine receives its
logical viewport. This check uses the display-list replay path.

Before the fix, the regression requested an 885×420 CSS viewport and received
885×480 because Telegram's desktop minimum height applied on mobile. The window
classified that frame as stretched and disabled touch scrolling. Its list tabs
also ended at Y=658, beyond the visible bottom at Y=404 after system insets.
Device screenshots reproduced the missing landscape tabs.

`BindMobile` now prepares responsive pages with native viewport sizing.
`LockView` keeps a fixed canvas minimum, and desktop pages retain their minimum
constraints. Telegram reuses the conversation's pinned bottom layer and system
bar strips for its list tabs. Phone lists use a block container and reserve tab
space. Tab changes, conversation Back, and list viewport height changes request
scroll offset zero. Keyboard insets only jump to the newest message in an open
conversation, so focusing list search does not scroll to the list end.

Regression tests cover portrait → landscape → portrait dimensions, desktop
minimums, invalid mobile preparation, locked game canvases, vertical touch scroll
through the window, tabs inside the viewport and after scrolling, tab navigation
scroll reset, list resize scroll reset, search keyboard scroll behavior, and
composer geometry after rotation with a keyboard inset.

The final follow-up source passed `xvfb-run -a make test`, `xvfb-run -a make
build`, runtime/Login/Platform WebAssembly vet, and the opt-in GPU replay parity
test. Each of the five mobile packages passed the existing serial
`scripts/android-release.sh <app>` pipeline. Signature, arm64-only native
libraries, and non-debuggable manifests were verified. The committed APK
checksums are in `examples/android-host/APK-SHA256SUMS`.

The final signed Telegram release was installed on the Pixel 7. Both 90-degree
landscape directions were tested with 30- and 50-device-pixel upward swipes,
roughly 12 and 19 CSS pixels at the configured density. A separate check used
three 540-device-pixel upward swipes in landscape before returning to portrait.
The deep-scroll portrait content matched its unscrolled baseline crop exactly,
excluding the changing system status bar. The process ID stayed the same through
these rotations. Settings and Contacts taps after deep scrolling opened their
content at the top and kept the bottom tabs visible. Native tap injection used
a stationary 120 ms swipe so the game loop saw both press and release. These
checks establish geometry and interaction, not frame pacing or latency.

The first composer rotation check exposed an IME ordering bug. On rotation, the
Android keyboard dismissal callback arrived with `FocusID="compose"` about
108 ms before the new viewport size. The IME adapter treated that callback as a
user dismissal and cleared focus before the resize.

The follow-up waits up to 250 ms for an orientation change before clearing focus.
If the viewport rotates during that interval, it keeps the focused field and
starts a fresh IME session. Otherwise, an ordinary keyboard dismissal still
clears focus. The updated signed Telegram release was installed on the Pixel 7.
Gboard remained visible through portrait → landscape → portrait and the opposite
landscape direction. Text entered after rotation appeared in the composer, and
Android Back still dismissed the keyboard and removed the field focus. The
original automatic-rotation preference was restored. IME composition and repeated
lifecycle cycles were not validated on the device.

One Home/launch cycle resumed the same Telegram process on the chat list, and
Android Back returned from the conversation to the list.

All five signed release APKs were installed with `adb install -r` through
Windows ADB. The updated Telegram device `base.apk` SHA-256 matched the built
artifact. The other four APK hashes were checked against their committed files
in the previous follow-up.

Local screenshots and logs stay under the ignored
`temp/android-runtime/pixel7/rotation/` directory. They are diagnostic artifacts,
not application data or committed release files. Android frame pacing, latency,
CPU/RSS/GPU use, and orientation performance have not been measured.

## Pending device checks and limits

The Pixel 7 follow-up covers Telegram rotation, vertical list scrolling, and
pinned list navigation. Full login keyboard/composition, dashboard Flexbox/Grid,
long settings scrolling, modal stacking/input, repeated navigation/resume, and
bitmap fallback checks remain incomplete. Test additional device densities as
well as logical sizes.

Input-to-render latency, scrolling frame pacing, Android CPU/RSS/GPU usage,
orientation performance, and navigation soak measurements need a device or
emulator. The desktop RSS interval is too short to establish memory stability.

Bitmap typography uses OpenType font drawing and has no full shaping contract.
FontFeatures and TextAutospaceGap currently fail with a rendering error. Bitmap
scrolling repaints a CPU viewport; full fallback pictures remain resident too.
Fixed projection applies at a committed normal viewport; fitted views and pending
resizes scale the full picture. Custom SetViewportPinZ counter-scroll behavior
requires replay. No persistent Activity/process restoration or accessibility
tree was added. Blinkless CSS semantics were not changed.

See [runtime behavior](../../documentation/android-runtime.md) and the
[build guide](../../documentation/android-build.md) for ownership and commands.
