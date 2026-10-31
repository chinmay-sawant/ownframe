# Accessibility tree

1.0 ships a Go accessibility snapshot for the operator who runs the process.
It does not ship a screen-reader product. On Linux, the window registers a live
AT-SPI application, so Accerciser and Orca can inspect it
([atspi.md](atspi.md)). NVDA, VoiceOver, and TalkBack cannot inspect an
ownframe window because the Windows, macOS, and Android adapters are not
shipped.

Linux has an experimental AT-SPI mapping, a live D-Bus client, and a fake-bus
test seam. The client connects to the accessibility bus, registers with
`org.a11y.atspi.Registry`, serves the mapped tree, and emits focus and
structure events. See [atspi.md](atspi.md) for interfaces and limits.

`Page.Access().AccessibilityTree` returns a detached semantic snapshot from the last
successful redraw. The tree follows HTML document order and includes supported
roles, names, descriptions, form state, actions, and bounds when the layout can
provide them.

```go
tree, err := page.Access().AccessibilityTree()
if err != nil {
	return err
}

for _, node := range tree.Nodes {
	log.Printf("role=%s id=%s", node.Role, node.ID)
}

return page.Access().AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
	Document: tree.Document,
	Node:     "id:save",
	Kind:     ownframe.AccessibilityActivate,
})
```

Call both methods on the page's owning loop. The window still calls the inner
screen. A native adapter must queue actions back to that loop. `Accessible` is
optional for hosts and does not add requirements to `host.Screen`. 1.0 does
not ship a screen reader.

## Try the example

Run `go run ./examples/accessibility` to open the desktop API lab. Its semantic
fixtures cover every exported role. The snapshot prints tree identity and
revision, parent and child links, names, descriptions, values, bounds, state
flags, and available actions. Password values are displayed as `[redacted]`.
The preview is hidden from the semantic tree so refreshing does not add its own
previous output to the next snapshot.

The API action buttons call Focus, SetValue, and Activate with the current
document token. The error probes exercise stale document, missing node,
unsupported action, disabled control, and invalid UTF-8 value errors. The demo
also checks that a tree is unavailable before the first successful render.
Automated example tests guard role, metadata, action, password, and error
coverage:

```sh
go test ./examples/accessibility/...
```

The demo is an API inspector. Linux also registers the same snapshot as a live
AT-SPI application, so the demo can be inspected with Accerciser or Orca.
Windows UI Automation, macOS VoiceOver, and Android TalkBack are not shipped.
See the [native accessibility status](accessibility-native.md) and
[AT-SPI live bridge](atspi.md).

## Live validation

Issue #48 tracks the end-to-end check against real AT-SPI clients. The run used
Ubuntu 22.04 with at-spi2-core 2.44 under Xvfb/X11, and read the tree through
libatspi 2.0 from Python 3.10. This matrix covers that X11 session; a Wayland
session is not part of it.

| Check | Result |
| --- | --- |
| Tree registration, roles, names, and bounds | **pass**. The application appears on the accessibility bus, and the page tree reports the mapped roles, names, and bounds. |
| Password value redacted, editable, no SetText | **pass**. A password field reads as an empty value, reports the editable state, and advertises no SetText action. |
| Focus events consumed by libatspi | **pass**. Tab focus emits `object:state-changed:focused`, and libatspi delivers the event to the listener. |
| Accerciser 3.38 | **pass**. It shows the application and the full page tree. This Accerciser version needed a local workaround for its zero-child-application population bug before it listed the children. That workaround is on the Accerciser side and is not an ownframe change. |
| Orca 42 | **partial**. Orca finds the active window and receives focus events, and Speech Dispatcher captured the startup utterance `Screen reader on.`. Per-control speech was not captured in this run; it is still pending a normal desktop run with Orca installed. |

## Node identity and updates

An element with a unique HTML `id` uses `id:<value>`. Other nodes use a
document path such as `path:1.0.2`. The document token changes after
navigation or template reload. Pass the token from the snapshot with each
action. A stale token returns `ErrAccessibilityStale`.

`Revision` is the page's successful redraw generation. Each call returns new
node and child slices, so changing a snapshot does not change page state. A
node's bounds are CSS pixels in the page content coordinate space. Ownframe
matches unique HTML IDs directly. For elements without IDs, it matches the
tag and normalized text against Blinkless's placed boxes. It sets `HasBounds`
only when the source and layout contain the same number of matching elements.

## Supported semantics and actions

The exported roles are Document, Main, Navigation, Group, Heading, Paragraph,
Text, Image, Link, Button, Textbox, Checkbox, Radio, Combobox, List, and
ListItem. HTML elements map to these roles where supported; an explicit
supported `role` attribute can also provide a role. The example contains at
least one fixture for each role.

Names can come from `aria-labelledby`, `aria-label`, associated labels, image
`alt`, element text, textbox placeholder, or title. Descriptions come from
`aria-describedby`. Nodes expose parent and child IDs, HTML element IDs, bounds,
disabled and focus state, focusability, checked state, password state, and the
actions available on that node. Text nodes and elements without a unique
matching layout box do not have bounds. Repeated anonymous elements with the
same tag and text get bounds in document order only if every matching element
has a placed box.

Form values and checked state come from live page controls. Password values
are always empty in the snapshot, and password nodes do not advertise SetValue.
They still report the editable state so a screen reader recognizes a password
field. Supported controls advertise Focus, Activate, or SetValue in `Actions`:

| Action | Supported targets |
| --- | --- |
| Focus | Focusable form controls, including buttons and disabled-by-ARIA controls |
| Activate | Buttons, selects, checkboxes, and radio buttons |
| SetValue | Editable text inputs and textareas that are not readonly |

Actions go through the page's focus, click, edit, binding, and change-handler
paths. The example tests all three actions on representative controls. Links
are represented in the tree, but link activation is not currently supported.

The exported errors describe unavailable trees, stale document tokens, missing
nodes, unsupported actions, disabled controls, and invalid UTF-8 values:
`ErrAccessibilityUnavailable`, `ErrAccessibilityStale`,
`ErrAccessibilityNode`, `ErrAccessibilityAction`,
`ErrAccessibilityDisabled`, and `ErrAccessibilityRange`.

## Limits

This API supplies semantic page data to a host. Linux serves that tree over
D-Bus and emits focus and structure events. AT-initiated actions are not wired
to the page yet: the live bridge exposes action metadata, but `DoAction` and
`GrabFocus` return false, link activation is not supported, and there is no
`EditableText`, `Document`, `Table`, `Selection`, or `Hyperlink` interface. 1.0
does not ship a screen-reader product.

The snapshot filters the `hidden` attribute, `aria-hidden="true"`, and
non-content tags. It does not read computed CSS visibility. A node hidden by a
stylesheet with `display:none` or `visibility:hidden` can remain in the tree.
The engine does not expose the computed style needed to filter those nodes.
Anonymous nodes get bounds only when their tag and text match placed boxes
without ambiguity. Screen-coordinate extents add Ebiten's monitor-relative
window position, so they can be offset on a secondary monitor or on Wayland.
Text ranges, live-region events, complete ARIA name
computation, virtualized offscreen rows, and custom widget actions are not
covered. This tree is not PDF/UA output and does not claim browser
accessibility parity.

See the [v0.0.2 accessibility and desktop input checklist](../plans/v0.0.2/accessibility-desktop-ime.md)
for native adapters and remaining platform work.
