## Summary

This PR integrates OwnFrame's v0.0.3 production-readiness work across the public API, page and window ownership, rendering, state handling, examples, documentation, and distribution checks. The branch changes 1,361 tracked files. It does not claim that every 1.0 validation item is closed, and it does not create a version tag or GitHub Release.

The 1.0 native-host scope remains desktop Linux, Windows, and macOS through `Run`, plus Android through `BindMobile`. Wasm and `-web` remain previews. The current support and validation status is in `plans/README.md`.

## Motivation / context

- Plan archive: `plans/v0.0.3/prod-ready-task/`. Its earlier checklists are superseded planning records, not the current status ledger.
- Current implementation and validation status: `plans/README.md` and `plans/validation/2026-10-23-session.md`.
- The goal is to make the root `ownframe` package the app-facing API, give page and host responsibilities clear owners, and make the documentation and checks match the code.
- This branch combines a long implementation sequence. The commit map below groups that sequence so reviewers can follow the work by responsibility.

## Changes

### Public API and package boundaries

- Replace the public `Page` alias with a defined wrapper that holds an unexported `*internal/page.Page` and delegates an explicit method set.
- Group page controls under `Form`, `Edit`, `Scroll`, `Window`, `Access`, and `Probe`. Keep `Testing` for tests.
- Keep common app calls directly on `Page`, including `SetData`, `Handle`, `Redraw`, navigation, sizing, `Click`, `Boxes`, frame requests, and the main scroll methods. `Page.Testing()` remains callable but is intended for test files; examples use the app-facing methods.
- Keep the Ebiten host loop private. Add `NewGame` for tests that need to drive the same game loop without exposing `host.Screen`.
- Split page responsibilities into packages such as `internal/page/a11y`, `click`, `dirty`, `draw`, `form`, `gesture`, `hit`, `nav`, `patch`, `reload`, `scroll`, and `source`. The `*_wire.go` files connect those subsystems through page-owned state.
- Migrate example callers to the public package where an API exists, and retain the explicit internal-import allowlist for the remaining cases.
- Keep the library and examples as separate Go modules. Add checks for the library archive and a consumer without the workspace or a local replacement.

### Page calls that moved

Because `ownframe.Page` used to be an alias, callers could use methods promoted from the implementation type. Those methods are no longer all on the root `Page`. Update calls like these:

| Before | After |
|--------|-------|
| `p.FormValue(id)` | `p.Form().FormValue(id)` |
| `p.Copy(ctx)`, `p.Undo(ctx)` | `p.Edit().Copy(ctx)`, `p.Edit().Undo(ctx)` |
| `p.AccessibilityTree()` | `p.Access().AccessibilityTree()` |
| `p.Stats()`, `p.Display()`, `p.PNG()` | `p.Probe().Stats()`, `p.Probe().Display()`, `p.Probe().PNG()` |
| `p.PollReload(ctx)`, `p.SetPerf(true)`, `p.SetDevTools(true)` | `p.Window().PollReload(ctx)`, `p.Window().SetPerf(true)`, `p.Window().SetDevTools(true)` |
| `p.SetWindowing(true)`, `p.SetAllowScroll(true)`, `p.SetViewportPinZ(z)` | `p.Scroll().SetWindowing(true)`, `p.Scroll().SetAllowScroll(true)`, `p.Scroll().SetViewportPinZ(z)` |

This is a method-set source change. It is not a requirement for external callers to import `internal/page`; Go already prevents that. `documentation/migrate-experimental.md` covers retained/reactive updates, not this wrapper migration, so this table records the relevant call changes here.

### Rendering and window behavior

- `Page.Redraw` remains synchronous. `internal/page/draw` owns template execution, form synchronization, styling, and display-list or bitmap commit for that call. This change does not move page redraw into Ebiten's `Draw` callback.
- `internal/render.Cache` keeps a parsed HTML tree and styled CSS document keyed by executed source. Viewport and state changes relayout that cached document instead of repeating template execution and cascade work.
- Replayable pages keep display-list operations and use partial repaint. Other pages keep the picture fallback. The 256 px tile grid covers page content, so a large page can use up to four times the window's pixel area. Oversized or unsafe paths fall back to a viewport buffer or full replay; the tile path does not guarantee lower total pixel memory for every page.
- A surface cleared every frame invalidates retained tiles, which the window now restores before blitting. Graphics tests compare eight retained updates with full replay and include the cleared-surface case.
- Viewport-pinned and fixed paint no longer increases the scrollable content bounds. Telegram's keyboard-inset and rotation regression checks the former composer gap.
- Desktop pacing parks a clean window after five clean frames and 250 ms without input. It wakes at `Page.WantFrameAt` when a page declares a deadline, or at the 125 ms keep-alive otherwise. `WindowOptions.NoIdlePace` opts out. Wasm and mobile keep their frame callback.
- Context cancellation ends the runtime cleanly. The renderer owns its caches and closes or evicts them independently. Pipeline timing, dirty counters, frame sampling, and allocation tracking remain opt-in.

### State, Android, and host fixes

- `internal/state.Binding` owns subscription cancellation, binding generations, and metadata synchronization. Page paint also invalidates retained PNG data when content changes.
- Reused Android activities reapply the debug capture policy. Normal launches keep secure capture settings.
- The window and replay tests cover dirty tiles, clean frames, dock clipping, focus, pointer and touch routing, scrolling, and input bounds.

### Examples, docs, and plans

- The 457 changed example files include API migrations, fixtures, tests, and runtime checks. `examples/ops-console` adds a five-screen yard-operations demo; `examples/maintenance.json` classifies it as experimental, not as a maintained support promise.
- Telegram adds a keyboard-inset and rotation regression. CSS preview and Spotify add broad test, benchmark, and runtime-observation coverage. Desktop cat adds preview-host and goroutine-safety work.
- The 48 documentation files update API, platform, rendering, accessibility, packaging, testing, performance, security, and migration claims. They identify desktop and Android as the 1.0 native hosts, and wasm and `Serve` as previews.
- The 45 plan files include historical phase plans, validation records, benchmark results, and maintenance status. The older phase checklists are superseded; `plans/README.md` is the current ledger.
- The accompanying `release-v0.0.3.md` is a draft note stored in the requested folder. It is not a GitHub Release body submission, and no release is published by this PR.

### Maintenance, CI, and distribution checks

- Add or tighten checks for both Go modules, examples, documentation snippets, plan claims, internal imports, graphics, module archives, generated site output, release archives, reproducibility, soak budgets, and performance measurements.
- Run separate CI jobs for race detection, cross-platform vet, native packaging, security and fuzz checks, secrets, workflow lint, examples race shards, and performance comparison.
- Add a commit-pinned Android SDK setup action before the existing Android platform and NDK installation. The prior hosted run failed before binding because `sdkmanager` was absent from `PATH`.
- Keep release candidate signing and attestation separate from publication. A candidate workflow has produced verified artifacts, but this branch does not create a version tag or GitHub Release.

### Commit map

The implementation history has these connected parts. The initial plan commit is followed by the API and page ownership work, then renderer/window fixes, then validation and distribution hardening. The final existing PR-preparation commits add the Android SDK setup and the first PR record; this update expands that record.

<details>
<summary>Commit subjects grouped by responsibility</summary>

**Plan and public API**

- `b187e513` `docs(plans): add v0.0.3 production-readiness checklists`
- `ca0a4ff3` `feat(api): wrap Page and ship v0.0.3 desktop host honesty`
- `925f98c3` `feat(api): hide inner Page and keep pinned chrome on tiles`
- `ac197154` `feat(api): move host loop off Page and tell the 1.0 story`
- `9162b794` `feat(api): keep Page as the app set and split page packages`
- `50d0f265` `feat(page): split draw, a11y, scroll and drop Testing.Host`
- `522b1aa5` `refactor(api): put app calls on Page and reserve Testing for tests`
- `2a43ff3e` `refactor(api): move measurement onto Probe and extract the click sequence`
- `7b06d014` `refactor(api): split Page into doors and move page subsystems out`
- `309a8c52` `refactor(page): move half of the flat page package into subsystems`

**Rendering, window, and state**

- `c59f1a51` `refactor(core): consolidate rendering responsibilities and fix retained PNG invalidation`
- `b0d428cb` `fix(state): own listener lifetimes and enforce maintenance validation`
- `ae5ddbbf` `refactor(window): isolate pacing policy and validate desktop graphics`
- `b32ea940` `refactor(state): own subscription lifecycle and metadata in Binding`
- `279294a2` `fix(window): rebuild tiles when the host clears every frame`
- `69626f88` `fix(window): exclude viewport chrome from scroll bounds`
- `acef739d` `refactor(page): keep one form document and move redraw into draw`
- `78306469` `refactor(draw): run template, style, and bitmap paint in draw`
- `24d4267a` `fix(runtime): treat cancel as a clean stop and idle tall desktop pages`

**Validation, Android, and distribution**

- `b5838e15` `test(clipboard): add Wayland route tests and a11y split leftovers`
- `f8a2100a` `refactor(maintenance): replace file caps with verified documentation and graphics gates`
- `96353481` `test(gates): validate graphics, modules, releases, and mobile devtools`
- `69f083de` `test(maintain): lock host entry points, docs, and the module zip`
- `31e17f5f` `test(validation): select retained desktop surfaces and run signed candidates`
- `a1d66b2d` `fix(release): provision offline dependencies and refresh generated site`
- `1b4fe4b1` `fix(android): apply debug capture policy to reused activities`
- `a17d16b7` `test(validation): record reviewed Pixel session and signed candidate proof`
- `0521ff10` `fix(readiness): correct performance evidence and strengthen release validation`
- `a5e5b6ad` `Merge production-readiness API work with validated mobile fixes`

**Plans and PR preparation**

- `e870762e` `docs(plans): add phase-wise roadmaps and Hello World framework benchmarks`
- `590f54b6` `fix(ci): install Android SDK for bind checks`
- `fde428a7` `docs(plans): add v0.0.3 production readiness PR record`

</details>

### Documentation and examples

- Update the platform, API, rendering, accessibility, packaging, testing, performance, security, and migration guidance to match the shipped code and its limits.
- Compile documentation snippets and check selected operational claims. Rebuild and check generated website output.
- Keep Android and desktop demos and the example module aligned with the wrapped public API and current host support.
- Add phase plans and framework benchmark records as historical planning material. The current ledger remains `plans/README.md`.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Desktop pacing avoids continuous frames while a clean window waits. Same-runner CI comparison passed on code head `fde428a`. The dated local benchmark report comes from `24d4267a` with a dirty source tree, so it is not a clean final-branch baseline. |
| **Memory** | Renderer caches have independent owners and close/eviction paths. Content-sized tiles can use up to four times the window pixel area; oversized and unsafe cases use other paths. Soak records separate redraw and PNG budgets. The validation record reports Linux RSS, not native Windows or macOS RSS. |
| **Behavior / correctness** | Pinned viewport paint no longer expands scroll bounds. Cleared-surface frames restore retained tiles. Cancellation ends cleanly. Android rotation/inset coverage checks the Telegram composer gap. |
| **API / CLI** | `ownframe.Page` is now a defined wrapper with explicit method groups. `NewGame` exposes the Ebiten game for tests while keeping `host.Screen` private. This repository has no CLI. |
| **Dependencies** | The library keeps one pinned blinkless engine dependency. Examples remain a separate module. The workspace replacement is for this checkout only; archive and consumer checks verify distribution without it. |
| **Binary size / build time** | No before/after size or build-time claim is made. CI compiles native package targets, checks the library archive, and packages release candidates. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| The root `Page` changed from an alias to a wrapper, so methods formerly promoted from the implementation type are no longer all direct methods. | Move form, edit, accessibility, probe, window, and scroll calls to the accessors shown above. Keep `SetData`, `Handle`, `Redraw`, navigation, sizing, click, boxes, frame requests, and the main scroll methods directly on `Page`. |
| The host screen type is no longer exposed through the root package. | Applications use `Run` or `BindMobile`. Tests that need the game loop use `ownframe.NewGame(ctx, page)`. |
| The package ownership changed across the v0.0.3 development cycle. | Use only the root `ownframe` API from applications. No v0.0.3 tag is created by this PR. |

## Public API boundary review

The following checks remain visible for review. Automated tests and lint cover several gates, but do not establish every naming, error, or invariant claim.

- [x] No new root aliases outside documented re-exports (`Box`, `Config`, and `Handlers`)
- [x] No new example `ownframe/internal/` imports outside the recorded allowlist (`make internal-import-check` passed in hosted CI)
- [ ] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [ ] No synonym names such as `PreviousX` for `X`
- [ ] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [ ] Library code panics only for invariants; new user-facing failure paths return `error`
- [ ] `context.Context` is the first parameter when a function already does I/O, cancel, or deadline; no hidden `time.After` ignores the caller's context
- [x] The new ops-console example is classified as experimental, not as a maintained support promise

## Test plan

### Hosted PR checks

The complete CI run for code head `fde428a705dc64d724f3c7c0a65a84e15ec17f19` passed. This includes `gates`, `test-race`, `android-bind`, `sec`, `secrets`, cross-vet, all examples-race shards, native Windows and macOS packaging, reproducibility, performance, and workflow lint. The Android bind compiled after the SDK setup fix. See [the passing run](https://github.com/chinmay-sawant/ownframe/actions/runs/39939413447).

This PR record update adds documentation only. GitHub will rerun checks for the updated PR head; the run above is evidence for the code head shown, not a claim that a later head has already completed CI.

### Recorded validation evidence

- `plans/validation/2026-10-23-session.md` records `make test`, `make test-race`, `make lint`, `make graphics-check`, `make module-check`, `make release-check`, and `make site-check` on an isolated snapshot containing `69f083de` and `31e17f5f`. It records the reviewed Pixel 7 session on Android 17. These checks belong to that named snapshot.
- `plans/validation/2026-10-23-review-fixes.md` records a wider passing local gate set from `24d4267a` with review changes uncommitted. Do not treat it as proof for an arbitrary clean branch checkout.
- The performance JSON and text report also identify source `24d4267a` as dirty. Its samples are not a final-branch performance baseline. The hosted same-runner comparison at `fde428a` is a separate check.
- The signed and attested candidate at `a1d66b2d` passed artifact and provenance verification. Later Android changes are outside that attested snapshot. No production tag or GitHub Release was published.
- The Pixel 7 record includes an APK built from `69626f88` with SHA-256 `cf3ebe8b53b594d143529c78b290dc1ee3fe1c3b0d0a35dbef4aa2704450a221`. The captures show no final portrait gap. They do not prove keyboard visibility is preserved across every rotation or complex IME composition.

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

These are recorded validation gates, not commands run locally while preparing this PR update. The updated PR head's hosted checks are the final run for the pushed description.

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
- Verify native Windows preedit, cancellation, conversion, and commit with a composing input method. The English typing session does not close this check.
- Revisit macOS input validation when a suitable host is available.
- Keep live AT-SPI D-Bus, Windows UIA, macOS VoiceOver, and Android TalkBack support out of 1.0 claims until implementation and real platform evidence exist.
- Publish a version tag and GitHub Release only through a separately authorized release step.

## Reviewer checklist

- [ ] Behavior matches the summary and the cited test evidence
- [ ] Scope and remaining platform limits are clear
- [ ] Public API migration examples match the new method set
- [ ] Redraw, tile, and desktop pacing descriptions match the implementation
- [ ] Open and deferred platform items remain visible
- [ ] No unrelated changes are in the diff
- [ ] CI passes on the updated PR head
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
| `.md` | 104 | 3591 | 503 |
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
| **Total** | **1361** | **35864** | **14692** |
