# v0.0.4 - Phase 04: P2 CI Coverage, Modules & API Hygiene

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P2
> **Status:** Open (0 / 11 complete)

---

## Overview

This phase closes CI platform blind spots, cleans up module dependency leaks and version drift, removes dead/duplicate internal helpers, fixes hot-reload/IPC/test-helper rough edges, and gates removal of `internal/render/engineMu` behind `blinkless` Phase 01 completion.

---

## 4.1 Platform CI Coverage & Module Hygiene

- [ ] `HYG-CI-01` Run native Windows accessibility and window tests on `windows-latest` (`amd64`) in CI (`.github/workflows/ci.yml:469-471`).
  - **Root cause:** `native-package` on `windows-latest` currently runs only 3 memory-clipboard tests (`TestMemory*`), skipping `./internal/uia` and `./internal/window` on `windows/amd64`; `windows-arm64` runs them but skips all PRs.
  - **Fix:** Add `go test -race ./internal/uia ./internal/window ./internal/filepick` (headless-safe subset) to `windows-latest` PR CI, and add pure-Go unit coverage for Windows/macOS `filepick` and `clipboard` argument/error paths.
  - **Proof:** Passing GitHub Actions `native-package` run showing `./internal/uia` and `./internal/window` executed on `windows-latest`.

- [ ] `HYG-MOD-01` Move `golang.org/x/perf` (`internal/benchscore` / `cmd/benchscore`) and `tool govulncheck` out of the shipped root library `go.mod` (`go.mod:9, 59`).
  - **Root cause:** Consumers importing `ownframe` inherit `golang.org/x/perf`, `aclements/go-moremath`, and `x/vuln`/`x/tools`/`x/telemetry` in their module graph even though `scripts/go.mod` already exists for repo tooling.
  - **Fix:** Relocate `cmd/benchscore` + `internal/benchscore` (and `tool govulncheck` or `make vuln`) into `scripts/go.mod` (or a dedicated tools module), updating `Makefile` and CI accordingly.
  - **Proof:** `go mod tidy` + `make module-check` + `make vuln` + `go list -m all` showing `golang.org/x/perf` absent from root `go.mod`.

- [ ] `HYG-MOD-02` Reconcile `go.mod` version drift and `ebitenmobile` version mismatch (`assets/go.mod`, `docs/go.mod`, `documentation/go.mod`, `plans/go.mod`, `skills/go.mod`, `examples/*/mygo/go.mod`, `.github/workflows/ci.yml:580`).
  - **Root cause:** 13 `go.mod` files span 3 Go versions (`1.26.4`, `1.26.11`, `1.27.3`), and `ci.yml:580` installs `ebitenmobile@v2.10.4` while `go.mod:7` pins `ebiten/v2 v2.10.6`.
  - **Fix:** Align sentinel zip-boundary modules to `go 1.26.11`, bump `ebitenmobile` in `ci.yml` to `v2.10.6`, and verify/document any intentional `mygo` comparison module version.
  - **Proof:** `git grep '^go 1\.' -- '**/go.mod' go.mod` clean + passing `android-bind` CI job.

## 4.2 API Consistency, Interface Assertions & Cleanup

- [ ] `HYG-API-01` Unify nil-`Page` error returns, guard nil `Option` in `NewWithOptions`, and align `context.Context` handling across `clipboard`, `filepick`, and `ui.View` (`api.go:88-93`, `page*.go`, `internal/clipboard/clipboard.go:64, 91`, `internal/filepick/filepick.go:29`, `ui/mount.go:110-168`, `ui/builder.go:248-255`).
  - **Root cause:** `Redraw`/`Load`/`Form().Focus` return `ErrNilPage` on nil/zero `*Page`, whereas `Click`, `MenuAction`, `Edit.*`, `Access.*`, `Window.PollReload`, `Probe.EndRedrawBatch`, and `Testing.*` return `nil`; `NewWithOptions(cfg, nil)` panics; `clipboard.WriteContext`/`ReadContext` panic on `ctx == nil`; `filepick.Pick` mixes `ErrCanceled` and raw `ctx.Err()`; `ui.View.Update`/`patch` wastes `renderFull` or falls through to `ReplaceCurrent` on `context.Canceled`; `ui.Render(nil)` returns `("", nil)` while `View.Update(ctx, nil)` errors.
  - **Fix:** Skip nil `Option`s in `NewWithOptions`, enforce one consistent nil-`Page` and nil/canceled-`ctx` contract across `page`, `clipboard`, `filepick`, and `ui.View`, and update `page_nil_test.go`.
  - **Proof:** Expanded `page_nil_test.go` + `clipboard`/`filepick`/`ui` context tests passing under `make test`.

- [ ] `HYG-API-03` Track pre-v1.0 API naming & cross-repo coupling reductions (stuttering door methods `Form().FormValue`/`Scroll().ScrollOffset`/`Access().AccessibilityTree`, top-level clipboard names `ownframe.Write`/`Read`/`Clear`, `layout.{Box,Display,DisplayOp}` alias exposure, triple HTML parsing, and `internal/motion/values.go` CSS color/transform duplication + string round-tripping).
  - **Root cause:** Public aliases expose `blinkless/layout` internals directly; domain-door and clipboard names carry pre-wrap stuttering/ambiguity; `ownframe` parses HTML 3× (`html.Parse`, `markup.Parse`, `form/scan.go`) and round-trips interpolated CSS strings back into `blinkless`.
  - **Fix:** Decide pre-v1.0 naming cleanup vs compatibility freeze, replace `internal/frame/anim_write.go:13-16` zero-value `DisplayOp` hack with `layout.Matrix2D`, and coordinate typed CSS value/parse helpers with `blinkless` where warranted.
  - **Proof:** Recorded pre-v1.0 API decision + `make test` and `alias_gate_test.go` passing.

- [ ] `HYG-API-02` Add compile-time `var _ host.X = (*Page)(nil)` assertions for all remaining optional `host` interfaces and move `atspi.Actions` to a platform-neutral internal package (`internal/page/`, `internal/window/a11y_*.go`).
  - **Root cause:** Only ~11 of ~25 runtime-asserted `internal/host` interfaces have compile-time checks on `*page.Page`, risking silent signature drift; Windows (`a11y_uia_windows.go`) and Android (`a11y_android.go`) import `internal/atspi` solely for `*atspi.Actions`.
  - **Fix:** Add compile-time assertions for every `host` interface implemented by `*page.Page` and extract the bounded action queue out of `internal/atspi` into a neutral package (e.g., `internal/a11ybridge` or `internal/a11yqueue`).
  - **Proof:** `go vet ./...` across Linux, Windows, macOS, and Android build tags (`make build`).

- [ ] `HYG-CLEAN-01` Remove dead synonym bloat, duplicate helpers, stale `//nolint` directives, and slice-aliasing hazards (`internal/page/source.go:15-24`, `internal/fetch/scheme.go:34-45`, `internal/web/page_http.go:17`, `dom_http.go:55`, `document.go:273-287`, `components/chadcn/helpers.go:11-19`, `interactions.go:150-182`, `internal/motion/relayout.go:540`, `BENCHMARKS.md:9`).
  - **Root cause:** 10 identical `maxSourceSize` aliases in `internal/page/source.go`, 7 error/const aliases in `internal/fetch/scheme.go`, duplicate `webControls`/`supportsSelection`/`debugf` helpers, `PlaceCaret` calling `focusedEditable(p)` twice (`source.go:188-195`), 10 stale `//nolint` directives for disabled linters (`relayout.go:540`, `examples/code-view/{watch,workspace}.go`, `examples/blinkless-all/snapshot.go:160`), `BENCHMARKS.md:9` citing `"Go 1.24"`, `chadcn.classes` mutating variadic backing slices in place, and `chadcn.Interactions.Handle()` panicking before `Mount()` / `Close()` not restoring handlers.
  - **Fix:** Collapse internal synonyms to single canonical names, reuse `host.WebDOM` and one shared `debugf`, remove stale `//nolint` directives, allocate/copy cleanly in `chadcn.classes`, and make `chadcn.Interactions` safe before `Mount` and clean on `Close`.
  - **Proof:** `make lint` + `make test`.

## 4.3 Hot Reload, IPC, Test Helpers & Gated Engine Decoupling

- [ ] `HYG-REL-01` Fix hot-reload validation, theme starvation, unbounded `ReadFile`, and dead `watched.data` (`internal/page/reload.go:75-80`, `internal/page/reload/poll.go:35-37, 86`, `state.go:40, 94`, `internal/page/source/source.go:99`).
  - **Root cause:** `installWatched` skips `validateHTML` for Go templates; `State.Poll` starves theme reloads when HTML has a syntax error; `watched.data` retains full file copies unused; `os.ReadFile` after `os.Stat` is unbounded on growing files/FIFOs.
  - **Fix:** Validate hot-reloaded templates consistently, poll theme independently from a broken pending HTML retry (and skip re-parsing unchanged broken bytes), remove unused `watched.data`, and read with `io.LimitReader(..., maxSize+1)`.
  - **Proof:** Unit tests in `internal/page/reload/` and `internal/page/` covering theme reload during broken HTML and oversized file reads.

- [ ] `HYG-IPC-01` Fix `internal/ipc` listener/panic handling and `pagetest` bitmap/nil behavior (`internal/ipc/send.go:24-65`, `request.go:43-52`, `pagetest/box.go:27-30`, `pagetest/click.go:29-30`).
  - **Root cause:** `ipc.Listen` silently drops registrations past `maxListeners` (256); `ipc.Send` swallows panics completely (`_ = recover()`); `ipc.removeEar` retains removed closure pointers in the backing array; `ipc.Request` drops the panic value instead of wrapping it; `pagetest.BoxByID` reads `Probe().Display()` (nil on bitmap fallback) instead of falling back to `p.Boxes()`; `pagetest.Type(ctx, nil, ...)` returns `ErrNoBox` instead of `ErrNilPage`.
  - **Fix:** Nil tail slot in `removeEar`, wrap recovered panic values in `ipc.Request`, log/surface `ipc.Send` panics & listener cap behavior explicitly, fall back to `p.Boxes()` in `pagetest.BoxByID`, and return `ErrNilPage` on nil `p`.
  - **Proof:** Unit tests in `internal/ipc/` and `pagetest/`.

- [ ] `HYG-CRASH-01` Harden crash directory validation, Windows UID fallback, and report header sanitization (`internal/crash/crash.go:29-45, 65`, `internal/crash/validate.go:56-62`).
  - **Root cause:** `os.Getuid()` returns `-1` on Windows (`ownframe-crashes--1`); `validateDir` rejects `C:` (len 2) but allows Windows drive root `C:\` (len 3); `crash.go:29-37` accepts pre-existing loose/symlinked dirs in `/tmp` without verifying permissions/symlink status; `title`/`reason` newlines are not sanitized before writing crash headers.
  - **Fix:** Use a safe Windows username/SID fallback instead of `-1`, reject drive roots (`C:\`), verify existing crash dirs with `Lstat` (rejecting symlinks and tightening mode on Unix), and sanitize newlines in `title`/`reason`.
  - **Proof:** Unit + fuzz tests (`FuzzSetCrashDir`) in `internal/crash/`.

- [ ] `HYG-ENG-01` **(Gated on `blinkless` Phase 01)** Remove stale `internal/render/engineMu` serialization and `statePseudoReporter` reflection shim (`internal/render/engine.go:5-12`, `internal/render/state_pseudos.go:8-35`).
  - **Precondition (MANDATORY):** Do **not** remove `engineMu` until `blinkless` fixes `Node.EnsureAttrs()` and `prepare.SheetCache` concurrency races (`blinkless` Phase 01 `BLK-CONC-01`, `BLK-CONC-02`) and `ownframe/go.mod` + `examples/go.mod` bump to that `blinkless` commit/tag.
  - **Fix:** Replace `statePseudoReporter` duck-typing with direct `doc.UsesStatePseudo()` (`blinkless/css`), then remove `engineMu` once upstream race fixes are pinned.
  - **Proof:** Parallel multi-page render + relayout stress test passing under `go test -race ./internal/render ./internal/page`.

---

## Closure Gates

- [ ] `P2-GATE-01` `make lint`, `make docs-check`, `make plans-check`, `make examples-check`, `make tools-check`, `make repo-size-check`, and `make internal-import-check` all exit `0`.
- [ ] `P2-GATE-02` `make test` and `make test-race` pass on Linux and Windows CI (`native-package` covering `./internal/uia` and `./internal/window`).
- [ ] `P2-GATE-03` `make module-check` and `make vuln` exit `0` with `golang.org/x/perf` absent from the shipped library module.
