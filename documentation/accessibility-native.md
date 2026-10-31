# Native accessibility status and validation

1.0 does not ship a screen-reader product. On Linux, the window registers a
live AT-SPI application, so Accerciser and Orca can inspect it. NVDA, VoiceOver,
and TalkBack cannot inspect an ownframe window because those adapters are not
shipped.

1.0 ships the Go snapshot plus an experimental Linux AT-SPI bridge. Ownframe
exposes page semantics through `host.Accessible`. Linux maps that snapshot to
AT-SPI objects (`internal/atspi`), registers with the accessibility bus, and
serves the tree.

There is no UI Automation provider, no NSAccessibility / VoiceOver bridge, and
no TalkBack bridge.

## Current integration boundary

`page.Access().AccessibilityTree` returns a detached snapshot. Each node has a stable ID for
the current document, a role, a name, state, bounds when available, and its
supported actions. `page.Access().AccessibilityPerform` accepts an action with the document
token. Call both on the page's owning loop. The window still calls the inner
screen through `host.Accessible`. A host must queue actions back to that loop.
This is not a screen reader.

`internal/atspi` maps that snapshot to AT-SPI roles, states, and actions.
Tests use `FakeBus`. `NewLiveBus` connects to the accessibility bus when the
window publishes its first tree. Linux `internal/window` constructs the
publisher, passes the window origin and size for screen bounds, and publishes
on redraw. See [atspi.md](atspi.md).

The API does not provide a native window handle. `internal/window` draws the
page through Ebitengine, so the Linux bridge uses the Ebiten window position
and size for root extents instead of a toolkit window object.

## Adapter requirements

A platform adapter must map the detached tree to native roles and properties.
It must preserve parent and child order, names, descriptions, values, focus,
checked and disabled state, and the actions listed on each node. It must not
expose password values. It must replace stale native nodes when the document
token changes and update properties when the revision changes.

An adapter must convert page-content CSS-pixel bounds to the coordinate system
used by the platform API. That conversion must account for page scrolling,
window position, content origin, and display scaling. It must dispatch native
actions on the page's owning loop and report errors from
`page.Access().AccessibilityPerform` to the platform layer. The window still
calls the inner screen.

The Linux bridge does the read side and event stream. It does not dispatch
native actions yet, so `DoAction` and `GrabFocus` return false, and no
`EditableText` interface is served.

The platform APIs need different host code:

- Linux has an AT-SPI accessible application and object tree over the
  desktop accessibility bus.
- Windows needs UI Automation providers attached to the native window tree.
- macOS needs NSAccessibility elements associated with the native window or
  view.
- Android TalkBack is not a 1.0 host feature.

## Validate the page API

Run the accessibility demo to inspect the page tree and exercise its actions:

```sh
go run ./examples/accessibility
```

Run the page, mapping, and demo tests:

```sh
go test -p 1 ./internal/page ./internal/atspi
cd examples
go test ./accessibility/...
```

These checks validate the Go tree, the AT-SPI mapping, and the fake bus. The
live bus check is in [atspi.md](atspi.md), and the recorded live results are in
[accessibility.md](accessibility.md#live-validation).

## Validate a native adapter

For Linux, run an example, then inspect it with the Atspi 2.0 client library,
Accerciser, or Orca. The automated check in [atspi.md](atspi.md) walks the
application tree and listens for focus events.

The manual checklist:

- Accerciser shows the app window and page nodes, with role, accessible name,
  description, value, focus, checked state, disabled state, bounds, and
  available actions on representative nodes.
- Orca announces control names and roles as Tab moves focus, reads the email
  value, and does not read the password value.
- Page changes update the native tree, and native actions update the page. The
  second half is not implemented yet.
- Bounds stay correct after scrolling and resizing.
- Password fields expose neither their value nor typed characters through the
  accessibility API.

Passing the Go tests or inspecting the demo's tree does not prove native
accessibility support. Do not treat the fake bus as a screen reader.
