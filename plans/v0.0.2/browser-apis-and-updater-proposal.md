# Browser API and updater proposal

This proposal closes the planning work for Canvas, video, WebGL, and update
ownership in v0.0.2. It makes no code or dependency changes. JavaScript,
live DOM mutation, and browser compatibility remain out of scope.

The current renderer places HTML and CSS through Blinkless, then replays its
display list through Ebitengine. `Page.SetImage` supplies image assets, and
`Page.SetTick` runs application code before a window draws. These boundaries
favor application-drawn frames over browser APIs that need JavaScript or a
second graphics context. See the [window lifecycle](../../documentation/window.md),
the [frame callback guide](../../documentation/frames.md), and the dated
[gap inventory](electron-gaps.md).

## Canvas

### Decision

Plan a Go image surface. Do not implement the Canvas 2D API. A Go application
draws into an `image.Image` and submits the result for a `<canvas>` element to
display.

The initial API candidate is `Page.SetCanvas(id string, img image.Image)`,
beside `Page.SetImage`. Blinkless owns canvas element sizing and placement.
Ownframe owns the image submission and repaint. The application owns all
drawing and animation code.

### Ownership and dependencies

- Blinkless adds `<canvas>` as a replaced element. The `width` and `height`
  attributes set its intrinsic bitmap size, with HTML defaults of 300 by 150.
  CSS controls its displayed box.
- Ownframe matches a submitted image to a unique canvas ID and paints it in
  that box. A changed frame repaints the canvas bounds without reparsing HTML
  or rerunning CSS layout. A changed intrinsic size triggers layout.
- The application creates and updates the image. It submits a stable image
  snapshot on the window update thread, such as from `Page.SetTick`.
- This uses Go's `image` package and the existing Ebitengine image path. It
  adds no Go module.

### Acceptance criteria for a future implementation

- A canvas with default dimensions, explicit dimensions, CSS dimensions, and
  CSS resizing gets the expected layout box.
- A Go app can submit an `image.RGBA` and see the pixels in the canvas box.
- Replacing pixels repaints only the canvas region and does not parse or lay
  out the document again when its dimensions stay the same.
- Changing the intrinsic dimensions updates the canvas layout and clips or
  scales pixels according to a documented rule.
- The same example runs on desktop and wasm. The documentation says that the
  API is a Go image surface, not the browser Canvas 2D API.

## Video

### Decision

Do not add a decoder or browser `<video>` implementation to ownframe. If an
application needs video, it can decode frames outside the library and submit
them to the Go image surface proposed above. The application owns media
controls, supported formats, playback time, and audio synchronization.

The app or an app-specific adapter owns demuxing and decoding. It chooses
codecs that fit its platform and distribution method. Ownframe only displays
the submitted frame. A video example must identify its decoder and tested
formats; the library makes no format guarantee.

### Dependencies and module impact

- A reusable frame surface depends on the Canvas proposal.
- Decoding depends on a decoder selected by the application. The Go standard
  library does not decode video.
- Audio synchronization depends on an application-owned playback clock and
  audio output. Ownframe does not expose a common audio clock across its hosts.
- The core module gains no decoder dependency. An application that adds a
  decoder accepts its module, codec, licensing, and platform costs itself.

### Acceptance criteria for an application example

- The example names its decoder, versions, supported formats, and supported
  operating systems.
- Playback selects frames from presentation timestamps using the
  application's playback clock. It drops late frames instead of building an
  unbounded queue.
- Audio and video stay within 100 ms of each other during a ten-minute run
  on each claimed platform.
- Seeking, pause, resume, end-of-stream, decode failure, and window shutdown
  release decoder and audio resources.
- The example reports frame drops and decode errors. It makes no claim that
  ownframe itself supports a video format.

## WebGL

### Decision

Do not expose WebGL or raw OpenGL through the current ownframe renderer.
Ebitengine owns the graphics context, and its public API does not expose GL
calls. A second binding cannot safely assume ownership of that context. WebGL
also does not fit the current no-JavaScript scope.

An application that needs custom GPU drawing should use a separate renderer
project and present its result as an image if it can do so without sharing
Ebitengine's graphics context. WebGPU follows the same decision. The Canvas
image surface is the supported proposal for app-drawn pixels.

### Dependencies and module impact

- A real WebGL-style API would need a renderer that owns context creation,
  GPU resources, shader compilation, synchronization, and device loss.
- That renderer would need a platform matrix for desktop, wasm, and mobile.
- A graphics binding alone does not meet these needs and would add a module
  while leaving context ownership unresolved.
- No ownframe module or graphics-context hook is proposed.

### Criteria to reopen the decision

- A separate design names the renderer and context owner for every supported
  platform.
- A runnable prototype demonstrates resource creation, drawing, resize,
  shutdown, and context or device loss without accessing Ebitengine's private
  context.
- The prototype names its module and licensing costs and keeps those costs
  outside the core module unless the no-new-modules rule changes.
- The proposal defines a software fallback and tests output on every claimed
  backend.

## Update ownership

The application or its distribution channel owns update checks, signing,
downloads, installation, and restart. OS stores and package managers should
update apps they install. An app distributed outside those channels may ship
its own updater process and update policy.

The [package guide](../../documentation/packaging.md) describes the current
archive layouts and `SHA256SUMS`. Those checksums detect archive changes; they
do not authenticate the publisher. The package script builds and checksums
archives. It does not contact an update service, sign releases, install an
update, or restart an application.

### Application responsibilities

A custom updater belongs to the application. It must verify a signed release
manifest and the downloaded artifact before installation. It must account
for platform-specific replacement rules, preserve or roll back the current
version on failure, and restart only after the application has saved its
state and closed its window.

The app can coordinate this through the existing `Run` lifecycle. `Run`
returns after the window closes, its context is cancelled, or a handler fails.
The app's `main` function can then save state, start its updater, and exit or
relaunch. See [window errors and shutdown](../../documentation/window.md#errors)
and [showing a page](../../documentation/window.md#showing-the-page).

### Library support

No updater-specific library hook is needed now. The app already owns process
startup and receives the result of `Run`. A generic shutdown callback would
add lifecycle state without helping the library install an update.

Revisit a generic shutdown hook only if an application cannot reliably save
or release its own resources before `Run` returns. Do not add update checks,
download code, signature policy, install logic, or restart calls to ownframe.

### Acceptance criteria for packaging documentation

- The package guide names the app, store, or package manager as update owner.
- The guide distinguishes a SHA-256 checksum from a publisher signature.
- Package scripts make no network calls and do not install or restart apps.
- Any custom updater example identifies its signing key policy, supported
  platforms, rollback behavior, and restart sequence.
