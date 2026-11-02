# Experimental themed controls

Package `components` adds a small set of styled controls above the neutral
`ui` builder. Each constructor returns a fresh `*ui.Element`. You can add IDs,
classes, attributes, callbacks, children, and CSS using the existing builder.

| Constructor | Contract |
|---|---|
| `Button(label, ButtonOptions)` | **Filled, tonal, outlined, text, or danger button**, with sizes and disabled state |
| `Switch(label, checked)`, `Checkbox(label, checked)` | **Controlled selection controls** |
| `Scrollbar(ScrollbarOptions)` | **Vertical pane scrollbar** with shared range, paging, wheel, keyboard, and drag state |
| `Progress(label, percent)` | **Bounded progress indicator** |
| `Notice(title, body)` | **Inline status message** |
| `Card(children...)` | **Surface container** with border, padding, and text color |
| `TextField(id, label, value)` | **Labeled text input**; value initializes the control |
| `Badge(label)` | **Non-interactive label** |
| `Screen(children...)`, `Panel(children...)` | **Application canvas and unpadded surface** |
| `SearchField(id, label, value)` | **Accessible search control** |
| `Choice(label, caption, selected)` | **Selectable option with a description** |
| `Dialog(DialogOptions, children...)` | **Controlled modal surface**, with an action on its backdrop |

There is no automatic state store, React lifecycle, or data-table controller.
Table markup remains available through `ui.Tag`. Sorting, selection, and row
windowing belong to the application until a shared table contract is defined.
Glass appearance is not a preset in this package.

## Apply a custom theme

This example assumes imports for `ownframe`, `ui`, `ui/components`, and `ui/theme`:

    stylesheet := theme.Dark + `:root {
        --of-primary: #7655c5;
        --of-on-primary: #ffffff;
        --of-surface: #222035;
        --of-radius: 12px;
    }`
    view, err := ui.Mount(ownframe.Config{
        Title: "Custom theme", Width: 640, Height: 480, Theme: stylesheet,
    }, components.Card(
        components.TextField("name", "Name", ""),
        components.Button("Save", components.ButtonOptions{Variant: components.ButtonFilled}),
    ))
    if err != nil {
        return err
    }
    // Run or Serve view.Page.

`theme.Light` and `theme.Dark` are stylesheet constants. They keep the same
geometry. Append overrides to the preset within the same stylesheet, or supply
a complete custom stylesheet. `Page.SetTheme` replaces the whole extra sheet;
it does not append to the previous one. Apply it with `Page.Redraw(ctx)`.

| Variable | Meaning |
|---|---|
| `--of-background` | **Application background** |
| `--of-surface` | **Cards and panels** |
| `--of-control` | **Secondary controls** |
| `--of-text`, `--of-muted` | **Regular and secondary text** |
| `--of-primary`, `--of-on-primary` | **Primary control fill and text** |
| `--of-border` | **Container and control borders** |
| `--of-scrollbar` | **Scrollbar thumb** |
| `--of-selected` | **Selected item surface** |
| `--of-radius`, `--of-spacing` | **Corners and card padding** |
| `--of-button-padding-x`, `--of-button-padding-y` | **Horizontal and vertical button padding**, default `12px` and `9px` |
| `--of-backdrop` | **Modal overlay**, default translucent black |

Style reusable controls with variables rather than fixed inline colors.
Inline literals still take precedence over ordinary theme rules. Components
have variable fallbacks and can render without a preset. CSS support and
cascade limits remain those of the [theme API](../../documentation/theming.md).

## Control a dialog

The host owns visibility and close actions. Use `DialogFocus.Activate(page,
ids...)` before the update that opens the dialog. List all focusable dialog
controls, with the first ID receiving initial focus. Tab and Shift+Tab stay
inside that set in document order. `DialogFocus.Deactivate(page)` before the
closing update restores the saved control if it still exists.

Handle Escape in the host's `KeyDown` callback, change the open state, and
update the view. The backdrop dispatches `DialogOptions.CloseAction`; panel
children dispatch their own actions. A panel ID prevents an interior click
from reaching the backdrop action. `Dialog` sets `role="dialog"`,
`aria-modal="true"`, and an accessible label. It uses absolute positioning, so
mount it within the intended containing block. Nested dialogs are not
supported. The gallery and mail example exercise this complete sequence.

`Form.SetFocusScope()` clears the scope. It restricts keyboard traversal and
explicit `Form.Focus` calls; it does not disable pointer actions behind a modal.
The overlay supplies that pointer barrier. Keep all focusable dialog IDs in
scope and deactivate the scope when removing the dialog.


## Override an application or a component

The precedence is OwnFrame defaults, then application properties, then the
nearest component's properties. A component override is scoped to that
component and its descendants. Sibling components continue to inherit the
application theme. Switching light and dark keeps local overrides intact.

Define application properties with `theme.Overrides` and append their
`Stylesheet()` output to `theme.Light` or `theme.Dark`. Set the resulting string
as `Config.Theme`. `theme.Surface()`, `theme.Text()`, and the other typed tokens
resolve against the current scope. They do not copy colors into a component.

For a local override:

    save := components.Button("Save", components.ButtonOptions{
        Variant: components.ButtonFilled,
        Overrides: theme.Overrides{
            "--of-primary": "#21694e",
            "--of-radius": "18px",
        },
    })

For any other tree, apply the same property set with
`properties.Apply(components.Card(children...))`. An empty set inherits all
properties. `DialogOptions.Overrides` scopes properties to the dialog panel.
Ordinary builder methods remain available for layout and explicit style changes.

`theme.LoadOverrides(files, path)` reads a JSON object from an `fs.FS`:

    {
      "--of-primary": "#21694e",
      "--of-radius": "18px",
      "--mail-row-height": "104px"
    }

Pass `os.DirFS("themes")` for files on disk, or an `embed.FS` for compiled assets.
The returned property set can configure either the application or a component.
This uses the standard filesystem interface; theme properties themselves are
data and require no custom provider interface. Unknown custom property names
are allowed when they follow the CSS `--name` form. Invalid declaration
boundaries are rejected before a file is accepted.

## Supply an icon

`ButtonOptions.Icon` accepts a builder node. Use `ui.Text` for a glyph or an
`img` element backed by `Page.SetImage` for application-owned image bytes.
Decorative icon descendants should have no IDs or actions: OwnFrame routes a
click through the nearest ID-bearing ancestor, which is the action button.
The builder supplies an ID for a `data-action` control that lacks one.
The mail example keeps its settings icon in `icons.go`, outside its layout.


## Button variants and sizes

`ButtonOptions.Variant` selects `ButtonTonal`, `ButtonFilled`, `ButtonOutlined`,
`ButtonText`, or `ButtonDanger`. The zero value is tonal. `Size` selects
`ButtonSmall`, `ButtonMedium`, or `ButtonLarge`; medium is the zero value.
The earlier experimental `Primary` boolean has been replaced by
`Variant: components.ButtonFilled`.

All variants consume semantic theme properties. Destructive actions read
`--of-danger` and `--of-on-danger`. Sizes choose padding fallbacks and font
sizes; application or component padding overrides still take precedence.

## Controlled state and feedback

`Switch` and `Checkbox` render HTML buttons with the appropriate role and
`aria-checked` state. Give them an ID and an `OnClick` callback or `data-action`.
The host updates its Go state and rebuilds the view after activation. These
are controlled components, rather than form inputs that store a checked value
on `Page.Form`. Keyboard activation follows the normal button input path.

`Progress` clamps its percentage to 0 through 100 and exposes a labeled
`progressbar` with an accessible value. `Notice` supplies a themed heading and
body for inline feedback. Both can be composed within custom controls.

The [component gallery](../../examples/builder/ui-builder/README.md) exercises these
APIs, light/dark/Studio palettes, and a separately packaged custom project card.

## Scroll a pane

`ScrollState` owns a vertical scroll position. Its zero value is empty. Keep
one state per pane on the page loop. `SetRange(content, viewport)` uses the
same logical unit for both sizes, such as pixels for a document or rows for
a windowed list. `Metrics()` returns a snapshot with the clamped offset and
`MaxOffset()`. A smaller range clamps the position and cancels an active drag.

`Scrollbar` draws the track and thumb inside a positioned viewport. It returns
nil when content fits or no measurements exist. `TrackPixels` is the actual
viewport height in CSS pixels, even when the state uses rows. IDs must be
unique across the page. The thumb defaults to `ID + "-thumb"`.

```go
var scroll components.ScrollState // retain this in the host app
scroll.SetRange(contentHeight, viewportHeight)
bar := components.Scrollbar(components.ScrollbarOptions{
    ID: "reader-scroll", ControlsID: "reader", Label: "Message",
    State: &scroll, TrackPixels: viewportHeight,
    OnChange: func(ctx context.Context) error {
        page.Form().RequestFocus("reader-scroll-thumb")
        return redraw(ctx)
    },
})
```

The host measures content and viewport after layout and translates or windows
its content using `scroll.Offset()`. Put the bar beside that content inside a
`PositionRelative` viewport with `OverflowHidden`. Reserve a right gutter so
content cannot overlap the thumb. Measurement must use intrinsic content
height, independent of its translated top.

Track clicks page by one viewport and call `OnChange` after a change. The host
routes `Handlers.Wheel` to `scroll.Wheel(dy, unitsPerNotch)` when the pointer is
over that pane. Positive wheel deltas move toward the start. Fractional input
accumulates. Redraw when these methods return true.

Claim a native `DragStart` on the thumb with `scroll.BeginDrag(trackPixels)`.
Forward `DragMove`'s vertical pixel delta to `scroll.DragBy(dy, trackPixels)`
and finish with `scroll.EndDrag()`. While the thumb has focus, forward
`KeyDown` to `scroll.Key(key, step)`. It handles arrows, Page Up/Down, Home,
and End; `step` uses the state's logical units. Block these handlers while a
modal covers the pane. The PNG web preview supports wheel, track clicks, and
keyboard input; thumb dragging uses the native host gesture handlers.

The thumb inherits `--of-scrollbar`. Override `--of-scrollbar-width`,
`--of-scrollbar-inset`, or `--of-scrollbar-radius` with CSS lengths to change
its geometry. These default to `8px`, `2px`, and `4px`. Use
`ScrollbarOptions.Overrides` for one instance or global theme properties for
all instances. OwnFrame's Light and Dark themes also provide a focus outline.

See [JMAP mail](../../examples/builder/jmap-mail/scroll.go) for independent row and
pixel ranges and [the gallery](../../examples/builder/ui-builder/scroll.go) for a
measured scrolling catalog. The host owns content layout and event routing;
the shared component owns thumb geometry and scroll state.
