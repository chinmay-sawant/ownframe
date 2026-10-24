# Phase 01 - inventory and guardrails

Goal: the migration has a written inventory and a check that stops the next
hand-written driver before it lands.

## Where it lives

- `scripts/example-internal-allowlist.txt` lists 40 pairs, including three rows
  for `examples/animations/page.go` added by commit `e48b91af`.
- The file header says "Do not add rows for new work", and AGENTS.md says no new
  example import of `internal/`. The three rows contradict both.
- `scripts/check-example-internal.sh` compares the set of imports with the
  allowlist. It passes today because the rows are listed.
- `examples/maintenance.json` classifies `animations` as `experimental`.
- `examples/readme.md` does not mention `examples/animations`.

## Design

1. Add the inventory table from the track README to `examples/readme.md`, with a
   `motion driver` column: `auto`, `manual`, `none`, or `data-driven`.

2. Add `scripts/check-motion-imports.sh`, called from `make examples-check`, that
   fails when a file under `examples/` imports any of:
   `internal/cssanimation`, `internal/csstransition`, `internal/animation`,
   `internal/transition`, `internal/csstrigger`, `internal/motion`. The
   allowlist does not exempt it. This is a ratchet: the current rows fail, so the
   script lands with a `--baseline` mode that lists today's three rows, and the
   list shrinks to empty as the track lands.

3. `cmd/animation-preview` and `cmd/transition-preview` are exempt because they
   live in the library module, do not ship to users, and exist to exercise the
   low-level sampler. Say so in a comment in the script and in
   `examples/readme.md`.

4. Freeze the allowlist. Add a comment that no motion row may be added, and make
   the new check the enforcement rather than the allowlist.

5. Record which examples legitimately keep a `SetTick`: Spotify (audio position),
   dino, eyes, flappy-bird, tetris, departure-board, perf-benchmarks. Each gets a
   one-line reason in the readme. A reviewer can then tell a game loop from a
   driver at a glance.

## Checklist

- [ ] Inventory table in `examples/readme.md`.
- [ ] `scripts/check-motion-imports.sh` with a baseline mode.
- [ ] Wire it into `make examples-check` and CI.
- [ ] Exempt `cmd/*-preview` with a written reason.
- [ ] Freeze the allowlist with a comment.
- [ ] List the legitimate `SetTick` examples with reasons.
- [ ] Run the check and confirm it fails on today's three rows, then passes with
      the baseline.

Gate: `make examples-check` fails if a new example imports a motion internal
package, and the failure message names the file and the package.

## Risks

- A ratchet that is easy to bypass is worse than no ratchet. Put the check in
  `make lint` as well as `make examples-check`.
- The baseline list must shrink, not grow. Fail the check if the baseline count
  increases.
