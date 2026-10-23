# Spotify player redraw performance

## Result

I measured `SetData` followed by `Redraw` for all eight Spotify screens. Each
screen kept the same data and produced the same HTML on every iteration. The
table shows the median of three runs, with three redraws per run.

| Screen | Time, before → after (ms/op) | Time saved | Allocated bytes, before → after (MB/op) | Allocations, before → after | Parses, before → after |
|---|---:|---:|---:|---:|---:|
| Home | 324.38 → 30.46 | 90.6% | 76.84 → 3.54 | 47,286 → 11,163 | 1 → 0 |
| Search | 204.50 → 17.87 | 91.3% | 54.77 → 2.82 | 37,446 → 7,602 | 1 → 0 |
| Library | 147.09 → 16.93 | 88.5% | 42.91 → 2.52 | 32,506 → 6,322 | 1 → 0 |
| Liked | 218.87 → 30.04 | 86.3% | 56.62 → 3.79 | 40,062 → 10,361 | 1 → 0 |
| Browse | 144.60 → 16.12 | 88.9% | 41.86 → 2.42 | 32,888 → 6,922 | 1 → 0 |
| Radio | 220.62 → 14.89 | 93.3% | 57.81 → 2.45 | 37,350 → 5,922 | 1 → 0 |
| Queue | 207.65 → 21.26 | 89.8% | 54.05 → 3.09 | 37,275 → 8,419 | 1 → 0 |
| Profile | 128.04 → 14.66 | 88.5% | 39.13 → 2.47 | 30,280 → 5,547 | 1 → 0 |

Redraw time fell by 86.3% to 93.3%. Allocated bytes fell by 93.3% to 95.8%.
The measured run changed from one parse per redraw to none when the rendered
HTML stayed the same.

## What changed

The player embedded SVG assets and passed them to Blinkless. The SVG backgrounds
were rasterized and PNG-encoded again during layout. The baseline CPU profile
spent most of its time in PNG filtering and compression.

The player now embeds transparent PNG variants at four times the source SVG
size and registers those images with the page. The SVG files remain as the
editable source. To regenerate the PNG files with ImageMagick, run this command
from the repository root:

```sh
for src in examples/spotify-player/player/assets/*.svg; do
  convert -background none -density 384 "$src" -strip "${src%.svg}.png"
done
```

`Page.SetData` also keeps the parsed document cache. `Redraw` already checks the
executed HTML before reusing that cache, so clearing it on every `SetData` call
forced a parse and cascade even when the output did not change.

After these changes, a CPU profile of the home redraw placed the largest cost in
Blinkless CSS selector matching. PNG filtering and compression no longer led
the profile. A redraw that changes the template output still needs layout work.

## Reproduce the measurement

The benchmark lives in `player/render_bench_test.go`. Run it from the repository
root:

```sh
go test -p 1 -run '^$' -bench '^BenchmarkScreenSetDataRedraw$' -benchmem -benchtime=3x -count=3 ./examples/spotify-player/player
```

The before run used the same Go sources with `SetData` cache invalidation and
SVG asset loading restored. The after run used the current PNG assets and cache
reuse. Both runs used the same benchmark command.

The measurements came from Go 1.26.4 on Linux amd64 with an Intel Core
i7-13700HX. They time Go template execution, page rendering, and display-list
creation. They do not measure window presentation, audio playback, network
requests, or frame rate. `Page.Probe().PNG` is not called by this benchmark.
