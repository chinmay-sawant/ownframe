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

- A three-pane layout: mailboxes, messages, and a reading pane, with visible\n  spacing between panes and a vertically aligned mail-search control.\n- A Settings button with a working light/dark theme toggle.\n- A draggable native custom-titlebar window with minimize, maximize, and\n  close controls. Browser preview does not move the OS window.
- Eight seeded messages with JMAP-shaped IDs, mailbox membership, keyword
  flags (`$seen`, `$flagged`, `$draft`), addresses, and timestamps.
- Clickable folder navigation, selected-message reading, unread status,
  star/unstar, archive, deletion into the local Trash mailbox, and local draft
  creation.
- A search field that filters by sender, subject, preview, and body.
- A large UI authored exclusively with the **public**
  `ownframe`, `ownframe/ui` APIs, no import of `internal/`.
- CSS expressed through the fluent builder. Property names are validated by
  the registry generated from the matrix's 555 implemented properties.

## Scope and security

This is a **JMAP-inspired mock client**, not a real RFC 8621 protocol client.
It makes **no network calls**, stores no credentials, and never sends, deletes,
or modifies mail on a server. All state is in memory; restarting restores
the sample messages. The example deliberately avoids adding HTTP or JMAP
authentication to the builder spike.

A future JMAP implementation should add session discovery, authentication,
`Mailbox/get`, `Email/query`, `Email/get`, `Email/set`,
`Email/changes`, pagination, error recovery, and a storage/cache layer.

## Architecture

```text
fixtures (Go structs)
       |
public ui.Element builders
       |
ui.NewPage / ui.Render
       |
public ownframe.Page + Handlers.Click + Page.Load
       |
OwnFrame / Blinkless rendering pipeline
```

The app uses explicit `data-action` values for stateful routing and the
public `Page.Load` API when local state changes. It does not use a private
builder API, a DOM, JavaScript, or any Blinkless internals.

Related: [builder issue #59](https://github.com/chinmay-sawant/ownframe/issues/59)
and [draft PR #60](https://github.com/chinmay-sawant/ownframe/pull/60).
