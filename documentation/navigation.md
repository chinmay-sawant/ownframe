# Navigation

`New` records the template source as history entry 0. `Load` parses a new template, drops every entry after the current one, appends the new source, and redraws. `Back` and `Forward` reparse an earlier or later entry and redraw. `HTML` returns the source at the current index.

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

`Click` hit-tests the last picture and draws the page again; an error returns before the redraw. A box without an id falls back to the innermost id-bearing element under the point, so a click on a child icon reaches its control. A hit box with a non-empty action and a registered route calls `Load` with that HTML and does not call the click handler. An action with no route falls through to the click handler. A form control activates and never follows a route. The layout box has no `href` field, so an `<a href>` does not navigate. Put the route name in `data-action`.

`SetData` still applies to whichever template is current. `Load` does not change the data. `Load`, `Back`, and `Forward` reset form state, because the current document changed ([forms.md](forms.md)). The theme is untouched, and the window keeps the scroll offset, pulled back inside a page that got shorter.

The [history example](../examples/history) drives `Load`, `Back`, `Forward`, `HTML`, and `Route` from toolbar boxes.

## Receive app deep links

Register each scheme your app accepts in `Config.DeepLinkSchemes`. Add a
`Handlers.DeepLink` callback, then pass OS-delivered URLs to
`Page.DispatchDeepLink` on the page's owning event loop.

```go
import (
    "context"
    "net/url"

    "github.com/chinmay-sawant/ownframe"
)

func newPage() (*ownframe.Page, error) {
	p, err := ownframe.New(ownframe.Config{
		HTML: `<h1>Home</h1>`, Width: 480, Height: 320,
		DeepLinkSchemes: []string{"ownframe", "https"},
	})
	if err != nil {
		return nil, err
	}
	p.Handle(ownframe.Handlers{DeepLink: func(ctx context.Context, link *url.URL) error {
		if link.Scheme == "ownframe" && link.Host == "chat" {
			return p.Load(ctx, `<h1>Chat</h1>`)
		}
		return nil
	}})
	return p, nil
}
```

Set at least one scheme to accept URLs. An empty list rejects every URL.
`New` rejects malformed schemes and blocks `javascript`, `data`, `vbscript`,
`file`, `blob`, and `about`. Dispatch rejects relative URLs and schemes outside
the allowlist. It lowercases the scheme before the callback runs.
`DispatchDeepLink` does not draw. The callback can call `Load`, or update app
data and call `Redraw`.

Ownframe does not register URL protocols with the operating system. The app
and its package installer own that registration. They must deliver both
cold-start and later URLs, and queue those events onto the page's loop. See
the [v0.0.2 navigation checklist](../plans/v0.0.2/accessibility-desktop-ime.md#phase-6-navigation-and-text-services).

## Open clicked links

`Handlers.OpenLink` receives `http`, `https`, and `mailto` targets from clicked
anchors. It does not open a browser or mail app itself. The application chooses
the OS action. Relative and fragment links are not handed off. `data-action`
routes keep precedence. The current Blinkless layout does not emit custom
scheme targets such as `ownframe://...`; use a `data-action` handler for an
app-local action.

## Keep shortcuts and spellcheck in the app

Ownframe key handlers only receive keys while the window has focus. Ebitengine
does not provide the OS-global registration, conflict handling, or unregister
lifecycle needed for global shortcuts. An app that needs them must use a
platform adapter and deliver shortcut actions to the page loop.

Spellcheck belongs in an application or platform adapter. Ownframe edits
canvas-rendered controls rather than native text widgets, so it has no system
dictionary or native suggestion UI to call. An app can choose a checker and
language policy, then present suggestions in its own controls.
