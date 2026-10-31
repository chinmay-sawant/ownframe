# Experimental OwnFrame UI builder

This package prototypes a GPUI-inspired **fluent Go API**, without making
Blinkless a component framework. OwnFrame still owns application windows,
input, and callbacks; Blinkless still parses HTML, cascades CSS, and lays out
the generated document.

## Run the three sample components

From the repository root:

    go run ./examples/ui-builder -web

Or open the native window:

    go run ./examples/ui-builder

The demo covers a styled card, a flexbox toolbar with native click handlers,
and a file-explorer list. These correspond to the three use cases in issue #59.

## Fluent API

    page, err := ui.NewPage(ownframe.Config{
        Title: "Builder experiment",
        Width: 640, Height: 480,
    }, ui.Div().BackgroundColor("#20242b").Padding("16px").Children(
        ui.Text("Hello OwnFrame"),
        ui.Button("Save").OnClick(func(ctx context.Context) error {
            log.Println("Save")
            return nil
        }),
    ))

The example assumes the caller imported context, log, OwnFrame, and the
OwnFrame ui package. NewPage gives OwnFrame ordinary HTML as its Config.HTML,
then connects callbacks through the existing data-action/Handlers.Click path.
It does not install a JavaScript engine or DOM. To render a fragment only, use
ui.Render(root) instead; click callbacks are not wired by Render.

For custom HTML attributes use Attr, and for custom styles use CSS:

    ui.Div().
        Attr("aria-label", "results").
        CSS("grid-template-columns", "1fr 2fr").
        CSS("column-gap", "12px")

## Public examples

- [`examples/ui-builder`](../examples/ui-builder/main.go) contains the original card, toolbar, and file explorer samples.
- [`examples/jmap-mail`](../examples/jmap-mail/README.md) is a three-pane JMAP-inspired mail demo with seeded mailboxes and working local selection/actions. It imports the public `ownframe` and `ownframe/ui` packages only, not private `internal` packages.

## Coverage of 555 implemented CSS properties

The public builder method CSS(property, value) (also called Style) accepts
**every one of the 555 CSS property names marked Implemented** in
[the OwnFrame compatibility matrix](../documentation/compatibility-matrix.md).
SupportedProperties lists the names, and SupportsProperty checks a name.
Convenience methods such as Display, Padding, Gap, BackgroundColor, and
BorderRadius simply call the same CSS method. There is intentionally no
separate Go method for each of the 555 properties.

The matrix is a 2026-10-20 snapshot of Blinkless: 555 Implemented, 3 Partial,
219 Unsupported, 8 intentionally ignored. This is property-name coverage,
**not** a claim that every value or every browser behavior works. A builder
call with an unsupported property name fails at Render/NewPage; supported
property values still go through Blinkless's normal CSS validation and
rendering path. No new CSS layout algorithms are in the builder.

Refresh the generated registry when the matrix changes:

    go generate ./ui

The generator lives in `scripts/gen-ui-properties`, outside this public UI package. `go generate ./ui` invokes it. The test TestAllImplementedPropertiesMatchCompatibilityMatrix checks the
generated set against the matrix, one-for-one. The generator and tests
currently enforce the requested 555-property snapshot.

## Experimental limits

- Nodes are mutable builders and not safe to edit concurrently.
- NewPage requires Config.HTML and Config.File to be empty. Existing
  template-based OwnFrame apps continue using ownframe.New unchanged.
- NewPage owns the click handler it installs. A later Page.Handle call
  replaces callbacks; avoid doing so without composing the builder handler.
- The builder serializes to HTML rather than keeping a reconciled component
  tree. Updates, lifecycle, state hooks, and partial subtree invalidation are
  outside this first experiment.
- Text/attribute output escapes HTML and Go template action delimiters.
  Manual style attributes and inline JavaScript handler attributes are
  rejected; use CSS and OnClick instead.
- The experimental API has no compatibility guarantee until reviewed.

See [issue #59](https://github.com/chinmay-sawant/ownframe/issues/59)
for the scope and follow-up design questions.
