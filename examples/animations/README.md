# CSS animation showcase

A gallery of 33 animation cards. Every card's motion is a CSS declaration in
`animations.html`: the `@keyframes` rules and the `animation` and
`-webkit-animation` declarations live there, while `page.go` holds the play
switch and one click handler. The page registers no tick callback, builds no
player, and imports no `internal` package. OwnFrame's automatic CSS driver binds
each target at the initial `Redraw` and samples it on every window tick
([css-motion.md](../../documentation/css-motion.md)).

## Sections

| Section | Cards |
|---|---|
| Timing | `ease`, `linear`, `cubic-bezier(0.68, -0.55, 0.27, 1.55)`, `steps(5, jump-end)` |
| Direction | `normal`, `reverse`, `alternate`, `alternate-reverse` |
| Delay and fill | positive and negative `animation-delay`, then `animation-fill-mode: none`, `forwards`, and `both` |
| Iteration | one cycle, three cycles, and `infinite` |
| 2D transforms | `translateX`, `rotate`, `scale`, `skewX`, and `matrix` |
| 3D transforms | `rotateX`, `rotateY`, `translateZ`, and `matrix3d` |
| Perspective | `perspective` and `perspective-origin` |
| Paint | `opacity`, `background-color`, `color`, and `border-color` |
| Path and alias | a `corner-shape: bevel` fill and `@-webkit-keyframes` reached through the `-webkit-animation` longhands |

Every card has one stage, one animated target, the property name, the
declaration that drives it, and an **Expect** line that says what the eye
should see. The 2D cards move the operation matrix. The 3D and
perspective cards go through the engine's re-projection pass, so
`Config.RetainState` is on from the first layout. The paint cards write only the
operations their element owns, which is why the background card can animate its
fill while its label keeps its text colour.

## Pause and play

The header button (`#toggle-play`) flips `ViewState.IsPlaying` and publishes the
`is-paused` class on the body. The stylesheet turns that class into
`animation-play-state: paused` for every target, so pausing is CSS too. A paused
player stops asking for frames. Pressing play again continues from the frozen
phase instead of restarting the keyframes. There is no reset or per-card control.

A redraw paints the declared base values. The window runs the motion tick before
it draws, and that tick writes the carried pose back, so the paused card stays
where it stopped.

## Running the example

Run as a native desktop window:

```sh
go run ./examples/animations
```

The developer tools dock is open by default. It starts on the Frame tab, and
the Elements and Ops tabs are available beside it. The Frame tab records the
pipeline and frame timings by default, so the PERFORMANCE, PIPELINE DETAIL, and
MEMORY rows report numbers instead of `-`. Pass `-performance=false` to measure
the gallery without that tracking. Close the dock with:

```sh
go run ./examples/animations -dev=false
```

A virtual display or remote session whose compositor holds a large window below
its refresh rate can cap the frame rate even when the page is idle. Turn vsync
off on that kind of host:

```sh
go run ./examples/animations -vsync=false
```

Leave vsync on for a normal desktop, where it prevents tearing. `-width` and
`-height` set the window size and `-monitor` chooses the display index:

```sh
go run ./examples/animations -width=1920 -height=1080 -monitor=1
```

To measure without watching the dock, run a fixed window of time. The summary
prints the frame rate, the page pipeline, the motion counters, and memory, then
closes the window:

```sh
go run ./examples/animations -profile-seconds=30 -vsync=false
```

The samplers wait for the first drawn frame and then for a two second warmup,
so the window's opening does not land in the samples.

The tall gallery warms its background caches in narrow bands over successive
frames. Scrolling beyond the viewport buffer reuses the overlapping pixels and
paints the newly exposed strips. Large viewport textures stay outside the
sprite atlas.

Serve the picture page in a browser:

```sh
go run ./examples/animations -web
```

`Serve` does not tick, so `-web` shows the committed frame rather than moving
cards. Use the desktop window or the WebAssembly canvas build to watch the
motion. Optionally specify a custom listen address with `-addr`:

```sh
go run ./examples/animations -web -addr=127.0.0.1:8142
```

## Running tests

```sh
go test -v -p 1 ./examples/animations/...
```

The tests assert that every card declares an animation and binds with no tick
callback, that the translate card advances its matrix across two ticks, that the
rotateY card restamps and changes projection on both ticks, that the background
keyframe leaves the label colour alone, that the corner path recolours the fill
and leaves the border, and that pause holds the pose while resume continues.

To measure native scrolling with VSync off, run this command from the repository
root. It opens a 1908 by 999 window, scrolls to the bottom and back, then repeats
the downward pass:

```sh
OWNFRAME_ANIMATIONS_GPU_TEST=1 go test -p 1 -count=1 -run '^TestNativeScrollProfile$' ./examples/animations
```

The test writes `frames.json` and `bottom.png` under
`temp/animations-native-profile/`. Frame gaps include GPU submission and
presentation waits between Draw calls. Work timings cover Update and Draw.
Memory samples include the running heap and frame rate once per simulated
second; the final heap value is recorded after garbage collection with the
game and its caches still retained. These values measure this native session,
not the browser or Android hosts.
