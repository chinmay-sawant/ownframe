# Runtime behavior helpers

Package `internal/runtimebehavior` holds browser-like UX helpers used by the
retained and partial update paths.

## Loading markers and suspense

- `LoadingMarker` tracks async loads and reports when they become slow.
- `SuspensePlaceholder` embeds a marker and resolves once deferred data arrives.

## Focus and errors

- `FocusRestorer` saves and restores caret and selection across partial updates.
- `PromiseError` surfaces state errors as page errors via `WrapPromiseError`.

## Transitions and clipping

- `Transition` and `PhaseTransition` describe CSS-like and mount/update/unmount phases.
- `ClipRect` provides component-scoped clipping and scrolling containers.

## Observers and watchdogs

- `Observer`, `VisibilityObserver`, and `ResizeObserver` notify about visibility and resize.
- `Watchdog` samples update frequency to catch feedback loops.
- `DevtoolsCounter` tracks patched vs repainted operations.

## Visual compare hooks

- `VisualComparer` (alias `VisualCompare`) stores full and partial repaint bytes.
- Compare with `Compare`, `CompareImages`, or `CompareImage`; register `OnMismatch`.

## Accessibility

- `AccessibilityFallback` keeps fallback state tied to retained element identity.
