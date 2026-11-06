# Invoice benchmark: testing/ownframe vs testing/mygo

Same thousand-row invoice table in two toolkits: ownframe with the UI
builder, and MyGo native UI from `github.com/egoist/mygo`.

Both show a header (title, subtitle, selection, theme, grand total), a theme
toggle, a dark column header row (Num, Invoice, Client, Date, Item, Qty,
Total, Status), and 1000 deterministic random rows (seed 1, so both sides
render identical data). Clicking a row selects it, tints it blue, and reports
it in the header readout.

Alternative row theme: `gray` (default) stripes odd rows `#f3f4f6` against
white; `light` renders every row white. The toggle rebuilds the whole view,
so it is a second full-rebuild benchmark next to row selection.

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
go run ./examples/testing/ownframe
go run ./examples/testing/ownframe -web
```

MyGo needs Go 1.27.3 or newer and runs from its own module:

```sh
cd examples/testing/mygo && GOWORK=off go run .
```

## Benchmarks

Headless, no display. Each harness measures wall time per op (headless, so
wall time is CPU time per frame), derived FPS (`1e9 / nsPerOp`), memory
(`B/op`, `allocs/op`, retained heap), and frame sizes. The ownframe harness
additionally prints the pipeline stage split (template, style, layout,
paint, other) and the HTML/box/op counts, which is where the MyGo gap
comes from.

```sh
go run ./examples/testing/ownframe -bench
go test ./examples/testing/ownframe/ -run '^$' -bench . -benchmem
cd examples/testing/mygo && GOWORK=off go run . -bench
cd examples/testing/mygo && GOWORK=off go test -run '^$' -bench . -benchmem
```

`redraw-same-data` (ownframe) is the cache-hit floor: the page skips layout
on an identical frame. `rebuild-same-data` / `build-full-view` is the
apples-to-apples full build both toolkits pay per data change.
`click-select-rebuild` / `click-select-reframe` is the interactive path:
select a row and rebuild the whole view.

`click-select-patch` is the same click through the fine-grained path: two
`PatchElement` class edits plus a readout text edit, no reparse, no
recascade. `theme-toggle-patch` is the same idea scaled to all 1000 rows in
one redraw batch. `theme-toggle-style` flips the theme by swapping the row
stylesheet (`SetTheme` + `Redraw`): one cascade, layout, and paint, no
builder HTML render. MyGo has no patch path to add: its immediate-mode
reframe already is the fine-grained path, so it is the comparison target.

## Results (i7-13700HX, headless, `go test -bench`)

| Path | ownframe | MyGo | Gap |
| --- | --- | --- | --- |
| Full build, changed data | 304.1 ms/op · 3.3 fps · 45.9 MB/op · 316k allocs/op | 31.6 ms/op · 31.7 fps · 40.6 MB/op · 25k allocs/op | MyGo ~9.7x faster |
| Interactive click + rebuild | 304.1 ms/op · 3.3 fps | 31.0 ms/op · 32.3 fps · 56 KB/op · 6k allocs/op | MyGo ~10x faster |
| Interactive click + patch | 55.7 ms/op · 18 fps · 42.5 MB/op · 34k allocs/op | (same as above) 31.0 ms/op · 32 fps | Patch 5.5x better than rebuild, MyGo still ~1.8x ahead |
| Rebuild, same data | 15.5 ms/op · 64 fps | 12.3 ms/op · 81 fps | Tie (different work) |
| Theme toggle + rebuild | 265.1 ms/op · 3.8 fps | n/a | Baseline for global changes |
| Theme toggle + 1000 patches | 19650 ms/op · 0.05 fps · 12.9 GB/op · 13M allocs/op | n/a | Patching loses 74x: never patch row by row |
| Theme toggle + stylesheet swap | 264.8 ms/op · 3.8 fps | n/a | Same as rebuild minus the HTML render |
| Idle re-render | 0.014 ms/op · 0 allocs | 12.3 ms/op (always reframes) | ownframe ~850x cheaper |

Stage split inside a full ownframe rebuild (`-bench` harness, Perf on):
template 0 ms, style 7.9 ms, layout 256.7 ms, paint 0 ms (headless). The
layout and display-list build over 9021 boxes is the cost, not the HTML
render (~9 ms, ~4%) and not the parse.

Takeaway: patch touches, don't rebuild them. A selection is two class
edits plus a text edit and runs 5.5x faster than a rebuild. But each
`PatchElement` re-resolves styles for the whole document, so a thousand
patches cost a thousand re-resolves: global changes (theme, and later
charts or material swatches) belong in a stylesheet swap or a rebuild,
never in a per-row patch loop. That is the rule to carry into new
components: fine-grained edits for small touches, one restyle for
full-surface changes.

Data parity is checked by construction (same seed-1 generator in both
modules) and both harnesses print `grand-total=$1420529.43`.

