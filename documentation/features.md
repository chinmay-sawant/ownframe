# Features

OwnFrame renders HTML and CSS screens while your app logic runs in Go. Blinkless lays out the page, and Ebitengine opens the native window. OwnFrame does not embed a browser or JavaScript runtime.

1.0 native hosts are desktop `Run` and Android `BindMobile`, for an operator you control. Native means a compiled Go process and an Ebiten window or Activity, not OS widgets and not a screen reader. wasm is a canvas preview. `-web` is a PNG debugger. See [platforms.md](platforms.md).

This page summarizes supported features and current limits. Use the linked guides for setup and examples. See [Android runtime limits](android-runtime.md) and [all examples](../examples/readme.md).

## CSS and HTML support

This matrix is the 1.0 language. Holes are blinkless. ownframe's workaround is `Page.SetScrollWindow` and `Page.SetViewportPinZ`. Those are not tickets on ownframe. A long list is a slice window plus spacers, not `content-visibility`. See `examples/list`, `examples/ops-console`, and [tricks-tips.md](tricks-tips.md).

| Feature | 1.0 |
|---|---|
| Flex | yes |
| Grid | yes |
| Tables | yes; collapsed borders replay as `OpGridRun` ([screen.md](screen.md)) |
| Circular corner radii | replay path |
| `var()` | yes; theme sheets can define custom properties ([theming.md](theming.md)) |
| `rgb()`, `hsl()` | yes in templates and theme sheets ([theming.md](theming.md)) |
| `linear-gradient` | yes; engine rasterizes to `OpImage` |
| `:hover`, `:focus`, `:focus-visible`, `:active`, `:checked` | host state ([forms.md](forms.md)) |
| CSS masks | first pass: border box, simple sizes, `repeat` variants (see CSS masks below) |
| CSS motion records | sampled from Blinkless; the window does not start them ([css-motion.md](css-motion.md)) |
| `@media` width and orientation | yes ([theming.md](theming.md)) |
| `overflow` clipping | yes; clipped ops become `OpNoop` ([screen.md](screen.md)) |
| Overflowing page | the window scrolls content larger than the frame ([scrolling.md](scrolling.md)) |
| `z-index` | yes; `Display.Order` ([screen.md](screen.md)) |
| `position: relative` | yes; containing block ([screen.md](screen.md)) |
| `position: absolute` | yes on a block container |
| Letter-spacing, `text-transform` | replay path ([screen.md](screen.md)) |
| `<img>`, `background-image` | PNG, JPEG, SVG via `SetImage` ([screen.md](screen.md)) |
| `input`, `textarea`, `select`, `button` | yes ([forms.md](forms.md)) |

| Absent or app workaround | 1.0 |
|---|---|
| `position: absolute` on a flex child | blinkless hole: the child stays in flow, so `top` and `left` do nothing. Use a block container (Telegram `app-thread`). `internal/page` tests record this. ([screen.md](screen.md)) |
| `position: sticky` under window scroll | blinkless hole: the header stays in flow and has no `Fixed` ops, so the window translates it with the page. Not CSS sticky. `internal/page` tests record this. ([screen.md](screen.md)) |
| `position: fixed` under window scroll | engine `Fixed` ops stay untranslated; sticky does not get that mark. |
| Pinned chrome | `Page.SetViewportPinZ` and `Page.SetScrollWindow` (Telegram header and composer). Do not fake sticky in ownframe. ([screen.md](screen.md), [tricks-tips.md](tricks-tips.md)) |
| `@media print` | never matches ([theming.md](theming.md)) |
| `@media (height)` | never matches. `TestMediaHeightFollowsTheFrame` is skipped. |
| `prefers-color-scheme` | never matches ([theming.md](theming.md)) |
| Elliptical fill radii | bitmap fallback ([screen.md](screen.md)) |
| Non-separable blend modes | bitmap fallback ([repaint.md](repaint.md)) |
| Rotated glyphs, fake oblique, font features, autospacing, missing font | bitmap fallback ([screen.md](screen.md)) |
| `@container` vs element patch | a document with `@container` refuses every patch ([optimizations.md](optimizations.md)) |
| `content-visibility` | not faked. Window the rows and pad with spacers. |

## Not in 1.0

These stay out of 1.0. Do not plan them as library work for this tag:

- JavaScript
- a live DOM
- `<a href>` navigation
- video
- canvas
- WebGL
- `<audio>` as HTML (`Run` and `BindMobile` still create an Ebiten audio context for Go playback)

## Screen paint

OwnFrame updates the page from its template and draws it in the window. A page can repaint only the area changed by an action. Read [how a screen is rendered](screen.md) and [how repaint works](repaint.md).

## Element patches

`Config.RetainState` keeps the layout state a single element patch needs: the resolved styles of the document and the box tree the placement produced. It is off by default, so a page that never patches an element pays nothing for it.

With it on, `Page.PatchElement(ctx, id, edits...)` edits an element in the cached tree, asks the engine to re-lay out that element alone, and adopts the patched display the engine hands back. The engine hands over the whole display because a patch that changes the operation count moves every later operation index, and with it the paint order and the owners. The page rebuilds its box index, marks the changed operations dirty the way a redraw does, drops the cached picture, and refreshes the accessibility tree, so the window repaints the card and not the canvas. The engine refuses the cases it cannot represent, such as a subtree that holds a float, a sticky box, an anchored element, a named flow, or an out-of-flow box whose containing block lies outside the subtree, an element that changed size or place, or a document that uses a container query. `PatchElement` then returns an error wrapping `ErrPatchRefused`; check with `errors.Is`. The patch is atomic: a refused patch drops the cached tree, so the next redraw parses the source again and the page shows the source's state. An app that keeps state only in the tree must re-apply it, which is what `Page.SetAfterRedraw` is for.

Text set through a patch does not go through the emoji pass. The pass runs on the source before it is parsed, and a text edit lands on the parsed tree, so an emoji in patched text paints as text rather than as the color image the same emoji gets from the template. The engine's tree API only writes a text node, so the image cannot be built inside the tree.

`Handlers.ClickPaint` runs after a click handler and after a control's own change. When it reports that it painted the frame, the page skips the redraw, so a click that patched an element does not also pay for a layout.

`Page.SetAfterRedraw(fn)` registers a callback that runs after a committed redraw, so an app that keeps its state in the parsed tree can reapply it after the page rebuilt.

Detail is in [optimizations.md](optimizations.md). The measured cost of an
element patch against a full redraw is in
[PATCH_COST.md](../examples/css-preview/PATCH_COST.md).

## CSS masks

Blinkless resolves mask image bytes and stores the CSS values on a display
group (`blinkless/internal/layout/mask_paint.go:25-56,86-124`;
`blinkless/internal/layout/blend_group.go:13-15,38-55`). OwnFrame keeps masked
pages on the display-list path. Replay paints the element and its
children into a GPU layer, applies a mask texture with destination-in
compositing, then draws that layer back into the page
([internal/replay/group_paint.go:37-45](../internal/replay/group_paint.go),
[internal/replay/group_layer.go:16-36](../internal/replay/group_layer.go),
[internal/replay/group_mask.go:10-25](../internal/replay/group_mask.go)).
`Page.PNG` uses the software painter when PNG pixels are requested
([internal/page/page_image.go:35-40](../internal/page/page_image.go)). The
mask painter reads raster alpha or luminance and supports `add`, `subtract`,
`intersect`, and `exclude` layer composition. The first pass uses the element
border box and accepts simple percentage or pixel sizes and positions, plus
`no-repeat`, `repeat`, `repeat-x`, and `repeat-y`. Blinkless preprocesses SVG
`<mask>` content using its `mask-type`; OwnFrame reads that result from the
alpha channel.

Values outside this first pass, including `content-box`, `padding-box`,
`cover`, `contain`, `space`, and `round`, return a paint error. The
[CSS masking example](../examples/css-masking) shows the ten standard
properties and checks the output pixels.

Detail is in [../examples/css-masking/README.md](../examples/css-masking/README.md).

## Theming

`Config.Theme` and `Page.SetTheme` store one extra stylesheet that the render state hands to the engine as an extra sheet. The engine walks its rules last, but each sheet numbers rule order from zero, so an equal-specificity tie can still go to a late template rule; theming.md spells out the order rule. Any property the engine implements can appear in the theme, including custom properties the template reads with `var()`. `SetTheme` does not draw; the next `Redraw` applies the sheet. An empty theme is removed.

Detail is in [theming.md](theming.md).

## CSS motion

OwnFrame can sample CSS animation and transition records from Blinkless and
write each frame into retained drawing operations. The caller supplies time,
timeline progress, trigger events, and before and after displays where needed.
The Page window does not start this runtime automatically.

The [CSS motion guide](css-motion.md) explains the inputs and limits. The
[CSS property catalog example](../examples/css-preview) has a separate Start
button for each of its 43 motion properties. Its
[profiling notes](../examples/css-preview/PROFILING.md) cover running the
example and measuring a click.

## Hot reload

`Config.File` reads the template from disk at `New` and watches it; `Config.ThemeFile` does the same for the theme. The window polls the files every 250 ms and swaps the template without touching history, so form values, focus, hover, and active survive while their elements do, and `SetImage` and `SetTick` stay. A parse error keeps the last good picture and prints one line to stderr; the bytes are retried on the next poll. `Serve` polls before `GET /` and `GET /frame.png`, and the shell page refreshes the image while a watch is active. wasm and mobile read the file once. Setting `HTML` and `File` together is `ErrBadSource`.

Detail is in [hot-reload.md](hot-reload.md).

## Frames

`Page.SetTick` registers one function the window calls before it draws each frame. The function can change an operation in the retained display list, or call `Redraw`, so a page can animate without parsing the HTML again. `Serve` does not tick.

Detail is in [frames.md](frames.md).

## Keys

`Handlers.KeyDown` and `Handlers.KeyUp` receive each key press and release with a lowercase key name, such as `"space"` or `"arrowdown"`. A key handler does not draw; the page paints from its tick or calls `Redraw` itself. The window sends one pair per real key event and drops auto-repeat pulses. Tab and Shift+Tab move focus when the page has fields, and F11 toggles fullscreen on desktop; [interaction.md](interaction.md) covers both.

Detail is in [keys.md](keys.md).

## IPC

`Send`, `Listen`, `Handle`, and `Request` pass strings between callers in this process. `Send`, `Listen`, and `Handle` do nothing for an empty channel. `Request` returns `ErrNoHandler` for an empty channel or when nobody is handling that channel. There is no socket and no page-process bridge.

Detail is in [ipc.md](ipc.md).

## Navigation

`Load` replaces the template and pushes history. `Back` and `Forward` move through that history. `Route` maps a `data-action` value to HTML, and a click on that action loads it instead of calling the click handler. `HTML` returns the current template source. There is no `loadURL`, no `<a href>` navigation, and no custom protocol.

Detail is in [navigation.md](navigation.md).

## Crash reports

`Run` stores every failure in a text file and returns an error that includes the file path: panics go through the recover, and startup or window errors go through the same writer. `BindMobile` recovers a panic the same way; a browser build recovers and returns an error, but writes no file. `Report` writes the same kind of file without a panic. Nothing is uploaded.

Detail is in [crash.md](crash.md).

## Fetch and XHR

`Fetch` sends GET. `XHR` sends a method, headers, and a body. Only `http` and `https` are accepted. A redirect to any other scheme returns `ErrScheme`, and a redirect that steps from `https` down to `http` returns `ErrDowngrade`. Hosts that resolve to loopback, private, link-local, multicast, or unspecified addresses are rejected with `ErrPrivateAddress`. The client stores no cookies and no cache.

Detail is in [fetch.md](fetch.md).

## Clipboard

`Write` stores the memory copy in memory-only mode and otherwise writes only the OS clipboard, so a failed read cannot resurface older text. `Read` returns the OS text, or `""` plus `ErrUnavailable` when the OS read fails. On Wayland, `Write` and `Read` run `wl-copy` and `wl-paste` when those binaries are on PATH. If they are missing, Linux uses the X11 `CLIPBOARD` selection when `DISPLAY` is set. Without the helpers and without `DISPLAY`, a write is a no-op and a read returns `ErrUnavailable`. Windows uses `CF_UNICODETEXT`. macOS with cgo uses `NSPasteboard`; without cgo it stays on the memory copy. wasm and Android stay on the memory copy. Tests call `UseMemory` and do not touch the desktop clipboard. `Serve` has no clipboard route.

Detail is in [clipboard.md](clipboard.md).

## Drag and drop

The window reads Ebiten's dropped files once per frame and hands them to a screen that implements `host.Dropper`. `Handlers.Drop` receives each file's name, size, directory flag, absolute path on desktop, and a `Read` function that serves the bytes during the call. A dropped directory arrives as one entry. `-web` has no window loop, so it receives no drops; the wasm canvas build does.

Detail is in [drag-drop.md](drag-drop.md).

## Forms

An `input`, `textarea`, or `select` with an id is stored on the page. A click focuses a text field or a textarea, toggles a checkbox, checks a radio, or cycles a select. Typing edits the focused text field even when the type handler is nil. A file input opens the desktop file dialog under `Run` and stores the chosen path; wasm, mobile, and `-web` keep the typed name. `SetFormValue` and `SetFormChecked` do not redraw. A control with `data-bind` is tied to a field on the pointer passed to `SetData`; an edit writes through before the redraw, and `Handlers.Change` receives the changed control's box. `:focus`, `:hover`, and `:active` match with host state, and `:checked` follows the control's `checked` attribute; `data-ownframe-*` remains the attribute alternative. The engine's default stylesheet gives a `<button>` a face when the author does not style it, and a submit-like `input` is rewritten to a `button`. The login example uses a button.

Detail is in [forms.md](forms.md). Select all, undo, and redo are in [editing.md](editing.md).

## Accessibility tree

`Page.AccessibilityTree` returns semantic roles, names, descriptions, live
form state, document order, and bounds where an element has a unique HTML ID.
`Page.AccessibilityPerform` routes supported focus, activation, and text-value
actions through the page's existing input paths. The snapshot is a host
contract for the operator who runs the process. 1.0 does not ship a
screen-reader product. Linux maps the tree to AT-SPI objects in process;
`Live` is false, so Orca cannot inspect the window. CSS-only hidden content
and anonymous-node bounds remain unresolved.

Detail is in [accessibility.md](accessibility.md) and [atspi.md](atspi.md).

## Interaction

Tab and Shift+Tab move focus in document order when a page has fields, and Escape clears it. A mouse press in a field places the caret, a drag extends the selection, a double-click selects a word, a triple-click selects a line, and a press held in place selects the word under the point or reaches `Handlers.LongPress`. A right click opens a shell menu with cut, copy, paste, select all, undo, and redo. The hovered shape picks the cursor, a touch drag scrolls and a pinch zooms by default; `Config.DisableTouchZoom` or `Page.SetTouchZoomAllowed(false)` disables only pinch zoom. A page without that option can still pinch on devices that provide touch input. `Ctrl+=` and `Ctrl+-` zoom the page while `Ctrl+0` resets it, `Page.ScrollTo` and `Page.ScrollBy` move the offset, and F11 toggles fullscreen on desktop.

Detail is in [interaction.md](interaction.md).

## Window options

`RunWithOptions` opens the desktop window with a `WindowOptions` value.
`Transparent` leaves unpainted pixels transparent, `Borderless` removes the
frame, `Floating` keeps the window above normal windows, and `FixedSize`
stops resizing. `BottomRight` places the window at the current monitor's
bottom right, inset by `Margin` CSS pixels. `MousePassthrough` sends every
pointer event to the application underneath, `Interactive` keeps input only
where its callback returns true, and `Draggable` reserves a region for moving
the window. The zero value opens the same window as `Run`; placement,
decorations, and passthrough do not apply to a browser canvas or a phone
view.
`examples/desktop-cat` uses the transparent, click-through settings. Detail
is in [window.md](window.md).

## DevTools

The window can draw an inspector over the page. F12 or Ctrl+Shift+I toggles
it, `Config.DevTools` starts it on, and `Page.SetDevTools` changes it at
runtime. The inspector is a right-side dock with three tabs. It opens on
Frame, which shows the frame and pipeline counters; Elements shows the pinned
or hovered element as pretty JSON, and Ops lists the display operations
colour by kind. The dock shrinks the page viewport while it is open, so the
page relayouts to the strip left of it. It outlines the box under the cursor,
and a pick opens the Ops tab on the operation inside the box. A click on an
Ops row outlines that operation and copies its line to the clipboard. It is
window chrome, so `Page.PNG`, the display list, and the box list never
change. A custom screen implements `host.Inspector` to opt in; a screen
without it gets no overlay.
`GET /debug/state` serves the same data in web mode when that endpoint is enabled (`ServeOptions{Debug: true}` or `{Perf: true}`).

Detail is in [devtools.md](devtools.md).

## Performance

The perf checklist lives in [performance.md](performance.md), with per release numbers in [performance-baseline.md](performance-baseline.md). The Frame tab carries PERFORMANCE, PIPELINE DETAIL, and MEMORY counters for frame stages, repaint work, and memory, and `GET /debug/state` serves the same data in web mode when that endpoint is
enabled (`ServeOptions{Debug: true}` or `{Perf: true}`). Sampling is off by default and costs nothing; `Config.Perf`, `ownframe.WithPerf(true)`, `WindowOptions.Perf`, or `ServeOptions{Perf: true}` opts in. The three benchmark apps sit in `examples/perf-benchmarks`, and `scripts/perf-idle.sh`, `scripts/perf-longrun.sh`, and `scripts/perf-baseline.sh` run the idle, soak, and baseline passes.

## Still absent

These Electron pieces are not in this branch. The scan that listed them is [../plans/v0.0.1/compare.md](../plans/v0.0.1/compare.md).

- Chromium, V8, preload, `contextBridge`, and Node.
- Cross-process IPC, native menus, tray, and more than one window. The library ships no notification API; `examples/desktop-cat` draws its own message bubble, takes agent messages over a local HTTP endpoint, and reads Chrome's media session on Windows. File dialogs exist on desktop `Run` only; wasm, mobile, and `-web` keep the typed name.
- Session, cookies, cache, and web storage.
- An ownframe auto-updater, installer, and uploaded crash dump. Applications
  or their distribution channels own update checks and installation
  ([packaging.md](packaging.md#auto-update)).
- Video, document canvas, WebGL, and a context menu were the v0.0.1 absences. The context menu shipped in v0.0.2 ([interaction.md](interaction.md)); video, document canvas, and WebGL are still out. There is no library audio API or `<audio>` element; the examples play audio through `examples/music`. `Run` and `BindMobile` create a 48 kHz Ebiten audio context, `Serve` does not ([window.md](window.md)).
- A live AT-SPI D-Bus client, UI Automation, VoiceOver, and TalkBack. Linux maps the Go tree to AT-SPI objects and records them on a fake bus in tests; it does not register on the accessibility bus ([accessibility.md](accessibility.md), [atspi.md](atspi.md)). Native spellcheck and OS-global shortcuts are also out. `Page.DispatchDeepLink` accepts app-delivered URLs under a scheme allowlist. The app and its installer still own OS protocol registration and delivery. Clicked `http`, `https`, and `mailto` links can be handed to an app callback. Android uses Ebiten's experimental `exp/textinput`; desktop builds have a Composer path with temporary preedit painting, pending native input-method checks ([interaction.md](interaction.md)). iOS is not shipped. The phase plan is [here](../plans/v0.0.2/accessibility-desktop-ime.md).
- Sandbox, CSP, and context isolation.

Paint-only hover handlers and scroll callbacks that reuse overscan are described in [frames.md](frames.md).
