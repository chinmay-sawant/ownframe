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

Framework paths, without invoice detail. Few element updates use the
default patch path with refresh fallback. A batched patch applies N edits
with one shared style re-resolve and one draw. Windowing lays out the
visible viewport plus spacers, so build cost scales with viewport, not
collection size. Global restyles use a stylesheet swap with one cascade,
layout, and paint. Per element loops over full surfaces stay forbidden;
the 1000 patch row below is the regression tripwire for that rule.
Batched and windowed rows in Results are measured by collation.

A refresh with no data change skips the render. The harness calls refresh
twice with nothing changed between the calls, so the second call has no
rebuild work.

Four more paths round out the comparison. Cold build is a fresh app plus
its first headless frame, the proxy for first paint (toolkit-level caches
persist in-process, so treat it as a cold page on a warm engine). Resize
alternates two window sizes with a redraw each. Typing is one keystroke
into the filter box plus the rebuild it triggers, alternating a typed char
with backspace so every iteration does constant work. Heap growth runs ten
full rebuilds, collects, and reports retained bytes plus the growth delta.

## Results (i7-13700HX, headless, `go test -bench`)

| Path | ownframe | MyGo | Gap |
| --- | --- | --- | --- |
| Full build, changed data | 225.8 ms/op · 4.4 fps · 96.4 MB/op · 292k allocs/op | 28.9 ms/op · 34.6 fps · 40.9 MB/op · 32.7k allocs/op | MyGo ~7.8x faster |
| Interactive click + rebuild | 225.8 ms/op · 4.4 fps | 16.6 ms/op · 60 fps · 0.1 MB/op · 6.1k allocs/op | MyGo ~14x faster |
| Interactive click + batch | 30.3 ms/op · 33 fps · 41.7 MB/op · 34k allocs/op | (same as above) 16.6 ms/op · 60 fps | Batch 7.5x better than rebuild, MyGo ~1.8x ahead |
| Rebuild, same data | skip floor, 0 allocs | 6.9 ms/op · 145 fps | ownframe skips, MyGo reframes |
| Cold build, first paint | 243.8 ms/op · 4.1 fps · 91.5 MB/op · 297k allocs/op | 25.5 ms/op · 39 fps · 41.3 MB/op · 39k allocs/op | MyGo ~9.6x faster |
| Resize + redraw | 247.6 ms/op · 2 resizes | 8.4 ms/op · 59.8 fps · 74 KB/op · 4.1k allocs/op | MyGo ~30x faster |
| Typing, one keystroke | 180.8 ms/op · 5.5 fps · 69.5 MB/op · 223k allocs/op | 16.5 ms/op · 60.6 fps · 8.6 MB/op · 10k allocs/op | MyGo ~11x faster |
| Heap, 10 full rebuilds | retained 1.41 GB · growth +5.8 MB · +38k objects | retained 4.5 MB · growth +6.4 KB · +17 objects | ownframe retains ~3000x, grows ~900x per rebuild |
| Theme toggle + rebuild | 224.1 ms/op · 4.5 fps | n/a | Baseline for global changes |
| Theme toggle + 1000 patches | 14432 ms/op · 0.1 fps · 13.0 GB/op · 13M allocs/op | n/a | Patching loses 64x: never patch row by row |
| Theme toggle + stylesheet swap | 214.3 ms/op · 4.7 fps | n/a | Same as rebuild minus the HTML render |
| Batched patch, few edits | 30.3 ms/op · 33 fps · 41.7 MB/op · 34k allocs/op | n/a | One shared re-resolve, one draw |
| Windowed build, visible viewport | skip floor after first windowed build, 0 allocs | n/a | Cost scales with viewport, not collection size |
| Clean refresh, no change | 0.2 ms/op · 1.6 KB/op · 1.6k allocs/op | n/a | Second of two back-to-back refresh calls skips the render |
| Idle re-render | 0.0009 ms/op · 0 allocs | 6.9 ms/op (always reframes) | ownframe ~7400x cheaper |

Stage split inside a full ownframe rebuild (`-bench` harness, Perf on):
template 0 ms, style 7.9 ms, layout 256.7 ms, paint 0 ms (headless). The
layout and display-list build over 9021 boxes is the cost, not the HTML
render (~9 ms, ~4%) and not the parse.

Takeaway: patch touches, don't rebuild them. A selection is two class
edits plus a text edit and runs 7.5x faster than a rebuild. But each
`PatchElement` re-resolves styles for the whole document, so a thousand
patches cost a thousand re-resolves: global changes (theme, and later
charts or material swatches) belong in a stylesheet swap or a rebuild,
never in a per-row patch loop. That is the rule to carry into new
components: fine-grained edits for small touches, one restyle for
full-surface changes. The new cold, resize, typing, and heap rows show
where the next round goes: first paint and resize are full builds, typing
costs a full rebuild per keystroke until the filter path learns patches,
and retained heap grows ~0.6 MB per rebuild against MyGo's near zero.

Data parity is checked by construction (same seed-1 generator in both
modules) and both harnesses print `grand-total=$1420529.43`.

