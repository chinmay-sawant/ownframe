# Phase 09: docs and gates

Goal: make the example part of the catalog and keep its evidence current.

Status: landed. The example README, catalog rows, screenshot matrix, fixture
regeneration notes, and gate run are in place.

## Design

1. Add `code-view` to `examples/maintenance.json` as experimental.
2. Add a row to the `examples/readme.md` table with the description and the
   `-web` port 8145.
3. Keep the example README as the source of run flags, keys, languages, and
   limits.
4. Document how to regenerate `fixtures/sample.diff`: create the old tree,
   commit it in a temporary repository, copy the new sample over, and run
   `git diff --cached HEAD`.
5. Record the screenshot matrix at 1280x720, 1500x900, and 1920x1080 for the
   explorer, changes, split, details, and scrolled views.
6. Run the full gates in one pass.
7. Audit the example against `AGENTS.md`: no `ownframe/internal/` import, no
   script tag, no canvas, and the public API only.

## Checklist

- [x] Add the catalog row and the maintenance entry. Evidence:
      `examples/maintenance.json`, `examples/readme.md`.
- [x] Write the example README.
- [x] Add the phase plan and the master checklist.
- [x] Document the diff fixture regeneration. Evidence:
      `examples/code-view/README.md`.
- [x] Record the screenshot matrix. Evidence: `TestScreenshotMatrix`; fifteen
      PNGs under `/tmp/opencode/code-view-shots/`.
- [x] Add the example to the website examples page if the site lists this
      catalog. Evidence: `examples/readme.md` feeds `frontend/src/content.js`;
      the Home showcase cards are curated and unchanged.
- [x] Run `make test`, `make build`, `make lint`, `make docs-check`,
      `make plans-check`, and `make examples-check`. Evidence: all passed on
      the `examples/code-view` branch.
- [x] Confirm `make internal-import-check` shows no new example import.
      Evidence: `make lint`.

## Gate

Every gate is green and the example appears in the catalog.

## Risks

- The examples module is separate. A change to the root library tag needs the
  examples `require` and `replace` review before a release.
- A screenshot is not a correctness test. Keep the box and token assertions as
  the gate and use the PNGs for review.
