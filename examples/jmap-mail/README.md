# Postmark · JMAP Mail Preview

This demo is a **three-pane desktop mail client mockup**, built with the
public Go API at `github.com/chinmay-sawant/ownframe/ui`. It demonstrates
how downstream application authors can compose substantial interfaces with
the builder added in OwnFrame PR #60.

## Run

From the repository checkout root:

```bash
go run ./examples/jmap-mail
```

Or open the existing OwnFrame browser preview:

```bash
go run ./examples/jmap-mail -web -addr 127.0.0.1:8132
```

## What it demonstrates

- A three-pane layout: mailboxes, messages, and a reading pane, with visible
  spacing between panes and a vertically aligned mail-search control.
- A Settings button with a working Light, Dark, and Custom theme selector.
- Independent list and reader panes using shared `components.Scrollbar` and
  `components.ScrollState`. The list measures rows; the reader measures pixels.
  Both support wheel input, track paging, native thumb dragging, and arrow,
  Page Up/Down, Home, and End keys while the thumb has focus.
- A draggable native custom-titlebar window with a developer-tools toggle,
  minimize, maximize, and close controls. Browser preview does not move the OS
  window.
- Eight seeded messages with JMAP-shaped IDs, mailbox membership, keyword
  flags (`$seen`, `$flagged`, `$draft`), addresses, and timestamps.
- Clickable folder navigation, selected-message reading, unread status,
  star/unstar, archive, deletion into the local Trash mailbox, and local draft
  creation.
- A search field that filters by sender, subject, preview, and body.
- A large UI authored exclusively with the **public**
  `ownframe`, `ownframe/ui` APIs, no import of `internal/`.
- CSS expressed with generated methods for every implemented property.
  Go completion offers property names such as `JustifyContent`; typed helpers
  offer values such as `ui.JustifyContentCenter` and `ui.Px(12)`. Use
  `ui.RawValue` for complex CSS values and `CSSRaw` for runtime property names.
  Typed theme tokens remove raw CSS from the mail view.

## Scope and security

This is a **JMAP-inspired mock client**, not a real RFC 8621 protocol client.
It makes **no network calls**, stores no credentials, and never sends, deletes,
or modifies mail on a server. All state is in memory; restarting restores
the sample messages. The example deliberately avoids adding HTTP or JMAP
authentication to the builder example.

A future JMAP implementation should add session discovery, authentication,
`Mailbox/get`, `Email/query`, `Email/get`, `Email/set`,
`Email/changes`, pagination, error recovery, and a storage/cache layer.

## Architecture

```text
fixtures (Go structs)
       |
public ui.Element builders
       |
ui.Mount / View.Update
       |
public ownframe.Page + View.Handle + themed components
       |
OwnFrame / Blinkless rendering pipeline
```

The app uses explicit `data-action` values for stateful routing and the
public `View.Update` API when local state changes. Updates keep navigation
history and live form edits. Clearing the search explicitly resets its form
value. Light and dark appearances use shared CSS variables; custom stylesheets
can be applied through `Page.SetTheme`. It does not use a private
builder API, a DOM, JavaScript, or any Blinkless internals.

Related: [builder issue #59](https://github.com/chinmay-sawant/ownframe/issues/59)
and [draft PR #60](https://github.com/chinmay-sawant/ownframe/pull/60).


## Extend the appearance

OwnFrame supplies `theme.Light` and `theme.Dark` from `ui/theme`. The mail app's
`theme.go` defines the third Sage palette and the Compose button's local corner
radius. Settings switches the global palette; Compose keeps its component
radius in every palette. `icons.go` supplies the Settings icon, and
`components.go` composes shared controls with mail-specific choices.

The order is OwnFrame defaults, then application properties, then component
properties. Use `theme.Overrides` for custom `--name` properties. Append its
validated `Stylesheet()` to a global preset, or pass it as
`ButtonOptions.Overrides` to configure only one button. `properties.Apply(tree)`
scopes the same overrides to any component tree and its children.

To supply a file, call `theme.LoadOverrides(files, "custom.json")` with an
`fs.FS`, such as `os.DirFS("themes")` or an `embed.FS`. The JSON object maps
custom property names to declaration values. The same loader supports both
global properties and local component overrides. No renderer changes are
needed. See the [component reference](../../ui/components/README.md).
