# v0.0.4 / 02 - Linux accessibility actions

> **Parent:** [v0.0.4 index](README.md)
> **Status:** all four phases complete. Linux live focus and activation are validated.
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

- [x] `A11Y-01` Trace D-Bus request handling, snapshot publication, page actions, window wakeups, and shutdown; record which goroutine owns each mutable object. Evidence: ownership and lifecycle are recorded in section 1.2 below, traced through `live_linux.go`, `action_queue.go`, and `a11y_atspi.go`.
- [x] `A11Y-02` Define the internal queued request and completion contract for focus and activation, including bounded capacity, cancellation, timeout, and shutdown outcomes. Evidence: `atspi.Actions` has bounded admission, cancellation handoff, a per-frame batch, and shutdown cancellation; queue tests pass.
- [x] `A11Y-03` Define how an AT-SPI object and action index resolve to a current page target; reject removed targets and unsupported actions without invoking another control. Evidence: the mapped document and node identity reach `AccessibilityPerform`; `TestA11yActionsUsePageLoopAndRejectRemovedTargets` rejects a removed target without another activation.
- [x] `A11Y-04` Define reply semantics so success means the page accepted the action, rather than merely admitting a request to the queue. Evidence: `TestActionQueueWaitsForActualResult` proves admission does not return success and the page error reaches the requester.

The implemented flow is a bus request, a bounded queue, a window wakeup, current
target validation, `AccessibilityPerform`, and the actual result. Page mutation
stays on the window loop. Do not hold a publisher or bus lock while awaiting
page completion. A cancelled request must not remain queued for later mutation.

### 1.2 Action ownership and validation record

| Owner | Responsibility |
|---|---|
| **D-Bus reader** | Decode traffic and dispatch bounded action reply workers; keep publication replies and ordinary queries moving. |
| **Reply worker** | Resolve the published object and action index, retain its document/node identity, and await the actual page result with a two-second deadline. |
| **Action queue** | Admit up to 16 pending calls, wake the window, and arbitrate cancellation against the execution handoff. |
| **Window loop** | Execute up to four requests per Update through the current page accessibility API, then publish resulting state through normal sync. |
| **Shutdown** | Cancel waiting and executing contexts, reject new admission, and release bus workers without holding publisher locks while awaiting the page. |

Cancellation before the execution handoff prevents mutation. After the handoff,
cancellation is cooperative through the page context and does not undo an edit.
A successful reply means execution completed without error, not that the queue
accepted the request. Text editing is not dispatched by this bridge.

The source base is `0a77375ad10c51c8f9dab2d9c075701e7a2249bb` plus the local
track 01 and track 02 changes. The live JSON records the dirty-source digest and
hashes of untracked source files. It contains no accessible text or field values.
The session used the actual Linux accessibility bus and a native window on
WSLg DISPLAY :0. The separate graphics run uses isolated Xvfb sessions. Neither
result establishes Orca speech, Windows UIA, VoiceOver, or TalkBack behavior.

Final commands, all exit 0, from the absolute OwnFrame checkout:

```sh
go test -race -p 1 ./internal/atspi ./internal/page ./internal/window
python3 scripts/check-atspi-actions.py --out temp/v004-performance/atspi-live.json
make lint
make test
make graphics-check
```

## Phase 2: Implement delivery through the existing page API

### 2.1 Queue and execution

- [x] `A11Y-05` Add internal request delivery between `internal/atspi` and `internal/window` without exporting a host loop or inner page. Evidence: `NewLiveBus(ActionHandler)` connects the bus to the internal action queue; no root Page or host-loop API was added.
- [x] `A11Y-06` Wake an idle window when a request arrives so action processing does not depend on incidental user input. Evidence: `TestWindowAccessibilityActionWake` passes inside the real paced game and paints the activated control before its later page deadline.
- [x] `A11Y-07` Drain admitted requests on the owning loop and invoke the existing accessibility action path after current-target validation. Evidence: `runA11yActions` runs in the window Update sequence; the host integration test focuses and activates through the existing page API.
- [x] `A11Y-08` Route `GrabFocus` and supported `DoAction` calls to delivery, returning false or the documented error for refused requests. Evidence: `TestServeActionsDeliverFocusAndActivation` and `TestServeActionsRejectInvalidDisabledAndUnsupported` pass; the live check also observes both actions.
- [x] `A11Y-09` Cancel pending requests and release waiters when the window or bus closes, with no action executed after cancellation wins the execution handoff. Evidence: queue shutdown tests pass for queued and executing requests; bus disconnect and Close cancel reply-worker contexts.
- [x] `A11Y-10` Publish resulting focus and control-state changes through the existing snapshot and event path. Evidence: the live check observes focused state and the activation counter in the newly published tree.

## Phase 3: Prove failure and concurrency behavior

### 3.1 Automated checks

- [x] `A11Y-11` Add bus tests for successful focus and activation, invalid action indexes, disabled controls, removed targets, and unsupported operations. Evidence: `serve_action_test.go`, existing redaction tests, and `a11y_actions_test.go` cover focus, activation, invalid indexes, disabled controls, removed targets, and refused operations.
- [x] `A11Y-12` Add queue tests for saturation, timeout, cancellation before execution, cancellation during execution, and shutdown while replies wait. Evidence: `action_queue_test.go` covers saturation, deadlines, pre-execution cancellation, cancellation during execution, and shutdown; the reader test proves queries remain serviceable.
- [x] `A11Y-13` Add a host integration test proving an action wakes the idle loop and changes focus or application state through the existing page path. Evidence: `TestWindowAccessibilityActionWake` passes in the graphics gate; the action changes the framebuffer while the host was idle.
- [x] `A11Y-14` Verify password redaction and confirm action diagnostics do not log accessible text. Evidence: the live check confirms an empty password value; `scripts/check-atspi-actions.py` saves booleans and provenance without accessible text or field values.
- [x] `A11Y-15` Run `go test -race -p 1 ./internal/atspi ./internal/page ./internal/window` and retain the successful result. Evidence: exit 0 for the exact three-package race command; [retained output](evidence/atspi-race.txt).

## Phase 4: Validate live behavior and document the exact scope

### 4.1 Closure gates

- [x] `A11Y-16` Run a live Linux session following `documentation/atspi.md`; record the host, source commit, accessibility tools, and bus setup. Evidence: `python3 scripts/check-atspi-actions.py --out temp/v004-performance/atspi-live.json` exited 0 on WSL2 Linux, DISPLAY :0, using gdbus and the discovered accessibility bus; [record](evidence/atspi-live.json).
- [x] `A11Y-17` Invoke focus through a live AT-SPI client and observe focus in the application. Evidence: the live record reports both `focus_accepted` and `focus_state_observed` as true.
- [x] `A11Y-18` Invoke activation through a live AT-SPI client and observe the intended application state change. Evidence: the live record reports both `activation_accepted` and `activation_counter_changed` as true.
- [x] `A11Y-19` Update `documentation/atspi.md`, `documentation/accessibility-native.md`, and the feature guide with proven action behavior and remaining limits. Evidence: the AT-SPI, native accessibility, page accessibility, feature, platform, testing, and README references now describe supported focus/activation and retained text-editing limits.
- [x] `A11Y-20` Run `make lint` on the final implementation and record its exit status. Evidence: exit 0; [lint output](evidence/atspi-lint.txt).
- [x] `A11Y-21` Run `make test` on the final implementation and record its exit status. Evidence: exit 0; [full test output](evidence/atspi-tests.txt).
- [x] `A11Y-22` Run `make graphics-check` for the window changes and record its exit status. Evidence: exit 0; all 14 required Linux graphics checks passed, including the new accessibility wake test; [output](evidence/atspi-graphics.txt).

## Dependencies

- A Linux desktop with a working accessibility bus is required for live closure.
  Keep those rows open if that host is unavailable.
- This track adds focus and activation delivery. EditableText, link activation,
  per-glyph text geometry, CSS-hidden-node filtering, and other platform adapters
  require their own scope and proof. Do not advertise them as completed here.
- Preserve existing password restrictions and the public `Page.Access()` door.
