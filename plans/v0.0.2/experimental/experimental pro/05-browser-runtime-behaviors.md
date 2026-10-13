# Phase 5: Runtime Behaviors and Resilience

Checklist for browser-like UX after patching is available.

- [ ] Add async loading markers in place of whole-page fallback.
- [ ] Add suspense-style placeholders for deferred data.
- [ ] Add CSS transitions and phase transitions where supported.
- [ ] Add component-scoped clipping and scrolling containers.
- [ ] Add focus restoration across partial DOM updates.
- [ ] Add unhandled promise/state errors surfaced as page errors.
- [ ] Add observer-like notifications for visibility and resize.
- [ ] Add devtools counters for patched vs repainted operations.
- [ ] Add visual tests comparing full vs partial repaint outputs.
- [ ] Add watchdog frequency sampling to catch feedback loops.
- [ ] Document migrate-in instructions for existing template apps.
- [ ] Keep accessibility fallback tied to retained element state.

## Acceptance criteria

- Apps can show "loading slowly" frames instead of blank repaints.
- A repaint failure falls back to full paint with logged diagnostics.
- Existing performance tests remain useful and catch regressions.