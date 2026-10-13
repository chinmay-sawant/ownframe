# Phase 5: Runtime Behaviors and Resilience

Checklist for browser-like UX after patching is available.

- [x] Add async loading markers in place of whole-page fallback.
- [x] Add suspense-style placeholders for deferred data.
- [x] Add CSS transitions and phase transitions where supported.
- [x] Add component-scoped clipping and scrolling containers.
- [x] Add focus restoration across partial DOM updates.
- [x] Add unhandled promise/state errors surfaced as page errors.
- [x] Add observer-like notifications for visibility and resize.
- [x] Add devtools counters for patched vs repainted operations.
- [x] Add visual tests comparing full vs partial repaint outputs.
- [x] Add watchdog frequency sampling to catch feedback loops.
- [x] Document migrate-in instructions for existing template apps.
- [x] Keep accessibility fallback tied to retained element state.

## Expanded helpers (phase 5 follow-up)

- [x] Add visual compare hooks for full vs partial repaint bytes.
- [x] Add watchdog frequency sampling helpers and reset support.
- [x] Expand transition and phase transition helper coverage.
- [x] Expand observer helper coverage for visibility and resize.
- [x] Add accessibility and focus restoration helper coverage.
- [x] Add runtime behavior documentation and migration notes.

## Acceptance criteria

- Apps can show "loading slowly" frames instead of blank repaints.
- A repaint failure falls back to full paint with logged diagnostics.
- Existing performance tests remain useful and catch regressions.