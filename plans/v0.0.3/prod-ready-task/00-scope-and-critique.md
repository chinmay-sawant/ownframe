# v0.0.3 - Scope, critique, and 1.0 cutoff

> **Parent:** [README.md](README.md)
> **Status:** reference. No implementation rows. Do not mark this file complete.
> **Estimated effort:** none. Read before Phase 0.

---

## Overview

This file records why the kitchen-sink production list was rejected, what 1.0 is, and which hosts are in. Implementation lives in the phase files. If a later PR disagrees with this scope, edit this file in the same PR as the phase row it changes.

## What 1.0 is

A stranger runs `go get github.com/chinmay-sawant/ownframe` and ships a **native Go app** on desktop and on Android. Same `Page`, same handlers.

**Native** here means one compiled Go process, blinkless layout, Ebiten drawing, no Chromium, no WebView, no JavaScript VM. Desktop `Run` opens an OS window on Linux, Windows, and macOS. Android `BindMobile` plus `examples/android-host` is a real `EbitenView` Activity and an installable APK. That is the same kind of native as Gio, Fyne, and Flutter.

It is not AppKit, WinUI, GTK, or Jetpack Compose widgets. It is not TalkBack, AT-SPI, UIA, or VoiceOver.

The public package is small. Errors are `error` values. Examples compile in a third-party module. The README names two native hosts (desktop and Android) and names wasm plus `-web` as previews. iOS is absent.

The engine (blinkless layout, Ebiten replay, dirty tiles, idle park) already exists. 1.0 does not invent a browser. It makes the library adoptable.

## What 1.0 is not

- A Fyne widget kit.
- A Wails/Electron replacement for existing web apps.
- A Play Store or App Store program (no TalkBack, no process restore contract, no iOS).
- Native screen readers.
- OS widget toolkits.
- A 500-row chat that re-executes the whole template.
- wasm or `Serve` as a shipped product.

## Critique of the first production list

The first list treated every wound as P0: wrap `Page`, native a11y on three OSes, a widget kit, Android process restore, Wayland paste, signed notarized packages, i18n, per-screen handlers. That is a company roadmap. It is not a Go 1.0.

It fought the product. Dropdowns, date pickers, and tables are widgets. This library paints HTML. Production is a deep `Page` (few methods, lots of behavior), a CSS contract, and OS glue that does not lie.

It used the wrong lever on files. The 2000-character gate made 381 files in `internal/page`, synonym methods, and comment-only `.go` files. Go wants related methods on one type, unexported until needed. Raising the cap without splitting packages grows the god type.

It was an API trap. Wrapping `Page`, deleting aliases, and "freezing v0.1" in one breath is a breaking change. At v0 that is legal. After v1 those names are forever. The first list never said "break now, while the major is 0."

It ignored Go's production bar: `go test -race` as CI, fuzz at trust boundaries, `govulncheck` as a kept gate, exported-API tests, `errors.Is` sentinels, one blinkless pin, examples `replace` during development.

It kept bad shapes: `type Page = page.Page`, examples importing `internal/`, `PatchElement` returning `bool`, `Stats()` returning `host.Stats`, Fetch and IPC in the root package, global `SetCrashDir` without a constructor story.

## The first list (kept for contrast)

Do not execute this list as a 1.0. Rows that survive live in the phase files with IDs.

**API and SDK**
- Wrap `Page`. ~20 user methods.
- One name per idea.
- Export clipboard, crash, click-test helper at root.
- `PatchElement` returns `(bool, error)`.
- Freeze v0.1.

**Layout and paint**
- CSS matrix. Written "we will not implement X."
- Sticky/fixed without folklore, or document the language.
- Tile path paints pinned layers.
- Virtualized list helper.
- Ticks dirty tiles by default.
- Tagged blinkless releases.

**Desktop OS**
- Native a11y or an AT banner.
- Wayland clipboard and wasm clipboard.
- File picker without tribal knowledge.
- Signed three-OS packages.
- IME verified on a native IM.

**Hosts**
- Honest README.
- Drop iOS from peer lists.
- Android restore, TalkBack, IME, or examples-only.
- `Serve` as loopback debugger.

**App building blocks**
- 5-screen non-clone sample.
- Login uses `data-bind`.
- Dropdown, date, dialog, table, or out of scope.
- Per-screen handlers.
- i18n.

**Quality bar**
- Kill or raise the 2000-char cap on examples.
- Keep visual tests. Stop `*_more_test.go`.
- Docs match `go.mod` and `.golangci.yml`.
- Move plans/skills/PHASES out of the library.
- Freeze the clone factory.

**Security and ops**
- Close the last audit before writing another.
- Crash recover on all `Run` paths.
- Document `Drop.Path` in public docs.

Old cutoff: API wrap + CSS matrix + Wayland paste + a11y + list windowing + honest README.

## What changed in the new plan

- Scope is **native desktop** (`Run` on Linux, Windows, macOS) and **native Android** (`BindMobile` + `examples/android-host`). Wasm and `-web` are documented previews. iOS is absent.
- Android 1.0 is the host that already ships: shared Java helpers, `mobile/` bind packages, signed arm64 APKs (login, platform, dino, lg-remote, telegram, eyes). Play Store, TalkBack, and process restore stay in [out-of-1.0.md](out-of-1.0.md).
- Sequence: stop the bleed, wrap the API, tell the truth, then paint, then teaching.
- `PatchElement` returns `error`, not `(bool, error)`. Callers use `errors.Is`.
- Clipboard and tests live in **named packages**, not a junk drawer on `Page`.
- Native a11y (OS screen readers) is a banner in 1.0 and a separate project after. The Go snapshot stays.
- List windowing is an example recipe, not a virtual DOM.
- File-size rule: split packages, then decide the cap for examples. Do not violate AGENTS.md in the same breath as "kill the cap."
- Security gate is `-race`, existing fuzz, `gosec`, `govulncheck`. Not SECURITY_AUDIT_4.
- v0 may break. Document that in CONTRIBUTING before the wrap PR.

## Current source baseline (2026-10-22)

| Claim | Evidence | 1.0 gap |
|---|---|---|
| Public API is a type alias | `api.go` `type Page = page.Page` | Wrapper |
| Synonym blur ids | `internal/page/form_focus_last.go` `LastFocusID`, `PreviousFocusID`, `BlurredFocusID`, `LastFocused` | Delete three |
| Synonym state | `internal/page/page_patch_state.go` `State` / `StateStore` / `Store` / `SetStateStore` / `AttachState` / `BindStore` | Keep `State` / `SetState` / `BindState` |
| Synonym props | `internal/page/page_patch_props.go` seven readers of one slice | Keep `ChangedProperties` and `TakeChangedProperties` |
| Synonym elements | `internal/page/page_element.go` `Element`, `ElementHandle`, `RetainedElement`, `RetainedNode`, `HandleElement` | Keep `Element` |
| Patch is bool | `internal/page/page_patch_element.go` `PatchElement(...) bool` | `error` + sentinels |
| Stats leaks host | `internal/page/page_stats.go` `Stats() host.Stats` | Named public type or hide |
| State leaks internal | `State() *state.Store` | Re-export or hide |
| Login imports internal | `examples/login/login/login_logout.go` `internal/clipboard`; `login_state.go` `internal/secret` | Public clipboard/secret |
| No root clipboard | no `clipboard.go` at module root | Export |
| Fetch and IPC at root | `fetch.go`, `ipc.go` | Subpackage or freeze, do not grow |
| Examples pin v0.0.2, second blinkless, no replace | `examples/go.mod` lines 8-11 | `replace` in repo, one pin |
| CONTRIBUTING Go version stale | `CONTRIBUTING.md` says 1.26.4; `go.mod` is 1.26.10 | Sync |
| README a11y stale | `README.md` "no accessibility tree" while `accessibility.go` re-exports a tree | Banner |
| Tiles bail on pins | `internal/window/replay_partial.go` calls `hasViewportOps()` | Phase 3 |
| Tick dirty is opt-in | `internal/page/page_frame.go` `UseFrameDirty` | Phase 3 |
| Race not library-wide CI | `.github/workflows/ci.yml` `make test` has no `-race` | Phase 1 CI |
| Fuzz exists for some targets | `ci.yml` sec job: emoji, web, fetch, page DOM/image | Extend crash dir and template |
| govulncheck exists | `make vuln`, `.github/workflows/vuln.yml`, `dependency-review.yml` | Keep, do not replace with audit markdown |
| Android native host | `BindMobile` in `run.go`; `examples/android-host` (insets, Back, lifecycle, touch cancel); `scripts/android-release.sh`; APKs in `examples/android-host/APPS.md` (login, platform, dino, lg-remote, telegram, eyes) | Docs name it native; Play/TalkBack/process restore stay out of 1.0 |
| Eyes Android | `examples/eyes`: landscape, device motion, tick, `scripts/android-eyes.sh`, `eyes-arm64-release.apk` | Exhibit next to telegram |

## Host policy

| Host | 1.0 role | Owning file |
|---|---|---|
| Desktop `Run` / `RunWithOptions` | Native product (Linux, Windows, macOS) | Phase 1 + 2 |
| Android `BindMobile` | Native product (Activity + APK via android-host) | [phase-02-docs-and-hosts.md](phase-02-docs-and-hosts.md) |
| Wasm canvas | Preview | same |
| `Serve` PNG | Loopback debugger | same |
| Play Store / TalkBack / process restore | Out of 1.0 | [out-of-1.0.md](out-of-1.0.md) |
| iOS | Absent | same |

## Dependencies

None. This file is the scope lock for the folder.

## Risks

If this file and a phase file disagree, this file wins until someone edits both in one PR.
