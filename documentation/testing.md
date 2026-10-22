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

`make graphics-check` runs three tests in separate processes and separate Xvfb sessions:

- `TestPartialFrameLeavesTheScreen` checks damage-only blits, full rebuilds, scroll copies, and hiding then restoring content under the devtools dock.
- `TestWindowRepaintPixels` uses a real page and compares the retained window framebuffer with a fresh full replay across a color change crossing tile boundaries, element removal, scrolling with a fixed header, a clean frame, and an edit while scrolled. It requires the color change to take the partial repaint path.
- `TestGPUVisibleParity` checks replay pixels, shape and text caches, transformed operations, and composited groups inside Ebiten's `Draw` loop.

All pixel readback occurs in the running game. Replay comparison targets use standalone textures so atlas placement does not introduce different interpolation rounding between the two targets. Pixel equality remains exact. Each process has a 90-second timeout and runs with software OpenGL for repeatability. Ubuntu needs `xvfb`, `xauth`, and `libgl1-mesa-dri`. The script fails when these prerequisites are missing. It leaves the user's `~/.Xauthority` untouched.

CI and release gates run the graphics target. Xvfb validates an X11 software graphics session; it does not establish native Wayland behavior, hardware GPU driver behavior, desktop IME correctness, or Android device behavior. Those require their own platform sessions.

To run the same pixel checks against your current Linux desktop session, use `GRAPHICS_SESSION=desktop make graphics-check`. This mode requires `DISPLAY`, opens test windows, and preserves your graphics driver selection. Under WSLg it checks the WSLg X11 path, not a native Windows host. A passing desktop run proves these rendering scenarios in that session; it does not prove hardware acceleration, input-method behavior, or other operating systems. The default remains isolated software rendering under Xvfb.

## Frame pacing ownership

`internal/window/pacing` owns idle transitions, declared frame deadlines, keep-alive scheduling, and diagnostic counters. Its tests use supplied times and the same interface as the window adapter. They run without a display or Ebiten dependency. Deadline replacement, cancellation, exactly-once wake consumption, timer sleep, and snapshot reads have explicit checks.

`idle_game.go`, `idle_frame.go`, and `idle_timer.go` retain the Ebiten mode changes, timer goroutine, and locking. The policy tests prove scheduling decisions. Pixel checks prove rendering; neither check simulates native keyboard input or establishes actual event-wait timing.

## State subscription lifecycle

The internal retained-state tests cover rebinding a page to a store, repeated binding to the same store, detaching one of two pages, and rejecting a former store's already-dispatched callback. Listener cancellation uses registration identity rather than a closure's code address. Callbacks execute outside the store lock and may reenter the store. An already-dispatched callback may finish after cancellation, so the page also checks the binding generation before recording metadata.

`go test -race -p 1 ./internal/page ./internal/state` exercises concurrent store patches, listener registration and cancellation, page rebinding, and metadata reads. These are internal engine seams, not additions to the public `ownframe.Page` method set.

## Review responsibility

There is no character cap. Keep related implementation and test scenarios together. Extract a package when it can own a responsibility with explicit inputs and outputs, rather than to reduce a file count. Use the cohesion questions in [CONTRIBUTING.md](../CONTRIBUTING.md).
