## Summary

Add Quarry, a local-first SQLite and PostgreSQL client, as a complete builder example. The PR also improves grid editing and adds six native screenshots to the frontend showcase.

## Motivation / context

- Plan: `plans/v0.0.8/quarry-db-client/README.md`
- Issues: none found.

## Changes

### Quarry application

- Add SQLite, PostgreSQL, and offline mock PostgreSQL drivers, along with schema browsing, SQL execution, query history, connection management, and monitoring.
- Add a typed result grid with filtering, sorting, inline edits, import/export, a cell inspector, and transactional change review.
- Keep Quarry state such as connections, snippets, history, theme, and pane sizes in its local config database.

### Frontend showcase

- Add six screenshots for query results, the Orders table, monitoring, connections, history, and settings.
- Add Quarry to the frontend showcase after Excel.

### Grid behavior

- Fix grid clipping and text sizing, keep per-column filters and toolbar actions usable, and make modal keyboard focus visible in both themes.
- Add coverage for inline value editing, validation, selection colors, and grid clipping.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No performance measurements were made. |
| **Memory** | No memory measurements were made. |
| **Behavior / correctness** | Adds a standalone database client example. Grid edits are staged and committed transactionally. |
| **API / CLI** | No root library API changes. Quarry adds its own command flags for database, web preview, and snapshots. |
| **Dependencies** | Adds pgx and its transitive dependencies to the examples module. The root module is unchanged. |
| **Binary size / build time** | No measurements were made. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

## Library freeze (until Page wrap)

- [x] No new module-root type aliases.
- [x] No new `internal/page` methods added to grow `ownframe.Page`.
- [x] No example imports of `ownframe/internal/` (`make internal-import-check` passed).
- [x] No new root `Page` API or synonym methods.
- [x] No library user-facing failure paths or exported return types added.
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr.

## Test plan

- [x] `make test`
- [x] `make lint`
- [x] `make build`
- [x] Frontend Vite build completed and emitted the Quarry screenshot assets.
- [x] Captured six Quarry screens from the running Windows application.

### Commands

```sh
make test
make lint
make build
cd frontend && npm exec vite build
```

## Screenshots / sample output

Six 1440x960 WebP captures are included in `assets/screenshots/demos/quarry/`: `query-results.webp`, `orders.webp`, `monitor.webp`, `connections.webp`, `history.webp`, and `settings.webp`.

## Related issues

None found.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] Labeled `enhancement` and `documentation`.
- [x] Related issues checked; no matching issue found.
- [x] Filled body saved under `plans/PR/pr-quarry-db-client.md`.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 3 | 774 | 0 |
| `.go` | 95 | 24012 | 5 |
| `.js` | 1 | 15 | 0 |
| `.md` | 11 | 824 | 0 |
| `.mod` | 1 | 4 | 0 |
| `.sum` | 2 | 23 | 1 |
| `.webp` | 6 | Binary | Binary |
| **Total** | **119** | **25652** | **6** |
