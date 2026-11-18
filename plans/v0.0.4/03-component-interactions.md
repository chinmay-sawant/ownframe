# v0.0.4 / 03 - Reusable component interactions

> **Parent:** [v0.0.4 index](README.md)
> **Status:** complete. Shared interactions, caller migrations, and closure gates passed.
> **Estimated effort:** size tabs, slider, and dialog slices after caller inventory. Validate each slice before starting the next.

---

## Overview

Keep component values in the host application while consolidating repeated
keyboard, pointer, and focus mechanics in `components/chadcn`.

## Executive Summary

At `085ba7ea`, `tabs.go` delegates keyboard traversal, `slider.go` provides
`SliderValueAt` for host pointer handling, and `dialog.go` provides `DialogFocus`
for focus containment and restoration. Those helpers exist and must be reused
or deliberately replaced. Controlled state is an intentional design choice.
The goal is less repeated interaction wiring, with an explicit ownership contract.

## Phase 1: Design around existing callers

### 1.1 Contracts and ownership

- [x] `COMP-01` Inventory tabs, sliders, and dialogs in `examples/builder/ui-builder` and `examples/builder/loglens`, including keyboard, pointer, focus, and dismissal handlers.
- [x] `COMP-02` Identify duplicated mechanics and specify which behavior remains application policy, including state updates and dialog dismissal rules.
- [x] `COMP-03` Write caller usage for an install-once interaction API using the existing public input hooks; preserve application handlers and builder callbacks.
- [x] `COMP-04` Define registration, update, removal, and teardown behavior so rebuilding a view cannot leave duplicate handlers or stale targets.

Keep selected tabs, slider values, and dialog open state in the host. Avoid
adding root `Page` forwarding methods just to move component wiring. A control
must not receive the same key or pointer action twice through composed handlers.

### 1.2 Caller inventory and contract

| Caller | Controls | Previous mechanics | Shared mechanics and remaining host policy |
|---|---|---|---|
| **ui-builder** | Catalog tabs; catalog and widgets sliders | `slider_drag.go` swapped drag handlers and stored a gesture in widgets state. Tabs supplied click callbacks. | `Interactions` routes keyboard and pointer input. Host callbacks still select the catalog tab, set volume/progress, update the notice, and rebuild. |
| **ui-builder modals** | Dialog demo, compare preview, catalog Dialog and AlertDialog | `app.go` manually activated `DialogFocus` and handled Escape for two modals; catalog modals lacked that shared focus wiring. | Controller derives enabled focus targets from the rendered panel. Existing CloseAction or OnOpenChange retains dismissal policy; AlertDialog backdrop remains inert. |
| **Loglens** | Main and drawer tabs; sample-rate, settings-threshold, alert-threshold sliders | Tabs supplied clicks; sample-rate clicks cycled values; other sliders depended on separate preset/nudge buttons. | Controller owns traversal and pointer mapping. Hosts own active screens, threshold/rate values, patching labels and fills, and preset actions. |
| **Loglens modals** | Create alert, delete view AlertDialog, on-call Dialog | Open flags and callbacks lived in the host, with a generic Escape handler and no shared focus scope. | One managed modal receives focus and Escape. Host callbacks own validation, confirmation, cancellation, toast messages, and open flags. Other overlay families retain their host handlers. |

`components/chadcn/interactions.go` owns one view's registrations and lifecycle.
Constructors register during Mount/Update. A fresh snapshot replaces the old one
only after the builder update succeeds. Stable IDs identify controls. Removed or
disabled sliders cancel their gesture; removed focus targets move to the selected
surviving tab or the first normal control after a modal closes. `Close` restores
focus and leaves application callbacks usable. No new root Page API or internal
example import is needed. The install-once usage is in
[Shared interactions](../../components/chadcn/README.md#shared-interactions).

Horizontal tabs use automatic activation and wrapping Left/Right, Home/End.
Slider arrows use the declared Step and Home/End clamp to the declared range.
Pointer mapping uses the actual rail, including a press before drag recognition.
Hosts remain responsible for applying callback values and publishing the matching
paint and ARIA attributes. Dialog focus uses the existing DialogFocus contract;
simultaneous or nested managed modals are rejected.

The page focus collector now recognizes explicit generic tabindex targets,
without treating them as editable text or swallowing nested fields. This is
required for the slider and tab panel semantics to be usable by keyboard.

The input conventions follow the W3C patterns for
[tabs](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/),
[sliders](https://www.w3.org/WAI/ARIA/apg/patterns/slider/), and
[modal dialogs](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/).

## Phase 2: Complete tab keyboard behavior

### 2.1 Focus and selection

- [x] `COMP-05` Implement standard tab arrow navigation and Home/End through the chosen interaction contract, with a documented activation policy.
- [x] `COMP-06` Keep focus, the active tab, `tabindex`, and ARIA state consistent after selection, rebuild, or removal.
- [x] `COMP-07` Migrate the inventory's tab callers to shared mechanics and remove their duplicated traversal code.
- [x] `COMP-08` Verify actual keyboard input selects or focuses the intended tab, respects the documented boundaries, and leaves unrelated application keys working.

## Phase 3: Complete slider pointer and keyboard behavior

### 3.1 Value changes

- [x] `COMP-09` Reuse `SliderValueAt` to implement pointer press and drag updates with coordinates transformed into the slider track's local space.
- [x] `COMP-10` Implement arrow increments and Home/End using Min, Max, and Step while preserving host ownership of Value.
- [x] `COMP-11` Handle disabled controls, release outside the track, cancellation, scrolling, and removal during drag without stale interaction state.
- [x] `COMP-12` Migrate the inventory's slider callers and remove duplicated pointer and keyboard mechanics.
- [x] `COMP-13` Verify pointer and keyboard changes reach the host callback and match the painted thumb and advertised value after a rebuild.

## Phase 4: Consolidate dialog focus and dismissal

### 4.1 Lifecycle

- [x] `COMP-14` Integrate the existing `DialogFocus` behavior with the chosen installation contract for initial focus, Tab containment, and focus restoration.
- [x] `COMP-15` Route Escape and configured backdrop dismissal through the controlled open-state callback without leaking the action to underlying controls.
- [x] `COMP-16` Define behavior when the trigger or focused control disappears, and explicitly retain or revise the current nested-modal restriction.
- [x] `COMP-17` Migrate the inventory's dialog callers and remove duplicate focus or dismissal wiring.
- [x] `COMP-18` Verify open, Tab, Shift+Tab, Escape, close, and reopen using real input, including a removed trigger and a rebuild while open.

## Phase 5: Validate the public caller experience

### 5.1 Closure gates

- [x] `COMP-19` Update `components/chadcn/README.md` with complete install-once examples and the host-owned state contract.
- [x] `COMP-20` Verify the teaching example uses public APIs and contains no duplicated tab, slider, or dialog mechanics supplied by the new contract.
- [x] `COMP-21` Run `make internal-import-check` and retain its result.
- [x] `COMP-22` Run `make lint` on the final implementation and record its exit status.
- [x] `COMP-23` Run `make test` on the final implementation and record its exit status.
- [x] `COMP-24` Run `make graphics-check` after paint or window changes and retain the result; record an explicit scope reason if no such change is made.
- [x] `COMP-25` Retain native interaction captures and observed results for all three controls, with the source commit and host recorded.

## Validation evidence

Source commit `0a77375ad10c51c8f9dab2d9c075701e7a2249bb` plus the local
implementation. [Native provenance](evidence/components-native.json) records
tracked and new source hashes, Go and machine details, input scope, and capture
observations. This is Linux native Ebitengine under Xvfb with software OpenGL;
it does not establish physical-device input or other desktop hosts.

| Command | Exit | Retained result |
|---|---|---|
| **make test** | 0 | [Library and example suite](evidence/components-tests.txt), including registration failure, removed tabs/triggers, modal reopen, slider clamping and cancellation, and migrated caller tests. |
| **make lint** | 0 | [All gates](evidence/components-lint.txt), including internal imports and documentation snippets. |
| **make graphics-check** | 0 | [15 required scenarios](evidence/components-graphics.txt), including the new dedicated native interaction process. |

The native fixture runs input on the update loop and captures the real
framebuffer after each asserted state: [initial](evidence/components-0.png),
[tab selection](evidence/components-1.png), [slider drag](evidence/components-2.png),
[open dialog](evidence/components-3.png), [restored focus](evidence/components-4.png).
The focused slider uses a 4 px corner radius so its focus outline remains a
rectangle around the control. Reproduce captures with
`OWNFRAME_INTERACTION_CAPTURE_DIR=/absolute/output/path make graphics-check`.
`TestManagedInteractionsNative` is also required by the graphics gate inventory
test, so silently dropping or skipping it fails the gate.

## Dependencies

- Track 02 owns Linux assistive-technology action delivery. Semantic component
  attributes alone do not establish working native accessibility.
- Preserve `ui.View.Handle` callback composition and incremental rebuild behavior.
- Nested modals and additional control families are separate decisions. Do not
  silently expand this track to redesign the whole component library.
