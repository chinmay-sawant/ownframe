# Phase 08, OOP-R rasterization

> **Parent:** `plans/v0.0.3/optimizations/README.md` - optimization family 8,
> "OOP-R (out of process raster)".
> **Status:** planning only. No behavior changes with this file.
> **Related:** `plans/v0.0.2/multi-process-isolation.md` is the broader process
> split plan. `internal/ipc` is in-process only and stays that way
> (`documentation/ipc.md`, `internal/ipc/ipc.go` re-exports
> `Send`, `Listen`, `Handle`, and `Request`).
> **Estimated effort:** 6 to 9 weeks after the frame transport and partial
> repaint contracts are stable, as sized in the multi-process isolation plan.

---

## Overview

OOP-R moves software rasterization and image decode work out of the process
that owns the page and out of the process that owns GPU presenting. The page
still builds `layout.DisplayList`, the window still presents through Ebiten, and
the renderer process only paints accepted frames into shared buffers. The output
stays `layout.DisplayList`; shared-memory RGBA frames are a transport artifact,
not a committed golden baseline and not a PDF or page bitmap product.

This phase is last in the optimizer order because it depends on frame transport,
damage lists, tile painting, and partial repaint being stable. It is a consumer
of `plans/v0.0.2/multi-process-isolation.md`, not a replacement for that plan.
The multi-process plan owns process roles, restart policy, platform mappings,
and recovery. Phase 08 narrows one part of that plan into a raster service
contract and a measurement gate.

Constraints carried from the multi-process plan:

- `internal/ipc` is in-process only. Page-internal channels are not routed across
  the process boundary. The renderer service needs its own protocol.
- wasm, Android, iOS, and `Serve` stay single-process. The build must say so
  explicitly, not degrade quietly.
- The page process remains the single writer of page state. The renderer only
  accepts frame and raster requests.
- Pixels first. Display list wire transport is a later option behind the same
  frame protocol, because `DisplayOp` carries an unexported payload
  (`layout/layout.go:400-480` in blinkless).
- Host crashes still close the window. The renderer process has no GPU access and
  no raw window handle.

## Where it lives

ownframe only. blinkless stays untouched. No dependency is added; the renderer
reuses the existing software painter and tile grid.

Likely ownframe files:

- `run.go`, `options.go`, `window_options.go`: process mode option and role re-exec.
- `internal/window/replay_tile*.go`: tile grid, paint, blit, plan, lifecycle.
- `internal/window/replay_partial.go` and `internal/window/replay_partial_state.go`:
  partial repaint entry point and counters.
- `internal/replay/*`: replay primitive, image cache, visible op selection.
- `internal/page/page_draw*.go`: display build, dirty list, and frame publish.
- `internal/render/*`: cache, state identity, styled document.
- New internal package for transport, mapping, supervisor, watchdog, and restart
  budget, owned by the multi-process plan.
- `documentation/performance.md`, `documentation/repaint.md`, and
  `documentation/platforms.md`: renderer process and single-process limits.

## Proposed renderer process

The renderer process is separate from both the host and the page. It does not
parse HTML, run templates, hold page state, or call Ebiten. It receives raster
requests, paints the requested region through the existing replay tile path,
and publishes the result into shared memory. It also keeps image decode work on
its side, so a slow or failing decode does not stall the page.

Roles:

- Host process: owns the window, GPU, input translation, chrome, frame upload,
  and OS services.
- Page process: owns template, data, layout, display list, hit testing, and the
  dirty list.
- Renderer process: owns tile grid painting, image decode, and shared frame
  buffers. It receives page data only through the transport.

The renderer never reads files, fetches URLs, or executes template code. It
returns status, not content decisions. A crash in the renderer restarts that
process and leaves the last good frame on screen while the page continues.

The renderer can be embedded in the page process behind a build or runtime flag
during development. That keeps tests and debugging simple and proves the protocol
without a second process.

## Transport

The transport is the one from `plans/v0.0.2/multi-process-isolation.md`,
section "Frame transport", with renderer-specific messages added beside it.
`internal/ipc` is not reused for cross-process transport.

Frame buffer:

- Shared mapping with one header and N buffers.
- Linux uses `memfd_create` plus `SCM_RIGHTS`; Windows uses a named mapping;
  macOS uses `shm_open`.
- Each buffer is device-pixel RGBA.
- Each published frame carries sequence number, page generation, size, scale,
  and dirty rect list.
- Backpressure keeps at most one unpublished buffer. The renderer merges dirty
  rects instead of allocating another buffer.

Control channel:

- Socket pair with length-prefixed messages.
- Host to renderer: hello, resize, buffer release, shutdown.
- Page to renderer: raster request with frame id, display generation, viewport,
  clip, scale, damaged rects, and image payload handles.
- Renderer to page or host: ready, frame published, status, heartbeat.
- Timeouts stay on the renderer side. The page never waits for a raster result
  inside `Update` or `Redraw`.

Raster request contract:

- A request names one frame and one content size.
- Dirty rect lists use the same cap and merge behavior as the tile plan.
- The renderer paints into its tile grid, then reports the dirty rects it
  actually painted.
- Raster output is compared against the in-process tile path during development
  before the in-process path is removed.

## Raster service

The raster service is the part that turns `layout.DisplayList` paint work into
shared pixels. It can start as an in-process service behind an interface, then
move across the process boundary after parity is proven.

Service operations:

1. `Open(width, height, scale)` allocates or resizes the tile grid.
2. `Paint(frameID, display, damage)` paints the damaged tiles and records
   paint and pixel counters.
3. `Decode(images)` decodes image payloads off the draw path and caches them by
   payload identity.
4. `Publish()` returns frame sequence, dirty rects, and a buffer handle.
5. `Close()` disposes images and tile memory deterministically.

Service rules:

- The renderer keeps the same tile size, clip, dock behavior, and chrome
  handling as `internal/window`.
- It does not change op order, z order, transforms, groups, or the output list.
- It does not introduce a raster golden. `make golden` still validates display
  lists in `./layout`.
- Decode errors fall back to the current empty or cached image behavior; they
  do not crash the page.
- Counts are available for devtools and paced dumps without per-frame allocation.

Migration rule:

The old `drawReplayPartial` path stays authoritative until a parity test proves
the renderer produces identical pixels for the supported cases. Only then does
the simple path become a fallback or get removed.

## Measurement gate

No OOP-R change closes without numbers. The baseline and after measurement must
use the same command and the same machine load. Record both in this ledger.

Baseline command:

`go run ./examples/departure-board -dump -pace 16ms`

After command:

`go run ./examples/departure-board -dump -pace 16ms`

Required measurements:

- worst frame time before and after;
- mean frame time before and after;
- raster paint count before and after;
- raster pixel count before and after;
- image decode time before and after;
- dirty rect count before and after;
- process RSS and allocation rate before and after;
- restart and heartbeat overhead during idle frames.

Proof gates:

- `make golden` after raster, damage, or transport changes.
- `make test` at session end.
- A synthetic slow decode test that keeps the page interactive.
- A synthetic renderer crash test that keeps the last frame visible.
- A packed layout parity test between in-process raster and renderer raster.

## Phases

### Phase 08.0, baseline and constraints

- [ ] Record the OOP-R baseline with the departure-board dump command.
- [ ] Document which platforms stay single-process: wasm, Android, iOS, `Serve`.
- [ ] Confirm no dependency is added and the allowlist test stays green.
- [ ] Confirm `internal/ipc` remains in-process only in docs and code.
- [ ] Read `plans/v0.0.2/multi-process-isolation.md` and mark inherited constraints.

### Phase 08.1, in-process raster service

- [ ] Define a raster service interface inside the window or replay package.
- [ ] Implement it with the current `drawReplayPartial` behavior.
- [ ] Route normal desktop painting through the service behind a flag.
- [ ] Keep tile counters and dock clipping behavior identical.
- [ ] Add service lifecycle tests.

### Phase 08.2, decode off the draw path

- [ ] Move image decode work into the raster service.
- [ ] Cache decoded images by payload identity.
- [ ] Preserve cache behavior under partial repaint.
- [ ] Add a slow decode test and a decode failure test.
- [ ] Measure decode time before and after.

### Phase 08.3, shared buffer publisher

- [ ] Allocate shared RGBA buffers per the multi-process plan.
- [ ] Publish a frame header with sequence and dirty rects.
- [ ] Treat buffer release as a host or page duty with deadlines.
- [ ] Keep at most one unpublished buffer under backpressure.
- [ ] Test header fields and dirty rect merge behavior.

### Phase 08.4, protocol and transport

- [ ] Define renderer request and response structs.
- [ ] Add the socket control channel and platform shared-memory mapping.
- [ ] Reject unknown protocol versions with a clear error.
- [ ] Keep input and page state off the renderer channel.
- [ ] Add a loopback fake renderer for protocol tests.

### Phase 08.5, renderer role

- [ ] Add a renderer role and role flag to `Run`.
- [ ] Build the service without templates, handlers, or window state.
- [ ] Publish frames into shared memory only.
- [ ] Send heartbeats with and without raster work.
- [ ] Prove the renderer restarts cleanly after a panic.

### Phase 08.6, page integration

- [ ] Send raster requests from the page when a frame is ready.
- [ ] Keep dirty lists and `TakeDirty` unchanged across the boundary.
- [ ] Replace `Page.Image` behavior only after host upload is proven.
- [ ] Verify tile painting and blit parity against the in-process path.
- [ ] Measure publish and upload cost.

### Phase 08.7, partial repaint parity

- [ ] Compare renderer pixels to in-process pixels for scroll cases.
- [ ] Compare renderer pixels to in-process pixels for dock cases.
- [ ] Compare renderer pixels to in-process pixels for two separated changes.
- [ ] Keep fallback image behavior when renderer output is unavailable.
- [ ] Document any case that still requires a full frame.

### Phase 08.8, recovery

- [ ] Detect renderer socket closure or heartbeat miss.
- [ ] Keep the last published frame on screen.
- [ ] Restart the renderer within the budget and clear the error state.
- [ ] Forward renderer crash paths to the page and host logs.
- [ ] Test a renderer kill without killing the page or window.

### Phase 08.9, measurement and closure

- [ ] Run the paced dump command before and after on the same hardware.
- [ ] Record worst frame, mean frame, paint, pixel, decode, and dirty metrics.
- [ ] Run `make golden` and `make test`.
- [ ] Update `documentation/performance.md` and `documentation/repaint.md`.
- [ ] Leave each unproven renderer claim as an open row.

## Exit criteria

- The renderer process paints a frame without page code or GPU access.
- Dirty rectangles, counters, and tile behavior match the in-process path.
- Image decode runs outside the page draw path and reports failures safely.
- A renderer crash does not close the window or lose page state.
- wasm, mobile, and `Serve` remain explicitly single-process.
- The measurement gate shows ratio-based improvement rather than an assumed win.
- `make golden` and `make test` pass.

## Risks

- The protocol can become the slowest part of the frame if request copies grow with
  page size rather than dirty area.
- Shared-memory buffers cost about 7.6 MB per frame at 1908x999, so buffer count
  must stay small and measured.
- Tile parity can fail on sub-pixel offsets, dock clip, and chrome handling.
- Moving decode outside the page may need payload identity rules that blinkless
  does not yet expose.
- The renderer may duplicate state if page and renderer both cache images.
- Cross-process debugging is harder than single-process debugging; keep a
  loopback service and traces from the start.
- The host still owns Ebiten and the GPU, so GPU faults remain outside this
  phase.

## Notes for future work

- Display list wire transport remains optional and requires either blinkless
  serialization support or an engine change because of the unexported
  `DisplayOp` payload.
- Directly composited images and solid color (phase 09) may fit naturally as
  renderer-side layer promotion, but that wait until tile painting is stable.
- Async image decoding (phase 10) can share the decode queue introduced here.
- Keep any binding or public API change out of OOP-R until the process boundary
  is proven.
