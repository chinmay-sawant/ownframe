# Experimental themed controls

Package `components` adds a small set of styled controls above the neutral
`ui` builder. Each constructor returns a fresh `*ui.Element`. You can add IDs,
classes, attributes, callbacks, children, and CSS using the existing builder.

| Constructor | Contract |
|---|---|
| `Button(label, ButtonOptions)` | **Primary or secondary button**, with optional disabled state |
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
        components.Button("Save", components.ButtonOptions{Primary: true}),
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
        Primary: true,
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
