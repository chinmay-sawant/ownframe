# Web mode

`Serve` shows the page's latest picture in a browser and sends clicks and
typing back to the same page. The window modes are in
[platforms.md](platforms.md).

## Starting it

```
go run ./examples/login -web
```

`-web` calls `Serve` instead of `Run`, and `-addr` changes the listen address.
`Serve` prints the URL and blocks. The examples and their defaults:

| Example | Address |
|---------|---------|
| login | 127.0.0.1:8091 |
| web | 127.0.0.1:8110 |
| platform | 127.0.0.1:8115 |

[examples/web](../examples/web) is the web mode example: a counter, a note
field, and a reset button on one page.

## Routes

| Route | What it does |
|-------|--------------|
| `GET /` | Returns the shell HTML: the frame image with a usemap, a text form, and a backspace form. `Cache-Control: no-store`. |
| `GET /frame.png` | Returns the PNG with `Cache-Control: no-store`, or 404 `no frame` when the page has no picture. |
| `GET /click` | Reads the query floats `x` and `y` in CSS pixels. 400 `bad coordinates` when either fails to parse, 303 to `/` on success, 500 when the page returns an error. |
| `POST /type` | Reads the urlencoded form field `text`. 400 `bad form` when the body does not parse, 303 to `/`, or 500. |
| `POST /backspace` | Deletes one character in the focused text control. 303 to `/`, or 500. |

One mutex serializes the handlers, so two requests never touch the page at
once.

## The image map

`GET /` emits one `<area>` for each box with a non-empty `data-action`, and
the href is `/click?x=<center>&y=<center>` in CSS pixels. The areas are listed
in reverse box order, so an inner element's area comes first and the browser
picks it; the engine's boxes are in document order with the inner element
last.

Plain controls and plain buttons have no `data-action`, so they get no area.
`GET /click` takes any coordinates and hit-tests them against the boxes, so a
click anywhere can still focus a control or reach a click handler.

## The frame

`Serve` draws the page first when it has not been drawn yet. `GET /` reads the
current boxes and size, and `GET /frame.png` returns the latest PNG. A page
kept as a display list has no bitmap, so `Page.PNG` paints once from the
stored template source and caches the bytes until the next `Redraw`
([screen.md](screen.md)). The shell page states that the screen is the image,
not the HTML around it.

## What it drops

`Serve` has no window loop:

- No tick ([frames.md](frames.md)).
- No key events ([keys.md](keys.md)).
- No clipboard, undo, redo, select-all, delete-word, or submit route. The
  only text input routes are `/type` and `/backspace`.
- No audio context ([features.md](features.md), [window.md](window.md)).
- No file dialog, so a file input keeps the typed name ([forms.md](forms.md)).
- No crash recovery ([crash.md](crash.md)).

A click still focuses a control and `/type` still edits it, so a form takes
text. Values survive a redraw like any other page.
