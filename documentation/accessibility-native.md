# Native accessibility status and validation

Ownframe exposes page semantics through `host.Accessible`. It does not
register those semantics with an operating system. Screen readers cannot
inspect an ownframe window through this API yet.

## Current integration boundary

`AccessibilityTree` returns a detached snapshot. Each node has a stable ID for
the current document, a role, a name, state, bounds when available, and its
supported actions. `AccessibilityPerform` accepts an action with the document
token. Call both methods on the page's owning loop. A host must queue actions
back to that loop.

The API does not provide a native window handle or native accessibility
objects. `internal/window` draws the page through Ebitengine and does not expose
a provider that AT-SPI, UI Automation, or NSAccessibility can query. A native
adapter needs a platform window integration in addition to the page tree.

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
  desktop accessibility bus.
- Windows needs UI Automation providers attached to the native window tree.
- macOS needs NSAccessibility elements associated with the native window or
  view.

The existing page API can supply semantic data to each adapter. It does not
implement any of these platform integrations.

## Validate the page API

Run the accessibility demo to inspect the page tree and exercise its actions:

```sh
go run ./examples/accessibility
```

Run the page and demo tests:

```sh
go test ./internal/page
cd examples
go test ./accessibility/...
```

These checks validate the Go tree and action API. They do not validate a
screen reader or a native accessibility inspector.

## Validate a native adapter

After an adapter exists, run its application on the target operating system.
Use that system's accessibility inspector to confirm that the app window and
page nodes appear. Check the role, accessible name, description, value, focus,
checked state, disabled state, and available actions on representative nodes.

Then use the platform screen reader to navigate the page, move focus, activate
a control, and edit a text field. Confirm that page actions update the native
tree and that native actions update the page. Check bounds after scrolling and
resizing. Confirm that password fields expose neither their value nor typed
characters through the accessibility API.

Run these checks separately on Linux with AT-SPI, Windows with UI Automation,
and macOS with NSAccessibility. Passing the Go tests or inspecting the demo's
tree does not prove native accessibility support.
