# Native accessibility status and validation

Currently (`v0.0.x`; no `v1.0` tag yet), ownframe ships the Go accessibility
snapshot, an experimental Linux AT-SPI bridge (`internal/atspi`), a preview
Windows UI Automation provider (`internal/uia` on `amd64` and `arm64`), and an
Android TalkBack bridge (`androidaccess`, `internal/a11ybridge`,
`internal/window/a11y_android.go`, and `dev.ownframe.android.AccessibilityBridge`
in `examples/android-host`). On Linux, the window registers a live AT-SPI
application, so Accerciser and Orca can inspect it. On Windows (`amd64` and
`arm64`), the window answers UI Automation (see
[Windows UI Automation](#windows-ui-automation)); NVDA and Narrator announce its
controls (see the recorded validation below), while JAWS has not been tried. On
Android, an Activity that attaches `AccessibilityBridge` to its `EbitenView`
exposes a virtual node tree to TalkBack. VoiceOver cannot inspect an ownframe
window because the macOS VoiceOver adapter is not shipped.

What ships today is the Go accessibility snapshot plus the Linux AT-SPI,
Windows UIA, and Android TalkBack bridges. Ownframe exposes page semantics
through `host.Accessible`. Linux maps that snapshot to AT-SPI objects
(`internal/atspi`), registers with the accessibility bus, and serves the tree.
Windows maps it to UI Automation providers (`internal/uia`). Android publishes
the encoded snapshot from the window loop (`internal/window/a11y_android.go`)
through `internal/a11ybridge` and `androidaccess`, and the Java
`AccessibilityBridge` serves `AccessibilityNodeProvider` virtual views and
queues focus, activation, and text actions back to the page loop.

There is no NSAccessibility / VoiceOver bridge.

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
- Android publishes the snapshot through `androidaccess` (`internal/a11ybridge`)
  and serves a virtual tree from `dev.ownframe.android.AccessibilityBridge`
  (`examples/android-host`).
- macOS needs NSAccessibility elements associated with the native window or
  view.

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
`AutomationFocusChanged`, followed by an `EVENT_OBJECT_FOCUS` WinEvent for the
clients that follow the focus that way. Raising is skipped while no client
listens.

Two details decide whether any event reaches a client, and getting either wrong
drops every event without an error:

- `get_ProviderOptions` must report `ProviderOptions_ServerSideProvider`
  (`0x2`). The enum is `ClientSideProvider = 1`, `ServerSideProvider = 2`; a
  provider hosted in the window it serves that reports `1` is treated as a
  client-side one.
- The window must answer `WM_GETOBJECT` for more than `UiaRootObjectId`. When
  UI Automation delivers an event it sends object IDs from zero up, and
  `UiaReturnRawElementProvider` resolves them. `OBJID_CLIENT` gets the MSAA
  view of the same tree, which NVDA reads. Other negative IDs (title bar, scroll
  bars) stay with the window.

`internal/uia/live_events_windows_test.go` hosts the provider in a plain Win32
window and keeps changing focus and names. Run it with `OWNFRAME_UIA_LIVE=1`
while a UIA client listens; it isolates the provider from Ebiten.
`go test -c` for `GOOS=windows`, then run the test binary on a desktop.

Hit testing: `ElementProviderFromPoint` returns the deepest element whose
rectangle contains the screen point, and the later one in document order where
two siblings overlap; the root is never returned, so UIA falls back to the
window when nothing is there (`Provider.ElementAt`). The method takes two
`double` arguments, which the Windows x64 and arm64 conventions pass in
floating-point registers that a Go syscall callback cannot read. A small
assembly thunk per architecture (`hit_windows_amd64.s`, `hit_windows_arm64.s`)
copies their bits into the integer registers of the same argument positions and
jumps to the Go callback.

Known limits:

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
- Event delivery: with the two fixes above, a .NET UI Automation client
  received the name-changed and focus-changed events of a standalone Win32
  window hosting the provider (`live_events_windows_test.go`). Before them it
  received none, although every raise call returned `S_OK`.
- NVDA 2026.3 (debug log) on `examples/accessibility` announced the focused
  controls as the focus moved with Tab: "Display name, edit, Ada", "Private
  value, edit, protected" (the password text is not exposed), "Receive updates,
  check box, checked", and "Email, radio button, checked".
- The NVDA result was also heard by ear on a real desktop, not only read from
  NVDA's log.
- Hit testing: `AutomationElement.FromPoint` at the center of every control
  rectangle UIA reported returned that control or one of its descendants for 44
  of 44 elements on `examples/accessibility`, with no mismatches. At the center
  of the checkbox and the radio button it returns the checkbox and the radio
  button, where it used to return the window.
- DPI: on the 150 percent monitor (window DPI 144) the checkbox rectangle is
  19 by 19 pixels, the 13 by 13 box at 100 percent scaled by 1.5, and real
  screen pixels inside the rectangle UIA reports contain the blue accent of the
  checked box and of the selected radio button (229 and 195 pixels at 150
  percent, 116 and 94 at 100 percent). A rectangle of the same size shifted 150
  pixels to the right contains none, so the check can fail.
- Narrator on Windows 11 read the controls of `examples/accessibility` as focus
  moved with Tab, confirmed by ear.
- JAWS has not been tried. It reads the same UI Automation and MSAA views as
  NVDA and Narrator, so no JAWS-specific code exists, but its behavior has not
  been verified, and it needs a Freedom Scientific license (the demo mode runs
  40 minutes per Windows session). Treat JAWS as untested until someone with a
  license runs it.
- Not run: Accessibility Insights and arm64 execution. The arm64 thunk is built
  and vetted on every change (`GOOS=windows GOARCH=arm64 go vet`), and the
  `windows-arm64` CI job runs the suite on a native arm64 runner. The Go suite
  (`go test ./internal/uia`) runs the portable layer; the two regression tests
  for the details above (`host_object_windows_test.go`) run on a Windows host,
  and the COM layer has no other automated test.

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
