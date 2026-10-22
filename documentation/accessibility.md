# Accessibility tree

1.0 ships a Go accessibility snapshot for the operator who runs the process.
It does not ship a screen-reader product. `NewLiveBus` reports `Live: false`.
Orca cannot inspect an ownframe window. NVDA, VoiceOver, and TalkBack cannot
either.

Linux has an experimental AT-SPI mapping and a fake-bus test seam
([atspi.md](atspi.md)). There is no live D-Bus AT-SPI client. UI Automation
is not shipped.

`Page.AccessibilityTree` returns a detached semantic snapshot from the last
successful redraw. The tree follows HTML document order and includes supported
roles, names, descriptions, form state, actions, and bounds when the layout can
provide them.

```go
var accessible ownframe.Accessible = page
tree, err := accessible.AccessibilityTree()
if err != nil {
	return err
}

for _, node := range tree.Nodes {
	log.Printf("%s %s: %s", node.Role, node.ID, node.Name)
}

return accessible.AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
	Document: tree.Document,
	Node:     "id:save",
	Kind:     ownframe.AccessibilityActivate,
})
```

Call both methods on the page's owning loop. A native adapter must queue
actions back to that loop. `Accessible` is optional for hosts and does not add
requirements to `host.Screen`.

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

The demo is an API inspector. Linux maps the snapshot to AT-SPI objects in
process; it does not register them on the accessibility bus. Windows UI
Automation, macOS VoiceOver, and Android TalkBack are not shipped. See the
[native accessibility status](accessibility-native.md) and
[AT-SPI first slice](atspi.md).

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
are always empty in the snapshot. Supported controls advertise Focus,
Activate, or SetValue in `Actions`:

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

This API supplies semantic page data to a host. Linux maps that tree to
AT-SPI objects in process. `Live` is false, so Orca cannot inspect the
window. 1.0 does not ship a screen-reader product.

The snapshot filters the `hidden` attribute, `aria-hidden="true"`, and
non-content tags. It does not read computed CSS visibility. A node hidden by a
stylesheet with `display:none` or `visibility:hidden` can remain in the tree.
The engine does not expose the computed style needed to filter those nodes.
Anonymous nodes get bounds only when their tag and text match placed boxes
without ambiguity. Text ranges, live-region events, complete ARIA name
computation, virtualized offscreen rows, and custom widget actions are not
covered. This tree is not PDF/UA output and does not claim browser
accessibility parity.

See the [v0.0.2 accessibility and desktop input checklist](../plans/v0.0.2/accessibility-desktop-ime.md)
for native adapters and remaining platform work.
