# Meridian — trading terminal

A dark desktop trading terminal built with ownframe. One HTML template and one
generated dataset drive seven screens: watchlist, markets, portfolio, orders,
screener, news, and settings. Everything is synthetic: there is no network
feed, no persistence, no account, and no live order routing.

The developer tools dock opens by default so the Frame, Elements, and Ops tabs
can be inspected while the page is running.

## Run

From the repository root:

```sh
go run ./examples/trade
```

The default window is 1600 × 900 and the minimum is 1180 × 700. The page is
authored with `100vw × 100vh`; it follows the window instead of locking to a
fixed canvas. The watchlist, quote, and centre columns are percentage widths,
and the panels use fluid rows with compact rules for short windows.

Close the dock:

```sh
go run ./examples/trade -dev=false
```

Resize the window at any time; the shell and panels follow the new size. `-perf`
is on by default and feeds the pipeline sampling used by the Frame tab; the
CSS motion runs either way. Pass `-perf=false` for a still page profile.

Serve the current frame in a browser:

```sh
go run ./examples/trade -web
```

The address is `http://127.0.0.1:8144`. `Serve` does not tick the page, so a
browser shows the committed frame rather than the live animation. Use the
desktop window to watch the tape, pulses, scan lines, and entrance motion.

## Screens

| Rail item | What it shows |
|---|---|
| Watchlist | Three-column reference layout: watchlist, candlestick chart with chart tabs, order blotter, and the quote/order-book/time-and-sales rail. |
| Markets | Index cards, breadth meters, an 11-tile sector heatmap, movers, flows, session map, and economic calendar. |
| Portfolio | Account value and P&L, performance line against a benchmark, allocation, holdings, account stats, and activity. |
| Orders | Order blotter with tabs and filters, plus a working paper order ticket and recent fills. |
| Screener | Preset strip, a three-card filter builder, and a scored result table with sparkbars. |
| News | Source filters, lead story and feed, trending tickers, sentiment bars, and a calendar. |
| Settings | Trading, appearance, data, and advanced preference groups, keyboard map, session summary, and about card. |

Interactions worth trying:

- Rail items switch screens. The topbar back and forward chevrons walk that
  history.
- A watchlist row selects the instrument for the chart, quote rail, and ticket.
  The star toggles the row, and the `ALL / US / HK / EV / STARRED` chips filter.
- Chart timeframes (`1D` through `All`) rebuild the candle window. The
  `Chart / Options / News / Financial` tabs switch the panel under the quote.
- The blotter tabs switch Today, Past, and Holdings; the filter selects are
  read-only preview controls.
- On Orders, choose Buy/Sell, a type, quantity, and TIF, then Submit. A working
  row is added to the blotter and a toast confirms it.
- Screener presets, filter chips, and Run screen update the preview. Settings
  toggles apply for the session; the Developer tools toggle also opens or
  closes the dock.

## Data

`data.go`, `data_market.go`, and `data_page.go` build the sample market once at
startup with fixed seeds. The curated tickers and session prices are fixed;
OHLC bars, depth levels, prints, holdings, and order rows derive from a
`math/rand/v2` PCG stream. Order identifiers come from a seeded
`gofakeit.Faker`. The display clock is the only wall-clock input. The same
binary therefore draws the same market on every run.

The dataset covers 16 instruments (US and HK, plus an EV filter), 180 intraday
candles per instrument, ten depth levels, a time-and-sales tape, 18 order rows,
eight holdings, market internals, news, screener results, and settings. No
value is fetched, saved, or sent.

## HTML and CSS drawing

The candlestick plot is DOM markup, not a Go canvas and not a chart library.
Candle bodies and wicks, volume columns, the dashed crosshair, and the moving
average segments are positioned spans. The plot grid is HTML spans, the SMA
legend and axes are flex rows, and the last-price line is a styled border. The
plot and its spans scale with the window using percentages and flex rows. The same approach draws the portfolio performance line: one span per
segment over a CSS grid.

Go generates the fixture data and the data-derived positions those spans need.
It does not paint pixels, call a renderer, or start an animation. There is no
reference-image background or image asset anywhere in the example. The rail
icons are inline HTML SVG. The small transaction ring is inline SVG markup whose
arc paths are generated from the sample transaction split.

## Motion

Every movement is a CSS declaration. The example registers no `SetTick`, builds
no player, and imports no package under `internal/`. `Redraw` binds the CSS
players once; the window tick samples them.

`components/motion.css` is the shared catalog:

- `trade-pulse` and `trade-halo` for live dots
- `trade-blink` with `steps(2, jump-end)` for the live tape marker
- `trade-marquee` for the status-bar tape, plus `-webkit-animation` through
  `.anim-marquee-webkit`
- `trade-rise` and `trade-slide-right` with cubic-bezier timing and `both` fill
- `trade-sheen` for the run-screen and bid-side sweep
- `trade-eq` for the rail equalizer, alternating with negative phase delays
- `trade-tilt` for the order-ticket entrance, using perspective and rotateX
- `trade-scan` and `trade-scan-down` for chart and table scan lines
- `trade-grow-x` / `trade-grow-y` for breadth, allocation, and sparkbars
- `trade-flash-up` / `trade-flash-down` for the best book levels
- `trade-swing`, `trade-float`, `trade-spin`, `trade-breathe`, `trade-pop`, and
  `trade-shift-up` for small accents

Only transform, opacity, colour, and the supported 3D perspective path are
written. Layout, filters, shadows, and SVG internals are not animated. See
[`../../documentation/css-motion.md`](../../documentation/css-motion.md) for the
runtime's shipping limits.

`Page.Window().SetAnimations(false)` and `SetTransitions(false)` turn the two
drivers off page-wide if a profile run needs a still page.

## Tests

```sh
go test -p 1 ./examples/trade/...
```

The tests cover the screen walk, CSS binding with no motion error, watchlist
filtering and selection, chart tabs and timeframes, the simulated order path,
rail history, and a settings toggle.

To capture the page as PNGs for a visual check:

```sh
OWNFRAME_TRADE_SCREENSHOT=1 go test -run '^TestScreenshot$' ./examples/trade
```

The screenshots are written to `/tmp/opencode/trade-shots/`; nothing is added
to the repository.

For a fast responsive layout check:

```sh
OWNFRAME_TRADE_RESPONSIVE=1 go test -run '^TestResponsiveScreens$' ./examples/trade
```

This renders the watchlist at 1600x900, 2560x1440, and 1280x720, then the
other six screens at 1280x720. It asserts that the canvas, the status bar,
and every screen root stay inside the requested window. It does not
rasterize PNGs, and it turns CSS motion off for the sweep, so it is a layout
check rather than a visual diff. The four-size matrix also covers
1920x1080 in full mode.

The full seven-screen by four-size PNG matrix is still available. It is slow
because every capture paints the whole page in software:

```sh
OWNFRAME_TRADE_RESPONSIVE=1 OWNFRAME_TRADE_RESPONSIVE_FULL=1   go test -run '^TestResponsiveScreens$' ./examples/trade
```

## Limits

- Desktop only for the full experience. `-web` serves a still picture because
  `Serve` does not tick. There is no Android project or mobile binding.
- All values are fixture data. Order submission mutates the in-memory sample
  only. No real market data, brokerage, authentication, or persistence ships
  here.
- A very small window trades detail for fit instead of reflowing every panel.
- The data is deterministic, so a refresh does not invent a new market. The
  refresh control confirms the local feed and leaves the seeded snapshot in
  place.
