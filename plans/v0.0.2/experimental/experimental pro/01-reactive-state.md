# Phase 1: Reactive State

Checklist of app/runtime work needed before DOM patching.

- [ ] Add field-level binding events instead of full `SetData` rebuilds.
- [ ] Allow handlers to patch only projected state slices, not the whole `View`.
- [ ] Introduce stable IDs for repeated items so list diffs remain deterministic.
- [ ] Keep focus and scroll state across state-driven re-renders.
- [ ] Preserve caret/selection when the focused field does not change.
- [ ] Add route interceptors that do not trigger a full document load.
- [ ] Keep Undo/Redo references attached to stable element identity, not box hits.
- [ ] Surface changed-property metadata from handlers to the page layer.
- [ ] Unit test each state transition without render round-trips.
- [ ] Record which update came from user input vs. background state.

## Acceptance criteria

- A click that only changes one scalar value recomputes the minimal changed region.
- Existing simple apps continue to call `SetData` and work unchanged.