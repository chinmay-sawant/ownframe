# v0.0.3 - Phase 2 field: Desktop OS glue

> Superseded planning record. `[s]` marks an old proposal or assertion, not proof of implementation. `[-]` marks work explicitly deferred in the original plan. Current decisions and evidence live in [plans/README.md](../../README.md).

> **Parent:** [README.md](README.md)
> **Estimated effort:** Wayland clipboard is a real feature (days to a week). Banner and filepick errors are a day. AT-SPI is a separate project.

---

## Overview

Desktop is the 1.0 host. Production on Linux is blocked for many apps by Wayland clipboard (memory-only) and by no native a11y. This phase makes those facts either true in code or true in the README. It does not notarize binaries.

Go practice: OS seams behind interfaces (already: clipboard `UseMemory`, `page.InstallPicker`), errors from syscalls, no shell concatenation (`internal/filepick` must keep `exec.Command` args separate).

## Executive summary

- Wayland paste works **or** clipboard docs say X11/Win/macOS+cgo.
- File picker failures return `error` with the fallback reason.
- A11y: Go tree plus banner. No AT-SPI/UIA/NSAccessibility in 1.0.
- `package.sh` stays unsigned. Signing is a runbook.

## Current source baseline

| Feature | Implementation | Hole |
|---|---|---|
| Clipboard | `internal/clipboard` X11, Win, macOS+cgo | Wayland, wasm, Android, macOS without cgo: memory |
| Filepick | zenity/qarma/matedialog/kdialog, WSL powershell, comdlg32, osascript | wasm/mobile/Serve: type a name |
| A11y | `Page.AccessibilityTree`, `accessibility.go` enums | no AT-SPI/UIA/NSAccessibility/TalkBack |
| IME | `exp/textinput` | native IM unverified |
| Idle | `internal/window/idle_policy.go` desktop | wasm/Android off |
| Crash | 0700 dir, 0600 files | wasm no file; Serve no recover |
| Package | `scripts/package.sh` trimpath `-s -w` SHA256 | unsigned, no icon |
| Drop | `Drop.Path` desktop absolute | sensitive; wasm empty |

## Checklist

### 2.1 Clipboard

- [s] `OS-01` inventory: document in `documentation/clipboard.md` the exact build tags and what `Read`/`Write` do on X11, Wayland, Windows, macOS cgo, macOS !cgo, wasm, Android. Proof: table with file cites (`clipboard_*.go`).
- [s] `OS-02` Wayland: implement OS clipboard **or** mark `[~]` with reason (no data-device protocol in-tree) and next gate "wl-clipboard helper vs native protocol." 1.0 may close with `[~]` **only if** OS-01 table is `[x]` and README does not imply paste works on GNOME Wayland.
- [-] `OS-03`  wasm `navigator.clipboard`. Next gate: [out-of-1.0.md](out-of-1.0.md). Do not advertise copy/paste on wasm.
- [s] `OS-04` public `Read`/`Write` after `API-15`. Tests stay on `UseMemory`. Proof: TEST-13.
- [s] `OS-05` never log clipboard contents. Proof: grep `clipboard` log/print in `internal/clipboard` empty of payload.

### 2.2 File picker

- [s] `OS-06` public error type or wrapped error when no backend is found (zenity missing, etc.). Today `OWNFRAME_FILEPICK_DEBUG=1` prints reasons to stderr. Move that into the returned `error`. Proof: test with picker stub / missing binary; stderr not required for the test to know why.
- [s] `OS-07` keep WSL powershell path conversion via `wslpath`. Proof: existing tests or a documented manual step; do not regress.
- [s] `OS-08` Command args stay separate (`exec.Command(name, args...)`). No `bash -c` with user paths concatenated. Proof: grep `bash -c` / `sh -c` in `internal/filepick` empty or reviewed.
- [s] `OS-09` wasm/mobile/Serve: typed name remains. Docs say so (DOC-05). Proof: `pick_other.go` still the stub.

### 2.3 Accessibility (1.0 banner)

- [s] `OS-10` README + `documentation/accessibility.md` + `documentation/accessibility-native.md` first paragraph: 1.0 ships the Go snapshot only. Native bridges are a later tagged project. Proof: three files agree.
- [-] `OS-11`  AT-SPI adapter. Next gate: new folder `plans/v0.0.3/a11y/` or out-of-1.0. Do not open a 1.0 `[ ]` row that blocks cutoff.
- [-] `OS-12`  UI Automation (Windows). Same.
- [-] `OS-13`  NSAccessibility (macOS). Same.
- [s] `OS-14` keep the Go tree accurate for the Elements/devtools path. Proof: existing a11y tests still pass after the Page wrap (`AccessibilityTree` may stay on inner page, exposed if public).

### 2.4 IME, idle, crash, package

- [-] `OS-15`  native IM verification (ibus/fcitx, Windows IME, macOS). Record one manual session or keep "unverified" in platforms.md. 1.0 does not require a new IME stack.
- [s] `OS-16` idle park stays desktop-only. Docs already. Proof: `idle_supported.go` / `idle_policy.go`; platforms.md idle row.
- [s] `OS-17` `Run` crash file 0600. Proof: crash tests. Wasm: recover, no file (document). Serve: `[~]` recover in 1.0 or document handler panic → 500. Next gate: Serve recover if Serve stays a debugger (low priority).
- [s] `OS-18` `scripts/package.sh` unsigned is accepted for 1.0. Add a **runbook** sentence: signing and notarization are release-operator steps, not library code. Proof: `documentation/packaging.md` or `documentation/release-checklist.md`.
- [-] `OS-19`  icons in package.sh. Next gate: packaging follow-up. Do not block 1.0.

### 2.5 Drop.Path

- [s] `OS-20` public docs (DOC-12) plus a code comment on `host.Drop` / `Drop.Path`: copy bytes in the callback. Proof: comment on the struct field.
- [s] `OS-21` keep 16 MiB read cap if present. Proof: cite the constant and a test.

## Dependencies

- Public clipboard names: `API-15`.
- README banner: `DOC-01`, `DOC-02`.
- Filepick public error: may need a root wrapper next to the Page wrap.

## Risks

- Implementing Wayland badly (shelling out to `wl-copy` with concatenated args) is a command-injection bug. If OS-02 is code, use `exec.Command` args only, or a protocol library reviewed in that PR.
- A11y theater (Java-style empty providers) is worse than the banner. 1.0 chooses the banner.

## Closure gates

1.0 closes this phase when OS-01, OS-04, OS-06, OS-08, OS-10, OS-16, OS-18, OS-20 are `[x]`, and OS-02 is either `[x]` or `[~]` with docs that match.

- [s] `make lint` and `make test` if Go changed
- [s] clipboard tests still `UseMemory`
- [s] platforms.md clipboard row matches OS-01
