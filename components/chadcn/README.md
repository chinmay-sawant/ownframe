# ChadCN controls

ChadCN is the first reusable component set in `components/`. It is built on
the `ui` builder, keeps state in the host application, and ships one stylesheet
plus light and dark palettes.

## Quick start

```go
sheet := chadcn.Dark + chadcn.Stylesheet
view, err := ui.Mount(ownframe.Config{
    Title: "ChadCN", Width: 960, Height: 720, Theme: sheet,
}, chadcn.Screen(
    chadcn.Card(
        chadcn.CardHeader(
            chadcn.CardTitle("Deployments"),
            chadcn.CardDescription("Three environments"),
        ),
        chadcn.CardContent(
            chadcn.Button("Deploy", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}),
        ),
    ),
))
```

`chadcn.Light` and `chadcn.Dark` set the `--of-*` tokens. `chadcn.Stylesheet`
carries every component's paint and state rule. An application that supplies
its own tokens can append `chadcn.Stylesheet` to its existing sheet.

## Open states are parameters

A component that can open takes `Open bool` plus its data. The same constructor
renders the closed trigger and the open surface. There is no `DialogOpen`
constructor.

```go
chadcn.Dialog(chadcn.DialogOptions{
    Open:  a.editOpen,
    Title: "Edit profile",
    Trigger: chadcn.Button("Edit profile", chadcn.ButtonOptions{}),
    OnOpenChange: func(ctx context.Context, open bool) error {
        a.editOpen = open
        return a.refresh(ctx)
    },
}, children...)
```

The 15 open-state captures in the reference set map to `Open: true` on these
constructors: `AlertDialog`, `Combobox`, `ContextMenu`, `DatePicker`, `Dialog`,
`Drawer`, `DropdownMenu`, `HoverCard`, `Menubar`, `NavigationMenu`, `Popover`,
`Select`, `Sheet`, `Toast`, and `Tooltip`.

## Components

### Primitives and display

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Aspect Ratio** | `AspectRatio` | static | Percentage padding box, no `aspect-ratio` dependency. |
| **Avatar** | `Avatar` | static | Image source, initials, fallback, status dot. |
| **Badge** | `Badge` | static | Default, secondary, outline, and danger tones. |
| **Card** | `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter` | static | Surface container with composable parts. |
| **Direction** | `Direction` | static | LTR and RTL wrapper. |
| **Empty** | `Empty` | static | Icon, title, description, actions, and footer slots. |
| **Item** | `Item` | static | Media, content, and actions slots. |
| **Kbd** | `Kbd` | static | Key label. |
| **Label** | `Label` | static | Label with optional `for` target. |
| **Progress** | `Progress` | static | Clamped percentage, progressbar semantics. |
| **Separator** | `Separator` | static | Horizontal or vertical rule. |
| **Skeleton** | `Skeleton` | static | Shimmering placeholder block. |
| **Spinner** | `Spinner` | static | Animated ring. |
| **Table** | `Table`, `TableHeader`, `TableBody`, `TableFooter`, `TableRow`, `TableHead`, `TableCell`, `TableCaption` | static | Themed table markup. |
| **Typography** | `H1`, `H2`, `H3`, `H4`, `P`, `Lead`, `Muted`, `InlineCode`, `Blockquote`, `List` | static | Type scale helpers. |

### Actions and forms

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Button** | `Button` | controlled | Tonal, filled, outlined, text, danger; sizes, icon, loading, disabled. |
| **Button Group** | `ButtonGroup` | static | Joined horizontal or vertical buttons. |
| **Toggle** | `Toggle` | controlled | `aria-pressed`; host flips it. |
| **Toggle Group** | `ToggleGroup` | controlled | Single or multiple selection; host owns `Selected`. |
| **Field** | `Field` | static | Label, description, error, required, control slot. |
| **Input** | `Input`, `InputWithOptions` | form | Text value stays on `Page.Form`. |
| **Textarea** | `Textarea`, `TextareaWithOptions` | form | Rows and min-height. |
| **Native Select** | `NativeSelect` | form | Native `<select>` with `<option>` children. |
| **Checkbox** | `Checkbox`, `CheckboxWithOptions` | controlled | `role=checkbox`, `aria-checked`. |
| **Radio Group** | `RadioGroup` | controlled | `role=radiogroup`; host owns the value. |
| **Switch** | `Switch`, `SwitchWithOptions` | controlled | `role=switch`, `aria-checked`. |
| **Slider** | `Slider` | controlled | Track, fill, thumb, ARIA range values. |
| **Input Group** | `InputGroup` | form | Prefix, suffix, and addon slots. |
| **Input OTP** | `InputOTP` | form | One input per digit, derived IDs. |

### Feedback and disclosure

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Alert** | `Alert` | static | Info, success, warning, danger; icon and action slots. |
| **Toast** | `Toast` | open | Absolute bottom-right surface; returns nil while closed. |
| **Accordion** | `Accordion` | open set | Multiple open IDs, `aria-expanded`, labelled region. |
| **Collapsible** | `Collapsible` | open | Trigger plus content. |
| **Tabs** | `Tabs` | active | Tablist, tab, and panel roles; one active panel. |
| **Breadcrumb** | `Breadcrumb` | static | Separators and `aria-current="page"`. |
| **Pagination** | `Pagination` | active | Bounds, current page, ellipsis window. |

### Overlays and menus

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Dialog** | `Dialog` | open | Backdrop, panel, header, footer, close control. |
| **Alert Dialog** | `AlertDialog` | open | Confirmation surface with cancel and confirm. |
| **Sheet** | `Sheet` | open | Side panel, left/right/top/bottom. |
| **Drawer** | `Drawer` | open | Bottom panel with grab handle. |
| **Popover** | `Popover` | open | Trigger plus anchored panel. |
| **Tooltip** | `Tooltip` | open | `role=tooltip`; host drives hover. |
| **Hover Card** | `HoverCard` | open | Rich preview card; host drives hover. |
| **Dropdown Menu** | `DropdownMenu` | open | Trigger plus menu, groups, shortcuts, submenus. |
| **Context Menu** | `ContextMenu` | open | Positioned at host X/Y coordinates. |
| **Menubar** | `Menubar` | open | One open top-level menu. |
| **Navigation Menu** | `NavigationMenu` | open | Trigger row plus one open panel. |
| **Command** | `Command` | query | Search field, groups, active row, empty state. |

### Selection

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Select** | `Select` | open | Trigger, listbox, selected marker. |
| **Combobox** | `Combobox` | open + query | Filtered list, empty state. |
| **Calendar** | `Calendar` | month + selected | 6x7 grid, outside days, leap-year safe. |
| **Date Picker** | `DatePicker` | open | Trigger plus calendar panel. |

### Host-owned composites

| Component | Constructor | State | Notes |
|---|---|---|---|
| **Scroll Area** | `ScrollArea` | scroll | Reuses `ScrollState` and `Scrollbar`. |
| **Resizable** | `Resizable` | ratio | Two panes and a focusable divider. |
| **Sidebar** | `Sidebar`, `SidebarNavItem` | open + collapsed | Navigation surface, active rows, badges. |
| **Carousel** | `Carousel` | index | One visible slide, controls, dots. |
| **Data Table** | `DataTable` | host | Header, rows, empty state, optional selection. |
| **Chart** | `Chart` | host | Bar frame, legend, totals; no canvas. |
| **Code Block** | `CodeBlock` | static | Header and `<pre><code>` body. |
| **Message** | `Message` | static | Chat row with avatar and meta. |
| **Bubble** | `Bubble` | static | Default, primary, and muted bubbles. |
| **Attachment** | `Attachment` | host | File tile with progress and action hook. |
| **Marker** | `Marker` | static | Timeline marker with status. |
| **Message Scroller** | `MessageScroller` | host | Scrollable chat viewport with top fade. |
| **Questionnaire** | `Questionnaire` | host | Questions, choices, progress. |
| **Trace** | `Trace` | host | Timing rows with status colors. |

## Layout and compatibility helpers

These helpers predate the ChadCN set or support host-owned compositions. They
remain exported so existing applications keep working.

| Helper | Purpose |
|---|---|
| **Screen**, **Panel** | Full viewport canvas and an unpadded surface. |
| **TextField**, **SearchField** | Labeled text and search controls; values stay on `Page.Form`. |
| **Choice** | Selectable option with a secondary description. |
| **Notice** | Inline status message. |
| **Scrollbar**, **ScrollState** | Vertical pane scrollbar and its shared range state. |
| **DialogFocus** | Tab scope and focus restore for a modal. |

## Builder dependency

The set uses the `ui.Element` surface: `Tag`, `Div`, `Span`, `P`, `Button`,
`Input`, `Section`, `Nav`, `Header`, `Ul`, `Li`, `Text`, `Children`, `Attr`,
`Class`, `ID`, `OnClick`, the generated CSS methods, and the value helpers
`Px`, `Percent`, `Vh`, `Number`, `Edges`, `SolidBorder`, `LengthVar`,
`ColorVar`, `Hex`, `Transparent`, `Var`, and `RawValue`.

`ui` stays experimental. A change to those methods is a component-set change
and needs a migration for every constructor.

## State and events

The host owns everything that changes. Constructors take the current state and
return a fresh tree. A click runs a Go callback, the callback updates host
state, and `ui.View.Update` rebuilds the tree. There is no component lifecycle,
hidden state, or global registry.

Use stable IDs on interactive controls. The component derives part IDs from the
base ID, such as `id + "-panel"`. Do not generate random IDs.

Text inputs stay on `Page.Form`. Checkbox, radio, switch, toggle, select,
combobox, tabs, accordion, and pagination state is passed back into the
constructor.

Tooltip and HoverCard need host hover state. Read `Handlers.Hover`, compare the
next box ID, set `Open`, and rebuild.

Overlay focus, Escape handling, and click-away stay with the host. Use
`DialogFocus` to scope Tab traversal and restore focus when the panel closes.

## Theming

Components read `--of-*` variables with fallbacks. `chadcn.Light` and
`chadcn.Dark` set the full set. Override tokens on `:root` or scope them to one
subtree with `theme.Overrides.Apply`:

```go
save := chadcn.Button("Save", chadcn.ButtonOptions{
    Variant: chadcn.ButtonFilled,
    Overrides: theme.Overrides{"--of-primary": "#21694e"},
})
```

## Tests and gallery

Each component has a focused render test in `components/chadcn`. The tests
check classes, roles, ARIA state, text, and IDs. Pixel tests cover the theme and
button geometry.

The gallery is `examples/builder/ui-builder`. Run it from the repository root:

```sh
go run ./examples/builder/ui-builder
```

## Limits

- No JavaScript, live DOM, canvas, or `<a href>` navigation.
- Charts, data tables, carousels, sidebars, and message scrollers are host-owned
  compositions. They do not sort, page, animate, or fetch by themselves.
- Code block syntax highlighting is not part of the set.
- `:disabled` and `::placeholder` do not match. Components use `[disabled]` and
  the page's placeholder attributes.
- `:focus-visible` matches like `:focus`.
- The engine does not apply `position: absolute` to a flex child. Overlays use
  a block containing block or a fixed backdrop.
