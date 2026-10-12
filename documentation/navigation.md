# Navigation

`New` records its HTML string as history entry 0. `Load` parses a new template, drops every entry after the current one, appends the new source, and redraws. `Back` and `Forward` reparse an earlier or later entry and redraw. `HTML` returns the source at the current index.

```go
if err := page.Load(ctx, `<h1 id="next">Next</h1>`); err != nil {
    return err
}
if err := page.Back(ctx); err != nil {
    return err
}
_ = page.HTML()
```

`Back` on the first entry, and `Forward` past the last entry, return `ErrNoHistory`. `Load` of blank HTML returns `ErrEmptyHTML`. A template that does not parse is returned as the parse error and does not move history.

`Route` maps a `data-action` value to HTML for `Click` to load. It returns nothing and does not parse or draw. An empty action is ignored. Routing the same action again replaces the HTML.

```go
page.Route("inbox", `<h1 id="in">Inbox</h1>`)
```

`Click` hit-tests the last picture and draws the page again when it succeeds. A hit box with a non-empty action and a registered route calls `Load` with that HTML and does not call the click handler. An action with no route falls through to the click handler. A form control activates and never follows a route. The layout box has no `href` field, so an `<a href>` does not navigate. Put the route name in `data-action`.

`SetData` still applies to whichever template is current. `Load` does not change the data. `Load`, `Back`, and `Forward` reset form state, because the current document changed ([forms.md](forms.md)). The theme and the window scroll offset are untouched.

The [history example](../examples/history) drives `Load`, `Back`, `Forward`, `HTML`, and `Route` from toolbar boxes.
