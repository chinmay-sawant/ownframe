## Summary

Adds the live Linux AT-SPI bridge that issue #48 was opened to validate. A Linux `Run` now discovers the accessibility bus, registers the window, serves the page tree over D-Bus, and emits focus and structure events, so Accerciser and Orca can inspect the application.

## Motivation / context

- Issue: #48 tracks live validation with Orca and Accerciser.
- Plan: `plans/v0.0.3/examples/code-view/phase-08-accessibility-and-mobile.md`
- Before this change `internal/atspi` mapped the snapshot and tests used `FakeBus`; `NewLiveBus` reported `Live: false`.

## Changes

### Bus transport and registration

- Discover `org.a11y.Bus` through `AT_SPI_BUS_ADDRESS`, `org.a11y.Bus.GetAddress`, or the runtime socket paths.
- Authenticate with `EXTERNAL`, claim `org.a11y.atspi.ownframe.app` with a `-<pid>` fallback, embed with `org.a11y.atspi.Registry`, and reconnect after a dropped connection.
- Match replies by serial. Add the value codec and fix array lengths, variant alignment, and the `siiva{sv}` event shape.

### Served interfaces

- Serve `Accessible`, `Application`, `Component`, `Text`, `Value`, `Action`, and `Cache` with their properties.
- Return the application root at index -1 with the page document as its only active child.
- Map extents from the Ebiten window origin and size.
- Redact password values, keep password nodes editable, and withhold `SetText`.
- Return false for AT-initiated `DoAction` and `GrabFocus`.

### Events and lifecycle

- Emit focused, checked, enabled, sensitive, editable, active, and children-changed events.
- Fix text offset alignment and boundary handling.
- Close the bus when `Run` returns.
- Skip the accessibility walk unless the page generation changed.
- Keep window and Telegram tests off the developer's real bus.

### Documentation and tests

- Document the live bridge and its limits in `documentation/atspi.md`.
- Record the issue #48 validation matrix in `documentation/accessibility.md`.
- Update the generated site assets from this branch's documentation sources.
- Rebuild the six demo wasm bundles because `internal/window` changed.

## Impact

| Area | Impact |
| --- | --- |
| **Performance** | Linux only. The tree walk is skipped unless the page generation changed. No benchmark or engine hot path changed. |
| **Memory** | One D-Bus connection per running application, closed when `Run` returns. |
| **Behavior / correctness** | `Live` is true on a Linux desktop. Windows, macOS, Android, and `-web` behavior is unchanged. Password values stay empty. AT-initiated actions return false. |
| **API / CLI** | No public API change. |
| **Dependencies** | None. The client uses the standard library. |
| **Binary size / build time** | Linux-only code is build-tagged. Other targets keep the stub. |

## Breaking changes / migration

| Item | Migration |
| --- | --- |
| None | The bridge is experimental and uses the existing Linux `Run` wiring. |

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root.
- [x] No new `func (p *Page)` on `internal/page` for the public `Page`.
- [x] No new example `ownframe/internal/` imports.
- [x] No synonym names.
- [x] No new exported methods return host or internal state types.
- [x] Library code panics only for invariants. AT-SPI failures are logged or retried.
- [x] The new bus code is internal and bounded by deadlines, so the public `context.Context` rule does not apply.
- [x] No new example whose purpose is resembling Telegram, Teams, Spotify, or Wispr.

## Test plan

- [x] `make test`
- [x] `make lint`
- [x] `make build`
- [x] `make test-race`
- [x] `make docs-check plans-check examples-check`
- [x] `make site-check`
- [x] Manual bus check on Ubuntu 22.04 under X11: two instances register, the application root reports index -1, and libatspi reads the tree and focus events.
- [x] Accerciser 3.38 shows the application and the page tree.
- [ ] Orca 42 per-control speech still needs a normal desktop run. The startup utterance was captured, but control speech was not.

### Commands

```sh
make test
make lint
make build
make test-race
make docs-check plans-check examples-check
make site-check
```

## Screenshots / sample output

No screenshot. The live validation matrix and the libatspi listener snippets are in `documentation/accessibility.md#live-validation` and `documentation/atspi.md#validate`.

## Related issues

- Relates to #48. The live bridge is implemented; Orca per-control speech remains open.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`).
- [x] Labels applied.
- [x] Related issue uses `Relates to`, not `Closes`.
- [x] Filled body committed under `plans/PR/pr-issue-48-atspi-live.md`.

## Reviewer checklist

- [ ] Behavior matches the summary and test plan.
- [ ] No performance or benchmark changes in the diff.
- [ ] No unrelated changes.
- [ ] Public API unchanged.
- [ ] Related issue uses the correct keyword.
- [ ] No secrets or generated build output beyond `docs/`.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 30 | 3125 | 268 |
| `.gz` | 2 | Binary | Binary |
| `.html` | 1 | 1 | 1 |
| `.js` | 2 | 211 | 106 |
| `.md` | 11 | 354 | 107 |
| `.wasm` | 10 | Binary | Binary |
| **Total** | **56** | **3691** | **482** |
