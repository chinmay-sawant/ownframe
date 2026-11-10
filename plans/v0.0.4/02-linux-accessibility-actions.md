# v0.0.4 / 02 - Linux accessibility actions

> **Parent:** [v0.0.4 index](README.md)
> **Status:** planned. Live action delivery is unimplemented.
> **Estimated effort:** size after the bus request lifecycle and window wake mechanism are traced. Closure requires a live Linux accessibility session.

---

## Overview

Connect Linux AT-SPI focus and activation requests to the page's existing action
implementation while preserving page-loop ownership and truthful replies.

## Executive Summary

At `085ba7ea`, `internal/atspi/serve.go` returns false for `GrabFocus` and
`DoAction`. `internal/page/a11y_wire.go` already provides
`AccessibilityPerform` on the page-owning loop. `internal/window/a11y_atspi.go`
publishes snapshots through a worker so bus traffic does not stall rendering.
The new inbound path must respect that separation.

## Phase 1: Define the action and lifecycle contract

### 1.1 Ownership

- [ ] `A11Y-01` Trace D-Bus request handling, snapshot publication, page actions, window wakeups, and shutdown; record which goroutine owns each mutable object.
- [ ] `A11Y-02` Define the internal queued request and completion contract for focus and activation, including bounded capacity, cancellation, timeout, and shutdown outcomes.
- [ ] `A11Y-03` Define how an AT-SPI object and action index resolve to a current page target; reject removed targets and unsupported actions without invoking another control.
- [ ] `A11Y-04` Define reply semantics so success means the page accepted the action, rather than merely admitting a request to the queue.

The proposed flow is a bus request, a bounded queue, a window wakeup, current
target validation, `AccessibilityPerform`, and the actual result. Page mutation
stays on the window loop. Do not hold a publisher or bus lock while awaiting
page completion. A cancelled request must not remain queued for later mutation.

## Phase 2: Implement delivery through the existing page API

### 2.1 Queue and execution

- [ ] `A11Y-05` Add internal request delivery between `internal/atspi` and `internal/window` without exporting a host loop or inner page.
- [ ] `A11Y-06` Wake an idle window when a request arrives so action processing does not depend on incidental user input.
- [ ] `A11Y-07` Drain admitted requests on the owning loop and invoke the existing accessibility action path after current-target validation.
- [ ] `A11Y-08` Route `GrabFocus` and supported `DoAction` calls to delivery, returning false or the documented error for refused requests.
- [ ] `A11Y-09` Cancel pending requests and release waiters when the window or bus closes, with no action executed after cancellation wins the execution handoff.
- [ ] `A11Y-10` Publish resulting focus and control-state changes through the existing snapshot and event path.

## Phase 3: Prove failure and concurrency behavior

### 3.1 Automated checks

- [ ] `A11Y-11` Add bus tests for successful focus and activation, invalid action indexes, disabled controls, removed targets, and unsupported operations.
- [ ] `A11Y-12` Add queue tests for saturation, timeout, cancellation before execution, cancellation during execution, and shutdown while replies wait.
- [ ] `A11Y-13` Add a host integration test proving an action wakes the idle loop and changes focus or application state through the existing page path.
- [ ] `A11Y-14` Verify password redaction and confirm action diagnostics do not log accessible text.
- [ ] `A11Y-15` Run `go test -race -p 1 ./internal/atspi ./internal/page ./internal/window` and retain the successful result.

## Phase 4: Validate live behavior and document the exact scope

### 4.1 Closure gates

- [ ] `A11Y-16` Run a live Linux session following `documentation/atspi.md`; record the host, source commit, accessibility tools, and bus setup.
- [ ] `A11Y-17` Invoke focus through a live AT-SPI client and observe focus in the application.
- [ ] `A11Y-18` Invoke activation through a live AT-SPI client and observe the intended application state change.
- [ ] `A11Y-19` Update `documentation/atspi.md`, `documentation/accessibility-native.md`, and the feature guide with proven action behavior and remaining limits.
- [ ] `A11Y-20` Run `make lint` on the final implementation and record its exit status.
- [ ] `A11Y-21` Run `make test` on the final implementation and record its exit status.
- [ ] `A11Y-22` Run `make graphics-check` for the window changes and record its exit status.

## Dependencies

- A Linux desktop with a working accessibility bus is required for live closure.
  Keep those rows open if that host is unavailable.
- This track adds focus and activation delivery. EditableText, link activation,
  per-glyph text geometry, CSS-hidden-node filtering, and other platform adapters
  require their own scope and proof. Do not advertise them as completed here.
- Preserve existing password restrictions and the public `Page.Access()` door.
