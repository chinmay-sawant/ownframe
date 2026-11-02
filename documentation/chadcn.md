# ChadCN components

`components/chadcn` is a reusable component set built on `ui` and `ui/theme`.
It supplies buttons, forms, feedback, overlays, menus, selection controls, and
host-owned composites. The full constructor table is in
[components/chadcn/README.md](../components/chadcn/README.md).

## Set up a page

```go
sheet := chadcn.Dark + chadcn.Stylesheet
view, err := ui.Mount(ownframe.Config{
    Title: "ChadCN", Width: 960, Height: 720, Theme: sheet,
}, chadcn.Screen(
    chadcn.Card(
        chadcn.CardContent(
            chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}),
        ),
    ),
))
if err != nil {
    return err
}
_ = view
```

`chadcn.Light` and `chadcn.Dark` set the `--of-*` tokens. `chadcn.Stylesheet`
carries the component paint and state rules. Append both to the theme. An
application with its own palette can keep its tokens and append only
`chadcn.Stylesheet`.

## Keep state in the host

Components do not store state. Pass the current value into the constructor,
handle the action in Go, change the state, and rebuild with `View.Update`.

```go
selectControl := chadcn.Select(chadcn.SelectOptions{
    ID:    "fruit",
    Open:  fruitOpen,
    Value: fruit,
    Items: []chadcn.SelectItem{
        {Value: "apple", Label: "Apple"},
        {Value: "banana", Label: "Banana"},
    },
    OnOpenChange: func(ctx context.Context, open bool) error {
        fruitOpen = open
        return nil
    },
    OnSelect: func(ctx context.Context, value string) error {
        fruit = value
        fruitOpen = false
        return nil
    },
})
_ = selectControl
```

This is the same pattern for `checked`, `active`, `open`, `month`, `page`, and
`query`. Text inputs are the exception: the page stores their value on
`Page.Form`, so give each input a stable ID.

## Open a component

Every openable component takes `Open bool`. The closed call renders the trigger
if the component owns one. The open call renders the trigger plus the surface.
The 15 open-state screenshots in the reference set are `Open: true` renders of
the same constructors.

The host owns Escape, click-away, and focus. `DialogFocus` scopes Tab traversal
for a dialog and restores the previous control when it closes:

```go
var focus chadcn.DialogFocus
focus.Activate(page, "dialog-save", "dialog-cancel")
// Render the open dialog.
focus.Deactivate(page)
```

For overlays that use a backdrop, mount the component at the screen root so
the backdrop covers the window. For anchored menus and popovers, mount the
component in a block container with `position: relative`.

Tooltip and HoverCard need host hover state. Read `Handlers.Hover`, compare the
next box ID, set `Open`, and rebuild.

## Component families

| Family | Constructors |
|---|---|
| **Primitives** | Card, Badge, Avatar, Separator, Skeleton, Spinner, Progress, Empty, Item, Kbd, Label, Table, Typography, AspectRatio, Direction |
| **Actions** | Button, ButtonGroup, Toggle, ToggleGroup |
| **Forms** | Field, Input, Textarea, NativeSelect, Checkbox, RadioGroup, Switch, Slider, InputGroup, InputOTP |
| **Feedback** | Alert, Toast, Accordion, Collapsible, Tabs, Breadcrumb, Pagination |
| **Overlays** | Dialog, AlertDialog, Sheet, Drawer, Popover, Tooltip, HoverCard, DropdownMenu, ContextMenu, Menubar, NavigationMenu, Command |
| **Selection** | Select, Combobox, Calendar, DatePicker |
| **Composites** | ScrollArea, Resizable, Sidebar, Carousel, DataTable, Chart, CodeBlock, Message, Bubble, Attachment, Marker, MessageScroller, Questionnaire, Trace |

Composites keep their domain state in the host. A data table does not sort or
page by itself. A chart does not fetch data. A carousel does not auto-advance.
The component supplies structure and paint; the application supplies behavior.

## Gallery

`examples/builder/ui-builder` is the component gallery. It switches between the
ChadCN light and dark palettes and shows the controls in a native window.

## Limits

- No JavaScript, live DOM, canvas, or `<a href>` navigation.
- `:disabled` and `::placeholder` do not match. Use `[disabled]` and the page's
  placeholder attributes.
- `:focus-visible` matches like `:focus`.
- The engine does not apply `position: absolute` to a flex child. Use a block
  containing block for anchored panels and a fixed backdrop for modals.
- Code block syntax highlighting is outside this set.
