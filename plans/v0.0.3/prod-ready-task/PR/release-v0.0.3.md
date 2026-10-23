# Draft release note for v0.0.3

This is a proposed release note for review alongside the production-readiness PR. OwnFrame remains in the v0.0.3 development cycle. This note does not announce a published release.

## What changed

- `ownframe.Page` is now a public wrapper with a defined method set and private page state. Page controls are grouped under `Form`, `Edit`, `Scroll`, `Window`, `Access`, and `Probe`.
- The app-facing package no longer exposes the host screen type. `NewGame` gives tests a supported way to drive the same Ebiten game loop used by native hosts.
- Page drawing, form state, accessibility snapshots, scrolling, host pacing, and repaint work have clearer package owners.
- Replayable pages keep a display list and repaint dirty tiles. Other pages use the picture fallback. Performance counters and pipeline timing remain opt-in.
- Desktop windows can wait while clean and wake for input or a page's requested frame deadline. The window also handles the documented focus, scroll, touch, drop, text input, context menu, and devtools behavior.
- The library and examples remain separate Go modules. The checks validate the library archive, a consumer outside the workspace, generated website output, release archives, checksums, and reproducible Linux and wasm packages.
- Documentation and examples describe the supported hosts and the limits that remain before 1.0.

## Supported host status

The 1.0 native-host scope is desktop Linux, Windows, and macOS through `Run`, plus Android through `BindMobile`. Wasm is a canvas preview. `-web` is a loopback PNG debugger.

OwnFrame does not ship a live screen-reader adapter. The AT-SPI mapping has no live D-Bus client. UI Automation, VoiceOver, and TalkBack are not shipped. Native Windows composing-input validation remains open, and macOS input validation remains deferred pending a host.

## Validation record

The branch contains passing two-module tests, race detection, lint, graphics, archive, website, reproducibility, performance, and Android device evidence on the named snapshots in `plans/validation/`. Hosted CI for the final PR head must also pass the Android bind job after the SDK setup fix. The record in `plans/README.md` identifies which results are current and which checks remain open.

No version tag, production artifact publication, or GitHub Release is part of this PR.
