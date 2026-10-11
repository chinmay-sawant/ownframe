# Android runtime and Blinkless upgrade

Target engine commit: `34d93862f9ed2eec613b0cdea11d3d6da089b574`.

## Findings

The runtime already caches parsing and styles, reuses display lists, bounds GPU
buffers for long content, and supports forms, history, routes, touch scrolling,
flicks, pinch, long press, selection, and mobile IME composition. Android examples
already embed EbitenView and suspend/resume it. The shared Android host reports
all four safe insets and supports predictive Back. Telegram already uses that
host; lateral cutout insets needed applying to its enclosing layout. The initial upgrade had no connected device. The follow-up uses a USB-connected
Pixel 7 through Windows ADB from WSL2.

The requested engine corrects the swapped PointsPerPixel and PixelPerPoint fields.
OwnFrame consumers need corresponding updates. Operation kinds remain compatible.
The existing software bitmap painter drops borders, grids, blend groups, rounded
geometry, and full transforms. Its fallback needs real paint checks.

## Execution

- [x] Inspect architecture, engine API diff, Android hosts, input, and buffer reuse.
- [x] Pin both modules to v0.0.0-20261010174808-34d93862f9ed.
- [x] Tidy both modules; preserve go.work; isolate ignored historical temp probes.
- [x] Correct unit consumers and cover real engine output in regression tests.
- [x] Improve bitmap fallback and verify painted geometry, opacity, and transforms.
- [x] Correct touch cancellation and duplicate press delivery.
- [x] Map IME bounds through scrolling, zoom, and resize.
- [x] Retain shared Android host; apply Telegram lateral safe areas and queue keyboard blur.
- [x] Run both module tests, vet, GPU comparisons, and mobile/wasm compile checks.
- [x] Build Android with sh scripts/android.sh.
- [x] Record benchmarks and profiling boundaries.
- [x] Document responsibilities, runtime behavior, build steps, and limitations.
- [x] Reproduce and correct Pixel 7 landscape sizing and vertical list scrolling.
- [x] Cover native sizing, fixed game canvases, pinned tabs, and rotated composer
      geometry with regression tests.
- [x] Preserve IME focus across rotation; verify keyboard return, text entry,
      both landscape directions, and ordinary keyboard dismissal on Pixel 7.
- [ ] Complete the remaining keyboard/composition, density, modal input, navigation,
      resume, bitmap fallback, and performance checks on a device.

Evidence and remaining validation: [validation record](android-runtime-validation.md).
