# v0.0.4 / 03 - Reusable component interactions

> **Parent:** [v0.0.4 index](README.md)
> **Status:** planned. Existing helpers are inputs to the design.
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

- [ ] `COMP-01` Inventory tabs, sliders, and dialogs in `examples/builder/ui-builder` and `examples/builder/loglens`, including keyboard, pointer, focus, and dismissal handlers.
- [ ] `COMP-02` Identify duplicated mechanics and specify which behavior remains application policy, including state updates and dialog dismissal rules.
- [ ] `COMP-03` Write caller usage for an install-once interaction API using the existing public input hooks; preserve application handlers and builder callbacks.
- [ ] `COMP-04` Define registration, update, removal, and teardown behavior so rebuilding a view cannot leave duplicate handlers or stale targets.

Keep selected tabs, slider values, and dialog open state in the host. Avoid
adding root `Page` forwarding methods just to move component wiring. A control
must not receive the same key or pointer action twice through composed handlers.

## Phase 2: Complete tab keyboard behavior

### 2.1 Focus and selection

- [ ] `COMP-05` Implement standard tab arrow navigation and Home/End through the chosen interaction contract, with a documented activation policy.
- [ ] `COMP-06` Keep focus, the active tab, `tabindex`, and ARIA state consistent after selection, rebuild, or removal.
- [ ] `COMP-07` Migrate the inventory's tab callers to shared mechanics and remove their duplicated traversal code.
- [ ] `COMP-08` Verify actual keyboard input selects or focuses the intended tab, respects the documented boundaries, and leaves unrelated application keys working.

## Phase 3: Complete slider pointer and keyboard behavior

### 3.1 Value changes

- [ ] `COMP-09` Reuse `SliderValueAt` to implement pointer press and drag updates with coordinates transformed into the slider track's local space.
- [ ] `COMP-10` Implement arrow increments and Home/End using Min, Max, and Step while preserving host ownership of Value.
- [ ] `COMP-11` Handle disabled controls, release outside the track, cancellation, scrolling, and removal during drag without stale interaction state.
- [ ] `COMP-12` Migrate the inventory's slider callers and remove duplicated pointer and keyboard mechanics.
- [ ] `COMP-13` Verify pointer and keyboard changes reach the host callback and match the painted thumb and advertised value after a rebuild.

## Phase 4: Consolidate dialog focus and dismissal

### 4.1 Lifecycle

- [ ] `COMP-14` Integrate the existing `DialogFocus` behavior with the chosen installation contract for initial focus, Tab containment, and focus restoration.
- [ ] `COMP-15` Route Escape and configured backdrop dismissal through the controlled open-state callback without leaking the action to underlying controls.
- [ ] `COMP-16` Define behavior when the trigger or focused control disappears, and explicitly retain or revise the current nested-modal restriction.
- [ ] `COMP-17` Migrate the inventory's dialog callers and remove duplicate focus or dismissal wiring.
- [ ] `COMP-18` Verify open, Tab, Shift+Tab, Escape, close, and reopen using real input, including a removed trigger and a rebuild while open.

## Phase 5: Validate the public caller experience

### 5.1 Closure gates

- [ ] `COMP-19` Update `components/chadcn/README.md` with complete install-once examples and the host-owned state contract.
- [ ] `COMP-20` Verify the teaching example uses public APIs and contains no duplicated tab, slider, or dialog mechanics supplied by the new contract.
- [ ] `COMP-21` Run `make internal-import-check` and retain its result.
- [ ] `COMP-22` Run `make lint` on the final implementation and record its exit status.
- [ ] `COMP-23` Run `make test` on the final implementation and record its exit status.
- [ ] `COMP-24` Run `make graphics-check` after paint or window changes and retain the result; record an explicit scope reason if no such change is made.
- [ ] `COMP-25` Retain native interaction captures and observed results for all three controls, with the source commit and host recorded.

## Dependencies

- Track 02 owns Linux assistive-technology action delivery. Semantic component
  attributes alone do not establish working native accessibility.
- Preserve `ui.View.Handle` callback composition and incremental rebuild behavior.
- Nested modals and additional control families are separate decisions. Do not
  silently expand this track to redesign the whole component library.
