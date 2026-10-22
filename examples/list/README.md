# list

Fixed-height row window math for ownframe lists.

ownframe has no virtual DOM. Slice a window of rows, put that slice in the
template, and pad with spacers so the document still scrolls to the full
height. `list.Window` returns the clamp: start index, count, top spacer,
bottom spacer.

```go
w := list.Window(offsetY, viewH, gridY, rowH, totalRows, overscan)
view.Rows = all[w.Start : w.Start+w.Count]
view.Top = w.Top
view.Bottom = w.Bottom
```

Call `Page.SetWindowing(true)` and `Page.SetScrollWindow` so the host redraws
only when the window moved. Return false from the callback while overscan
still covers the viewport.

CSS `position: sticky` and `position: fixed` do not pin under window scroll.
blinkless does not mark sticky as fixed. Pin chrome with
`Page.SetScrollWindow` and `Page.SetViewportPinZ`. Telegram is the worked
example.

`examples/perf-complex` is the full 480-row fixture.
`examples/ops-console` is a small ticket list, not a windowed virtualizer.

Honest numbers, recorded on the perf-complex machine: 48 of 480 rows redraw
in about 35 ms, not 16 ms. Teams 500 posts unwindowed is about 7.3 s.

The layout engine pin is `github.com/chinmay-sawant/blinkless` at the
pseudo-version in the repo `go.mod`. There is no tagged blinkless release yet.
