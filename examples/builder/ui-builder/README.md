# OwnFrame component gallery

This example is an interactive catalog for `ownframe/ui/components`.
The shared package owns reusable controls. The neutral `ui` package builds
HTML and layout. App-owned components live in `custom/`.

Run from the repository root:

```sh
go run ./examples/builder/ui-builder
```

Use `-web` for the PNG browser preview or `-dev` to open DevTools. The desktop
gallery supports windows from 900 by 610 pixels. The header and navigation
stay visible while the content pane scrolls. The pane uses the shared
`components.Scrollbar` and `components.ScrollState`. `scroll.go` measures its
content after layout and routes wheel, native thumb drag, and focused keyboard
input to that state. Clicking above or below the thumb pages the content.

## Explore the catalog

- Components demonstrates button variants and sizes, disabled state, badges,
  switches, labeled fields, checkboxes, progress, and notices. The existing
  navigation, pagination, avatar, and small table examples show compositions.
- Forms demonstrates labeled controls, validation, save, and reset.
- Layout previews typed layout values and their Go source.
- State & motion demonstrates controlled state, keyframes, and transitions.
- Collections demonstrates filtering and selection.
- Theme studio changes the global palette and token overrides.
- Source & preview shows a shared card, its Go composition, and generated HTML.
- Dialogs demonstrates a controlled modal with keyboard focus management.
- Custom component composes two independent `custom.ProjectCard` instances.

This is an OwnFrame component library example. It does not import Material UI
or claim its complete control set or behavior.

## Change a theme

The header switches between OwnFrame's Light and Dark themes and the app-owned
Studio palette. `theme.go` contains Studio and the custom card's local
properties. Theme studio can override the global accent, radius, and card
padding; Reset removes those overrides while retaining the selected palette.

The order is OwnFrame defaults, application properties, then component
properties. Every screen uses typed theme tokens for ordinary paint and
borders. Illustration palettes remain local data. The motion example uses
`RawValue` for animation declarations that have no typed helper.

## Build a custom component

`custom/project.go` is a separate Go package. `ProjectCard` accepts
`ProjectOptions` with data, pin state, callbacks, and optional overrides. It
composes shared cards, badges, progress, and buttons and creates a fresh tree
for each call. The host owns the state and supplies callbacks; absent callbacks
disable their action buttons.

`screen_custom.go` mounts two instances. The first inherits the global theme.
The second sets `--project-accent`, a local primary color, and a corner radius.
Their pin and open actions are independent, and switching palettes preserves
their state and local overrides.

The gallery uses `ui.Mount` and `View.Update` to replace the current view and
its callbacks without adding navigation history. Controls with stable IDs
keep live state. Explicit reset actions also update the live form values.

See the [shared component API](../../../ui/components/README.md) for public options,
file-loaded properties, icon slots, and controlled interaction contracts.
