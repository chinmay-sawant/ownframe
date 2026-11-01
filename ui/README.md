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
`FlexDirectionColumn`, and `JustifyContentCenter`. Helpers include `Px`, `Rem`,
`Percent`, `Hex`, `Number`, and `Edges`. `Edges` accepts one to four lengths in
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
those callbacks. Updates, state hooks, component lifecycle, and partial
subtree invalidation are outside this builder. Text and attributes escape
HTML and Go template delimiters. Manual style attributes and inline JavaScript
handler attributes are rejected. The API has no compatibility guarantee yet.
