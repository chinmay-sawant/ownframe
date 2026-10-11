# Read an accessibility tree

`Page.AccessibilityTree` returns a detached snapshot of the last successful
page redraw. The tree uses HTML structure for parent and child order. It maps
supported roles, names, descriptions, form state, and bounds for elements with
unique IDs.

```go
tree, err := page.AccessibilityTree()
if err != nil {
	return err
}

for _, node := range tree.Nodes {
	log.Printf("%s %s: %s", node.Role, node.ID, node.Name)
}

err = page.AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
	Document: tree.Document,
	Node:     "id:save",
	Kind:     ownframe.AccessibilityActivate,
})
```

The `Accessible` interface is optional for hosts. It exposes the same tree
and action method without adding requirements to `host.Screen`. Call both
methods on the page's owning loop. A native adapter must queue actions back
to that loop.

## Try the example

Run `go run ./examples/accessibility` to try desktop text composition, inspect
the semantic tree, invoke a set-value action through that tree, and scroll a
page with a visible scrollbar thumb. The demo is a tree inspector; native
screen-reader bridges remain platform work.

## Node identity and updates

An element with a unique HTML `id` uses `id:<value>`. Other nodes use a
document path such as `path:1.0.2`. The document token changes after
navigation or template reload. Pass the token from the snapshot with each
action. A stale token returns `ErrAccessibilityStale`.

`Revision` is the page's successful redraw generation. Each call returns new
node and child slices, so changing a snapshot does not change the page. A
node's bounds are CSS pixels in the page content coordinate space. Ownframe
sets `HasBounds` only when one layout box matches a unique HTML ID.

## Supported semantics and actions

The first mapping covers documents, main content, navigation, groups,
headings, paragraphs, static text, images, links, buttons, text fields,
checkboxes, radio buttons, selects, and lists. It reads names from
`aria-labelledby`, `aria-label`, associated labels, image `alt`, text,
placeholder, and title. Descriptions use `aria-describedby`.

Form values and checked state come from the live page controls. Password
values are always empty in the snapshot. Supported controls advertise
`AccessibilityFocus`, `AccessibilityActivate`, or `AccessibilitySetValue`.
The page routes those actions through its existing focus, click, edit,
binding, and change-handler paths. Disabled controls do not advertise
actions. Duplicate IDs do not receive bounds or actions.

## Limits

This API supplies semantic page data to a host. It does not expose the
Ebitengine window to AT-SPI, UI Automation, or NSAccessibility. ownframe has
no native screen-reader bridge yet, and the tree alone does not make an app
accessible.

The snapshot does not use computed CSS visibility. A node hidden only with
`display:none` or `visibility:hidden` can remain in the tree. Anonymous nodes
have no reliable layout bounds. Text ranges, live-region events, complete
ARIA name computation, virtualized offscreen rows, and custom widget actions
are not covered. This tree is not PDF/UA output and does not claim browser
accessibility parity.

See the [v0.0.2 accessibility and desktop input checklist](../plans/v0.0.2/accessibility-desktop-ime.md)
for native adapters and remaining platform work.
