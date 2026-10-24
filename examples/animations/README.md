# CSS animation showcase

A single page whose motion is declared entirely in CSS. `animations.html` holds
the `@keyframes` rules and the `animation` and `-webkit-animation` declarations;
`page.go` holds view state and click handlers only. The page registers no tick
callback, builds no player, and imports no `internal` package. OwnFrame's
automatic CSS driver binds each chip at the initial `Redraw` and samples it on
every window tick ([css-motion.md](../../documentation/css-motion.md)).

## What it does

The page presents six question cards with live chip stages, plus the 27
animation properties from the Blinkless compatibility matrix as a reference
table:

1. **Timing functions and curves**: `ease-in-out`, `linear`,
   `cubic-bezier(0.68, -0.55, 0.27, 1.55)`, and `steps(5, jump-end)`.
2. **Direction and cycles**: `animation-direction` and
   `animation-iteration-count`.
3. **Delays and offsets**: positive, zero, and negative start delays.
4. **Fill modes**: `none`, `forwards`, and `both`.
5. **WebKit prefixes**: the vendor aliases normalize to the standard properties.
6. **Advanced declarations**: `animation-composition`, `animation-timeline`,
   `animation-range`, and `animation-trigger`.

Answering a question, changing a filter, resetting the quiz, or pressing a
track's Play button updates the view and calls `Redraw`. The automatic driver
rebinds on that commit and keeps each chip's original start, so the keyframes
keep their phase instead of restarting. A real replay needs `animation-trigger`,
which the automatic driver does not dispatch.

Pause and resume are also CSS. The button toggles a body class whose rule sets
`animation-play-state: paused`; the paused player stops requesting frames, and
removing the class lets the kept start sample again.

The chips exercise a 2D `transform` (the matrix path), `opacity`, and colour
keyframes that write only the element's own fill, so the label keeps its text
colour. The disabled driver, the 3D re-projection path, and the fidelity-gap
list are covered in [css-motion.md](../../documentation/css-motion.md).

## Running the example

Run as a native desktop window:

```sh
go run ./examples/animations
```

Serve the picture page in a browser:

```sh
go run ./examples/animations -web
```

`Serve` does not tick, so `-web` shows the committed frame rather than moving
chips. Use the desktop window or the WebAssembly canvas build to watch the
motion. Optionally specify a custom listen address with `-addr`:

```sh
go run ./examples/animations -web -addr=127.0.0.1:8142
```

## Running tests

```sh
go test -v -p 1 ./examples/animations/...
```

The tests assert that the page binds its chips without a tick callback, that two
`Tick` calls advance a chip's matrix, and that a background keyframe leaves the
label's text colour alone.
