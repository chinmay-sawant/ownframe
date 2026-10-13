# Phase 2: Retained DOM / Element Model

Checklist for retaining element identity between parses.

- [x] Add a retained node tree keyed by IDs and template structure.
- [x] Keep element instances alive across redraws when matching keys stay stable.
- [x] Distinguish create, move, update, and delete operations.
- [x] Add immutable node patch records so page-level code can react.
- [x] Add event capture, bubble, and delegation semantics.
- [x] Add synthetic events covering keyboard, pointer, focus, and IME.
- [x] Add component lifecycle events: mount, update, unmount.
- [x] Separate element metadata from CSS layout and painted pixels.
- [x] Expose retained element handles on the page API.
- [x] Ensure stale IDs and removed elements are disposed before paint.

## Acceptance criteria

- A simple button keeps the same computed action and styles without re-hit testing.
- Focus, hover, and disabled states remain attached to the same element instance.
- Deleted list rows disappear immediately and leave no ghost in staging.