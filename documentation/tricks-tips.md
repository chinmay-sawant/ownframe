# Tricks and tips

Small patterns that keep apps fast. Each one names the problem, the fix, and where to see it.

## Infinite scroll with a local database

Problem: a channel has no end. Holding every post in memory makes each redraw slower until the app stalls. The Teams channel benchmark showed this directly: 10 posts redrew in 190 ms, 50 in 720 ms, 200 in 2.8 s, 500 in 7.3 s. Boxes and operations grow per post, about 15 boxes and 27 ops each here.

Fix: use three layers, each smaller than the last.

1. Remote database is the truth and keeps everything.
2. Local SQLite is a short cache. Create it at login, keep about one day, prune by time. A scroll past the cache fetches from remote and writes to SQLite first. Never render from the network directly.
3. Memory holds a working set, for example 100 posts around the viewport.
4. The screen renders a smaller window, for example 50 posts, with top and bottom spacers so the document keeps its full scroll height.

Scrolling up past the window slides the window. Scrolling past the cache triggers one paged fetch keyed by time or id, then the window slides. Cache expiry only drops the local copy. It never deletes server history.

The row window pattern lives in `examples/perf-complex/windowing.go`. It keeps 48 of 480 rows, reuses the window while overscan covers the viewport, and returns false when no rebuild is needed. That fixture went from 306 ms to 35 ms per redraw. Teams posts vary in height because threads expand, so anchor the window by post id and measured box position instead of fixed row math. A latest-50 view with a load-older control is a fine first step before full scroll windowing.

See `examples/teams/channelbench` for the synthetic post benchmark, `examples/teams/channels` for the posts template, and `examples/teams/store` for the SQLite layer.
