# Running and profiling the CSS preview

## Display and window size

The desktop exposes two X monitors on `DISPLAY=:0`: `XWAYLAND0` at 2560x1600
and `XWAYLAND1` at 1920x1080, placed at `+0+320`. The preview runs on the
secondary 1920x1080 monitor.

The window is 1908x999, which is the 1920x1080 monitor minus the title bar and
the frame border. The example takes the size from flags:

```sh
CSS_PREVIEW_WALK=1 CSS_PREVIEW_DEVTOOLS=1 \
  go run ./examples/css-preview -perf -width 1908 -height 999
```

- `-perf` turns on perf sampling and opens the devtools dock.
- `CSS_PREVIEW_WALK=1` enables the walk hook. It serves `127.0.0.1:6971` with
  `/goto`, `/aim`, `/cursor`, `/state`, `/reset`, and `/quit`.
- `CSS_PREVIEW_DEVTOOLS=1` keeps the dock open while the walk runs. Without it
  the walk hook closes the dock so both catalog columns stay clickable.
- F12 and Ctrl+Shift+I toggle the dock by hand.

## Click script

`scripts/css-preview-click-test.py` drives the real window. It moves the X11
pointer and sends real button press and release events, so every click goes
through the window pointer path.

```sh
python3 scripts/css-preview-click-test.py --runs 1 --no-close
```

- It calibrates the window origin from `/cursor`, scrolls each Start button
  under a fixed row with `/goto`, and reads the app's started timestamp and the
  next layout counter from `/state`.
- The timer starts at the press, so the 80 ms hover settle and the 50 ms hold
  stay outside the measured interval.
- `--no-close` keeps the window open for inspection. Without it the script
  closes the app with `/quit`.

## Measured numbers

One 43-click pass at 1908x999 with the devtools dock open:

| Metric | Before the fifth pass | Now |
|---|---:|---:|
| redraw p50 | 19.3 ms | 0.0 ms |
| redraw p95 | 39.6 ms | 0.0 ms |
| redraw max | 49.5 ms | 0.0 ms |
| redraw total | 0.85 s | 0 s |
| press to started p50 | 44.3 ms | 15.8 ms |
| clicks | 42/43 | 43/43 |
| Alloc/frame | 1.7 MB (per 500 ms, mislabelled) | 123.7 KB |
| Go heap | 350 MB | 56 MB |
| RSS | about 1000 MB | 261 MB |

The click patches the card it changed instead of rebuilding the page, so the
page reports no redraw for any of the 43 clicks and the redraw rows read zero.
The walk still waits for each painted frame, and that wait is the visible
latency. The canonical numbers for the click path and the latency breakdown
are in `documentation/optimizations.md`. The memory rows are from the earlier
run, before the element patch.

The dock takes the right 340 px of the window, so the page viewport reads
1568x999 in the Frame tab while the window reads 1908x999.

## Play all 43 previews

The header carries a `Play all 43 previews` button. One press starts every
preview through the same path its own Start button uses. The catalog stays
windowed: the rows on screen start now, and the rows off screen wait for
their first render, which gives the transition its starting paint. The
animations, the transitions, and the triggered samples all run together, and a
Start button still replays its own preview afterwards.

With all 43 running at 1908x999, dock open, measured when the press still
rendered the whole catalog:

| Metric | Value |
|---|---:|
| FPS | 60.5 |
| Frame | 17.3 ms |
| Draw | 0.4 ms |
| Alloc/frame | 121 KB |
| Go heap | 59.6 MB |
| RSS | 227 to 246 MB |
| Page redraw for the press | 7.6 ms |
| Click in that mode | 21 ms p50 |

The catalog is windowed now, so the press paints only the rows on screen and a
click keeps the windowed cost. The press stamps the shared start time after it
patched the cards, so a press that patches every card cannot leave a sample
finished before the first tick reaches it. The measured press cost and the
edge cases are in `documentation/optimizations.md`, in the section on the Play
All click cost.

## Scrolling

A long vertical scroll presents about 37 frames per second in a 1908x999
window on this WSL2 host. The Frame tab shows the reason: `Draw` is 0.1 ms and
the page redraws about eight times in six seconds, one per row the window
crosses. The rest of the frame is the WSLg software compositor uploading the
1908x999 surface, which is the same cost at any page size. A smaller window or
a GPU compositor is what holds 60 frames per second here; the page work is no
longer part of the frame.
