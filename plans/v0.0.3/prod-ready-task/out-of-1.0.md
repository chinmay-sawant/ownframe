# v0.0.3 - Out of 1.0

> Superseded planning record. `[s]` marks an old proposal or assertion, not proof of implementation. `[-]` marks work explicitly deferred in the original plan. Current decisions and evidence live in [plans/README.md](../../README.md).

> **Parent:** [README.md](README.md)
> **Estimated effort:** not estimated here. These are separate projects.

---

## Overview

The first production list mixed year-scale OS work and a widget kit into 1.0. This file holds that work so it cannot stall the cutoff. Promoting a row means: move it to a phase file, add it to the README cutoff table, and delete the `[~]` here with a pointer.

**Already in 1.0 (do not park here):** native desktop `Run` on Linux, Windows, and macOS; native Android `BindMobile` plus `examples/android-host`, `mobile/` packages, and signed arm64 APKs listed in `examples/android-host/APPS.md` (login, platform, dino, lg-remote, telegram, eyes). Docs for that host live in [phase-02-docs-and-hosts.md](phase-02-docs-and-hosts.md) (`DOC-01`, `DOC-18`..`DOC-21`).

This file is Play-grade Android, iOS, wasm-as-product, OS screen readers, engine holes, and widgets.

## Hosts that are not 1.0 product

- [s] `OUT-01` Android **Play** path only: process restore after death, TalkBack node, IME as a supported contract (not `exp/textinput` folklore), `package.sh` folding in APK builds, GOWORK-safe `ebitenmobile bind`. Next gate: a dedicated `plans/v0.0.3/` Android Play ledger. Owner: BindMobile + `examples/android-host`. Telegram and eyes stay 1.0 exhibits (`DOC-19`). Do not read this row as "Android is a demo."
- [s] `OUT-02` iOS Xcode project and `ebitenmobile bind -target ios`. Next gate: a repo you can `open`. Until then comparison docs must not list iOS as shipped (`DOC-06`).
- [s] `OUT-03` wasm `navigator.clipboard`, file picker, idle park, crash files. Next gate: wasm product review. 1.0 docs call wasm a canvas preview (`DOC-01`).
- [s] `OUT-04` `Serve` as a ship target: TLS, auth, tick, relayout, audio. Next gate: none. 1.0 keeps loopback PNG debugger (`DOC-07`). Serve recover (`OS-17`) may land as a small fix without promoting Serve.

## Native accessibility

- [s] `OUT-05` AT-SPI (Linux). Next gate: `plans/v0.0.3/a11y/` or a host adapter package. 1.0 is the banner (`OS-10`).
- [s] `OUT-06` UI Automation (Windows). Same.
- [s] `OUT-07` NSAccessibility (macOS). Same.
- [s] `OUT-08` Android accessibility node / TalkBack. Same, plus Java. Depends on OUT-01. 1.0 Android still ships without TalkBack (`DOC-20`).

## Engine and widgets

- [s] `OUT-09` `position: sticky` / `fixed` under window scroll in blinkless. Next gate: blinkless PR then pin (`PAINT-16`).
- [s] `OUT-10` `position: absolute` inside flex in blinkless. Same.
- [s] `OUT-11` `@container` + element patch coexistence. Engine.
- [s] `OUT-12` compositor layers for 43 CSS motion previews. See [optimizations/](../optimizations/README.md) planning only. Do not dual-track `[ ]` rows.
- [s] `OUT-13` dropdown that is a dropdown, date field, dialog, table widget kit. Next gate: HTML patterns in examples, not engine widgets. Ownframe paints HTML.
- [s] `OUT-14` per-screen handler registration framework. Apps switch on `box.Action`. A router package can wait.
- [s] `OUT-15` i18n in the engine. Apps use `golang.org/x/text` and their templates.

## Packaging and OS extras

- [s] `OUT-16` signed/notarized three-OS `package.sh`, icons, SmartScreen/Gatekeeper. Runbook only in 1.0 (`OS-18`).
- [s] `OUT-17` native IM verification matrix (ibus, fcitx, Windows IME, macOS). `OS-15`.
- [s] `OUT-18` Wayland clipboard protocol (if `OS-02` closed as `[~]` docs-only). Next gate: protocol or `wl-clipboard` with `exec.Command` args, security review.
- [s] `OUT-19` multi-window, tray, native menu, auto-update.

## Process theater

- [s] `OUT-20` SECURITY_AUDIT_4 50-item generic list. Next gate: `gosec` / `govulncheck` / `-race` findings only (`GO-BAD-10`).
- [s] `OUT-21` Chromium compositor phases 01-10 as 1.0. Stay in `optimizations/` as planning.
- [s] `OUT-22` deleting existing brand-clone examples. 1.0 freezes new ones (`HYG-01`), keeps the old trees (`HYG-03`).

## How to promote a row

1. Copy the ID into a phase file as `[ ]` with paths, proof, and gate.
2. Add it to the README cutoff only if 1.0 must wait for it.
3. Change this row to `[~]` with "moved to `phase-XX.md` ID" or delete it if fully moved.
4. Do not leave two active `[ ]` copies.

## Dependencies

None for 1.0 cutoff. README cutoff must not list these IDs.

## Risks

Someone will try to "just add TalkBack" inside the Page wrap PR. Reject that. Point here.

Someone will read OUT-01 as "Android is out of 1.0." It is not. Play-grade Android is out. The Activity, android-host, and APKs are in.
