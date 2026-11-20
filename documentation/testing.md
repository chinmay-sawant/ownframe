# Validate a change

Run these checks from the repository root with the Go version in `go.mod`.

```sh
make test
make lint
make graphics-check
```

`make test` runs the library and every example module package. Tests sit beside the packages they protect. Example tests remain compatibility checks even when an example is classified as experimental. `make test-race` checks the library only.

## Documentation and process checks

`make docs-check` extracts Go fences from the README and every guide, plus the website quick-start snippets. It compiles them with `go vet -p 1` against the checkout. It does not run the snippets or open their windows.

[snippets.json](snippets.json) supplies imports and surrounding parameters for fragments. The Go inside the fence is compiled directly, rather than copied into a second example. A new fence or removed fence requires a manifest update. Third-party framework comparisons and optional keyring code are listed as external illustrations with reasons; the checker reports them. The internal replay fragment is explicitly marked as implementation code. App-facing snippets cannot import `ownframe/internal/`.

`make plans-check` checks the distinct states and evidence fields in [the current ledger](../plans/README.md). Its rejection tests cover ambiguous states, duplicate identifiers, missing evidence, and disabled Python assertions. The checks use explicit validation, so `PYTHONOPTIMIZE=1` cannot bypass them. It cannot establish whether prose evidence is true; reviewers must inspect the referenced result. Superseded production-readiness checklists are not proof of completion.

`make examples-check` requires every example root to have a maintenance classification in [maintenance.json](../examples/maintenance.json). The maintained teaching set is login, bind, forms, list, desktop-cat, and telegram. Other examples remain in the compatibility suite. Classification is a scope decision, not a claim that Android or desktop behavior was validated on every platform.

`make lint` prints one result per gate and includes complete diagnostics for any failed gate. Use `make lint VERBOSE=1` to see successful tool output, including each excluded third-party illustration. Every gate still runs after a failure. `make tools-check` tests the output runner and the Go consolidation tool. The output regression tests cover success suppression, failure diagnostics, missing tools, and verbose output.

`make lint` runs the exact golangci-lint version declared by `GOLANGCI_LINT_VERSION`, both Go module lint passes, the documentation and ledger checks, the example import gate, and website ESLint. An installed runner with a different version is replaced for this invocation by a versioned local tool under ignored `temp/tools/`. It does not replace the user's PATH binary.

## Software pictures and window pixels

`TestPagePictureMatchesRedraw` compares software PNGs after interactions with a fresh redraw and requires the interaction to change the picture. It protects retained page state and software-paint consistency. It does not capture the window's tile cache or prove browser parity. `TestRetainedPaintInvalidatesEncodedPNG` checks decoded pixels after operation patches, display patches, tick mutations, and explicit invalidation, without requiring a redraw. A separate no-op test verifies cache reuse.

`make graphics-check` runs the following tests in separate processes and separate Xvfb sessions:

- `TestPartialFrameLeavesTheScreen` checks damage-only blits, full rebuilds, scroll copies, and hiding then restoring content under the devtools dock.
- `TestWindowRepaintPixels` uses a real page and compares the retained window framebuffer with a fresh full replay across a color change crossing tile boundaries, element removal, scrolling with a fixed header, a clean frame, and an edit while scrolled. It requires the color change to take the partial repaint path.
- `TestWindowBackdropFilterReplaysFullFrame` checks that a backdrop-filter blur matches a full replay and does not allocate the tile grid.
- `TestStickyViewportPixelsAndHits` checks viewport activation, containing-block release, scroll return, overlap paint/hit priority, edits while scrolled, and zoom. Its canonical HTML fixture also runs in Chrome through `node scripts/check-sticky-reference.mjs`; `scripts/compare-sticky-captures.py` requires equal RGBA pixels for eight unzoomed frames. Set `OWNFRAME_STICKY_CAPTURE_DIR` to retain native PNGs.
- `TestManagedInteractionsNative` drives shared tabs, slider press/drag, and modal focus/Escape on the native update loop; set `OWNFRAME_INTERACTION_CAPTURE_DIR` to retain its five framebuffers.
- `TestViewportScrollPixels` compares the viewport cache with a fresh replay across vertical and diagonal scrolling, edits during a buffer move, and jumps with no overlap. It also requires an overlapping scroll to repaint less than a full viewport.
- `TestViewportAnimatedReturnPixels` and `TestViewportDirectAnimationReturnPixels` check that a card moving while offscreen leaves no old pose when scrolled back into view, with and without retained layers.
- `TestWindowStaticEvents` requires zero OwnFrame draws during a quiet interval, then wakes the real window through a background frame request and checks its changed pixels.
- `TestWindowIdleDeadline` runs the actual pacing wrapper and timer. It requires the blocking idle FPS mode, a quiet interval with a declared deadline, a deadline-driven tick that paints blue, and restoration of the original tick rate. A watchdog wakes a stuck loop only to fail the test. Keep the test window free of input during this check.
- On Linux, `TestWindowAccessibilityActionWake` submits an accessibility action
  while the real window loop is idle and checks that it paints the changed
  pixels before a later page deadline. The test also requires a completed
  action reply. Live D-Bus delivery is checked separately by
  `scripts/check-atspi-actions.py`.
- `TestWindowDevToolsIdle` compares cached dock pixels with direct rendering, requires the open dock to allow idle pacing, and checks scheduled metric refreshes without continuous drawing.
- `TestSidebarHoverPixels` moves hover between Loglens sidebar controls, the header, and the content panel with a scrolled page and an open dock, then checks the sidebar text remains painted on the native framebuffer.
- `TestDrawPaintsVerticalThumb` checks visible scrollbar pixels inside the running game.
- `TestGPUVisibleParity` checks replay pixels, shape and text caches, transformed operations, composited groups, and translucent fills, images, and grids inside Ebiten's `Draw` loop. Replay fixtures live under `internal/replay/testdata`, so a library module zip contains them.
- `TestScrollbarThumbStaysOnePiece` scrolls the jmap-mail drafts list up and down and requires the list thumb to stay one run on the retained framebuffer. The wheel handler runs after the frame sync, so the check scrolls after `Update` and reads pixels in `Draw`.

All pixel readback occurs in the running game. Replay comparison targets use standalone textures so atlas placement does not introduce different interpolation rounding between the two targets. Pixel equality remains exact. Each process has a 90-second timeout and runs with software OpenGL for repeatability. Ubuntu needs `xvfb`, `xauth`, `libgl1-mesa-dri`, and `libx11-dev`. The script fails when these prerequisites are missing. It leaves the user's `~/.Xauthority` untouched.

CI and release gates run the graphics target. Xvfb validates an X11 software graphics session; it does not establish native Wayland behavior, hardware GPU driver behavior, desktop IME correctness, or Android device behavior. Those require their own platform sessions.

To run the same pixel checks against your current Linux desktop session, use `GRAPHICS_SESSION=desktop make graphics-check`. This mode requires `DISPLAY`, opens test windows, and preserves your graphics driver selection. Under WSLg it checks the WSLg X11 path, not a native Windows host. A passing desktop run proves these rendering scenarios in that session; it does not prove hardware acceleration, input-method behavior, or other operating systems. The default remains isolated software rendering under Xvfb.

## Frame pacing ownership

`internal/window/pacing` owns idle transitions, declared frame deadlines, keep-alive scheduling, and diagnostic counters. Its tests use supplied times and the same interface as the window adapter. They run without a display or Ebiten dependency. Deadline replacement, cancellation, exactly-once wake consumption, timer sleep, and snapshot reads have explicit checks.

`idle_game.go`, `idle_frame.go`, and `idle_timer.go` retain the Ebiten mode changes, timer goroutine, and locking. The policy tests prove scheduling decisions. `TestWindowIdleDeadline` checks that these decisions reach the running host and update its pixels. It rejects extra frames during the declared quiet interval, but does not measure energy use or promise precise wake latency. Native keyboard input remains outside this check.

## State subscription lifecycle

`internal/state.Binding` owns subscription cancellation, binding generations, and changed-property queues. Seven direct tests cover rebinding, independent subscriptions, concurrent metadata access, snapshot ownership, shared-store fallback, and callbacks from a detached binding. The callback checks include detaching and rebinding the same store, where pointer equality cannot identify a stale registration. Listener cancellation uses registration identity rather than a closure's code address. Callbacks execute outside the store lock and may reenter the store. An already-dispatched callback may finish after cancellation, so the binding checks its generation before recording metadata.

`go test -race -p 1 ./internal/page ./internal/state` exercises concurrent store patches, listener registration and cancellation, binding changes, and metadata reads. Page integration tests retain coverage for rendering metadata and the public testing facade. These changes add no methods to public `ownframe.Page`.

## Review responsibility

There is no character cap. Keep related implementation and test scenarios together. Extract a package when it can own a responsibility with explicit inputs and outputs, rather than to reduce a file count. Use the cohesion questions in [CONTRIBUTING.md](../CONTRIBUTING.md).

## Distribution and website checks

`make module-check` builds a module zip with Go's archive rules, including library fixtures and excluding nested example and tooling modules. It compiles a separate public-API consumer with `GOWORK=off` and rejects replacements in the resolved library graph. Dependencies may come from the existing module cache or the public proxy; this is not an empty-cache offline test. Before an examples tag, run `cd scripts && GOWORK=off go run ./cmd/module-check -root .. -examples-tag examples/vX.Y.Z`. It rejects local filesystem replacements and checks the examples against their published library requirement.

`make site-check` runs website unit tests, rebuilds the site and rejects stale tracked output, then checks browser layouts. Refresh intentional changes with `npm --prefix frontend run build` first. CI runs this gate alongside module and graphics checks.

`make site-check` builds the production site into `frontend/dist/`, verifies the nine demo payloads, and drives the built site in Chrome. The [Pages workflow](../.github/workflows/pages.yml) builds and deploys its own artifact on `master`. Commit changes to the frontend sources and preview media.

`make release-check` packages the login candidate locally with dependency downloads disabled and verifies archive inventory, checksums, paths, executable modes, and wasm magic. It neither publishes a release nor proves OS signing. Tag CI prepares a signed and attested candidate after its gates pass; that workflow needs its own successful run before claiming signing evidence.

`make soak-check` runs a bounded page workload for ten minutes by default. See [performance.md](performance.md) for scope and limits. A web endpoint sampler and a page workload do not prove native input, Android lifecycle behavior, or window frame pacing.


## Review regression coverage

CI now shards example race tests with `scripts/test-examples-race.sh`, runs the
`ownframe_debug` web tests and a 60-second page soak, and compares repeated
parent/candidate benchmarks on the same runner. The release workflow reuses
these gates. Workflow definitions are not evidence that their native jobs have
run successfully.

The partial-frame graphics fixture has a committed 100 by 100 pixel golden in
`internal/window/testdata/partial-frame.png`. Its white background and red
12..32 pixel rectangle are independently specified; the first frame checks
every pixel, and partial frames check all red ink inside the reported damage.
Other parity comparisons still test consistency between replay paths.

Each window owns a `replay.Renderer`. Its six mutable caches cannot evict
another window's entries, and `Close` releases their GPU images. Package drawing
helpers retain a default renderer for existing internal callers and tests.

## Testing an app by driving it (click, scroll, drag, hover)

Use this when you must check that an app's screens match what they say they do: a demo app, a showcase, or any page with descriptions next to pictures. It works from a terminal with no mouse, so any agent harness can follow it. The worked example is `examples/blinkless-all`.

### Three things to compare

Every row on a screen can disagree in three places. Look at all three.

| Source | Where it comes from |
|--------|---------------------|
| **Description** | The sentence the page prints about the row |
| **Picture** | What is painted on the page |
| **Readout** | The text the host prints from its own layout of the same markup |

The readout is often laid out separately from the painted page. A bug that only the painted page has (an image the page never registered, a heading the demo never drew) passes every readout check. In `blinkless-all`, `shape-image-threshold` printed "text starts 40px" while the painted box started its text at 80px. Compare the picture to the readout, not the readout to itself.

### Level 1: drive the page headless in a Go test

This is the cheapest way to click, hover, drag and type, and it runs under `make test`. Prefer it for anything that must stay fixed.

- `ownframe.New(...)` then `page.Click(ctx, x, y)`. Find the box first with `page.Probe().Display().Boxes` (fields `ID`, `X`, `Y`, `W`, `H`) and click its center.
- `page.Probe().Display()` returns the painted display list. Assert on that, not on a readout string. `page.Probe().PNG()` returns the picture.
- `pagetest/` has box lookup and click helpers.
- `examples/blinkless-all/nav_scripted.go` defines `liveRig`, a manual-clock rig with `tap`, `press`, `dragTo`, `release`, `key`, `typ`, `pointAt` (hover), `idle(d)` (advance time without a window), `box`, `textOf`, and `writeFrame` (one PNG per step). `nav_live_test.go` shows each used in a test.
- `page_paint_test.go` in the same folder shows the painted-page pattern: build the page with `NewPage`, read `Display()`, assert on box and text positions.

When you fix a bug, write the test first and watch it fail without the fix. The `shape-image-threshold` test was checked by removing the one-line fix and confirming the test went red.

### Level 2: serve the page and read pictures over HTTP

Build the example into the scratchpad with `go build -o <scratch>/app ./examples/<name>`. Do not run a bare `go build` in an example folder: it leaves a binary there.

`-web ADDR` serves the page with no window. Routes are in [web.md](web.md). The ones that matter here:

| Action | Command |
|--------|---------|
| Picture of the viewport | `curl -o f.png http://ADDR/frame.png` |
| Click at a point | `curl -X POST -H 'Origin: http://ADDR' -d 'x=90&y=238' http://ADDR/click` |
| Scroll down | `curl -X POST -H 'Origin: http://ADDR' -d 'x=500&y=600&dx=0&dy=-12' http://ADDR/wheel` |
| Change counter | `curl http://ADDR/state` |

Notes that cost time once:

- `/click` and `/wheel` need an `Origin` header that matches the address, or they return 403.
- A negative `dy` scrolls down. One unit is about 48px. `dy=-12` moves about one screen of a 900px viewport and keeps a little overlap.
- To reset to the top, send `dy=3000` twice.
- `action=<id>` on `/click` did not switch tabs. Click the button's pixel position instead (read it from `Display().Boxes`).
- Scroll until two frames hash the same (`md5sum`), then stop. A blinking caret breaks that test on a page that has one; cap the loop instead.
- `-web` has no live session unless the app adds one. `blinkless-all` does (`web_live.go`): `/hover`, `/click`, `/key` and `/type` feed the same live sessions the window uses. Its limits: `/click` reports the box under the point, so the cursor is that box's center and a slider click lands mid-track; there is no pointer-down or pointer-up route, so dragging is a window-only test; and a click and a key sent in the same frame run the key first, so wait a moment between them. For an app with no live session, empty readouts and frozen sliders in `-web` are how `-web` works, not an app bug.

### Level 3: the real window, driven and read back

Run the app in a window with `-snapshot ADDR` (the `blinkless-all` flag). It serves `GET /frame.png`, a picture of the window as it is now, at full page height, not just the viewport. This shows live state: readouts filled in, sliders at their live values, animated frames.

`scripts/gui-drive.sh` sends real mouse and keyboard input to that window and saves its picture. It needs `xdotool`. Without root, `scripts/gui-drive.sh setup` downloads the package with `apt-get download` and unpacks it into a local folder, and the script finds it there. No `sudo` is needed. It was written and tested in WSL2 with WSLg, where the window is an XWayland window on `DISPLAY=:0`.

```sh
go build -o "$TMPDIR/app" ./examples/blinkless-all
"$TMPDIR/app" -tab nav -snapshot 127.0.0.1:8120 -height 800 &
scripts/gui-drive.sh info                 # window id, position, where page pixel 0,0 is
scripts/gui-drive.sh click 90 238         # page pixels, not screen pixels
scripts/gui-drive.sh type "hello"
scripts/gui-drive.sh key Down
scripts/gui-drive.sh drag 50 476 170 476  # press, six moves, release
scripts/gui-drive.sh scroll 6             # six wheel notches down, about 48px each
scripts/gui-drive.sh snap after-click     # writes $GUI_OUT/after-click.png
```

Coordinates are page pixels in the viewport. For a row you saw in the full-page picture, the viewport y is the picture y minus the scroll offset, which is 48 times the notches scrolled from the top. Scroll to the top first with `scroll -100`. Under the app's fixed header the visible area starts near y=305.

Gotchas that cost time on WSLg:

- **Keys need `--window`.** There is no window manager, so plain XTEST keys never reach the window. The script sends `key` and `type` with `xdotool --window`, which works.
- **The wheel hits whatever is under the pointer.** A box with its own scrolling (the action-scroll sheet) eats the notches, so the page does not scroll as counted. `scroll` parks the pointer in the right margin first. If a click lands on the wrong element after you worked in a scrollable box, suspect the scroll offset.
- **The click origin is measured, not derived.** `xdotool getwindowgeometry` reports a window larger than the page, and page pixel 0,0 sits at the window position minus 32 in x and plus 4 in y on the host this was written on. `info` prints the result. If a click misses, set `GUI_OX` and `GUI_OY` and check by clicking a known button.
- **A restart clears odd state.** Restart the app between unrelated tests so one test's focus or selection does not leak into the next.
- **Shell word splitting.** In zsh an unquoted `$var` is not split into words, so `set -- $xy` gives one argument. Pass `$x $y` separately or run the loop under bash.

`-demo -demo-out DIR` plays a scripted live session on a manual clock with no window and writes one PNG per step. It is slow (minutes). Run it in the background and wait for the finish notice.

### What a tester covers per row

Do each of these where the row has the control, and report any that you could not do:

- The action in the row's own "Try it" sentence, by mouse and by key.
- Both ends of a range: a slider dragged past each end, a sheet scrolled to its end and back, a field filled until it overflows and emptied with Backspace.
- The negative case the description names: the control that must not react (the close box and body that do not drag, the field that keeps its width).
- State after leaving and returning (a tab switch reloads the page, so fields return to their defaults).
- Anything that animates: two pictures a moment apart must differ, and a paused one must not.

### Reading the pictures

- Read each screenshot yourself. Do not trust a hash alone.
- Crop a region with `convert in.png -crop 1000x300+0+2400 +repage out.png` (ImageMagick) when a full-page picture is too small to read. The system Python `PIL` was broken on that host, so do not depend on it.
- Record each finding as description, picture, readout, and which of the three disagrees. Mark each finding as seen in `-web` only, in the window, or both. Web-only findings are usually missing live state, not bugs.
- A note that only appears after a click (the `will-change` paint count) is not a mismatch at rest. Reproduce it with `liveRig` before you call it a bug.

### Process hygiene

- Start servers in the background and stop them when done. Stop by process name with `pgrep -x <name>`. Do not use `pkill -f <text>`: the pattern matches your own shell command line and kills your shell.
- Put binaries, screenshots and scratch tools in the scratchpad directory, not in the repo.
- Run `go test ./examples/<name>` for the example you touched. Run `make test` and `make lint` before you call the work done.
