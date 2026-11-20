# v0.0.3 - Phase 07: Dependencies, Binary Size, and Pre-1.0 Platform Readiness

> **Parent:** [README.md](README.md)
> **Status:** In progress (12/24 done: `DEP-01`, `DEP-04`–`DEP-08`, `DEP-10`, `DEP-11`, `DEP-13`, `DEP-15`, `DEP-GATE-02`, `DEP-GATE-03`)
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

- [x] `DEP-01` Audit `go.mod` and `examples/go.mod` Go version directive (`go 1.26.11`). Evidence: verified via `go list -m` that **zero** dependencies require `> 1.26.0` (`blinkless` `1.26`, `ebiten/v2` & `golang.org/x/*` `1.26.0`, `golangci-lint v2.15.0` `1.26`). `1.26.11` was set in `66ec4898` to match CI `GOTOOLCHAIN=local` security patch hygiene; `go.mod` + `examples/go.mod` can safely lower to `go 1.26.0` for downstream consumers while keeping CI `setup-go` and `go.work` at `1.26.11`.
- [ ] `DEP-02` Tag a clean `blinkless` release (e.g., `v0.1.2` or `v0.2.0`) in the upstream `blinkless` repo and pin `go.mod` + `examples/go.mod` to that exact semver tag instead of a pseudo-version before cutting `ownframe v0.0.3`. Proof: `grep blinkless go.mod examples/go.mod` shows a clean semver tag (no `-0.2026...` suffix).
- [ ] `DEP-03` Track `ebiten/v2` `v2.11.0` stable release (or verify if `v2.10.x` stable suffices once `v2.11.0` features are audited); replace the `v2.11.0-alpha.0...` pseudo-version when stable lands, or document the exact upstream Ebiten commit/issue requiring alpha. Proof: `go.mod` uses a tagged Ebiten release OR explicit upstream blocker recorded with `[~]` and tracking URL.
- [x] `DEP-04` Run `go mod graph` to attribute every indirect dependency in `go.mod` to either `blinkless` or `ebiten/v2`. Evidence (`go mod graph` at `da4f12cc`): **all** of `codeberg.org/go-pdf/fpdf`, `github.com/tdewolff/canvas`, `github.com/tdewolff/font`, `github.com/yuin/goldmark`, `github.com/go-fonts/latin-modern`, `star-tex.org/x/tex`, `modernc.org/knuth`, `github.com/BurntSushi/freetype-go`, `graphics-go`, `xgb`, `xgbutil`, and `github.com/ByteArena/poly2tri-go` enter the module graph via `github.com/chinmay-sawant/blinkless` (and its `tdewolff/canvas` / `tdewolff/font` dependencies), not via `ownframe` direct imports.
- [x] `DEP-05` Run `make vuln` (`go tool govulncheck ./...`) on both root and `examples/` modules and record the full result. Evidence (2026-11-20, `da4f12cc` / `deca07d9`): `make vuln` exited 0 (`No vulnerabilities found. Your code is affected by 0 vulnerabilities.`; 2 uncalled vulnerabilities in imported packages).

### 7.2 Profile and reduce compiled binary size

- [x] `DEP-06` Measure exact baseline binary sizes (unstripped, stripped `-ldflags="-s -w"`, and wasm + gzipped) for `examples/hello` and `examples/login` (built into scratch dir). Evidence (2026-11-20, Linux x86_64, `go 1.26.11`):
  - `examples/hello` desktop: **45 MB** unstripped, **33 MB** stripped (`-ldflags="-s -w"`).
  - `examples/login` desktop: **45 MB** unstripped, **33 MB** stripped (`-ldflags="-s -w"`).
  - `examples/login` wasm (`GOOS=js GOARCH=wasm -ldflags="-s -w"`): **53 MB** uncompressed, **14 MB** gzipped (`.wasm.gz`).
- [x] `DEP-07` Run symbol and section profiling (`size -A` and `go tool nm -size`) on `login-unstripped` to pinpoint the top size contributors. Evidence (2026-11-20):
  - **ELF sections (on-disk):** `.text` **14.4 MB**, `.gopclntab` **11.0 MB** (kept even when stripped), `.noptrdata` **4.4 MB**, `.rodata` **4.1 MB**, `.debug_*` **~9.1 MB** (removed by `-s -w`). Note: `crypto/internal/fips140/drbg.memory` (32 MB) sits in `.noptrbss` (0 bytes on disk).
  - **Top on-disk package/asset contributors (`go tool nm -size`):** `blinkless/internal/fonts/assets` (**2.95 MB** embedded fonts), `blinkless/internal/layout` (**1.85 MB**), `tdewolff/canvas` (**0.91 MB**, incl. two 256 KB sRGB LUTs), `runtime` (**0.81 MB**), `andybalholm/brotli` (**0.67 MB** static dictionary), `net/http` (**0.47 MB**), `go-text/typesetting/font/opentype/tables` (**0.40 MB**), `go-text/typesetting/harfbuzz` (**0.34 MB**), `crypto/tls` (**0.34 MB**), `ownframe/internal/window` (**0.29 MB**), `ebiten/v2/internal/shader` (**0.27 MB**), `ebiten/v2/internal/gamepaddb` (**0.20 MB**), `golang.org/x/image/font/gofont/gomedium` (**0.16 MB**).
- [x] `DEP-08` Gate heavy optional bundled assets behind opt-in or lazy registration where practical. Evidence: audited all `//go:embed` assets across `ownframe` vs upstream `blinkless`:
  - Core `ownframe/internal` embedded assets total only **42.6 KB (1.4%)**: `internal/emoji/png/*.png` (42 Twemoji PNGs = **36.7 KB**) and `internal/replay/*.kage` (2 Kage shaders = **5.9 KB**), while `internal/window/devtools*.go` has **0 B** of embedded assets (pure Go display ops).
  - Optional `components/chadcn/css/*.css` (76 files = **128.6 KB**) is only linked by consumers importing `components/chadcn`, and the 7 chart skins (63 files / 7,331 LOC / 259 KB) were moved out of the root module in `MOD-03`.
  - Trimming `ui/properties_generated.go` (`MOD-GATE-02`) reduced `ui.a` archive size from **12.37 MB -> 4.24 MB (-8.13 MB / -65.7%)** and `ui` symbols from **40,759 -> 11,928 (-70.7%)**.
  - Upstream `blinkless/internal/fonts/assets/*.ttf` (14 TTF files = **3.09 MB / 98.6%** of embedded runtime assets) dominates the embedded floor and is tracked under `DEP-09`.
- [ ] `DEP-09` Investigate whether `blinkless` pulls `fpdf`, `star-tex.org/x/tex`, `modernc.org/knuth`, `go-fonts/latin-modern`, `tdewolff/canvas`, and `yuin/goldmark` into every `ownframe` binary even when only HTML+CSS layout is used; if so, file/track an upstream `blinkless` package-split issue so core HTML/CSS layout does not force PDF/TeX/Markdown dependencies onto GUI apps. Proof: upstream issue or PR link recorded + measured binary size impact once split.
- [x] `DEP-10` Add a CI size-reporting step to release builds so desktop and wasm binary-size regressions are visible in CI logs. Evidence: added `Report packaged binary sizes` step in `.github/workflows/release.yml` (`supply-chain` job); YAML validated cleanly.

### 7.3 Platform and accessibility readiness gates before any `1.0` claim

- [x] `DEP-11` Audit Wayland clipboard feasibility and surface missing `wl-clipboard` tools explicitly. Evidence: confirmed GNOME Mutter does not implement `zwlr_data_control_v1`/`ext_data_control_v1` and `wl_data_device` requires GLFW/Ebiten focus serials, so `wl-copy`/`wl-paste` stays the zero-dep backend; added `ErrWaylandToolsMissing`, `Status`, and `Diagnose()` across all OS build tags in `internal/clipboard/` (preserving `errors.Is(err, ErrUnavailable)`); `go test -v -race ./internal/clipboard/...` (14 tests) and cross-OS `go vet` passed.
- [ ] `DEP-12` Complete open `NATIVE-IME` validation on Windows (preedit, cancellation, conversion, commit, English + CJK layout switching) tracked in `plans/README.md`, and record host OS, IME, and test evidence. Proof: validation record added under `plans/validation/` and `NATIVE-IME` updated in `plans/README.md`.
- [x] `DEP-13` Define an explicit, written **v1.0 Accessibility Policy** consistent across `README.md`, `documentation/platforms.md`, `documentation/accessibility-native.md`, and `plans/README.md`. Evidence: added `## v1.0 accessibility policy` to `documentation/accessibility-native.md` stating `v0.0.x` ships the Go snapshot + live Linux AT-SPI bridge (no `EditableText`) while Windows UIA, macOS VoiceOver, and Android TalkBack require dedicated platform owners and real AT validation before any general-purpose `v1.0` screen-reader claim; `make docs-check` exits 0.
- [ ] `DEP-14` Verify macOS desktop host (`Run`, clipboard, filepick `osascript`, IME) and Windows desktop host (`Run`, clipboard, `comdlg32` filepick, custom chrome) on real hardware before tagging any `v0.1.0` or `v1.0` release, recording exact OS versions and results. Proof: validation records committed under `plans/validation/`.
- [x] `DEP-15` Audit `examples/go.mod` release checklist (`documentation/release-checklist.md`) and verify end-to-end via `make module-check` and `make release-check`. Evidence: `make module-check` (`1382 archive entries; isolated consumer compiled with no workspace or replace`), `cd scripts && GOWORK=off go test ./cmd/module-check`, and `make release-check` (`2 archives verified; no release published`) all exited 0.
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
- [x] `DEP-GATE-02` Minimal `hello` and `login` binary-size profile completed and documented, with CI size reporting active. Evidence: `DEP-06`–`DEP-08` and `DEP-10` complete.
- [x] `DEP-GATE-03` Honest, consistent pre-1.0 vs 1.0 platform & accessibility policy across `README.md`, `AGENTS.md`, `documentation/platforms.md`, `documentation/features.md`, `documentation/accessibility-native.md`, and `plans/README.md`. Evidence: `DOCS-05`–`DOCS-08` and `DEP-13` verified under `make docs-check`.
- [ ] `DEP-GATE-04` All verification commands in `DEP-17` exit 0 on the final commit.
