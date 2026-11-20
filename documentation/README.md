# Documentation

Build screens with HTML and CSS and handle data and actions in Go. Target native hosts (currently in the `v0.0.x` preview cycle; no `v1.0` tag yet) are desktop `Run` and Android `BindMobile`. Wasm and `-web` are previews.

## 1. Start here (3 docs)

1. **[Render a page](screen.md)** — templates, layout, replay, and images.
2. **[Connect controls to Go data](binding.md)** — `data-bind`, actions, and handlers.
3. **[Choose a platform](platforms.md)** — desktop `Run`, Android `BindMobile`, wasm preview, and `-web` debugger.

## 2. Core guides

### Window, input, and layout
- [Open a window](window.md)
- [Handle forms and edits](forms.md)
- [Text editing and selection](editing.md)
- [Interaction and IME](interaction.md)
- [Keyboard handling](keys.md)
- [Pointer and touch input](pointer.md)
- [Scrolling and windowing](scrolling.md)
- [Drag and drop](drag-drop.md)
- [Clipboard](clipboard.md)
- [Desktop shell integrations](desktop-shell-integrations.md)
- [Runtime behavior](runtime-behavior.md)

### Styling, components, and motion
- [Change the page theme](theming.md)
- [Style invalidation](styleinvalid.md)
- [Build with the UI builder](builder.md)
- [Reusable components](components.md)
- [ChadCN component set](chadcn.md)
- [CSS animations and transitions](css-motion.md)
- [Go-driven frame animations](frames.md)
- [Inline color emoji](emoji.md)
- [Tips and patterns](tricks-tips.md)

### Architecture, performance, and tooling
- [Repaint, tiles, and layers](repaint.md)
- [Optimization paths](optimizations.md)
- [Performance profiling](performance.md)
- [DevTools dock](devtools.md)
- [Hot reload](hot-reload.md)
- [Navigation and history](navigation.md)
- [In-process IPC](ipc.md)
- [Fetch and XHR](fetch.md)
- [Crash reports](crash.md)
- [Serve a page in a browser (`-web`)](web.md)

### Accessibility
- [Accessibility tree](accessibility.md)
- [Native accessibility status](accessibility-native.md)
- [Linux AT-SPI bridge](atspi.md)

### Android, packaging, security, and contributing
- [Create an Android app](android-app-creation.md)
- [Android build reference](android-build.md)
- [Android runtime limits](android-runtime.md)
- [Play Store gaps](play-store.md)
- [Package a release](packaging.md)
- [Release checklist](release-checklist.md)
- [Secrets handling](secrets.md)
- [Security model](security.md)
- [Supply chain](supply-chain.md)
- [Starter templates](starters.md)
- [Validate a change](testing.md)
- [Why ownframe](why-ownframe.md)

## 3. Comparisons and reference

- [Supported features and limits](features.md)
- [Feature-to-example index](features-examples.md)
- [Blinkless HTML/CSS compatibility matrix](compatibility-matrix.md) (synced from upstream Blinkless; feeds `scripts/gen-ui-properties`)
- [Performance baseline](performance-baseline.md)
- [Browse all examples](../examples/readme.md)
- [App showcase](../showcase.md)

### Framework comparisons
- [Syntax comparison across Go UI libraries](compare-syntax.md)
- [Compare with Electron](compare-electron.md)
- [Compare with Wails](compare-wails.md)
- [Compare with WebView](compare-webview.md)
- [Compare with Fyne](compare-fyne.md)
- [Compare with Gio](compare-gio.md)
- [Compare with Ebitengine](compare-ebitengine.md)
- [Compare with Rust GPUI](compare-rust-gpui.md)

## 4. Historical and point-in-time notes

- [Lint findings snapshot](lint-findings.md)
- [Loglens Windows profiling notes](loglens-profiling.md)
- [Independent maintenance review prompt](maintenance-review.md)
- [Code Review example map](mermaid-code-review.md)
- [Migrating existing template apps](migrate-experimental.md)
- [Rename and migration (`go-gpui` -> `ownframe`)](rebrand.md)
