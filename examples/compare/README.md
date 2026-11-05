# Compare table

Same thousand-row styled table in two toolkits: ownframe with the UI
builder, and MyGo native UI from `github.com/egoist/mygo`.

Both show a header, a dark column header row (Num, Title, Description),
and 1000 rows with alternating white and gray backgrounds. Clicking a row
selects it, tints it blue, and reports it in the header readout.

Apples-to-apples form on both sides: each app renders all 1000 rows and
rebuilds the whole view on every click, so this measures the default
full-rebuild path of each toolkit with the generic library fixes in place.
(A best-practice round with ownframe row windowing and the mygo
virtualized `List` was measured earlier; the numbers are in the session
notes, not in this file.)

- `ownframe/` is a normal example in this module. It uses `ui.Mount` and
  `View.Update`, the same builder pattern as `builder/ui-builder`.
- `mygo/` is an isolated module. MyGo `v0.7.4` needs Go `>= 1.27.3` while
  this repo pins Go `1.26.10` with `GOTOOLCHAIN=local`, so it cannot share
  `examples/go.mod`. It stays out of `make build`, `make test`, and
  `make lint` until the repo moves to a newer Go.

Both open a native GUI window. Neither uses a webview: the mygo table is a
Go view drawn on the GPU, and the ownframe table is builder HTML laid out
by blinkless and replayed on the Ebiten canvas. The ownframe app also
serves the same picture in a browser with `-web`.

## Run

```sh
go run ./examples/compare/ownframe
go run ./examples/compare/ownframe -web
go run ./examples/compare/ownframe -dev
go run ./examples/compare/ownframe -dev -perf
```

MyGo needs Go 1.27.3 or newer and runs from its own module:

```sh
cd examples/compare/mygo && GOWORK=off go run .
```

## Benchmarks

Headless, no display needed. The two harnesses measure different layers, so
compare trends within each app, not raw numbers across them. One toolkit
difference is real: the mygo tester only clicks rows inside the window, so
its click benchmark uses a visible row, while the ownframe dispatcher
accepts any row by action id.

```sh
go test ./examples/compare/ownframe/ -run '^$' -bench . -benchmem
cd examples/compare/mygo && GOWORK=off go test -run '^$' -bench . -benchmem
```

`BenchmarkListRedraw` times a warm ownframe redraw with unchanged data.
`BenchmarkListUpdate` times rebuilding the small windowed view with
unchanged state. `BenchmarkListClickUpdate` times the full data-update
path: alternating row clicks through the page dispatcher, select, rebuild
(no-op clicks skip the rebuild).
`BenchmarkListTester` builds the virtualized MyGo table in memory with
`ui.NewTester`. `BenchmarkListClick` times the full data-update path
there: alternating row clicks, select, reframe.
