## Summary

Add a cross-platform Pet Care example with shared Go screens, local SQLite storage, seeded records, and JSON backup and restore. Add Android packaging and shared host behavior, then document and present the example in the website showcase.

## Motivation / context

- Plan: `plans/v0.0.1/examples/pet-care-app-checklist.md`
- Related issues: None found.

## Changes

### Pet Care example

- Add sign-in, Today, pet profile, care schedule, history, and settings screens backed by local SQLite data.
- Seed an empty database with sample pets and care records. Add JSON backup and restore.
- Add Android bindings and a Gradle app that uses the shared Android host for platform behavior.

### Documentation and website

- Document creating Android apps with ownframe and update Android runtime and interaction guidance.
- Add a Showcase submenu. Group the Windows Pet Care capture with Pet Care, and present that capture on its own row.
- Organize screenshot assets by platform and category, convert PNG and JPEG assets to WebP, and update site references.
- Give the docs panes independent scrolling, constrain the sidebar and content widths, hide visible scrollbars, and size demo videos to their source ratio.
- Replace the command-only Get Started block with a highlighted Go example and widen the homepage helper text.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No performance measurements were made. |
| **Memory** | No memory measurements were made. |
| **Behavior / correctness** | Adds an example app and reusable Android host behavior. Existing Pet Care data stays local to the configured database. |
| **API / CLI** | No root package API or CLI changes. Adds the Android host library and example-specific entry points. |
| **Dependencies** | No root Go module dependency changes. Adds the Android Gradle project for the example. |
| **Binary size / build time** | No measurements were made. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

## Test plan

- [x] `npm run build` in `frontend` completed successfully.
- [ ] `make test` was not run during this PR preparation.
- [ ] Windows input behavior and the remaining app-level test cases are still tracked in the Pet Care checklist.

### Commands

```sh
cd frontend && npm run build
```

## Screenshots / sample output

Pet Care Android and Windows screenshots are included under `assets/screenshots/android/pet-care/` and `assets/screenshots/windows/pet-care/`. The website showcase also includes Telegram and desktop demo captures.

## Related issues

None found.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] Labeled `enhancement` and `documentation`.
- [x] Related issues checked; no matching issue found.
- [x] Filled body saved under `plans/PR/pr-pet-care-example.md`.

## Follow-ups (out of scope)

- Add storage/login and app-level tests, and verify Windows mouse, keyboard focus, Enter, Escape, and text editing as listed in the Pet Care checklist.

## Reviewer checklist

- [ ] Confirm the Pet Care desktop and Android paths use the same app and seeded data model.
- [ ] Confirm the shared Android host leaves app navigation and permissions with the example.
- [ ] Confirm screenshot paths and Showcase categories match the moved assets.
- [x] PR has assignee and labels.
- [x] Related issues contain no placeholder IDs.
- [x] Diff-stat-by-extension table is generated from `scripts/pr-diff-stat.sh` and included below.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.bat` | 1 | 0 | 0 |
| `.css` | 1 | 0 | 0 |
| `.go` | 1 | 0 | 0 |
| `.gradle` | 1 | 0 | 0 |
| `.html` | 1 | 0 | 0 |
| `.jar` | 1 | Binary | Binary |
| `.java` | 1 | 0 | 0 |
| `.jpg` | 1 | Binary | Binary |
| `.js` | 1 | 0 | 0 |
| `.jsx` | 1 | 0 | 0 |
| `.md` | 1 | 0 | 0 |
| `.png` | 1 | Binary | Binary |
| `.properties` | 1 | 0 | 0 |
| `.sh` | 1 | 0 | 0 |
| `.webp` | 1 | Binary | Binary |
| `.xml` | 1 | 0 | 0 |
| No extension | 1 | 0 | 0 |
| **Total** | **1** | **0** | **0** |
