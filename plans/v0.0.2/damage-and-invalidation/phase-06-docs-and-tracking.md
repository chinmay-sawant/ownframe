# Phase 06, docs and tracking

Goal: every phase's numbers live in one place, and the canonical docs describe
the shipped behavior instead of the old single-rect story.

## Checklist

- [ ] `documentation/repaint.md`: the damage list, the cap, the blit rule, and
      the measured numbers per phase. Keep the measured table current.
- [ ] `documentation/performance.md`: the DIRTY section wording follows the
      list, and a line that the departure-board dump exists and how to run it.
- [ ] `documentation/frames.md`: the retarget pattern from phase 02 next to the
      `UseFrameDirty` rule.
- [ ] `documentation/devtools.md`: which tab to read for FPS, and the dock's own
      cost note from phase 03.
- [ ] `examples/departure-board/PERFORMANCE.md`: before and after per phase,
      with the same command lines.
- [ ] `examples/readme.md` and `documentation/features-examples.md` if a new
      example, flag, or dump scenario appears.
- [ ] `../../../PHASES.md`: what shipped, with commit hashes.
- [ ] `../../../AGENTS.md`: the map line for a new folder or entry point.
- [ ] `../../README.md`: add this folder to the v0.0.2 plan table.

## Exit

- Each phase has one recorded before and after, produced by the same command on
  the same machine.
- The v0.0.2 README and `PHASES.md` point at this folder, and no doc still
  claims every frame with a tick replays in full.
