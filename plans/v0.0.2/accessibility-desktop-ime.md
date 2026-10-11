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
- [x] Verify content clicks, page scrolling, and scrollbar dragging do not
  consume one another's events in focused window tests.
- [x] Document the pointer behavior and add targeted window package tests.

The focused pointer tests pass. Running-window behavior still needs a manual
check; the phase exit remains open until that and `make test` pass.

Exit: hovering and dragging either thumb works in the running window, with
targeted pointer tests and `make test` passing.

## Phase 2: Desktop IME

- [x] Start an Ebiten text-input session for the focused text field on desktop.
- [x] Place the candidate window from the focused field's transformed bounds
  after scroll, zoom, and resize changes.
- [x] Place the candidate window at the glyph-level caret rather than the
  focused field box.
- [x] Keep preedit text separate from the committed field value; draw its
  selection and underline without firing field change callbacks.
- [x] Apply commits and reconversion through the existing edit path exactly
  once, including replacement around the caret.
- [x] End or cancel sessions on blur, Escape, and window shutdown without
  applying stale callbacks to a newly focused field.
- [x] Cover caret bounds and session lifecycle with focused tests.
- [ ] Verify composition with a desktop input method on each supported
  desktop OS.
- [x] Update the interaction, editing, and feature guides to name the current
  implementation and remaining limits.

Exit: a desktop user can compose and commit text in a focused ownframe field;
the stored value changes only on commit, and callbacks fire once per edit.
Focused tests cover caret geometry and session cleanup. Real input-method
composition on each supported desktop OS remains an open exit check.

## Phase 3: Accessibility tree

- [x] Define a document-scoped accessibility node model for roles, names, values,
  bounds, enabled state, focus, and parent-child order.
- [x] Map supported HTML semantics from Blinkless's parsed source, and join
  bounds through unique authored IDs; omit generic decorative wrappers.
- [x] Expose a detached tree snapshot and focused-node identity to host
  adapters without coupling the page to a platform API.
- [x] Add action routing for focus, activate, and text value edits where the
  existing page API supports them.
- [x] Add a runnable example for desktop composition, tree inspection, and a
  tree-driven text edit.
- [x] Test names, roles, order, live state, bounds, and action routing.
- [x] Document unsupported semantics and the difference between the tree and
  PDF/UA output.
- [x] Map anonymous nodes to conservative layout bounds when the rendered
  boxes match unambiguously.
- [ ] Filter CSS-hidden nodes after Blinkless exposes computed visibility.

Anonymous bounds are omitted when element-to-box matching is ambiguous.
Computed CSS visibility remains unavailable from the pinned Blinkless API.

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

- [ ] Choose the supported desktop platforms for native accessibility.
- [x] Document the native adapter boundary and validation route for AT-SPI,
  UI Automation, and NSAccessibility.
- [ ] Expose the tree to Linux assistive technology through AT-SPI.
- [ ] Expose the tree to Windows UI Automation.
- [ ] Expose the tree to macOS NSAccessibility.
- [ ] Verify role, name, focus, bounds, and actions with a screen reader or
  platform accessibility inspector on each supported OS.
- [x] Keep the tree usable by non-window hosts even when no native adapter is
  available.

Exit: a screen reader can inspect and operate the supported controls in the
running desktop window on each platform claimed by the documentation.

## Phase 5: Desktop shell integrations

- [x] Document multi-window ownership, per-window page state, and close
  handling as a design proposal.
- [ ] Add native menu hooks while retaining the current in-window context
  menu as a portable fallback.
- [ ] Add system tray lifecycle and click handling on supported desktop OSes.
- [ ] Add desktop notification requests and permission/failure behavior.
- [ ] Test window shutdown, menu actions, tray teardown, and notification
  failures with platform fakes where possible.
- [x] Document which integrations are native and which are portable fallbacks.

Exit: each shipped integration has a runnable example and an OS-specific
verification note.

## Phase 6: Navigation and text services

- [x] Define a host callback for deep links and validate accepted URL schemes.
- [x] Define an application callback for clicked `http`, `https`, and `mailto`
  links.
- [ ] Verify OS protocol registration and cold-start/later URL delivery in
  application adapters.
- [ ] Add opt-in OS-global shortcut registration with explicit unregister and
  conflict handling.
- [x] Decide that spellcheck belongs in an application or platform adapter;
  document the API and platform limits.
- [x] Test deep-link dispatch, scheme validation, clicked-link handoff, and
  IME focus/lifecycle cancellation.

Exit: supported integrations have documented registration, dispatch, and
shutdown behavior; unsupported items remain explicitly deferred.

## Phase 7: Browser APIs, planning only

- [x] Canvas: define whether the target is a Go image surface, a Canvas 2D
  compatibility API, or both; specify sizing, invalidation, and animation.
- [x] Video: define decoder ownership, supported formats, frame pacing, audio
  synchronization, and how applications supply or decode frames.
- [x] WebGL: decide whether exposing Ebiten's graphics backend is possible,
  and specify resource ownership and portability before proposing an API.
- [x] Keep JavaScript, DOM mutation, and browser compatibility claims outside
  these proposals unless a separate runtime project is approved.

Exit: each proposal has an owner, dependency list, no-module impact, and
acceptance criteria. No implementation is included in this cycle.

## Phase 8: Distribution and updater ownership

- [x] State that update checks, signing, and installation are owned by the
  application or its distribution channel.
- [x] Document that update checks, installation, and restart policy belong
  to the app or its distribution channel; no library updater hook is needed.
- [x] Keep packaging scripts focused on building and checksumming archives;
  document that checksums do not authenticate release publishers.

Exit: the feature guide names the owner and does not imply that packaging
provides automatic updates.

## Working rules

- Finish one phase slice, run its focused checks, and then update its rows.
- Run `gofmt` on changed Go files and `make test` before closing an
  implementation phase.
- Every Go file stays at or below 2000 characters.
- Close the Canvas, video, and WebGL planning rows only when each proposal
  records ownership, dependencies, module impact, and acceptance criteria.
  Those rows do not mean the APIs are implemented.
