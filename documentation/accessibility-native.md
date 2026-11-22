# Native accessibility status and validation

Currently (`v0.0.x`; no `v1.0` tag yet), ownframe does not ship a cross-platform
screen-reader adapter. On Linux, the window registers a live AT-SPI application,
so Accerciser and Orca can inspect it. On Windows (amd64 and arm64), the window
answers UI Automation (see [Windows UI Automation](#windows-ui-automation)); that
provider is a preview and has not been tried with NVDA or Narrator. VoiceOver
and TalkBack cannot inspect an ownframe window because the macOS VoiceOver and
Android TalkBack adapters are not shipped.

What ships today is the Go accessibility snapshot plus an experimental Linux
AT-SPI bridge. Ownframe exposes page semantics through `host.Accessible`. Linux
maps that snapshot to AT-SPI objects (`internal/atspi`), registers with the
accessibility bus, and serves the tree.

There is no NSAccessibility / VoiceOver bridge and no TalkBack bridge.

## v1.0 accessibility policy

During the `v0.0.x` preview cycle, ownframe ships the Go accessibility snapshot
(`page.Access().AccessibilityTree` and `page.Access().AccessibilityPerform`) and
the live Linux AT-SPI bridge in `internal/atspi`. The bridge supports read,
focus, activation, and `EditableText` (`SetTextContents`, `InsertText`,
`CopyText`, `CutText`, `DeleteText`, `PasteText`) for enabled non-password text
controls.

macOS VoiceOver (`NSAccessibility`) and Android TalkBack remain unshipped, and
the Windows UIA provider is a preview. Ownframe will not make a general-purpose
`v1.0` screen-reader claim across desktop or mobile platforms unless those
adapters have dedicated platform owners and recorded validation against real assistive
technologies (NVDA, VoiceOver, and TalkBack) matching the capability scope
tracked in [platforms.md](platforms.md) and [plans/README.md](../plans/README.md).

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

The Linux bridge serves the read side and event stream. `GrabFocus`, supported
`DoAction` entries, and `EditableText` methods dispatch to the owning window
loop and report the page's result. See the [live action check](atspi.md#validate)
for the supported operations and limits.

The platform APIs need different host code:

- Linux has an AT-SPI accessible application and object tree over the
  desktop accessibility bus.
- Windows uses UI Automation server-side providers attached to the native window
  (`internal/uia`).
- macOS needs NSAccessibility elements associated with the native window or
  view.
- Android TalkBack is not a 1.0 host feature.

## Windows UI Automation

`internal/uia` serves the same `host.Accessible` snapshot to Windows UI
Automation. It is a preview for Windows `amd64` and `arm64`; the `386` build
compiles without an adapter.

Structure:

- `snapshot.go` and `provider.go` are portable. They map the snapshot to UIA
  elements and run client operations. Linux tests drive them with a fake UIA
  client (`provider_test.go`).
- `com_windows.go`, `host_windows.go`, and `dll_windows.go` implement the COM
  providers with `syscall.NewCallback` vtables and `uiautomationcore.dll`. There
  is no cgo.
- `internal/window/a11y_uia_windows.go` wires it into the window. The window
  finds its Ebiten (GLFW) `HWND`, subclasses the window procedure to answer
  `WM_GETOBJECT`, and publishes a tree after each page generation change.

Mapping:

| Page role | UIA control type | Patterns |
| --- | --- | --- |
| button, link | Button, Hyperlink | **Invoke** (with an activate action) |
| textbox, combobox | Edit, ComboBox | **Value** (read-only without a set-value action) |
| checkbox | CheckBox | **Toggle** |
| radio | RadioButton | **SelectionItem** |
| document, group, main, navigation | Document, Group | none (main and navigation set the localized control type) |
| heading, paragraph, text | Text | none (headings set the localized control type) |
| image, list, listitem | Image, List, ListItem | none |

Each element reports name, help text (description), automation ID (element ID),
enabled, keyboard-focusable, has-focus, password, and offscreen. Bounds are the
page bounds times the display scale plus the client origin, in physical screen
pixels. The root is the window's fragment root and merges with the native
`HWND` provider, so the window keeps its title and `WindowPattern`.

Actions run on the page loop through the same queue as AT-SPI, with a 3 second
limit per call. Disabled controls answer `ElementNotEnabled`; stale or removed
nodes answer `ElementNotAvailable`. Password fields keep their name and
`IsPassword`, but expose no value and no Value pattern, and `SetValue` is
refused before it reaches the page.

Events: a changed document or tree shape raises a structure-changed
(`ChildrenInvalidated`) event on the root. Name, value, toggle, enabled, and
selected changes raise property-changed events, and a focus move raises
`AutomationFocusChanged`. Raising is skipped while no client listens.

Known limits:

- Hit testing is not implemented. `ElementProviderFromPoint` takes two `double`
  arguments, and Go syscall callbacks cannot read floating-point registers, so
  the provider answers "nothing here" and UIA falls back to the window.
- Page scroll offset and page zoom are not applied to bounds, as in the AT-SPI
  bridge. The built-in title bar inset is.
- No `Text`, `ExpandCollapse`, or `Scroll` patterns; a combobox exposes Value only.
- Windows `386` and `arm` are not supported by the provider.
- A host that installs its own window procedure after ours stays chained. On
  close, ours then keeps forwarding instead of unhooking.

Validation recorded for this provider (Windows 11 under WSL interop, a
cross-built `examples/accessibility`, probed with .NET `System.Windows.Automation`
from PowerShell):

- The raw tree matched the page: control types, names, automation IDs, enabled
  state, bounds, and the `WindowPattern` from the native window.
- Invoke on a button changed the page (its count went from 0 to 1). Toggle
  flipped a checkbox, `ValuePattern.SetValue` changed a textbox, `SelectionItem.Select`
  selected a radio button, and `SetFocus` focused a textbox.
- The password field reported `IsPassword`, no patterns, and no value.
- `SetValue` on a disabled field raised `ElementNotEnabledException`.
- The provider calls to raise property events returned `S_OK` and listeners were
  reported. The PowerShell client did not receive our property, structure, or
  focus events in the runs I made, so event delivery is **not validated**.
- Not run: NVDA, Narrator, Accessibility Insights, DPI above 100 percent checks,
  multi-monitor bounds checks, and arm64 execution. The Go suite
  (`go test ./internal/uia`) runs the portable layer only; the COM layer has no
  automated test.

Run the portable tests with `go test -p 1 ./internal/uia`, and compile the
Windows layer with `GOOS=windows go vet ./internal/uia ./internal/window`.

## Validate the page API

Run the accessibility demo to inspect the page tree and exercise its actions:

```sh
go run ./examples/accessibility
```

Run the page, mapping, and demo tests:

```sh
go test -p 1 ./internal/page ./internal/atspi ./internal/uia
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
- Page changes update the native tree, and Linux AT-SPI actions (`GrabFocus`,
  supported `DoAction` entries, and `EditableText` on enabled non-password text
  controls) update the page through the window loop. macOS VoiceOver and
  Android TalkBack adapters remain unshipped; Windows is covered below.
- Bounds stay correct after scrolling and resizing.
- Password fields expose neither their value nor typed characters through the
  accessibility API.

Passing the Go tests or inspecting the demo's tree does not prove native
accessibility support. Do not treat the fake bus or the fake UIA client as a
screen reader.
