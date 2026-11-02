# ChadCN component set

> **Status:** planned, not started.
> **Parent:** [maintenance ledger](../../../README.md)
> **Visual reference:** `examples/AI/shadcn/` (local, gitignored)
> **Library path:** `components/chadcn`
> **Package:** `chadcn`

## Goal

Make `components/chadcn` the first set in a multi-set component library. A
future set can live beside it at `components/<set>` without moving the package
or changing the import paths of the first set.

The reference directory holds 81 light and 81 dark images. They are 66
components plus 15 open states. The open states are not separate components.
They are extra gallery and test states for components whose default example
shows a trigger.

The visual reference is the shadcn/ui docs, Base UI variant. The implementation
is independent Go and CSS, uses no Tailwind, and claims no affiliation. The
package name is ChadCN, matching the directory.

## Count

| Item | Count |
|---|---|
| **Base components** | 66 |
| **Open states** | 15 |
| **Images per theme** | 81 |
| **Images for both themes** | 162 |

The 15 open states belong to 15 of the 66 base components: alert-dialog,
combobox, context-menu, date-picker, dialog, drawer, dropdown-menu, hover-card,
menubar, navigation-menu, popover, select, sheet, toast, and tooltip. They add
work for state, focus, and gallery coverage, but they do not add constructors.

## Current state

`ui/components` is the existing control package. It has Button, Card, Badge,
TextField, Switch, Checkbox, Progress, Notice, Dialog, DialogFocus, Scrollbar,
Screen, Panel, SearchField, and Choice. It has tests and two consumers:
`examples/builder/ui-builder` and `examples/builder/jmap-mail`.

Phase 0 moves it to `components/chadcn` and updates both consumers. Do not keep
a forwarding package. The library is pre-1.0 and a second component package
would split the public API.

`ui` stays the experimental builder. `ui/theme` stays the shared token package
for now. A later plan can promote the theme when a second set proves the need.

## Target tree

```text
components/
├── README.md                    how component sets are organized
└── chadcn/
    ├── README.md                API, status table, state, theming
    ├── doc.go                   package contract
    ├── styles.go                //go:embed css/*.css, exports Stylesheet
    ├── theme.go                 ChadCN light and dark palettes
    ├── css/
    │   ├── 00-base.css          focus ring, disabled, motion, reset
    │   ├── button.css
    │   ├── card.css
    │   └── ...                  one file per component
    ├── button.go
    ├── button_test.go
    ├── card.go
    ├── dialog.go
    ├── ...
    ├── internal/
    │   ├── ids.go               derived part IDs, ID validation
    │   ├── aria.go              shared state attributes
    │   └── testutil/
    │       ├── golden.go        HTML snapshots
    │       └── layout.go        render and box assertions
    └── testdata/
        └── golden/
            ├── button/
            └── dialog/
```

A future set stays a sibling:

```text
components/
├── chadcn/
└── <set>/
```

Shared helpers move to `components/internal` only when the second set needs
them. Until then they stay inside `chadcn`.

## Component contract

1. One package, one file per component, one test file beside it.
2. A constructor returns a fresh `*ui.Element`. Options use an options struct,
   and the zero value is the default:
   `func Button(label string, options ButtonOptions) *ui.Element`.
3. Content arrives through children. Named parts such as an icon, trigger, or
   footer arrive through options fields.
4. The host owns state. Constructors take `checked`, `open`, `active`, `page`,
   and render from those values. Actions go through `OnClick` or `data-action`.
   There is no hidden state, no global registry, and no component lifecycle.
5. Interactive components require a stable host ID. Derived part IDs use the
   base ID, for example `id + "-panel"`. No random IDs.
6. Constructors emit stable classes: `.of-button`, `.of-button--filled`,
   `.of-button__icon`. Variants and states are CSS, not duplicated Go paint.
7. Paint and states live in `css/`. Geometry and dynamic values stay inline.
   The app appends `chadcn.Stylesheet` to its theme:
   `theme.Dark + chadcn.Stylesheet`.
8. Theme values use `--of-*` tokens with fallbacks. `theme.Overrides` still
   works because it writes custom properties, not ordinary properties.
9. Accessibility is part of the constructor: roles, labels, `aria-*` state,
   focus scope, and the keyboard contract.
10. Every component has a Go doc comment, a README status row, and a state in
    the gallery.

Class names use the `.of-` prefix. A component root is `.of-<component>`,
a part is `.of-<component>-<part>`, and a variant is
`.of-<component>--<variant>`.

## Open states are parameters

The 15 open-state images do not need 15 extra constructors. Each one is the same
component with `Open: true` and its data passed through the options struct. The
closed state is the same call with `Open: false`.

The rule:

- A component that can open takes `Open bool` in its options.
- When `Open` is false, it renders its trigger if it owns one, otherwise nil.
- When `Open` is true, it renders the trigger plus the open surface.
- The host sets `Open` from a click, hover, or key handler, then rebuilds the
  view with `View.Update`.
- The `-open.png` image is a render test of the same constructor with
  `Open: true`. There is no `DialogOpen` constructor.

```go
func (a *App) profileDialog() *ui.Element {
    return chadcn.Dialog(chadcn.DialogOptions{
        Open:        a.editOpen,
        ID:          "edit-profile",
        Title:       "Edit profile",
        Description: "Make changes to your profile here.",
        Trigger:     chadcn.Button("Edit profile", chadcn.ButtonOptions{}),
        CloseAction: "profile:close",
        OnOpenChange: func(ctx context.Context, open bool) error {
            a.editOpen = open
            return a.refresh(ctx)
        },
    },
        chadcn.Field(chadcn.FieldOptions{Label: "Name", Control: chadcn.Input("name", "Pedro Duarte")}),
        chadcn.Field(chadcn.FieldOptions{Label: "Username", Control: chadcn.Input("username", "@peduarte")}),
    )
}
```

Data lists use the same parameter path:

```go
func (a *App) fruitSelect() *ui.Element {
    return chadcn.Select(chadcn.SelectOptions{
        ID:    "fruit",
        Open:  a.fruitOpen,
        Value: a.fruit,
        Items: []chadcn.SelectItem{
            {Value: "apple", Label: "Apple"},
            {Value: "banana", Label: "Banana"},
            {Value: "blueberry", Label: "Blueberry"},
        },
        OnOpenChange: func(ctx context.Context, open bool) error {
            a.fruitOpen = open
            return a.refresh(ctx)
        },
        OnSelect: func(ctx context.Context, value string) error {
            a.fruit = value
            a.fruitOpen = false
            return a.refresh(ctx)
        },
    })
}
```

| Open image | Constructor | State and data fields |
|---|---|---|
| `alert-dialog-open` | `AlertDialog` | `Open`, `Title`, `Description`, `ConfirmLabel`, `CancelLabel`, `OnConfirm`, `OnCancel` |
| `combobox-open` | `Combobox` | `Open`, `Query`, `Value`, `Items`, `OnQuery`, `OnSelect` |
| `context-menu-open` | `ContextMenu` | `Open`, `X`, `Y`, `Items`, `OnSelect` |
| `date-picker-open` | `DatePicker` | `Open`, `Month`, `Selected`, `OnMonthChange`, `OnSelect` |
| `dialog-open` | `Dialog` | `Open`, `Title`, `Description`, `children`, `CloseAction` |
| `drawer-open` | `Drawer` | `Open`, `Side`, `Title`, `Description`, `children` |
| `dropdown-menu-open` | `DropdownMenu` | `Open`, `Trigger`, `Items`, `OnSelect` |
| `hover-card-open` | `HoverCard` | `Open`, `Side`, `children` |
| `menubar-open` | `Menubar` | `Open`, `Menus`, `OnOpenChange` |
| `navigation-menu-open` | `NavigationMenu` | `Open`, `Items`, `OnOpenChange` |
| `popover-open` | `Popover` | `Open`, `Side`, `Trigger`, `children` |
| `select-open` | `Select` | `Open`, `Value`, `Items`, `OnSelect` |
| `sheet-open` | `Sheet` | `Open`, `Side`, `Title`, `Description`, `children` |
| `toast-open` | `Toast` | `Open`, `Title`, `Description`, `Action` |
| `tooltip-open` | `Tooltip` | `Open`, `Content`, `Side` |

Static components pass data the same way, with no open state:

```go
chadcn.Alert(chadcn.AlertOptions{
    Variant:     chadcn.AlertDanger,
    Title:       "Payment failed",
    Description: "Your card was declined.",
    Action:      chadcn.Button("Retry", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}),
})
```

Tooltip and HoverCard do not own hover. The host reads `Handlers.Hover`, sets
`Open` from the trigger ID, and rebuilds the view.


## Styling rule

The existing package paints with inline styles. Inline `style` beats normal
theme rules, so `:hover`, `:focus-visible`, and variants cannot change those
colors from a shared sheet. At 66 components that pushes state CSS into every
application.

ChadCN keeps ordinary paint in CSS:

```css
.of-button { background: var(--of-control); color: var(--of-text); }
.of-button:hover { background: var(--of-control-hover); }
.of-button[disabled] { opacity: .5; }
```

The builder still owns structure, geometry, and dynamic values such as a
progress width or an open panel offset.

## Phases

| # | Phase | Components | Depends on |
|---|---|---|---|
| 00 | **Move and foundation** | existing controls, package plumbing | none |
| 01 | **Primitives** | 15 static components | 00 |
| 02 | **Actions and forms** | 14 components | 00, 01 |
| 03 | **Feedback and disclosure** | 7 components, 1 open state | 01, 02 |
| 04 | **Overlays and menus** | 12 components, 11 open states | 02, 03 |
| 05 | **Selection** | 4 components, 3 open states | 02, 04 |
| 06 | **Host-owned composites** | 14 components | 01 to 05 |
| 07 | **Gallery, docs, gates** | all | all |

The component counts add to 66. The open states add to 15.

### Phase 00: move and foundation

- [ ] Add `components/README.md`, explaining the multi-set layout.
- [ ] Move `ui/components` to `components/chadcn`; package name `chadcn`.
- [ ] Update imports in `examples/builder/ui-builder` and
      `examples/builder/jmap-mail`.
- [ ] Remove `ui/components`; no forwarding shim.
- [ ] Add `components/chadcn/README.md`, `doc.go`, `styles.go`, and
      `css/00-base.css`.
- [ ] Add `theme.go` with light and dark palettes that match the reference
      captures.
- [ ] Add `internal/ids.go`, `internal/aria.go`, and `internal/testutil`.
- [ ] Move the existing tests and keep them passing.
- [ ] Record the `ui` methods the component set depends on. If `ui` stays
      experimental, those methods must not change without a migration.

### Phase 01: primitives

- [ ] `screen.go`: Screen and Panel.
- [ ] `card.go`: card, header, title, description, content, footer.
- [ ] `separator.go`: horizontal and vertical.
- [ ] `aspect_ratio.go`: fixed-ratio box.
- [ ] `badge.go`: tones and sizes.
- [ ] `avatar.go`: image, initials, fallback, status.
- [ ] `kbd.go`.
- [ ] `skeleton.go`.
- [ ] `spinner.go`.
- [ ] `progress.go`.
- [ ] `empty.go`.
- [ ] `item.go`.
- [ ] `label.go`.
- [ ] `table.go`: table, header, body, row, cell, caption.
- [ ] `typography.go`: heading, paragraph, inline code, blockquote, list
      wrappers.
- [ ] `direction.go`: left-to-right and right-to-left wrapper.
- [ ] Golden HTML, light and dark pixel checks, and one gallery screen.

### Phase 02: actions and forms

- [ ] `button.go`: variants, sizes, icon slot, disabled, loading.
- [ ] `button_group.go`.
- [ ] `toggle.go`.
- [ ] `toggle_group.go`.
- [ ] `field.go`: label, description, error, required.
- [ ] `input.go`: text, search, email, password, file.
- [ ] `textarea.go`.
- [ ] `native_select.go`.
- [ ] `checkbox.go`.
- [ ] `radio_group.go`.
- [ ] `switch.go`.
- [ ] `slider.go`.
- [ ] `input_group.go`: prefix, suffix, addons.
- [ ] `input_otp.go`.
- [ ] Keep text values on `Page.Form`. Checkbox, radio, switch, and toggle
      state stays in the host and is passed back into the constructor.
- [ ] Golden HTML, focus and keyboard tests, light and dark pixel checks.

### Phase 03: feedback and disclosure

- [ ] `alert.go`: info, success, warning, danger, action slot.
- [ ] `toast.go`: host-owned queue, open and close actions, `-open` state
      through `ToastOptions.Open`.
- [ ] `accordion.go`: single and multiple open sets.
- [ ] `collapsible.go`.
- [ ] `tabs.go`: active tab, keyboard traversal.
- [ ] `breadcrumb.go`: text, separator, current page.
- [ ] `pagination.go`: page, previous, next, disabled bounds.
- [ ] Golden HTML, state tests, and the `toast-open` gallery state.

### Phase 04: overlays and menus

- [ ] `dialog.go`: move the existing Dialog and DialogFocus.
- [ ] `alert_dialog.go`.
- [ ] `sheet.go`.
- [ ] `drawer.go`.
- [ ] `popover.go`.
- [ ] `tooltip.go`.
- [ ] `hover_card.go`.
- [ ] `dropdown_menu.go`.
- [ ] `context_menu.go`.
- [ ] `menubar.go`.
- [ ] `navigation_menu.go`.
- [ ] `command.go`: query, groups, active item, empty state.
- [ ] Overlays position inside a block container. The engine does not apply
      `position: absolute` to a flex child.
- [ ] The host owns open state, Escape handling, click-away, and focus scope.
      Reuse `DialogFocus`.
- [ ] Add the 11 open-state gallery states: alert-dialog, dialog, drawer, sheet,
      dropdown-menu, context-menu, menubar, popover, hover-card, tooltip,
      navigation-menu. Each one is the same constructor with `Open: true`.
- [ ] Golden HTML, focus order, Escape, and click-away tests.

### Phase 05: selection

- [ ] `select.go`: trigger, content, group, item, selected marker.
- [ ] `combobox.go`: input, filtered list, empty state.
- [ ] `date_picker.go`: trigger, calendar panel, selected date.
- [ ] `calendar.go`: month grid, weekday header, disabled days, outside days.
- [ ] Add the 3 open-state gallery states: select, combobox, date-picker. Each
      one is the same constructor with `Open: true`.
- [ ] Keyboard tests: arrows, Home, End, Enter, Escape.
- [ ] Date tests: month rollover, leap year, first weekday, disabled dates.

### Phase 06: host-owned composites

- [ ] `scroll_area.go`: move Scrollbar into ScrollArea; keep ScrollState.
- [ ] `resizable.go`: split panes, drag state helper.
- [ ] `sidebar.go`: layout, collapsible state, mobile state.
- [ ] `carousel.go`: host index, previous and next, slide slots.
- [ ] `data_table.go`: header, rows, empty state, selection column. Sorting,
      filtering, and paging stay in the host.
- [ ] `chart.go`: chart frame, legend, axes, and a data contract. Drawing
      strategy is a separate decision.
- [ ] `code_block.go`: frame, filename, copy affordance, plain text body.
      Syntax highlighting is not part of this phase.
- [ ] `message.go`, `bubble.go`, `attachment.go`, `marker.go`: chat primitives.
- [ ] `message_scroller.go`: windowing contract over the existing scroll state.
- [ ] `questionnaire.go`: question, choices, progress, submit slot.
- [ ] `trace.go`: trace rows, timing bars, status.
- [ ] Document which pieces are components and which are composition rules.
- [ ] Light and dark gallery states for each composite.

### Phase 07: gallery, docs, and gates

- [ ] `examples/builder/ui-builder` shows all 66 components and all 15 open
      states in light and dark.
- [ ] Add `documentation/chadcn.md`: import path, stylesheet, theme, state,
      focus, and component status.
- [ ] Add the package status table to `components/chadcn/README.md`.
- [ ] Update `AGENTS.md` with the `components/chadcn` ownership line.
- [ ] Run `make test`, `make lint`, `make docs-check plans-check examples-check`,
      and `make graphics-check`.
- [ ] Confirm no new example import of `ownframe/internal/`.
- [ ] Confirm the gallery and mail examples still build and run.

## Acceptance criteria

- 66 constructors, one file per component.
- 15 open states reachable from the gallery and covered by tests. Each one
  calls its constructor with `Open: true`; no separate open-state constructor.
- Light and dark palettes use the reference captures as the visual target.
- Focus rings, disabled states, hover states, and selected states match the
  intended component.
- Overlays keep focus inside the panel, close on Escape, and restore focus.
- No JavaScript, canvas, or `<a href>` navigation.
- Existing examples migrate to `components/chadcn`.
- The package has no import of `ownframe/internal/` from examples.
- The stylesheet is the only file an application must append for paint.

## Risks

- `ui` is experimental. The component set returns `*ui.Element`, so the builder
  methods used here are part of the component contract. Freeze that subset or
  accept churn.
- Inline styles beat theme rules. Moving paint to CSS is required for hover and
  focus states. A component that keeps paint inline cannot participate in the
  shared state rules.
- The engine does not apply `position: absolute` to a flex child. Overlay
  panels need a block containing block, as the existing Dialog tests show.
- `:disabled`, `::placeholder`, and `::selection` never match. Use `[disabled]`,
  `data-ownframe-placeholder`, and the existing selection span.
- `:focus-visible` matches like `:focus`. Keyboard focus and pointer focus look
  the same.
- Charts, data tables, carousels, and sidebars need host state. Calling them
  drop-in widgets would hide that contract.
- Code block syntax highlighting is a separate project. Plain text first.

## Out of scope

- Exact Tailwind or shadcn visual parity.
- JavaScript, a live DOM, video, canvas, WebGL, and `<a href>` navigation.
- A data-table controller with sorting, filtering, and paging.
- A chart renderer and chart tooltips.
- Native screen-reader adapters.
- Copying upstream component code. The implementation stays Go and CSS.

## Open decisions

- Should `ui/theme` move to root `theme` now, or after the second component set?
- Should `ui/components` be removed outright or kept for one release behind a
  deprecation note? The plan recommends removal.
- Should the reference screenshots move from `examples/AI/shadcn` to
  `examples/AI/chadcn`? The plan leaves them where they are.
