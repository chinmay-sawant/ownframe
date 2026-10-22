# Native accessibility status and validation

1.0 ships the Go snapshot plus an experimental Linux AT-SPI mapping. Ownframe
exposes page semantics through `host.Accessible`. Linux maps that snapshot
to AT-SPI objects in process (`internal/atspi`). It does not register those
objects on the accessibility bus, so screen readers cannot inspect an
ownframe window yet.

There is no live AT-SPI D-Bus client, no UI Automation provider, no
NSAccessibility / VoiceOver bridge, and no TalkBack bridge.

## Current integration boundary

`AccessibilityTree` returns a detached snapshot. Each node has a stable ID for
the current document, a role, a name, state, bounds when available, and its
supported actions. `AccessibilityPerform` accepts an action with the document
token. Call both methods on the page's owning loop. A host must queue actions
back to that loop.

`internal/atspi` maps that snapshot to AT-SPI roles, states, and actions and
offers a bus seam. Tests use `FakeBus`. `NewLiveBus` reports `Live: false`.
Linux `internal/window` constructs the publisher and skips the tree walk
until a live bus exists. See [atspi.md](atspi.md).

The API does not provide a native window handle. `internal/window` draws the
page through Ebitengine. A live adapter still needs a platform window
integration in addition to this mapping.

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
`AccessibilityPerform` to the platform layer.

The platform APIs need different host code:

- Linux needs an AT-SPI accessible application and object tree over the
  desktop accessibility bus. Mapping is present; the bus client is not.
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

These checks validate the Go tree, the AT-SPI mapping, and the fake bus.
They do not validate a screen reader.

## Validate a native adapter

After a live bus client exists, run the application on Linux. Use Accerciser
to confirm that the app window and page nodes appear. Check the role,
accessible name, description, value, focus, checked state, disabled state,
and available actions on representative nodes.

Then use Orca to navigate the page, move focus, activate a control, and edit
a text field. Confirm that page actions update the native tree and that
native actions update the page. Check bounds after scrolling and resizing.
Confirm that password fields expose neither their value nor typed characters
through the accessibility API.

Passing the Go tests or inspecting the demo's tree does not prove native
accessibility support. Do not treat the fake bus as a screen reader.
