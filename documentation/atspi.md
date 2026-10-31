# Linux AT-SPI live bridge

This is experimental. On a Linux desktop, `NewLiveBus` connects to the
accessibility bus and registers the window with the AT-SPI registry. Accerciser
and Orca can inspect the tree. Windows UI Automation, macOS VoiceOver /
NSAccessibility, and Android TalkBack are not shipped, and iOS is not a 1.0
target.

## What the bridge does

`internal/atspi` exposes a page snapshot as a live AT-SPI application:

- Bus discovery. `AT_SPI_BUS_ADDRESS` wins when set. Otherwise the client asks
  `org.a11y.Bus.GetAddress` on the session bus, then falls back to
  `$XDG_RUNTIME_DIR/at-spi/bus*` and `/run/user/<uid>/at-spi/bus*`.
- Registration. The client claims `org.a11y.atspi.ownframe.app`, or that name
  with a randomized per-process suffix when it is taken, uses the unique bus
  name for object references, and calls `org.a11y.atspi.Socket.Embed` on the
  registry root. The unique name lets the registry remove the application when
  the connection closes. If the claim fails, publication stays unavailable and
  the failure reason is reported.
- Wire protocol. `dbus_wire.go` frames D-Bus messages.
  `dbus_value.go` encodes and decodes the scalar, struct, array, variant,
  and dictionary values the AT-SPI interfaces need. Replies are matched by
  serial and sender, so a forged reply or a queued `NameAcquired` signal does
  not confuse a call.
- Interfaces. Every object serves `org.a11y.atspi.Accessible` and
  `org.a11y.atspi.Component`. Text-bearing nodes add
  `org.a11y.atspi.Text`; combo boxes add `org.a11y.atspi.Value`; nodes with
  actions add `org.a11y.atspi.Action`; the root adds
  `org.a11y.atspi.Application`. The application also serves
  `org.a11y.atspi.Cache` at `/org/a11y/atspi/cache`, so libatspi queries the
  tree in one call instead of one round trip per node.
- Root shape. The application root is a synthetic `application` object marked
  active, and the page document is its only child. This is the shape Orca
  expects when it looks for the active window.
- Events. Focus and state changes emit
  `org.a11y.atspi.Event.Object.StateChanged`; added and removed nodes emit
  `ChildrenChanged`. Both use the standard `siiva{sv}` signature.
- Redaction. Password values are empty. Password nodes stay editable so a
  screen reader reports a password field, but they do not advertise SetText
  and their value is never published.
- Bounds. Extents use CSS pixels from the page snapshot. Screen coordinates add
  the Ebiten window position; root extents use that window position and size
  when the document node has no placed box. A parent-relative request subtracts
  the parent's origin. Ebiten reports the window position relative to the
  current monitor, so screen extents can be offset on a secondary monitor and
  on Wayland. The primary X11 monitor is the validated case.

## Limits

The bridge is a read-mostly surface:

- `DoAction` and `GrabFocus` return false. AT-initiated edits do not reach the
  page yet.
- There is no `EditableText`, `Selection`, `Document`, `Table`, or `Hyperlink`
  interface, so `SetTextContents` is not exposed. Text caret offset and
  selection are reported as zero and empty.
- Link activation is not supported.
- Anonymous labels whose tag and text match more than one placed box have no
  bounds. The snapshot still contains them.
- Text ranges and character extents use the node box, not per-glyph geometry.
- The page snapshot does not read computed `display:none` or
  `visibility:hidden`, so a stylesheet-hidden node can remain in the tree.
- Scrolling changes the page snapshot bounds; the bridge does not maintain a
  separate scroll offset beyond what the snapshot reports.

## Validate

Unit and mock-bus checks:

```sh
go test -p 1 ./internal/atspi ./internal/page ./internal/window
```

Live tree check with the Atspi 2.0 client library (the same library Accerciser
and Orca use):

```sh
# Terminal 1: start a desktop app. The login example is small.
go run ./examples/login

# Terminal 2: read the desktop tree.
/usr/bin/python3.10 - <<'PY'
import gi
gi.require_version("Atspi", "2.0")
from gi.repository import Atspi
Atspi.init()
desktop = Atspi.get_desktop(0)
for i in range(desktop.get_child_count()):
    app = desktop.get_child_at_index(i)
    print(app.get_name(), app.get_role_name(), app.get_child_count())
PY
```

The app should appear with the page title (for the login example, `Sign in`),
the `document web` role, and its controls as children. Focus the window, press
Tab, and watch `object:state-changed:focused` with:

```sh
/usr/bin/python3.10 - <<'PY'
import gi
gi.require_version("Atspi", "2.0")
from gi.repository import Atspi

def cb(event):
    print(event.type, event.source.get_name(), event.source.get_role_name(), event.detail1)

Atspi.init()
listener = Atspi.EventListener.new(cb)
listener.register("object:state-changed:focused")
Atspi.event_main()
PY
```

The recorded live results, including the Accerciser workaround and the pending
Orca per-control speech check, are in
[accessibility.md](accessibility.md#live-validation). For your own manual
screen-reader check, install `orca` and `accerciser`, run Orca with
`orca --debug`, then tab through the window and confirm it announces the control
names, roles, and the email value. In Accerciser, the application should show
the page tree with roles, names, descriptions, states, bounds, and actions. Do
not treat the unit tests alone as screen-reader evidence.
