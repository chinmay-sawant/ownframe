# Multi-process isolation

Recorded 2026-10-13 against the tree after commit `679b7b9`.

ownframe runs the page, the replay, the Ebiten window, and the GPU stack in
one address space. A panic in layout is caught at the top of `Run`
(`run.go:46-65`) but only by ending the run. A hang in template execution,
layout, or a fetch never returns to `Update` and the window stops
responding. An out-of-memory in the page heap or a fatal fault inside a C
library the GPU stack loads takes the window with it. This plan splits the
window host from the page and renderer, keeps a shared-memory frame path
between them, and makes a broken page a restartable event instead of a dead
window.

## What exists today

- `Run` calls `runWindow` (`run.go:18-20`,
  `internal/window/run_options.go:11-54`). The window shell implements
  `ebiten.Game` and calls `s.app` directly for size, title, draw state,
  input, ticks, and the fallback picture. There is no boundary to put a
  fault behind.
- `internal/page` owns the template, the data, the handlers, the cached
  parse, the display list, and all interaction state
  (`internal/page/page_types.go:15-77`). `Prepare` draws it once before the
  loop (`internal/page/page_prepare.go:6-19`). `internal/render` and
  `internal/bitmap` run on the same goroutine as the Ebiten callback, with
  `engineMu` serializing engine use (`internal/render/engine.go:6-7`).
- `internal/host` defines the screen the window talks to
  (`internal/host/screen.go:20-48`) plus the optional capabilities the
  window calls during a frame: `Ticker`, `Inspector`, `Reloader`,
  `Dropper`, `Focuser`, `ContextMenu`, `CursorShape`, `ScrollRequester`,
  `IME`, `LongPresser` (`internal/host/`).
- A panic becomes a text file through `internal/crash`
  (`internal/crash/crash.go:31-46`) and then an error return. Nothing keeps
  the window alive.
- The GPU stack is in-process by construction. Mesa, libLLVM, and the GL
  driver are loaded into this binary.

The failure list this plan attacks: a layout panic or infinite loop, a
template or fetch hang that blocks paint, an OOM in layout, and a fatal
fault in a text or image decoding C dependency. It does not fix a GPU driver
crash, because the host still owns the GPU, and it does not sandbox hostile
content, which would need an OS sandbox, not just a second process.

## The split

Two processes, one window.

Host process. The Ebiten loop, the OS window, the GPU, input events, the
devtools dock, the shell context menu, scrollbars, and the operating-system
services that need the window: clipboard, file dialogs, printing, IME,
cursor shape, window title and size. It never parses HTML, never runs a
template, and never calls into the engine.

Page process. Template execution, parse, cascade, layout, the display list,
the bitmap fallback, hit testing, focus, caret, selection, fetch, the file
watch, and `SetTick`. It owns the page state and produces frames.

The page process runs the same `main` as the host. `Run` re-executes the
binary with an environment role flag and connects a socket pair. The page
role does not open a window; it builds the same `Page` because the
program's own setup code runs unchanged, then serves frames. This keeps the
public API shape, `Run(ctx, page)`, at the cost of one documented
requirement: page construction must be repeatable, because it happens in
both processes. A page whose setup reads a one-shot resource or starts a
listener needs a guard. If the re-exec proves fragile in practice, the
fallback shape is an explicit split API, a `Prepare` that starts the page
worker and a `RunHost` that opens the window, with the existing `Run` kept
as the single-process default.

## Frame transport

Shared memory with dirty rectangles.

- One header plus N frame buffers in a shared mapping. Linux uses
  `memfd_create` and passes the descriptor over the control socket with
  `SCM_RIGHTS`; Windows uses a named file mapping; macOS uses `shm_open`.
  Each buffer is device-pixel RGBA. At 1908x999 that is about 7.6 MB per
  buffer, so a triple buffer is about 23 MB per window.
- The page process rasterizes the page into the next free buffer with the
  existing software painter. It already produces exactly this: the
  fallback path stores an `image.Image` from `internal/bitmap`
  (`internal/page/page_draw_fallback.go:14-39`), and the host already
  uploads a viewport of one into an `ebiten.Image` with `WritePixels`
  (`internal/window/bitmap_viewport.go:34-49`). The shared-memory path is
  that code with the copy going through a mapping instead of a slice.
- Each published frame carries a sequence number, page generation, size,
  scale, and a dirty-rect list. The host uploads only the dirty rects and
  blits. Full frames happen on the first frame, a size change, a page
  change, and after a restart. Partial frames are the steady state.
- Backpressure: the page keeps at most one unpublished buffer. If the host
  has not consumed the previous frame, the page merges dirty rects into the
  pending one instead of allocating another.

The alternative, streaming the display list and replaying it in the host,
keeps text sharp at any device scale and sends kilobytes instead of
megabytes. It is the right end state and the wrong first step. `DisplayOp`
carries a pointer to an unexported payload (`layout/layout.go:400-480`), so
a wire form means either an engine change or a serializer that walks every
accessor and reconstructs ops with fonts, image bytes, transforms, groups,
and grid runs. Pixels first; the display-list path can come later behind
the same frame protocol, selected per frame.

## Control channel

One socket pair, length-prefixed messages. The socket carries the memfd and
nothing else needs a second channel.

- Host to page: hello with protocol version and device scale, resize,
  pointer move, pointer button, wheel, key down and up, text input, IME
  commit and cancel, focus in and out, drop files, reload request,
  shutdown, inspect request.
- Page to host: ready with page size and minimum size, title change, frame
  published, clipboard read or write request, file dialog request, print
  request, cursor shape, context menu request, crash report path,
  heartbeat.
- Ordering: input messages are ordered and acknowledged in bulk by frame
  sequence. Pointer moves coalesce; keys never do.
- Timeouts: the host waits at most a few hundred milliseconds for a frame
  after an input message. On timeout it keeps the last frame, forwards a
  cancel to whatever work is in flight, and starts the hang watchdog. The
  window never blocks inside `Update` waiting for the page.

The `Screen` interface (`internal/host/screen.go:20-48`) has about thirty
methods. The frame and input messages cover the ones the window calls every
frame. The rest become request messages with deadlines.

## Input, state, and who owns what

- The host normalizes an event and forwards it. It does not know what a
  click hit or what has focus, so hover, press, release, type, and submit
  are all page-side. The host keeps only the last known cursor shape and
  the pending clipboard result.
- Scroll: the page owns the scroll offset because layout height decides
  it. The host sends wheel deltas; the page answers with a frame and, if
  the page moved, the new offset for the scrollbars.
- `SetTick`: the page process calls it once per its own frame cadence and
  publishes the changed rects. The host blits and presents. The tick
  callback never touches the GPU.
- Devtools: the host draws the dock, the counters, and the overlay. The
  Elements and Ops tabs read the display list, which lives in the page, so
  the host sends an inspect message and the page replies with a serialized
  operation list on demand. The Frame tab counters ride every frame header.
- Clipboard, file dialogs, and print are host-side. `internal/clipboard`,
  `internal/filepick`, and `internal/print` move behind the host, and the
  page calls them by message. `internal/fetch` and the reload watcher stay
  page-side. `internal/ipc` stays in-process and is documented as such;
  routing page-internal channels across the boundary is not worth it.

## Crash and hang recovery

- The page sends a heartbeat with each frame and on a timer when idle.
- If the socket closes or the heartbeat misses the deadline, the host marks
  the page dead, keeps the last frame on screen, and draws a banner over
  it. The window stays interactive for moving, resizing, and closing.
- The host restarts the page: same executable, same role flag, fresh
  socket pair and shared mapping. The new page runs its setup, republishes
  a full frame, and the banner clears. A restart budget, for example three
  in sixty seconds, turns into a persistent error state with the last
  frame and the crash report path.
- A hang gets the same treatment with a longer deadline. Go cannot safely
  interrupt a stuck goroutine, so the host kills and restarts.
- A panic in the page is caught at the `Run` boundary in that process,
  written by `internal/crash` as it is today
  (`internal/crash/crash.go:31-46`), and the path is sent to the host for
  the banner and the log. Crash reports stay local.
- A host crash still closes the window. Keep the host thin: Ebiten calls,
  input translation, frame upload, chrome drawing. No parsing, no layout,
  no fetch.

## What it fixes

- A layout panic or an endless loop restarts the page and leaves the window
  open.
- A page heap that grows out of control can be killed and restarted
  without taking the window. The system OOM killer can still pick any
  process, so this is containment, not immunity.
- A fatal fault in a font or image decoding dependency that the page
  touches no longer takes the window, unless the same library is also
  loaded in the host by Ebiten itself.
- A slow fetch or a long cascade no longer freezes resize, move, and
  close, because the host does not wait on the page.
- Reload-and-keep-the-window becomes a restart of one process instead of a
  program design problem.

## The cost

- Latency: every input event is one socket hop. Local messages are tens of
  microseconds, but a frame can be one refresh late. For typing and
  clicking this is invisible. For a 60 Hz drag it is one frame behind,
  the same as a normal compositor pipeline.
- Copies: a dirty rect is a memcpy from the page's software paint into the
  mapping, then an upload from the mapping into the GPU. The fallback path
  already pays the second copy today. The new cost is the first one, and
  it scales with the dirty area, not the page.
- Memory: one to three shared buffers per window, about 8 to 23 MB, plus
  the page process's own heap, which is already the large one.
- Complexity: a protocol with a version handshake, shared-memory lifecycle
  per platform, a supervisor, and a second process to debug. `go test`
  cannot cover the transport with the race detector across processes;
  the fake-page tests need an in-process loopback implementation of the
  same protocol.
- Coverage: wasm has no second process, Android and iOS bind Ebiten in the
  app process, and `Serve` has no window at all. Those stay single
  process, and the build must make that explicit rather than silently
  degrade.
- The existing engines assume a single writer per page. The page process
  becomes that writer, which is cleaner, but every host-side cache keyed on
  `Generation()` or a display list needs the frame header instead of a
  direct call.

## Repo touchpoints

- `run.go`, `options.go`, `window_options.go`: a process-mode option and
  the role re-exec in `Run`.
- `internal/window`: input translation and the frame blit replace direct
  `s.app` calls for drawing. `shell` stops owning the display list; the
  devtools data path becomes inspect messages.
- `internal/page`: runs in the page role, keeps everything it has, gains
  the serve loop and the publish path.
- `internal/render`, `internal/bitmap`: unchanged in behavior, now in the
  page process.
- `internal/host`: the optional interfaces become message pairs.
- `internal/crash`: unchanged, plus forwarding the report path.
- `internal/clipboard`, `internal/filepick`, `internal/print`: host-side.
- `internal/ipc`: documented as in-process only.
- New package `internal/mp` or `internal/isolation`: shared mapping,
  protocol, supervisor, watchdog, restart budget.
- Scripts and packaging: the binary must know its own path for the re-exec
  and handle a symlinked install.
- Docs: AGENTS.md map, a new `documentation/isolation.md`, and a note in
  the platform pages.

## Phases

1. Spiked transport behind an environment flag. Page child rasterizes the
   login example into shared memory, host shows it, input ignored. Prove
   memfd plus `SCM_RIGHTS` plus `WritePixels`, the watchdog, and the
   restart. 1 to 2 weeks.
2. Input forwarding, dirty-rect frames, scroll and hover, chrome host-side.
   The examples run end to end under the flag. 2 weeks.
3. OS services over the control channel: clipboard, file dialogs, print,
   cursor, IME, context menu. Devtools inspect path. 1 to 2 weeks.
4. Recovery: restart loop, banner, budget, crash report forwarding. Tests
   with a deliberately panicking and a deliberately hanging page. 1 week.
5. macOS and Windows mappings, packaging, docs, and the explicit
   wasm/mobile single-process story. 1 to 2 weeks per platform.

Roughly 6 to 9 weeks for Linux, then platform work. The ordering is
deliberate: prove the frame path and the restart first, because if either
does not hold, the rest of the plan is not worth building.

## What this is not

- Not a sandbox. The page process has the same user and the same filesystem
  access. A real sandbox is a separate project.
- Not GPU crash isolation. Ebiten owns the GPU in the host process, and a
  driver fault still kills the window.
- Not a tab model. One window, one page process. Multiple windows each get
  their own page process, which is a natural extension but not in this
  plan.
- Not a replacement for `internal/web`. A served page has no window to
  save, and `Serve` stays in-process.
