# Phase 2: Retained DOM / Element Model

Checklist for retaining element identity between parses.

- [ ] Add a retained node tree keyed by IDs and template structure.
- [ ] Keep element instances alive across redraws when matching keys stay stable.
- [ ] Distinguish create, move, update, and delete operations.
- [ ] Add immutable node patch records so page-level code can react.
- [ ] Add event capture, bubble, and delegation semantics.
- [ ] Add synthetic events covering keyboard, pointer, focus, and IME.
- [ ] Add component lifecycle events: mount, update, unmount.
- [ ] Separate element metadata from CSS layout and painted pixels.
- [ ] Expose retained element handles on the page API.
- [ ] Ensure stale IDs and removed elements are disposed before paint.

## Acceptance criteria

- A simple button keeps the same computed action and styles without re-hit testing.
- Focus, hover, and disabled states remain attached to the same element instance.
- Deleted list rows disappear immediately and leave no ghost in staging.