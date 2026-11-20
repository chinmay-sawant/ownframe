# ownframe

Go library that renders HTML and CSS screens through [blinkless](https://github.com/chinmay-sawant/blinkless) in an Ebiten window, with no Chromium, WebKit, WebView, or JavaScript runtime.

## Where code lives

- **Public API (module root `ownframe`):** Keep the public API at the module root (do not move it into `api/`). `ownframe.Page` wraps an unexported `*page.Page` behind an explicit delegated method set (`api.go`, `page.go`, `page_*.go`, `run.go`, `serve.go`, `elements.go`, `events.go`, `clipboard.go`, `crash.go`, `fetch.go`, `ipc.go`). `NewGame` returns the Ebiten game for tests without calling `Run`.
- **Target native hosts:** Currently in the `v0.0.x` preview cycle (no `v1.0` tag yet). Target native hosts are desktop `Run` (Linux, Windows, macOS) and Android `BindMobile`. Wasm (`scripts/browser.sh`, `browser/index.html`) and `-web` (`Serve`, `internal/web`) are previews/debuggers, not release hosts. See `documentation/platforms.md`.
- **Modules & workspace:** `examples/` is its own Go module (`github.com/chinmay-sawant/ownframe/examples`) joined by `go.work` so library module zips carry no examples. `examples/go.mod` pins the last published `ownframe` tag and `replace`s it with `../` during development. Keep the `blinkless` pin in `examples/go.mod` equal to root `go.mod`. For a release, tag `vX.Y.Z`, bump `examples/go.mod` to that tag, drop `replace => ../`, and tag `examples/vX.Y.Z`.
- **Core internal packages:**
  - `internal/page`: Page state, template execution, `Load`/`Back`/`Forward`/`Route`, form values, caret/selection, focus order, dirty regions (`Probe().TakeDirty`), hot reload (`PollReload`), perf opt-in (`WithPerf`), and accessibility snapshots.
  - `internal/render`: Cached parsed HTML + styled CSS document; size/state changes relayout via `css.Relayout` without reparsing.
  - `internal/replay`: Display-list replay onto Ebiten (rounded fills, borders, grids, shaped text, linear-filtered images, raster/text/chunk caches, Kage blend/isolation layers, retained motion layers, `internal/emoji` Twemoji PNGs). Bitmap fallback shows a `bitmap fallback` badge.
  - `internal/window`: Ebiten loop for desktop, Android, and wasm (256 px tile grid or viewport buffer, retained layer compositing, 5-frame desktop idle park with `WantFrameAt`, DevTools dock via F12/Ctrl+Shift+I, touch scroll/pinch/long-press, F11 fullscreen, WSL top-edge snap, optional `-tags ownframe_debug` loopback endpoints).
  - `internal/motion` (+ `internal/viewtransition`): Automatic CSS `@keyframes` and `transition` driver. Toggle via `Page.Window().SetAnimations(false)` / `SetTransitions(false)`; inspect via `Page.Probe().Motion()`. Never drive the same op with both CSS motion and `internal/frame` (`Page.SetTick`).
  - `internal/frame`: Go-driven imperative animation over retained `layout.Display` ops (`documentation/frames.md`).
  - `internal/host`: Unexported `Screen` and optional host interfaces (`Ticker`, `Inspector`, `Reloader`, `Dropper`, `Focuser`, `Hoverer`, `ContextMenu`, `CursorShape`, `ScrollRequester`, `IME`, `LongPresser`). Never imports root `ownframe`.
  - `internal/atspi`: Linux AT-SPI accessibility bridge (`internal/window/a11y_atspi.go`). Password values stay empty; Windows UIA, macOS VoiceOver, and Android TalkBack are not shipped (`documentation/atspi.md`, `documentation/accessibility-native.md`).
  - `internal/clipboard`, `internal/filepick`, `internal/crash`, `internal/fetch`, `internal/ipc`, `internal/textrun`, `internal/web`: Platform and utility subsystems. `Serve` exposes no clipboard route; `ServeWithOptions` gates debug/pprof behind explicit `InsecureAllow*` flags.
- **Engine (`blinkless`):** Pinned in `go.mod` and mirrored in `examples/go.mod` (no `replace`). Text replay requires Ebiten v2.10.4+ (`go-text/typesetting` v0.3.4+ lookup cache).
- **Components & UI:** `ui/` + `components/chadcn` (chart skins live under `examples/builder/charts/`). See `documentation/components.md` and `documentation/chadcn.md`.
- **Examples & Android:** See `examples/readme.md` for all examples (`login` accepts `secret`/`secret`, `code-view`, `telegram`, `spotify-player`, `desktop-cat`, `wispr-flow-dashboard`, `examples/builder/ui-builder`, `examples/builder/loglens`, etc.). For Android `ebitenmobile bind` (`scripts/android.sh`, `scripts/android-release.sh`), **do not export `GOWORK`**: gomobile copies packages into a temporary module that `go.work` would disown.
- **Standalone CLI preview tools:** `cmd/animation-preview`, `cmd/transition-preview`, `cmd/benchscore`, and `cmd/internal/`.

Read `documentation/features.md` before changing paint, theming, frames, IPC, navigation, crash reports, fetch, clipboard, drag and drop, or DevTools.

## Public API after the Page wrap

`ownframe.Page` is a defined struct with an unexported `*page.Page`. Do not embed `*page.Page` again. Do not add an exported `Inner()`. Do not return `host.Screen`; tests needing a game use `ownframe.NewGame`.

- No new `type X =` alias at module root except documented allowlisted re-exports (`alias_allow_test.go`).
- No new `func (p *Page)` on `internal/page` added solely to grow `ownframe.Page`.
- Public `Page` delegates domain doors: `Form()`, `Edit()`, `Scroll()`, `Window()`, `Access()`, `Probe()`, and `Testing()`. `Page.Testing()` is for `*_test.go` files only (`api_testing_test.go`).
- No new example imports of `github.com/chinmay-sawant/ownframe/internal/` (`make internal-import-check`).
- Do not scan or inspect files outside the ownframe repository unless explicitly requested.

## Package and file cohesion

Group code by responsibility. Keep a type and its related methods together when that makes behavior easier to follow. Split packages at ownership boundaries with explicit inputs and outputs. File length is a review signal, not a character limit.

Keep related test scenarios and fixtures together. Do not split files solely to satisfy a size count, create synonym helpers, or move literals into detached constant files just for lint. Use `go run scripts/merge-go.go` for same-package merges and review cohesion with `CONTRIBUTING.md`.

## Lint

`make lint` is the lint gate (`scripts/run-lint.py`). It runs `make tools-check`, `scripts/lint_test.sh`, pinned `golangci-lint` (`GOLANGCI_LINT_VERSION` in `Makefile`) over both Go modules, `make docs-check`, `make plans-check`, `make repo-size-check`, `make examples-check`, `make internal-import-check`, and `npm --prefix frontend run lint` (ESLint 10).

Policy lives in `.golangci.yml` (curated enable list, not `enable-all`). `//nolint` is a last resort and requires a written reason.

CI (`.github/workflows/ci.yml`) runs `make test`, `make lint`, `make test-race` (library module), security, cross-compile, secrets, and website checks. Feature-branch pushes only trigger heavy jobs when the commit message carries `[ci only=<job>]`.

## Writing

- Load the `unslop` skill when available before writing prose, docs, commit messages, or PR text.
- Highlight key table columns with plain markdown bold (`**value**`). Never use HTML `<mark>` tags in tables.

## Builds

Do not run `go build ./...` (it links every example and leaves binaries in the repo root). Use `make build` (`go vet -p 1 ./... ./examples/...`, or `make build BUILD_P=4`). For one example: `go vet -p 1 ./examples/login`.

- **Website:** `npm --prefix frontend run build` builds the site and nine WebAssembly demos into `frontend/dist/`. GitHub Actions publishes that directory. Commit website sources and preview images. Generated WebAssembly and Android APKs stay out of Git.

## Testing an app by driving it

Follow `documentation/testing.md` ("Testing an app by driving it"):
1. Compare three things per row: description, painted picture, and readout. Assert on `page.Probe().Display()`, not readout text alone.
2. Prefer headless Go tests first (`liveRig` in `examples/blinkless-all/nav_scripted.go`). Write the failing test, then fix.
3. For static pictures, serve with `-web ADDR` and read `/frame.png`, `/click`, `/wheel` via `curl` (matching `Origin`; negative `dy` scrolls down).
4. For live state, run with `-snapshot ADDR`, drive via `scripts/gui-drive.sh`, and inspect `/frame.png`. Park pointer off scrollable boxes before scrolling. If the window cannot be reached, report it as untested.
5. Inspect screenshots directly; report mismatches as web-only, window, or both.
6. Put scratch binaries/screenshots in scratch directories (`go build -o <scratch>/app ./examples/<name>`); stop servers with `pgrep -x <name>`, never `pkill -f <text>`.

## Before you finish

- For docs or maintenance tooling: run `make docs-check plans-check examples-check tools-check repo-size-check`.
- For paint or window changes: also run `make graphics-check`.
- Run `gofmt` on every edited Go file and run `make test` (`make test TEST_P=1` if memory-constrained).
- Leave `~/.Xauthority` untouched.
- Keep clipboard tests off the desktop clipboard (`clipboard.UseMemory(true)`).

## Blocked work

If a task cannot be done right as stated, update the user and stop. Name the blocked part, the needed input or decision, and what can proceed once answered. Never fill gaps with guesses, stubs, or silent defaults, and never report blocked work as done.
