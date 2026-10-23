## Summary

This branch implements the v0.0.3 production-readiness work across the public API, page and window ownership, rendering, examples, documentation, and release checks. It keeps 1.0 native-host claims to desktop and Android, with wasm and `-web` described as previews; open validation items remain in the current ledger, and this PR does not tag or publish v0.0.3.

## Motivation / context

- Plan archive: `plans/v0.0.3/prod-ready-task/`. Its earlier checklists are superseded planning records, not the current status ledger.
- Current implementation and validation status: `plans/README.md` and `plans/validation/2026-10-23-session.md`.
- The work makes the root `ownframe` package the app-facing API, sets clear ownership for page and window behavior, and makes platform and validation claims match the evidence.

## Changes

### Public API and package boundaries

- Replace the public `Page` alias with a defined wrapper that holds an unexported `*internal/page.Page` and delegates an explicit method set.
- Group page controls under `Form`, `Edit`, `Scroll`, `Window`, `Access`, and `Probe`. Keep `Testing` for tests.
- Keep the Ebiten host loop private. Add `NewGame` for tests that need to drive the same game loop without exposing `host.Screen`.
- Move drawing, accessibility, scrolling, form, and drop responsibilities into their owning page subsystems. Migrate example callers to the public package where an API exists, and retain the explicit internal-import allowlist for the remaining cases.
- Keep the library and examples as separate Go modules. Add checks for the library archive and a consumer without the workspace or a local replacement.

### Rendering and window behavior

- Keep a styled document for the page lifetime and relayout it when the viewport or state changes. Move template execution, styling, and picture painting into the draw path.
- Preserve the display-list replay path for replayable pages and the picture fallback for other pages. Keep partial repaint work within the content-sized tile grid.
- Keep renderer caches under renderer ownership with independent eviction and close paths. Keep pipeline timings, dirty counters, and allocation tracking opt-in.
- Move desktop pacing policy into `internal/window/pacing`. A clean desktop window can wait for input or the page's next requested frame, with the documented keep-alive deadline.
- Retain the desktop devtools dock, mobile bottom panel, pointer and touch behavior, scroll-window pinning, and the blank-frame regressions in graphics coverage.
- Treat context cancellation as a clean stop and keep tall desktop pages idle when no input or animation is due.

### Maintenance, CI, and distribution checks

- Add or tighten checks for both Go modules, examples, documentation snippets, plan claims, internal imports, graphics, module archives, generated site output, release archives, reproducibility, soak budgets, and performance measurements.
- Run separate CI jobs for race detection, cross-platform vet, native packaging, security and fuzz checks, secrets, workflow lint, examples race shards, and performance comparison.
- Add a commit-pinned Android SDK setup action before the existing Android platform and NDK installation. The prior hosted run failed before binding because `sdkmanager` was absent from `PATH`.
- Keep release candidate signing and attestation separate from publication. A candidate workflow has produced verified artifacts, but this branch does not create a version tag or GitHub Release.

### Documentation and examples

- Update the platform, API, rendering, accessibility, packaging, testing, performance, security, and migration guidance to match the shipped code and its limits.
- Compile documentation snippets and check selected operational claims. Rebuild and check generated website output.
- Keep Android and desktop demos and the example module aligned with the wrapped public API and current host support.
- Add phase plans and framework benchmark records as historical planning material. The current ledger remains `plans/README.md`.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Pipeline timing and allocation tracking remain off by default. Desktop idle pacing and partial repaint reduce work when no frame is due or only part of the page is dirty. CI compares performance on the same runner and applies a coarse 1.5 regression ratio. |
| **Memory** | Renderer-owned caches have independent eviction and close paths. Soak checks track redraw and PNG budgets separately. The validation records report Linux RSS only, not native Windows or macOS memory behavior. |
| **Behavior / correctness** | Page and host responsibilities have clearer owners. The window retains tile clipping, focus, scroll, touch, drop, IME input plumbing, and cancellation behavior. Pixel and graphics regressions cover retained surfaces, dock clipping, rotation sizing, and clean frames. |
| **API / CLI** | `ownframe.Page` is now a defined wrapper. The grouped accessors and `NewGame` replace access through internal page and host types. This repository has no CLI. |
| **Dependencies** | Keep one pinned blinkless engine dependency in the library module. Keep examples in their own module, with the checkout workspace replacing the published library version only for local development. |
| **Binary size / build time** | No before/after size or build-time claim is made here. CI compiles the native package targets, checks the library archive, and packages Linux and wasm release candidates. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `ownframe.Page` is no longer an alias for `internal/page.Page`; its inner page is private. | Use the documented methods on `ownframe.Page` and its `Form`, `Edit`, `Scroll`, `Window`, `Access`, and `Probe` accessors. Code that depended on internal fields or methods must move to the public API. See `documentation/migrate-experimental.md`. |
| The host loop stays private and no longer exposes `host.Screen` through the root API. | Tests that need to run the host game can use `ownframe.NewGame`. Applications should use `Run` or `BindMobile`. |
| This is a v0.0.3 development-cycle change and may break APIs before 1.0. | Review the migration guide and current API docs before upgrading. No tagged release is created by this PR. |

## Library freeze (until Page wrap)

This branch performs the Page wrap described by the freeze note. Keep these checks visible for reviewer confirmation; passing CI does not establish every API naming or invariant claim.

- [ ] No new `type X =` at the module root except documented re-exports
- [ ] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [ ] No new example `ownframe/internal/` imports outside the recorded allowlist
- [ ] No synonym names such as `PreviousX` for `X`
- [ ] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [ ] Library code panics only for invariants; new user-facing failure paths return `error`
- [ ] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's context
- [ ] No new example whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr

## Test plan

### Hosted checks recorded before the Android SDK setup fix

At `e870762e907efe0a0eea43804034d47ba393c639`, the CI `gates` job passed `go mod verify`, both-module tests, `make lint`, `make graphics-check`, `make module-check`, and `make site-check`. The `test-race`, security, reproducibility, native Windows and macOS packaging, cross-vet, examples-race, performance, workflow-lint, and secrets jobs also passed. The Android bind job failed before compiling because the runner could not find `sdkmanager`.

This PR adds the missing SDK setup action, pinned to `be39fa834029ff78f1a44aa3bb0819b8fc2bd8fd`. The hosted CI rerun for the PR head must confirm the bind job. [Failed pre-fix run](https://github.com/chinmay-sawant/ownframe/actions/runs/39937138755).

### Recorded validation evidence

- `plans/validation/2026-10-23-session.md` records passing `make test`, `make test-race`, `make lint`, `make graphics-check`, `make module-check`, `make release-check`, and `make site-check` on the isolated validation snapshot. It also records a reviewed Pixel 7 session on Android 17. These results describe that named snapshot, not arbitrary later edits.
- `plans/validation/2026-10-23-review-fixes.md` records passing tests, build, race, lint, graphics, docs, plans, examples, module, release, reproducibility, website, soak, and targeted checks for its reviewed source snapshot. Its remaining platform and independent-review limits are recorded there.
- The validation session records a successful signed and attested candidate at `a1d66b2d`; no production tag or GitHub Release was published.
- Native Windows composition input remains unverified. macOS input validation is deferred because no host was available. The current ledger also leaves the independent clean-checkout handoff open.

### Commands

```sh
make test
make lint
make test-race
make graphics-check
make module-check
make release-check
make site-check
```

These commands are listed as recorded validation gates, not as new local runs performed while writing this PR record. The PR-head CI run is the final verification for the pushed branch.

## Screenshots / sample output

The reviewed Android session captures are in the validation record:

- [Keyboard with committed text](../../../validation/pixel7-2026-10-23/04-keyboard-qa.png)
- [Portrait after background and resume](../../../validation/pixel7-2026-10-23/06-resumed.png)
- [Session metadata and visual review](../../../validation/pixel7-2026-10-23/session.json)

## Related issues

No current issue tracks this v0.0.3 production-readiness integration. The only repository issue is closed issue [#5](https://github.com/chinmay-sawant/ownframe/issues/5), which covers the v0.0.2 documentation audit and does not match this work, so this PR does not claim to close or relate to it.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied: `enhancement`, `documentation`
- [x] Related issues reviewed; no applicable issue ID exists
- [x] Filled body committed under `plans/v0.0.3/prod-ready-task/PR/pr-production-readiness.md`

## Follow-ups (out of scope)

- Complete the independent clean-checkout review described in `documentation/maintenance-review.md`.
- Verify native Windows preedit, cancellation, conversion, and commit with a composing input method.
- Revisit macOS input validation when a suitable host is available.
- Publish a version tag and GitHub Release only through a separately authorized release step.
- Keep live screen-reader adapters and the deferred work in `plans/README.md` out of 1.0 claims until they have implementation and platform evidence.

## Reviewer checklist

- [ ] Behavior matches the summary and the cited test evidence
- [ ] Scope and remaining platform limits are clear
- [ ] Public API and migration changes match the code
- [ ] The page and host ownership boundaries remain coherent
- [ ] No unrelated changes are in the diff
- [ ] CI passes on the PR head, including Android binding after the SDK setup change
- [ ] PR metadata and related-issue handling are accurate
- [ ] No secrets or unintended generated artifacts are included
- [ ] Diff-stat-by-extension table matches the generator output

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 1183 | 27705 | 13298 |
| `.gradle` | 1 | 5 | 0 |
| `.html` | 5 | 456 | 3 |
| `.java` | 1 | 19 | 2 |
| `.jpg` | 2 | Binary | Binary |
| `.js` | 4 | 825 | 413 |
| `.json` | 4 | 650 | 0 |
| `.jsx` | 1 | 5 | 5 |
| `.md` | 104 | 3482 | 503 |
| `.mjs` | 1 | 1 | 1 |
| `.mod` | 2 | 13 | 7 |
| `.png` | 4 | Binary | Binary |
| `.ps1` | 1 | 160 | 0 |
| `.py` | 20 | 1476 | 0 |
| `.sh` | 14 | 428 | 386 |
| `.sum` | 2 | 10 | 2 |
| `.tsv` | 1 | 13 | 0 |
| `.txt` | 5 | 190 | 6 |
| `.xml` | 1 | 1 | 1 |
| `.yml` | 3 | 252 | 24 |
| No extension | 2 | 64 | 41 |
| **Total** | **1361** | **35755** | **14692** |
