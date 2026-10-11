# v0.0.2 accessibility and desktop input checklist

This checklist tracks the remaining desktop and platform gaps recorded in
[`electron-gaps.md`](electron-gaps.md) and [`documentation/features.md`](../../documentation/features.md).
It is the status source for the work below. Rows stay open until the listed
evidence exists.

## Scope

This plan covers accessibility, desktop IME, shell integrations, and the
browser APIs called out in the gap list. Video, document canvas, and WebGL are
planning-only in this cycle. Auto-update belongs to an application or package
manager, not the ownframe library. Accessibility and input must work with the
existing Go renderer and Ebiten window; they do not imply JavaScript support.

## Phase 1: Pointer and scrollbar behavior

- [x] Keep the vertical or horizontal resize cursor while the pointer is over
  its scrollbar thumb.
- [x] Route thumb presses and drags through the live pointer path.
- [ ] Verify content clicks, page scrolling, and scrollbar dragging do not
  consume one another's events.
- [x] Document the pointer behavior and add targeted window package tests.

Exit: hovering and dragging either thumb works in the running window, with
targeted pointer tests and `make test` passing.

## Phase 2: Desktop IME

- [x] Start an Ebiten text-input session for the focused text field on desktop.
- [x] Place the candidate window from the focused field's transformed bounds
  after scroll, zoom, and resize changes.
- [ ] Place the candidate window at the glyph-level caret rather than the
  focused field box.
- [x] Keep preedit text separate from the committed field value; draw its
  selection and underline without firing field change callbacks.
- [x] Apply commits and reconversion through the existing edit path exactly
  once, including replacement around the caret.
- [ ] End or cancel sessions on blur, Escape, and window shutdown without
  applying stale callbacks to a newly focused field.
- [ ] Cover the session lifecycle with tests and verify composition with a
  desktop input method on each supported desktop OS.
- [x] Update the interaction, editing, and feature guides to name the current
  implementation and remaining limits.

Exit: a desktop user can compose and commit text in a focused ownframe field;
the stored value changes only on commit, and callbacks fire once per edit.

## Phase 3: Accessibility tree

- [x] Define a document-scoped accessibility node model for roles, names, values,
  bounds, enabled state, focus, and parent-child order.
- [x] Map supported HTML semantics from Blinkless's parsed source, and join
  bounds through unique authored IDs; omit generic decorative wrappers.
- [x] Expose a detached tree snapshot and focused-node identity to host
  adapters without coupling the page to a platform API.
- [x] Add action routing for focus, activate, and text value edits where the
  existing page API supports them.
- [x] Test names, roles, order, live state, bounds, and action routing.
- [x] Document unsupported semantics and the difference between the tree and
  PDF/UA output.
- [ ] Filter CSS-hidden nodes and map anonymous nodes to reliable layout bounds.

Exit: callers can inspect a current semantic tree for a page and invoke the
supported actions through its nodes.

### Selected API shape

The optional `host.Accessible` interface returns a detached
`AccessibilityTree` and accepts one document-scoped `AccessibilityRequest`.
The page builds semantics from the executed template before `syncForm`
rewrites controls. It uses Blinkless's exported `markup.Parse` tree for real
HTML ancestry, then reads live field values and joins layout bounds only by a
unique HTML ID. This keeps parentage out of hit-box geometry and avoids a
different HTML parser. Candidate 2 supplied the parser and source model.
The implementation uses candidate 1's single action-request shape and its
path IDs, duplicate-ID guards, and password redaction. Candidate 3 correctly
identified CSS visibility and anonymous bounds as engine limits, but missed
the exported `markup.Parse` tree in the pinned version. The API stays a page
and host contract. Native OS adapters remain phase 4 work.

## Phase 4: Native accessibility adapters

- [ ] Choose and document the supported desktop platforms and adapter
  boundary.
- [ ] Expose the tree to Linux assistive technology through AT-SPI.
- [ ] Expose the tree to Windows UI Automation.
- [ ] Expose the tree to macOS NSAccessibility.
- [ ] Verify role, name, focus, bounds, and actions with a screen reader or
  platform accessibility inspector on each supported OS.
- [ ] Keep the tree usable by non-window hosts even when no native adapter is
  available.

Exit: a screen reader can inspect and operate the supported controls in the
running desktop window on each platform claimed by the documentation.

## Phase 5: Desktop shell integrations

- [ ] Design multi-window ownership, per-window page state, and close handling.
- [ ] Add native menu hooks while retaining the current in-window context
  menu as a portable fallback.
- [ ] Add system tray lifecycle and click handling on supported desktop OSes.
- [ ] Add desktop notification requests and permission/failure behavior.
- [ ] Test window shutdown, menu actions, tray teardown, and notification
  failures with platform fakes where possible.
- [ ] Document which integrations are native and which are portable fallbacks.

Exit: each shipped integration has a runnable example and an OS-specific
verification note.

## Phase 6: Navigation and text services

- [ ] Define a host callback for deep links and validate accepted URL schemes.
- [ ] Add opt-in OS-global shortcut registration with explicit unregister and
  conflict handling.
- [ ] Decide whether spellcheck belongs in ownframe or in an application
  adapter; record the API and platform limits before implementation.
- [ ] Add tests for dispatch, focus/lifecycle changes, and unsupported hosts.

Exit: supported integrations have documented registration, dispatch, and
shutdown behavior; unsupported items remain explicitly deferred.

## Phase 7: Browser APIs, planning only

- [ ] Canvas: define whether the target is a Go image surface, a Canvas 2D
  compatibility API, or both; specify sizing, invalidation, and animation.
- [ ] Video: define decoder ownership, supported formats, frame pacing, audio
  synchronization, and how applications supply or decode frames.
- [ ] WebGL: decide whether exposing Ebiten's graphics backend is possible,
  and specify resource ownership and portability before proposing an API.
- [ ] Keep JavaScript, DOM mutation, and browser compatibility claims outside
  these proposals unless a separate runtime project is approved.

Exit: each proposal has an owner, dependency list, no-module impact, and
acceptance criteria. No implementation is included in this cycle.

## Phase 8: Distribution and updater ownership

- [ ] State that update checks, signing, and installation are owned by the
  application or its distribution channel.
- [ ] Document the hooks an application can use to report an update and
  restart safely, if any library support is needed.
- [ ] Keep packaging scripts focused on building and checksumming archives.

Exit: the feature guide names the owner and does not imply that packaging
provides automatic updates.

## Working rules

- Finish one phase slice, run its focused checks, and then update its rows.
- Run `gofmt` on changed Go files and `make test` before closing an
  implementation phase.
- Every Go file stays at or below 2000 characters.
- Keep video, canvas, and WebGL rows open and planning-only for this work.
