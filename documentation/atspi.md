# Linux AT-SPI first slice

This is experimental. It maps `host.AccessibilityTree` to AT-SPI roles,
states, and actions. It does not register an accessible application on the
session bus. Orca, Accerciser, and other AT-SPI clients cannot inspect an
ownframe window through this slice.

Windows UI Automation, macOS VoiceOver / NSAccessibility, and Android
TalkBack are not shipped. iOS is not a 1.0 target.

## What landed

Package `internal/atspi` converts a page snapshot into AT-SPI objects:

- Roles: document web, landmark, panel, heading, paragraph, static, image,
  link, push button, entry, password text, check box, radio button, combo
  box, list, list item.
- Actions: `grab-focus`, `click`, `set-text`.
- Password nodes keep an empty value. Default logs use `Summarize`, which
  prints role and path only.
- The well-known name this process may claim is
  `org.a11y.atspi.ownframe.*`. `org.a11y.Bus` and the registry are rejected.

`FakeBus` records claim, publish, and event calls for tests. It opens no
socket. `NewLiveBus` always reports `Live: false`.

Linux desktop `Run` constructs a publisher in `internal/window/a11y_atspi.go`
and calls it from `syncImage`. Because the live bus is down, the window
does not walk the page tree each frame.

## What is blocked

A D-Bus AT-SPI client: connect to the accessibility bus, own
`org.a11y.atspi.ownframe.app`, export
`/org/a11y/atspi/accessible/*`, and register with
`org.a11y.atspi.Registry`. That client is not in this slice. No new D-Bus
dependency was added.

Do not set an environment flag that claims the window is exposed to a
screen reader.

## Validate

```sh
go test -p 1 ./internal/atspi ./internal/page ./internal/window
```

Those tests use the fake bus. They do not prove Orca can read the window.
After a live client exists, use Accerciser and Orca as
[accessibility-native.md](accessibility-native.md) describes.
