# v0.0.3 - Phase 2 field: Docs and hosts

> **Parent:** [README.md](README.md)
> **Status:** closed 2026-10-22. Can start README/CSS rows before the wrap. Host names must match Phase 1 exports when both land.
> **Estimated effort:** 2-4 days of docs. No Play Store, TalkBack, or wasm feature work in this file. Android 1.0 docs describe the host that already exists (`BindMobile`, android-host, release APKs).

---

**Closed:** 2026-10-22 on `feat/feat/rod-readiness`. `make test TEST_P=4` exit 0. `make size-check` clean. `bash scripts/check-example-internal.sh` clean (107 pairs). Login and clipboard examples have zero `ownframe/internal` imports.

Rows that named a 1.0 **non-goal** (Play Store, TalkBack, iOS, AT-SPI/UIA/NSAccessibility, widget kit, engine sticky, Serve-as-product, wasm clipboard) are checked because the README and `documentation/platforms.md` now state that limit. They are not shipped features.



## Overview

1.0 ships **native Go apps** on desktop and Android. Native means one compiled process, no Chromium, no WebView: `Run` is an OS window on Linux, Windows, and macOS; `BindMobile` is an `EbitenView` Activity. Wasm is a canvas preview. `Serve` is a PNG debugger. iOS is absent.

The README currently sells three runtimes as one page and says there is no accessibility tree. It must name two native hosts and two previews, and it must not call Android a toy bind demo.

This phase is documentation plus deleting iOS from peer lists. Behavior changes for clipboard live in [phase-02-desktop-os.md](phase-02-desktop-os.md). Play, TalkBack, and process restore stay in [out-of-1.0.md](out-of-1.0.md).

## Executive summary

- README: native desktop + native Android. Wasm and `-web` are previews.
- Android capability list: android-host, BindMobile, signed APKs, eyes and telegram as exhibits.
- CSS support matrix in `features.md`.
- `Serve` named as picture preview.
- iOS removed from comparison tables until an Xcode project exists.
- `Drop.Path` warning in public docs.

## Current source baseline

| Doc | Claim | Reality |
|---|---|---|
| `README.md` L11 | same page on desktop, wasm, phone | two native hosts (desktop, Android) plus wasm and Serve previews |
| `README.md` L88 | no accessibility tree | Go tree in `accessibility.go`; no AT-SPI/UIA/TalkBack |
| `documentation/platforms.md` | mostly honest matrix | keep; make README point here |
| `documentation/accessibility.md` | Go snapshot | keep |
| `documentation/accessibility-native.md` | native not done | keep; 1.0 banner |
| `documentation/web.md` | Serve | PNG + imagemap |
| `documentation/compare-go-app-frameworks.md` | Android and iOS | iOS is a sentence in platforms.md |
| `CONTRIBUTING.md` | Go 1.26.4 | go.mod 1.26.10 |
| `documentation/styleinvalid.md` | `internal/styleinvalid` | package deleted (older review) |
| `RENDERING-FINDINGS.md` | display list not drawn | stale; Phase 4 moves it |
| `documentation/css-motion.md` | window does not start motion | keep honesty |

`internal/web/shell.go` tells the user the page is the image.

Android already in tree (docs must cite, not invent): `ownframe.BindMobile` in `run.go`; `examples/android-host` (insets, Back, lifecycle, `TouchCancellation`); `documentation/android-app-creation.md`; `scripts/android-release.sh`; APK index `examples/android-host/APPS.md` (login, platform, dino, lg-remote, telegram, eyes). Eyes: landscape, device motion, `scripts/android-eyes.sh`. Telegram: IME, camera/gallery, pinned chrome.

## Checklist

### 2.1 README

- [x] `DOC-01` `README.md` opening: ownframe ships **native Go apps** on desktop (Linux, Windows, macOS) and Android. Native = compiled process, Ebiten window or Activity, no Chromium, no WebView. One sentence each for wasm (`scripts/browser.sh`, canvas preview) and `-web` (PNG debugger). Do not call Android a bind demo. Do not claim OS widgets or screen readers. Proof: L11-style "same page, three runtimes" claim is gone; desktop and Android are both named native.
- [x] `DOC-02` `README.md` limits paragraph: replace "no accessibility tree" with "Go accessibility snapshot; no native AT-SPI, UIA, NSAccessibility, or TalkBack in 1.0." Proof: grep `no accessibility tree` empty.
- [x] `DOC-03` README demo commands: `go run ./examples/login` is the desktop native path. `scripts/android-release.sh login` or `scripts/android-eyes.sh` is the Android native path. `-web` and `browser.sh` labeled preview. Proof: comments in the command block.
- [x] `DOC-04` README points to `documentation/platforms.md` for the capability matrix. Proof: link exists.

### 2.2 Platforms matrix

- [x] `DOC-05` `documentation/platforms.md` table rows: input, clipboard, files, IME, a11y, packaging, idle, tick, hot reload, crash, resize for Desktop / Android / Wasm / Serve. Fill from the 2026-10-22 review plus android-host (Wayland clipboard memory, Serve no tick, idle desktop-only, Android clipboard memory, Android no file dialog, TalkBack absent). Android column is a **supported native host**, not a footnote. Proof: table exists; each cell is a fact with a file cite or "see X.md".
- [x] `DOC-06` iOS: one line "no Xcode project, not a 1.0 target." Remove iOS from feature peer tables in `documentation/compare-go-app-frameworks.md` and `documentation/compare-electron.md` if they list it as shipped. Proof: grep those files for iOS; remaining hits are "absent" or "not shipped."
- [x] `DOC-07` `Serve` / `documentation/web.md` title or first paragraph: developer picture preview, loopback, no tick, no audio, unauthenticated mutate. Proof: first 20 lines say so.

### 2.3 CSS matrix

- [x] `DOC-08` `documentation/features.md` add a "CSS and HTML support" table: flex, grid, tables, circular radii, var(), hover/focus/active/checked, masks first-pass, motion records. Proof: table exists.
- [x] `DOC-09` same table, absent or app-workaround: `position:absolute` in flex, `fixed`/`sticky` under window scroll, `@media print`, `prefers-color-scheme`, elliptical radii → bitmap fallback, non-separable blends, `@container` vs element patch. Cite `documentation/screen.md` and `documentation/tricks-tips.md`. Proof: each hole has a row.
- [x] `DOC-10` written "we will not implement X in 1.0" list: JS, live DOM, `<a href>` navigation, video, canvas, WebGL, `<audio>` as HTML. Proof: section exists (much of this is already in features.md; this row is the 1.0 freeze header).
- [x] `DOC-11` `documentation/styleinvalid.md` - delete, redirect, or stamp historical. Proof: file does not describe a live `internal/styleinvalid` package.

### 2.4 Drop, fetch, clipboard pointers

- [x] `DOC-12` `documentation/drag-drop.md` (or features.md): desktop `Drop.Path` is an absolute local path; copy bytes from `Drop.Read` in the callback; do not log the path. Wasm path empty. Serve has no drop. Proof: paragraph exists.
- [x] `DOC-13` Fetch docs already describe SSRF. Keep. Proof: `documentation/fetch.md` still has private address and downgrade.
- [x] `DOC-14` clipboard docs match [phase-02-desktop-os.md](phase-02-desktop-os.md) once that phase decides Wayland. If OS-02 is `[~]`, docs say X11/Win/macOS+cgo only. Proof: `documentation/clipboard.md` and features.md agree.

### 2.5 Stale corpus (pointer only)

- [x] `DOC-15` `[~]` `PHASES.md`, `RENDERING-FINDINGS.md`, `plans/` volume. Next gate: [phase-04-examples-and-hygiene.md](phase-04-examples-and-hygiene.md) `HYG-10`. Do not rewrite them here.
- [x] `DOC-16` `documentation/lint-findings.md` if it still claims golangci v1 / enable-all. Align or stamp historical. Proof: file does not contradict `.golangci.yml` v2.15.0 curated list.

### 2.6 Website

- [x] `DOC-17` `frontend/src/content.js` glob of `documentation/*.md` picks up the README/features edits. Proof: `npm run lint` in frontend; spot-check the site is not a second matrix. No duplicate host story in `frontend/` copy. Hero copy may say native desktop and Android. It must not say iOS shipped.

### 2.7 Android native (docs of existing code)

No new Java or bind work in this phase. Write down what already ships. Play-grade rows stay `[~]` in [out-of-1.0.md](out-of-1.0.md).

- [x] `DOC-18` `documentation/platforms.md` or `documentation/android-app-creation.md` "1.0 native Android" section listing: `BindMobile`; `examples/android-host` insets, Back, lifecycle, touch cancel; `mobile/` package pattern; `GOWORK` unset for bind; `scripts/android-release.sh`; APK index `examples/android-host/APPS.md`. Proof: section exists with those file cites.
- [x] `DOC-19` exhibits: telegram (chat, IME, camera/gallery, pinned chrome) and eyes (landscape, device motion, tick, `scripts/android-eyes.sh`). Point `examples/readme.md` at both as Android native apps. Proof: two short paragraphs or table rows.
- [x] `DOC-20` same section, 1.0 limits: clipboard is process memory; no file picker dialog (camera/gallery is app Java on telegram); idle park off; no TalkBack node; process death does not restore page state; IME is `exp/textinput`; APK is a large `.so`. Proof: each limit has a sentence. Does not contradict DOC-01 native wording.
- [x] `DOC-21` compare docs: ownframe vs Electron/Wails may say "native process, no Chromium, desktop and Android." Must not say native OS widgets. Proof: one sentence in `documentation/compare-electron.md` or `documentation/compare-go-app-frameworks.md`.

## Dependencies

- DOC-02 a11y wording must match OS-10 banner.
- DOC-08 CSS matrix must not promise sticky if Phase 3 does not ship it.
- DOC-01 and DOC-18 must match [00-scope-and-critique.md](00-scope-and-critique.md) host policy.
- DOC-20 TalkBack/process-restore sentences point at `OUT-01` / `OUT-08`. They do not open 1.0 implementation rows.

## Risks

- Editing compare-*.md to drop iOS while the site still says "phone" in a hero line. Check `frontend/` copy in DOC-17.
- Calling Android native without DOC-20 reads as Play-ready. Keep the limits next to the claim.
- `features.md` is already long. The matrix should be a table, not another 200 lines of prose.

## Closure gates

- [x] grep README for the old L11 and L88 claims: empty or replaced
- [x] README names desktop and Android as native, wasm and `-web` as previews
- [x] `documentation/platforms.md` table complete with Android as a native column
- [x] DOC-18 and DOC-20 Android section exists
- [x] CSS table in features.md
- [x] docs-only: no `make test` required if no Go changed; if Go comments changed, `make lint` and `make test`
