# v0.0.3 - Phase 07: Dependencies, Binary Size, and Pre-1.0 Platform Readiness

> **Parent:** [README.md](README.md)
> **Status:** Open (baseline measured at `da4f12cc`)
> **Estimated effort:** Ongoing / milestone-gated (some items depend on upstream tags and OS hardware)

---

## Overview

At commit `da4f12cc`, three external-facing barriers stand between `ownframe` as an impressive personal toolkit and a durable, adoptable library:

1. **Bleeding-edge dependency pins:** `go.mod` and `examples/go.mod` require `go 1.26.11`, an alpha pseudo-version of `github.com/hajimehoshi/ebiten/v2` (`v2.11.0-alpha.0.20261103125056-58976b602532`), and a pseudo-version of `github.com/chinmay-sawant/blinkless` (`v0.1.2-0.20261118190733-7abfce32c375`), alongside ancient unmaintained indirects (`freetype-go` 2016, `graphics-go` 2016, `poly2tri-go` 2017, `xgb` 2021).
2. **Heavy compiled binary weight:** Compiled test binaries (`ownframe.test`, `page.test`, `window.test`, `loglens.test`, `code-view.test`) weigh **43–46 MB**, and wasm demos weigh **52–66 MB** uncompressed—weakening the "tiny binary vs WebView/Electron" value proposition.
3. **Platform & accessibility gaps for any "1.0" release:** Live screen-reader integration ships only on Linux AT-SPI (no Windows UIA, macOS VoiceOver, or Android TalkBack), GNOME/Wayland clipboard relies on shelling out to external `wl-clipboard`, and Windows/macOS IME validation remains open in `plans/README.md`.

## Executive Summary

- Move `blinkless` and `ebiten/v2` off pseudo-versions onto tagged releases before tagging `v0.0.3` / `v0.1.0`, audit the indirect dependency tree, and run `make vuln`.
- Profile minimal app binaries (`examples/login`, a bare `hello` app, and wasm builds) by symbol and embedded asset to quantify and shrink the 43–46 MB footprint.
- Replace external `wl-clipboard` dependency on Wayland with a native path (or clearly bounded fallback) and define concrete, measurable exit criteria for native accessibility adapters and IME validation prior to any `v1.0` claim.
- Run the complete verification battery and record before/after metrics in `plans/v0.0.3/chore-improvements/README.md`.

## Current Source Baseline (`da4f12cc`)

| Dimension | Baseline Measurement | Location / Notes |
|---|---|---|
| Go toolchain directive | `go 1.26.11` | `go.mod`, `examples/go.mod` |
| `blinkless` pin | `v0.1.2-0.20261118190733-7abfce32c375` | Pseudo-version in `go.mod` & `examples/go.mod` |
| `ebiten/v2` pin | `v2.11.0-alpha.0.20261103125056-58976b602532` | Alpha pseudo-version in `go.mod` & `examples/go.mod` |
| Compiled binary sizes | **43–46 MB** test binaries; **52–66 MB** `.wasm` demos | Unstripped test binaries (`ownframe.test` 43M, `page.test` 44M, `code-view.test` 46M) |
| Platform / a11y gaps | Linux AT-SPI only; GNOME/Wayland needs `wl-clipboard`; Windows/macOS IME open | Documented in `README.md`, `plans/README.md`, `documentation/accessibility-native.md` |

---

## Phase 7: Dependencies, Binary Size, and Pre-1.0 Platform Readiness

### 7.1 Stabilize Go, `blinkless`, and Ebiten dependency pins

- [ ] `DEP-01` Audit `go.mod` and `examples/go.mod` Go version directive (`go 1.26.11`): verify whether the library actually uses language/stdlib features added in patch `1.26.11` vs `go 1.26.0` (or latest stable minor), so external consumers on `1.26.x` are not forced into unnecessary toolchain downloads. Proof: test build at the minimum supported Go version recorded.
- [ ] `DEP-02` Tag a clean `blinkless` release (e.g., `v0.1.2` or `v0.2.0`) in the upstream `blinkless` repo and pin `go.mod` + `examples/go.mod` to that exact semver tag instead of a pseudo-version before cutting `ownframe v0.0.3`. Proof: `grep blinkless go.mod examples/go.mod` shows a clean semver tag (no `-0.2026...` suffix).
- [ ] `DEP-03` Track `ebiten/v2` `v2.11.0` stable release (or verify if `v2.10.x` stable suffices once `v2.11.0` features are audited); replace the `v2.11.0-alpha.0...` pseudo-version when stable lands, or document the exact upstream Ebiten commit/issue requiring alpha. Proof: `go.mod` uses a tagged Ebiten release OR explicit upstream blocker recorded with `[~]` and tracking URL.
- [ ] `DEP-04` Run `go mod graph` to attribute every indirect dependency in `go.mod` (specifically `codeberg.org/go-pdf/fpdf`, `tdewolff/canvas`, `yuin/goldmark`, `go-fonts/latin-modern`, `star-tex.org/x/tex`, `modernc.org/knuth`, `BurntSushi/freetype-go`, `graphics-go`, `xgb`) to either `blinkless` or `ebiten/v2`. Identify any indirects pulled in unintentionally by unused imports or test-only code in `blinkless` or `ownframe`. Proof: dependency attribution table recorded; any avoidable direct/indirect dependency pruned.
- [ ] `DEP-05` Run `make vuln` (`go tool govulncheck ./...`) on both root and `examples/` modules and record the full result. Proof: `make vuln` exits 0 with commit hash and date recorded.

### 7.2 Profile and reduce compiled binary size

- [ ] `DEP-06` Measure exact baseline binary sizes (unstripped, stripped `-ldflags="-s -w"`, and gzipped/UPX-excluded) for:
  1. Minimal `hello` desktop binary (no `components/`, no `ui/`),
  2. `examples/login` desktop binary,
  3. `examples/login` `GOOS=js GOARCH=wasm` binary (uncompressed + `.gz`).
  Proof: exact byte sizes recorded in a baseline table (built into scratch directory, not committed).
- [ ] `DEP-07` Run symbol and embedded-data profiling (`go tool nm -size <bin> | sort -rn | head -40` and `go build -work` / section breakdown with `size -A` / `readelf -S`) on the minimal `hello` binary to pinpoint the top 10 size contributors (suspected: embedded fonts in `go-fonts/latin-modern` / `blinkless`, 42 Twemoji PNGs in `internal/emoji`, `tdewolff/canvas`, `fpdf`, `star-tex`, `ebiten`, `internal/atspi`, `ui/properties_generated.go`). Proof: top-10 contributor breakdown with exact megabyte share recorded.
- [ ] `DEP-08` Gate heavy optional bundled assets behind opt-in or lazy registration where practical (for example: evaluate whether `internal/emoji` Twemoji PNGs, `internal/window` DevTools assets, or PDF/TeX/Markdown subpackages transitively imported from `blinkless` can be decoupled so minimal `ownframe` apps only pay for what they link). Proof: written savings estimate per item + implemented reductions with measured before/after binary sizes.
- [ ] `DEP-09` Investigate whether `blinkless` pulls `fpdf`, `star-tex.org/x/tex`, `modernc.org/knuth`, `go-fonts/latin-modern`, `tdewolff/canvas`, and `yuin/goldmark` into every `ownframe` binary even when only HTML+CSS layout is used; if so, file/track an upstream `blinkless` package-split issue so core HTML/CSS layout does not force PDF/TeX/Markdown dependencies onto GUI apps. Proof: upstream issue or PR link recorded + measured binary size impact once split.
- [ ] `DEP-10` Set an explicit target binary-size budget for a stripped minimal `ownframe` desktop binary (e.g., <15 MB stripped) and add a size-reporting step to CI release builds so regressions are caught immediately. Proof: CI job logs stripped binary sizes for `login` desktop and wasm builds.

### 7.3 Platform and accessibility readiness gates before any `1.0` claim

- [ ] `DEP-11` Investigate and implement (or scope cleanly) native Wayland clipboard support in `internal/clipboard` so GNOME/Wayland Linux desktop users do not silently lose copy/paste when `wl-clipboard` is absent from `PATH` (or surface a clear diagnostic error/capability check instead of silent failure). Proof: Wayland clipboard test / manual validation log on a real Wayland session.
- [ ] `DEP-12` Complete open `NATIVE-IME` validation on Windows (preedit, cancellation, conversion, commit, English + CJK layout switching) tracked in `plans/README.md`, and record host OS, IME, and test evidence. Proof: validation record added under `plans/validation/` and `NATIVE-IME` updated in `plans/README.md`.
- [ ] `DEP-13` Define an explicit, written **v1.0 Accessibility Decision** in `documentation/accessibility-native.md` and `plans/README.md`: either assign owners & milestones for Windows UI Automation (UIA), macOS NSAccessibility/VoiceOver, and Android TalkBack bridges before calling the library `v1.0`, OR explicitly scope `v1.0` to "internal/operator desktop tools (Linux AT-SPI live; Windows/macOS/Android screen-reader adapters post-1.0)" consistently across every doc. Proof: consistent policy across `README.md`, `documentation/platforms.md`, `documentation/accessibility-native.md`, and `plans/README.md`.
- [ ] `DEP-14` Verify macOS desktop host (`Run`, clipboard, filepick `osascript`, IME) and Windows desktop host (`Run`, clipboard, `comdlg32` filepick, custom chrome) on real hardware before tagging any `v0.1.0` or `v1.0` release, recording exact OS versions and results. Proof: validation records committed under `plans/validation/`.
- [ ] `DEP-15` Audit `examples/go.mod` release checklist (`documentation/release-checklist.md`) so the `replace github.com/chinmay-sawant/ownframe => ../` removal and `examples/v0.0.3` tag sequence is verified end-to-end by `make module-check` and `make release-check`. Proof: `make module-check` and `make release-check` exit 0.
- [ ] `DEP-16` Verify `scripts/android-release.sh` reproduces a working signed/debug APK cleanly after Phase 01 removes pre-committed APKs from git. Proof: fresh APK build + device/emulator install log recorded.

### 7.4 Final program closure gates

- [ ] `DEP-17` Run full verification suite across both modules:
  - `make test`
  - `make test-race`
  - `make lint`
  - `make docs-check plans-check examples-check tools-check`
  - `make internal-import-check module-check graphics-check site-check vuln`
  Proof: exact command list, commit SHA, and exit-0 outcomes recorded here.
- [ ] `DEP-18` Capture a final before-vs-after metrics table in [README.md](README.md) comparing `da4f12cc` against the closing commit across `.git` / `docs/` / `examples/` sizes, root file/method counts, `internal/page` & `internal/window` & `internal/replay` file counts, motion package count, `AGENTS.md` size, `examples/` internal-import count, and stripped binary size. Proof: completed metrics table in `plans/v0.0.3/chore-improvements/README.md`.
- [ ] `DEP-19` Reconcile completed rows back into the parent maintenance ledger (`plans/README.md`) with evidence pointers, without leaving duplicate active work across files. Proof: `make plans-check` exits 0.
- [ ] `DEP-20` Confirm `git status` is clean, no untracked `*.test` / `mem.out` / `.doc-snippets-*` leftovers remain, and the PR against `master` passes all GitHub Actions CI jobs. Proof: green CI run URL recorded.

---

## Dependencies

- `DEP-02` and `DEP-09` depend on upstream changes/releases in `github.com/chinmay-sawant/blinkless`.
- `DEP-03` depends on upstream `github.com/hajimehoshi/ebiten/v2` tagging `v2.11.0` (unless `v2.10.x` suffices).
- `DEP-12` and `DEP-14` require access to real Windows and macOS hosts (mark `[~]` with host blocker if unavailable, per `AGENTS.md` "Blocked work").
- `DEP-17`–`DEP-20` run only after Phases 01–06 have landed.

## Risks

- **Transitive binary weight owned by upstream:** If 30+ MB of the 43 MB binary comes from `blinkless` pulling PDF/TeX/font libraries or from Ebiten/purego, `ownframe`-only optimizations will have a hard floor until `blinkless` splits its optional renderers (`DEP-09`). Measure first (`DEP-07`) before optimizing blindly.
- **Unmaintained indirect dependencies (`freetype-go`, `graphics-go`, `xgb`):** These cannot be replaced inside `ownframe` alone; track them honestly via `make vuln` and upstream Ebiten/blinkless upgrades.

## Closure Gates

- [ ] `DEP-GATE-01` `blinkless` pinned to a tagged semver release; `make vuln` exits 0.
- [ ] `DEP-GATE-02` Minimal `hello` and `login` binary-size profile completed and documented, with CI size reporting active.
- [ ] `DEP-GATE-03` Honest, consistent pre-1.0 vs 1.0 platform & accessibility policy across all docs and `plans/README.md`.
- [ ] `DEP-GATE-04` All verification commands in `DEP-17` exit 0 on the final commit.
