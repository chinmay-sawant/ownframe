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

`make graphics-check` runs six tests in separate processes and separate Xvfb sessions:

- `TestPartialFrameLeavesTheScreen` checks damage-only blits, full rebuilds, scroll copies, and hiding then restoring content under the devtools dock.
- `TestWindowRepaintPixels` uses a real page and compares the retained window framebuffer with a fresh full replay across a color change crossing tile boundaries, element removal, scrolling with a fixed header, a clean frame, and an edit while scrolled. It requires the color change to take the partial repaint path.
- `TestViewportScrollPixels` compares the viewport cache with a fresh replay across vertical and diagonal scrolling, edits during a buffer move, and jumps with no overlap. It also requires an overlapping scroll to repaint less than a full viewport.
- `TestWindowIdleDeadline` runs the actual pacing wrapper and timer. It requires the blocking idle FPS mode, a quiet interval with a declared deadline, a deadline-driven tick that paints blue, and restoration of the original tick rate. A watchdog wakes a stuck loop only to fail the test. Keep the test window free of input during this check.
- `TestDrawPaintsVerticalThumb` checks visible scrollbar pixels inside the running game.
- `TestGPUVisibleParity` checks replay pixels, shape and text caches, transformed operations, composited groups, and translucent fills, images, and grids inside Ebiten's `Draw` loop. Replay fixtures live under `internal/replay/testdata`, so a library module zip contains them.

All pixel readback occurs in the running game. Replay comparison targets use standalone textures so atlas placement does not introduce different interpolation rounding between the two targets. Pixel equality remains exact. Each process has a 90-second timeout and runs with software OpenGL for repeatability. Ubuntu needs `xvfb`, `xauth`, and `libgl1-mesa-dri`. The script fails when these prerequisites are missing. It leaves the user's `~/.Xauthority` untouched.

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
