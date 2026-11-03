# Experimental OwnFrame UI builder

Package `ui` builds HTML through fluent Go calls. OwnFrame owns windows,
input, and callbacks. Blinkless parses the generated HTML, cascades CSS, and
lays out the page. The builder has no JavaScript runtime or reconciled
component tree.

## Construct a page

    page, err := ui.NewPage(ownframe.Config{
        Title: "Builder experiment",
        Width: 640, Height: 480,
    }, ui.Div().
        Display(ui.DisplayFlex).
        FlexDirection(ui.FlexDirectionColumn).
        JustifyContent(ui.JustifyContentCenter).
        Padding(ui.Edges(ui.Px(12), ui.Px(16))).
        Gap(ui.Px(8)).
        Width(ui.Percent(100)).
        CustomProperty("--accent", "#4677ed").
        BackgroundColor(ui.Hex("#20242b")).
        Color(ui.Var("--accent", "#ffffff")).
        Children(
            ui.Text("Hello OwnFrame"),
            ui.Button("Save").OnClick(func(ctx context.Context) error {
                log.Println("Save")
                return nil
            }),
        ))

This example assumes imports for context, log, OwnFrame, and OwnFrame's `ui`
package. Handle `err` before using `page`. `NewPage` owns `Config.HTML` and
connects `OnClick` through the existing action dispatcher. `Config.HTML` and
`Config.File` must be empty. `ui.Render(root)` returns a fragment without
connecting callbacks.

## Compile-time property and value checks

Every implemented property in the
[compatibility matrix](../documentation/compatibility-matrix.md) has a generated
method with its CSS name converted to Go spelling. For example,
`justify-content` becomes `JustifyContent`. The generated registry follows the
compatibility matrix.
An unknown method fails compilation. Go completion can list the methods and
show their property-specific argument types.

Each property has its own value interface. Strings cannot be passed directly:

    ui.Div().Display(ui.DisplayFlex)                 // Compiles.
    ui.Div().Display("flex")                         // Does not compile.
    ui.Div().Display(ui.PositionAbsolute)            // Does not compile.
    ui.Div().Color(ui.Px(10))                        // Does not compile.
    ui.Div().InventedProperty(ui.Px(10))             // Does not compile.

Common keyword values are constants such as `DisplayFlex`,
`FlexDirectionColumn`, and `JustifyContentCenter`. Helpers include `Px`, `Vh`, `Rem`,
`Percent`, `Hex`, `Number`, `Edges`, `SolidBorder`, and `TranslateY`. `Edges` accepts one to four lengths in
CSS shorthand order. Percentages are separate from lengths, so
`BorderWidth(Percent(10))` fails compilation. `Inherit`, `Initial`, `Unset`,
`Revert`, and `RevertLayer` can be supplied to every property; their rendering
semantics follow the engine.

The typed API checks property names and the documented value families, not
complete CSS grammars. Numeric ranges, context-dependent values, and engine
support for individual values remain runtime concerns. Invalid zero values,
nil values, malformed hex colors, and non-finite numbers produce errors from
`Render` or `NewPage`. An explicitly constructed invalid keyword number also
produces an error there. Typed convenience values do not cover every CSS value.

## Complex values and runtime input

Use `RawValue` when a known property needs CSS syntax without a typed helper:

    ui.Div().
        Width(ui.RawValue("calc(100% - 24px)")).
        Border(ui.RawValue("1px solid #42536b")).
        GridTemplateColumns(ui.RawValue("1fr 2fr"))

The property name is still checked by the compiler. `RawValue` opts out of
property-specific value typing. It checks declaration boundaries; it does
not validate the property's CSS grammar.

For a property name supplied at runtime, use `CSSRaw(property, value)`.
Unknown names produce an error from `Render` or `NewPage`. `SupportedProperties`
lists the accepted names, and `SupportsProperty` checks a name. Runtime strings
cannot receive compile-time support checks. `CSSRaw` does not add rendering
support for properties absent from the engine.

`RawValue`, `CSSRaw`, and `CustomProperty` accept one declaration value. They
reject empty values, control characters, unbalanced quotes or delimiters,
comments, rule delimiters, and extra declarations. Delimiters inside quoted
strings and functions, including data URLs, are allowed. CSS semantics still
go through Blinkless; an invalid semantic value can be ignored by the engine.

## Custom properties

`CustomProperty("--accent", "#4677ed")` declares a CSS variable.
`Color(Var("--accent"))` references it, and `Var("--accent", "#ffffff")` adds a
fallback. Custom property names are case-sensitive and use the ASCII subset
`--[A-Za-z0-9_-]+`. Invalid names or malformed declaration values produce an
error from `Render` or `NewPage`. Variable inheritance, fallback, and cycle
resolution follow Blinkless. A custom variable cannot implement a new layout
property. Use `CustomProperty`, rather than `CSSRaw`, to declare one.

## Extend the generated API

From the repository root, run:

    go generate ./ui

The generator reads implemented property names from the compatibility matrix.
Adding an implemented name generates its method and value interface without
changing a hardcoded property count. New methods initially accept CSS-wide
keywords, `Var`, and `RawValue`. Add typed value families and keyword choices in
[`scripts/gen-ui-properties/style.go`](../scripts/gen-ui-properties/style.go)
as the engine supports them. A metadata entry for an absent property fails
generation, as do duplicate names or Go method-name collisions.

The registry and generated-method tests compare coverage against the matrix.
Compiler tests check valid and invalid external callers. A page integration
test checks custom variable inheritance and fallback through engine layout.
After regeneration, publish a new library version so callers can compile
against the added methods. Methods cannot be added at runtime.

## Migration and limits

This API uses generated named methods with typed values. Use `RawValue` with a
named method for complex values, and use `CSSRaw` only when runtime input
chooses the property name. The examples use both forms and demonstrate CSS
custom properties.

To rerun the migration for the two examples, run this command from the
repository root:

    env GOWORK=off go -C scripts run ./cmd/migrate-ui-builder

The command migrates literal property names and the earlier convenience
methods. It reports runtime property names for manual migration to `CSSRaw`.

Nodes are mutable and must not be edited concurrently. `NewPage` installs its
click callbacks; replacing the page click handler later requires composing
those callbacks. State hooks, component lifecycle, and partial subtree
invalidation are outside this builder. Text and attributes escape
HTML and Go template delimiters. Manual style attributes and inline JavaScript
handler attributes are rejected. The API has no compatibility guarantee yet.


## Update a builder view

Use `Mount` when application state rebuilds the tree:

    view, err := ui.Mount(config, buildScreen())
    if err != nil {
        return err
    }
    view.Handle(ownframe.Handlers{Click: handleAppAction})
    // After changing Go state on the page loop:
    if err := view.Update(ctx, buildScreen()); err != nil {
        return err
    }

`view.Page` is the ordinary `*ownframe.Page` used by `Run` and `Serve`.
`View.Handle` keeps builder `OnClick` callbacks and forwards other actions to
its application click handler. `Page.Handle` and a non-nil `Page.AddHandlers`
click callback replace that dispatch, so use `View.Handle` for mounted views.
`Render` still emits markup only and does not connect callbacks.

`Update` replaces the current source and callback registry without pushing
navigation history or deleting forward history. It performs a full document
render through Blinkless, subject to normal redraw batching. Tree validation
and canceled-context errors leave the previous source and callbacks in place.
Inside a redraw batch, rendering errors can arrive later from the batch flush;
the accepted source and callbacks remain the current state for the retry.
Updates must run on the page loop and must not be called from an after-redraw
callback. They do not reconcile subtrees or promise a rendering speedup.

Controls with stable IDs and types retain live values, caret, selection, and
focus. Removed controls lose their state. Changing an input's initial `value`
in a rebuilt tree does not overwrite a live edit. Use
`view.Page.Form().SetFormValue(id, value)` to reset it explicitly before the
update. `Form.RequestFocus(id)` applies focus during the next form sync, so a
newly opened dialog can focus a control that was absent in the previous view.

## Styled controls and custom themes

The optional [`chadcn`](../components/chadcn/README.md) package supplies buttons,
cards, labeled text fields, badges, and controlled dialogs. The raw `ui`
builder keeps its neutral HTML/CSS API. The gallery and JMAP mail examples
share these controls and semantic theme variables.

Pass `theme.Light`, `theme.Dark` from `ui/theme`, or your own stylesheet
through `Config.Theme`. Append a `:root` rule to override tokens. Later,
`Page.SetTheme` stores another stylesheet and `Page.Redraw(ctx)` applies it.
Component defaults use variable fallbacks rather than inline variable
definitions, so root theme variables can reach every instance.


Theme tokens use `ColorVar` and `LengthVar` to preserve property-specific types.
For example, `ColorVar("--of-primary", Hex("#4677ed"))` can supply a color but
cannot supply a width. These helpers validate the fallback and variable name;
the stylesheet still supplies the runtime value. `Var` remains available for
CSS values whose type is only known at runtime.

The mail view uses shared controls and typed values without `RawValue`.
Built-in palettes live in `ui/theme`; application properties live in the mail
example's `theme.go`, and its custom icon is in `icons.go`. See the component
reference for scoped overrides and JSON property files through `fs.FS`.
