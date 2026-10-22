# Hot reload

A page built with `Config.File` watches that file. Save an edit and the open
window redraws within 250 ms. `Config.ThemeFile` watches the extra stylesheet
the same way. A page built from `Config.HTML` with neither file field keeps
the behavior it had before.

## Config

| Field | Meaning |
|-------|---------|
| `File` | Path to the page source. `New` reads it before the template parse. |
| `ThemeFile` | Path to the theme stylesheet. `New` reads it before `parseTheme`. |
| `DisableHotReload` | `true` turns the watch off. The default is on. |

Setting `HTML` and `File` together returns `ErrBadSource`; there is no
precedence rule. The same error covers `Theme` with `ThemeFile` and a file
that cannot be read at `New`. A blank string or a blank file returns
`ErrEmptyHTML`. `Page.Testing().SetHotReload(false)` turns a running watch off and
`Page.Testing().SetHotReload(true)` turns it back on. `Page.Testing().Watching()` reports the
current answer. `HTML()` returns the file bytes, not the path, so it matches
what the window shows.

## The poll

The watch is a poll, not an OS notification, and the library adds no module
for it. Every 250 ms the window asks the page to check its files. A stat that
shows the same modification time and size skips the read. A stat that changed
reads the file and compares a SHA-256 hash with the last accepted bytes. Only
a different hash parses. `internal/window` runs the poll after `tickFrame`
and before `syncImage`, so the existing generation sync picks up the new
picture.

`Serve` has no frame loop. It polls before answering `GET /` and
`GET /frame.png`, under the same mutex as the other routes, so the next
request sees an edit. While a watch is active the shell page reloads
`/frame.png` every 250 ms in the browser. A string page emits no script.

## Parse errors

A template that does not parse or a stylesheet that does not compile keeps the
last good template or theme. The page stores the bytes as pending and returns
the error. The next poll retries those bytes even when the stat did not move,
which covers a half-written save. A new edit that changes the file replaces
the pending bytes. A file that disappears keeps the last good picture and
reports the stat error once. A recreated file reloads.

The window prints each distinct error once, in this shape, and keeps drawing:

```
hot reload: <path>: <error>
```

`Serve` prints the same line to stderr, once per distinct message. A
successful poll clears the line, so the same error after a fix prints again.

## What survives

A reload replaces the current history entry and nothing else. The history
length and index do not move, so `Back` and `Forward` still walk older
entries. The template data from `SetData` stays.

- Form values and the focused control survive while the control still exists.
  A field removed from the new template loses focus. No `Change` callback
  fires: a reload is not an edit.
- Hover and active ids are checked against the new boxes and cleared when the
  element is gone.
- A scroll offset is pulled back inside a page that got shorter.
- `SetImage` entries stay, so a template that still uses the same `src` finds
  the same bytes.
- A `SetTick` callback stays registered.

## host.Reloader

The window polls any screen with this method, not only `*ownframe.Page`:

```go
type Reloader interface {
    PollReload(ctx context.Context) (bool, error)
}
```

A true return means the page changed and the window should show the new
generation. A custom screen without the method is never polled.

## Limits

- The watch covers the template and the theme. It does not watch Go source,
  so a changed handler still needs a restart. A tool such as `air` around
  `go run` handles that outside this library.
- Templates are trusted author content: a template action can call methods
  on the page data and ownframe does not sandbox execution. Ship production
  with the watch off (`Config.DisableHotReload: true`, `-reload=false` in
  the reload example) and keep template files on owner-only permissions, so
  a changed file cannot swap the running page.
- wasm and mobile read `Config.File` once at `New` and never watch. Keep
  `Config.HTML` there.
- A network share that updates content without changing the modification time
  or size is not seen. The pending retry covers a torn write, not that.
