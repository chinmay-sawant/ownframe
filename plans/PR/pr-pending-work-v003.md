# chore(v0.0.3): land pending perf, component, and example work

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../skills/PR/PR_TEMPLATE.md). Base: `origin/master` (`68d5252d`). Code head: `chore/pending-work-v003` (`5bda7e6b`). Diff at the code-fix head: 32 commits, 471 files, +44,521 / -1,229.

> **CI repair landed on this branch.** The local-only Blinkless replaces are gone, both modules resolve `dc1d518` from the proxy, the examples pin matches the library, and brotli moved to `v1.2.7` for GO-2026-6888. Lint, tests, race, build, graphics, security, reproducible, fuzz, website, docs, and the performance comparison pass locally. The Windows, macOS, and Android CI jobs are the remaining checks.

---

## Summary

- Rewrites the page patch and render pipeline: `Page.PatchElements` applies a batch through one cascade and one display build, incremental dirty regions replace full diffs, clean frames skip, adopted sources hash once, kid/box/click caches go lazy, form sync and navigation memoize, and `Display.Release` gives the patched display a lifetime. `Page.PatchElement` stays; `Page.PatchElements` and `ownframe.ElementPatch` join it.
- Rebuilds `ui` around allocation-controlled storage: attributes and styles use inline 4-entry sets with slice overflow, single-number lengths stay unformatted until render, `View.Update` diffs a rebuilt tree and applies a patch batch with a full-render fallback, and `ui.Document` assembles one buffer. The allocation suite pins 34,000 / 500 / 2 alloc ceilings and a golden render hash.
- Cuts renderer and runtime cost: generation-keyed replay plans, tile-bucketed plan lookup, cached chunk signatures, pooled `vector.Path` scratch, stack-only draw options, smaller GPU caches, transition capture filtered to transitioning ids, a cached `cssmotion.Index`, zero-allocation window box reads, and eleven new emoji (31 → 42).
- Adds seven HTML/CSS chart packages (`apexcharts`, `chartjs`, `echarts`, `nivo`, `recharts`, `shadcnchart`, `tremor`) and a broad ChadCN pointer/patch fix pass with stable part ids on nearly every option struct.
- Adds the GPU-raymarched `examples/blackhole`, the `examples/builder/discord` clone, the `examples/testing` ownframe-vs-MyGo invoice/black-hole harnesses, seven chart galleries in `ui-builder`, and builder-pattern site starters.
- Adds `internal/perfcheck`, `internal/benchscore`, the `benchscore` CLI, `make bench-gate`, and four profiling/measurement scripts. The Phase 07 plan records the current baseline table.
- Bumps blinkless to `dc1d518bff32` and adds `golang.org/x/perf` for benchmark parsing.

---

## Motivation / context

- This is the pending v0.0.3 work after the release PRs: the patch-path performance plan (`plans/v0.0.3/performance/phase-07-rebuild-and-patch-paths.md`) plus the component, demo, and site work that exercises it.
- The demos are the evidence base: a 1,000-row invoice, a windowed Discord transcript, inventory filters, and seven chart galleries all drive the same patch and allocation paths.
- Related issues: `Relates to #45` (reproducible benchmark harness) and `Relates to #59` (experimental `ui` builder).
- No issue fully closes with this branch.

---

## Changes

### Page patch and render pipeline

- `Page.PatchElements` (`internal/page/patch_batch.go`) accepts `[]ElementPatch{ID, Edits}`, checks `patch.Ready` for every entry, applies all edits, calls `cache.RelayoutFor` once, and runs `render.RelayoutElements` to clone the retained placement, splice every subtree, and build one display. It then adopts, captures/binds motion, refreshes accessibility per id, updates the styled cache, and remembers the new render key. Any failure rolls the batch back (`failBatch` → `rollbackPatch` → invalidate) so the next redraw reparses. An empty batch returns nil; a batch where no edit applies is refused, and a refused batch wraps `ErrPatchRefused`.
- `internal/render/relayout_element.go` adds `render.RelayoutElements` as the one-display batch path. `5f08ad45` removes the per-element display build.
- `page_adopt.go` adds `AdoptPatchedSource`, exposed as `LoadOptions.AdoptTree`. It records the source, keeps the cached tree, and does not parse, execute, or draw. It is for the builder after a patch; hand-written HTML is refused with `page.ErrAdoptSource` (not re-exported at the root).
- `draw.Run` returns after `SyncForm`/`Unchanged` when nothing changed. `Page.unchanged` (`page_render_key.go`) rejects on retained-state changes, active motion, caret/blink flips, or a changed `draw.Key`; the key uses source hash + length + revision, so zero-revision keys fall back to hash/length/memcmp. `remember` drops the stored source once a revision exists. A realistic idle `Redraw` is pinned at zero allocations.
- `internal/page/dirty/` replaces the full per-op diff with 64-op region stamps: `diff.go` hashes `SameOp` inputs per region and verifies `Font`/`Grid` identities, falling back to the per-index scan on a NaN or pointer mismatch. `DiffWithBounds` reuses the computed bounds for perf counters, `boxes.go`/`bounds.go` map changed ops to repaint boxes, and `op.go` carries per-op damage. `dirty_incremental_test.go` compares the fast path to the legacy comparator.
- Dirty marking uses a per-page `dirty.KidIndex`: `markElement` unions the element subtree and its previous rect without a full box scan. `setBoxes` deltas the content rect, repoints the hit id map when `hit.SameIDs` holds, and rewrites only the rects in the kid index when ids and parent links are stable. `maybeReleasePainted` releases the painted baseline after `TakeDirty` unless retained state, live motion, or a full mark needs it.
- Form sync (`form_wire.go`) snapshots shared state, takes a fast path when the body is unchanged and no handler ran, and otherwise hashes the body once and threads it through `form.ScanCached`, `form.RewriteCached`, and `draw.EmojiPassWithHash`. The scan memo is four entries keyed by hash plus exact source; the rewrite memo caches the head offset. `publishForm` records the output hash. `internal/page/form/otp.go` adds one-digit OTP slots that advance/retreat focus; disabled `<option>`s are skipped.
- Navigation (`internal/page/nav/history.go`, `page_nav.go`) stores an FNV hash and delimiter verdict beside each source. `Load`, `ReplaceHTML`, `Back`, and `Forward` compare hashes first and reuse the parsed template and style cache for identical sources. `isBuilderDocument` recognises the `ui.Document` envelope and keeps `tpl` nil, so a builder redraw returns `nav.Current` directly.
- Click resolution (`internal/page/click/label.go`) calls `boxAt` once, threads `hit.Spot` through `click.Run`, resolves the innermost id-bearing ancestor without rescanning, and maps a label click to its control through `LabelControl`/`LabelFor`. It rejects disabled or out-of-scope controls.
- Hover and motion capture are scoped in `page_hover_patch.go` and `motion_wire.go`: `ad86a640` adds stable patch ids and limits hover/transition capture to the affected scope. `styleStaged` reuses a cached styled document when `render.State.SameStyleState` proves changed hover/active/focus ids match no used pseudo, and `relayedDoc` reuses the cascade an element patch just relayed, so a hover patch plus its redraw share one cascade.
- `internal/render/state.go`, `internal/page/page_draw.go`, and `a7576b39` wire `Display.Release`: `commitDisplay`/`paintFallback` release the previous display unless it is the painted baseline, and `TakeDirty` releases the older painted list.
- Behavior fixes: caret/anchor movement now forces a repaint; label clicks focus the named control; disabled controls keep the default cursor; OTP slots accept one digit and move focus; disabled select options are never landed on.
- Public additions: `type ElementPatch = page.ElementPatch` (root alias), `func (p *Page) PatchElements(ctx, []ElementPatch) error`, and `LoadOptions.AdoptTree`. Internal seams: `host.BoxSlicer`, `page.(*Page).BoxesSlice`/`BoxAt`, `render.State.SameStyleState`, and `render.RelayoutElements`. `pagetest.BoxByAction("")` now returns false, and both pagetest helpers read `Probe().Display().Boxes`, so they return false on a bitmap-fallback page where `Page.Boxes()` used to work.

### UI builder

- `ui/builder.go` replaces per-element attribute and style maps with an inline 4-entry set plus a slice overflow. `Px`/`Rem`/`Vh`/`Percent`/`Number` store numbers and units; `appendScalar` reproduces `strconv.FormatFloat(n, 'f', -1, 64)` at render time. Attribute order is preserved and names are sorted only when written.
- `ui/diff.go` adds the retained-tree diff. It requires equal tags, child counts, child order, and `id`s at every node, rejects structural changes, id renames, attribute removal, callback-presence changes, cycles, and depth overflow. Changed text/attributes/styles become `ownframe.ElementPatch` values; text patches are leaf-only. Caps are 64 non-text patches or 256 text-only patches.
- `ui/mount.go` makes `View.Update` patch when `Page.HTML() == v.html`: it renders a snapshot, calls `PatchElements`, then `Page.Load(..., AdoptTree: true)`. A refused patch or failed adopt rolls back to a full parse with `ReplaceCurrent: true`. `Mount` forces `Config.RetainState = true`; `NewPage` does not.
- `ui.Document(title, fragment)` assembles a page in one buffer. `controlID` uses manual FNV-1a over scope, tag, action attributes, and subtree text, so inserting a conditional sibling does not renumber an existing control.
- Allocation tests: `TestAllocBuildThousandRowTree` (34,000 allocs for 1,000 rows × 9 elements), `TestAllocRenderTree` (500 allocs), `TestAllocDocumentAssembly` (2 allocs), plus a frozen `naiveRenderer` golden (`f2fb5fc61eec12c1`, 48,064 bytes) and a stable-`controlID` test.

### Replay, window, and motion

- `internal/replay/chunk_plan.go` keys plan reuse on display pointer plus fingerprint, seals cached chunk signatures from tick-writable fields, and buckets plan items into 256 px tiles. `plan_cache_test.go` proves reuse and rebuild through build counters rather than timing.
- `DrawRects` adds a per-rect paint path. The union viewport path was removed during CI cleanup because nothing called it. **The tile path is still not wired into the window:** `internal/window` paints through `paintChunked`, which never reads the tile buckets, so plan builds pay the tile-index cost without the window benefit.
- Vector fills, strokes, segments, and corner paths share a `sync.Pool` of `*vector.Path`; `DrawPathOptions`/`StrokeOptions` are stack values and colours go through `ColorScale.Scale(...)` without interface boxing. GPU cache caps drop to 8 MB chunk, 8 MB raster, 8 MB text, and 64 MB decoded images per renderer.
- `internal/motion/transition.go` filters transition capture to ids that declare a transition in either display; the cssmotion index is cached in `motion.Runtime` and reused through `BindAfterWithIndex`, with `Repoint` on op-count matches. `internal/cssmotion/index.go` splits the `Index` type out of `ownership.go`.
- `internal/window/boxes.go` adds a zero-allocation `boxes()` path that prefers `display.Boxes`, then `host.BoxSlicer.BoxesSlice()`, then the page boxes; devtools hover/pick, context menu, drag-on-control, and content size all use it.
- `internal/emoji` grows from 31 to 42 color emoji. The eleven PNGs are **added**, not deleted: 👀 🧠 📌 🔋 🌮 🍜 🧈 🦀 🎨 🎧 👥. The ASCII gate still admits only `F0 9F` or `E2` + `9A/9C/9D/AD`, and the table test walks every entry.

### Components

- Seven new packages emit charts as plain HTML plus CSS; grep confirms no canvas, SVG, or script. `apexcharts` (30 constructors), `chartjs` (8), `echarts` (23), `nivo` (31), `recharts` (12), `shadcnchart` (6), and `tremor` (13) each export constructors returning `*ui.Element` plus a `Stylesheet`.
- Five sets share a copied div-paint engine (`of-apx`, `of-cjs`, `of-rch`, `of-shc`, `of-trm` prefixes); `echarts` and `nivo` are separate. Marks are sized divs and labelled grid cells. Pie/donut are stacked tracks, line/area are stems plus fill, treemap/sunburst are indented rows. This is intentional HTML/CSS fidelity, not canvas parity.
- ChadCN gets a large behavior pass: search hit-testing and backdrop ordering, disabled paint/cursor, field ARIA linkage (`aria-required`, `aria-invalid`, `aria-describedby`), calendar outside days, questionnaire counting, negative area-chart bands, menu hover rows, dialog/sheet/drawer fallback ids and dismissal, scroll translation, resizable ratio mapping, carousel dots, data-table/chart row ids, and a live CodeBlock Copy action.
- Stable part ids are added to nearly every option struct so `ui.View.Update` can patch text and attributes instead of reparsing. `ChoiceWithID`, `ProgressWithOptions`, `ResizableRatioAt`, `SliderValueAt`, `ScrollMetrics.OffsetAt`, and `ScrollState.Track` are new; `OnSelect` on carousel, `OnChange` on scroll area/resizable, and `State` on message scroller are additive.
- ChadCN now has 270 test functions, including `probe_patch_*` suites that mount a real page with `Perf: true` and require either exactly one generation with no new parse or no generation and no parse. Eight skip sites remain blocked on engine limits (percentage-height resizes, table-internal text relayout, scroll offset/scrollbar geometry, `option` tag).

### Examples and site

- `examples/blackhole`: three embedded Kage shaders (360-step raymarch, bloom, post), readback/JPEG bridge through `SetImage`/`SetData`, adaptive resolution, drag orbit with inertia, wheel zoom, and a `PatchElements` HUD. Desktop `Run`, `-web` static preview on `127.0.0.1:8155`, and `-bench`. Seven tests plus camera/DRS/HUD/heap benchmarks.
- `examples/builder/discord`: seeded servers/channels/members, a windowed 500-message transcript with a host-owned `ScrollState`, search palette, fullscreen settings, create-server dialog, profile card, context menu, toast, typing simulation on `SetTick`, per-channel drafts through `Page.Form`, custom chrome, and light/dark appearance. 37 tests; `TestWindowBounds` is skipped with a stale reason.
- `examples/testing`: ownframe 1,000-row invoice (`ui.Mount`/`View.Update`, deterministic seed 1, filter, selection, theme, patch/rebuild paths) and an isolated MyGo module (`go 1.27.3`, MyGo `v0.7.4`) that now defaults to a black-hole app with the invoice behind `-invoice`. Both expose `-bench`/`-bench-file` and Go benchmarks; the README records `grand-total=$1420529.43` parity.
- `examples/builder/inventory` adds stable `stock-row-<code>` and `inv-pill-<filter>` ids plus hover/patch tests. `ui-builder` adds seven chart-set tabs, 20–30 cards per set, a `CHART SETS` sidebar group, and slider drag mapping through `chadcn.SliderValueAt`.
- `examples/frontend/hello-builder` and `builder-mount` are the builder-pattern starters. `examples/compare` adds `BenchmarkSelectiveRepeatClick` and a no-op re-click test; its `run.sh` currently passes undefined `-auto-click`/`-animate-frames` flags and aborts.
- Website: homepage gains an HTML-template/builder toggle and starter picker; `showcase.js` adds Login, Trade, Loglens, Inventory, JMAP Mail, MailPulse, and UI-builder cards; code-view and JMAP get autoplaying MP4s; `build-demo.mjs` builds the UI builder to `demos/ui-builder/demo.wasm` (ignored) with the committed loader.
- Assets: nine demo screenshots and two MP4s are **added**; `jmap-mail.png` was added and replaced by `.webp` inside `a3628621`. Nothing in this slice is deleted in the final diff.

### Tooling, benchmarks, and docs

- `internal/perfcheck` checks redraw time and allocation against per-box/op budgets. It has no production caller yet; only its own tests use it.
- `internal/benchscore` parses `go test -bench` text with `x/perf/benchfmt`, classifies rows (idle, heap, interaction, rebuild, resize, typing), applies 2.0x generic and 1.0x idle/same-data budgets, and fails missing reference rows. `cmd/benchscore` exits 1 on failure; `make bench-gate` wraps it and is not part of `test` or `lint`.
- `scripts/watch-pprof.sh` captures heap/CPU/state from a debug build; `invoice-guibench.py`, `invoice-webbench.py`, and `mygo-guibench.py` measure native, web, and MyGo click-to-visible paths.
- `plans/v0.0.3/performance/` records the Phase 07 baseline table (304.1 ms full build, 55.7 ms few patches, 19,650 ms 1,000 patches, 264.8 ms stylesheet swap) and targets (`<=152 ms`, `<=28 ms`). Four of seven checklist items are checked; `OPT-07-04`, `-06`, and `-07` remain open.
- Docs: `documentation/components.md` (110 ChadCN constructors, quick start, limits), `documentation/builder.md` (`ui` guide, patch-based update, theme composition, limits), `documentation/starters.md`, 42-emoji update, and snippet exceptions. `AGENTS.md` adds `PatchElements` to the frozen `Page` list.
- `Makefile` adds `bench-gate`; `.golangci.yml` adds `Vh` to `mnd`; `.gitignore` ignores generated UI-builder wasm copies.
- `Händen` is a 0-byte accidental root file added by `ec1ec680`; it should be deleted.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Patches batch into one display; clean frames skip; form sync skips control-free pages; hit/box/kid/nav caches remove repeated walks; the UI builder avoids map allocations and full reparses; replay plans reuse across repaints; vector paths pool; transition capture maps only transitioning ids; window box reads allocate zero. |
| **Memory** | Pooled builders, chunk plans, vector paths, and display chunks are bounded; GPU caches halve to 8/8/8/64 MB; inline attr/style sets avoid per-element maps; `Display.Release` returns engine chunks. The fingerprint map is capped at 1,024 entries; tab/replay caches remain bounded by their existing caps. |
| **Behavior / correctness** | Patch updates require stable ids and unchanged structure and fall back to full renders; disabled controls drop hover paint; chart marks are divs, not polylines/arcs; table-internal text relayout and scroll-offset patches still skip; `Display.Release` makes patched displays safe after the previous display is released. |
| **API / CLI** | Adds `Page.PatchElements`, `ownframe.ElementPatch`, `LoadOptions.AdoptTree`, internal `host.BoxSlicer`, and `ui.Document`; keeps `Page.PatchElement`; adds seven component packages, `perfcheck`/`benchscore`/`benchscore` CLI, and `make bench-gate`; changes `pagetest.BoxByAction("")` to false. Additive fields on many option structs. No removals. |
| **Dependencies** | blinkless `1b8814741a85` → `dc1d518bff32`; adds direct `golang.org/x/perf` and indirect `go-moremath`; `go.work.sum` grows for x/perf's test graph. The examples module now pins the same `dc1d518` commit. Bumps `github.com/andybalholm/brotli` to `v1.2.7` for GO-2026-6888. |
| **Binary size / build time** | New demos and chart packages grow example compile/test time. `blackhole` links Ebiten graphics; `ui-builder` pulls seven chart sets; the WASM demo builds ~56 MB (ignored). No library binary-size claim is made. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| **Blinkless resolution changed from a local checkout to the module proxy** at `dc1d518`. | Nothing to do for normal builds or CI; `go mod download` fetches the pinned pseudo-version. A developer who relied on editing the sibling Blinkless checkout should add a temporary local replace rather than commit it. |
| Additive fields on exported ChadCN option structs break unkeyed literals. | Use keyed literals. v0 allows this. |
| A patch whose edits do not change layout (for example a `title` edit) can adopt the previous display; `PatchElements` still refuses a batch where no edit applies. | Treat a successful call as a commit, not a guarantee that layout moved; read the display for damage instead of assuming payload fields are populated. |
| `LoadOptions.AdoptTree` wins silently over `ReplaceCurrent` and rejects non-builder source with `page.ErrAdoptSource`, which is not re-exported at the root. | Use it only from the builder path; tests can match the internal sentinel or the error text. |
| `pagetest.BoxByAction("")` now returns false, and pagetest helpers read the committed display, so they return false on a bitmap-fallback page. | Pass a non-empty action; use `Page.Boxes()` when a fallback page must be inspected. |
| Disabled controls no longer paint hover fills. | Update visual tests/screenshots if they assumed hover paint on disabled controls. |
| `Mount` forces `Config.RetainState = true`. | Callers that relied on `Mount` leaving the option alone must set it explicitly or use `NewPage`. |
| `View.Update` keeps the tree until the next call; mutating it in place is unsupported. | Rebuild the tree and call `Update` again. |
| Chart packages are HTML/CSS approximations, not canvas parity. | Do not port canvas-only chart features; read each package README for its mark model. |
| `Händen` and `examples/compare/run.sh` are accidental/stale. | Delete `Händen`; fix or remove `run.sh`. |

---

## Library freeze (until Page wrap)

The `Page` wrap has landed; the checklist is kept for reviewer confirmation.

- [x] No new `type X =` at the module root except documented re-exports (`alias_allow_test.go` records `page_patch_el.go:ElementPatch`).
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows; `PatchElement`/`PatchElements` are real public entry points.
- [ ] No new example `ownframe/internal/` imports (`make internal-import-check`).
- [x] No synonym names.
- [ ] New exported methods return named types in package `ownframe` or a public subpackage.
- [x] Library code panics only for invariants; new user-facing failure paths return `error`.
- [x] `context.Context` is first when the function already does I/O, cancel, or deadline.
- [ ] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr. `discord` is a new chat clone; it is not one of the named products, but reviewers should confirm the spirit of this rule.

---

## Test plan

- [x] `make test` (library and examples).
- [x] `make lint` (library, examples, docs, plans, import policy, website ESLint).
- [x] `make test-race` (library, including the transition sampling tests).
- [x] `make build` (`go vet -p 1 ./... ./examples/...`).
- [x] `xvfb-run -a make graphics-check`.
- [x] `go test -p 1 -count=1 -run Sec ./... ./examples/...`.
- [x] Fuzz smoke for `FuzzReplace`, `FuzzSanitizeLine`, `FuzzBlockedIP`, `FuzzMethodHeader`, `FuzzDOMLimits`, `FuzzImageSrc`, and `FuzzSetCrashDir`.
- [x] `make docs-check` (39 snippets compiled, 12 illustrations).
- [x] `make plans-check` (24 rows with distinct states).
- [x] `make examples-check` (7 teaching roots, 61 experimental/support roots).
- [x] `make internal-import-check` (clean, 16 allowlisted pairs).
- [x] `npm --prefix frontend test` (13 tests).
- [x] `bash scripts/reproducible-check.sh` (archive checksums match).
- [x] `scripts/perf-baseline.sh` parent/candidate comparison (`--max-ratio 1.5`).
- [ ] `native-package` on Windows and macOS (CI only).
- [ ] `android-bind` (CI only).
- [ ] `make bench-gate` (manual, needs a reference/candidate pair).

Dependency-side validation: `go test -count=1 ./...` passes in the blinkless checkout at `dc1d518`.

### Commands

```sh
make test
make lint
make test-race
make build
xvfb-run -a make graphics-check
go test -p 1 -count=1 -run Sec ./... ./examples/...
npm --prefix frontend test
bash scripts/reproducible-check.sh
```

---

## Screenshots / sample output

- Demo assets: `assets/screenshots/demos/{code-view.webp,jmap-mail.webp,vskode-demo.webp,inventory.png,login.png,loglens.png,mailpulse.png,trade.png,ui-builder.png}` and `frontend/public/demos/{code-view.mp4,jmap-mail.mp4}`.
- UI-builder WASM preview: `frontend/public/demos/ui-builder/index.html`.
- Black-hole fallback: `examples/blackhole/scene_fallback.jpg`.
- Invoice parity checkpoint in `examples/testing/README.md`: `grand-total=$1408316.22` for the shared PCG seed. That table predates the later allocation commits; treat it as historical.

---

## Related issues

- Relates to #45 (reproducible benchmark harness).
- Relates to #59 (experimental `ui` builder).

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`, `benchmark`, `documentation`)
- [x] Related issues filled (`Relates to #45`, `Relates to #59`)
- [x] Filled body saved under `plans/PR/`

---

## Follow-ups (out of scope)

- Drop the local blinkless replaces, tidy both modules, commit sums, and align the examples pin to `dc1d518`.
- Fix the `boxCaches` map: `map[*Page]*pageBoxes` holds a strong key, so the `SetFinalizer` cannot run and a discarded page pins its box list, kid index, and last display.
- Complete `Display.Release` coverage: `adoptSpliced` does not release the display it replaces, and `CommitList`/`CommitBitmap` drop the unused field; extend `page_release_test.go`.
- Replace the `isBuilderDocument` string check with a verifiable marker so crafted HTML cannot skip `checkDOMLimits` or leave `{{` unparsed.
- Clear `commitRoots` when transitions are off, and check op ids/owners/order before reusing the motion index on an op-count match.
- Narrow damage for identical `Load`/`Back`/`Forward`; `markFull` currently repaints the content rect even when the source is unchanged.
- Validate `AdoptTree`'s cached tree against the source, or document that it trusts the caller; it ignores `ctx` and does not rescan form controls.
- Delete the 0-byte `Händen` file.
- Fix or remove `examples/compare/run.sh`; its flags are undefined.
- Refresh `examples/testing/README.md` for the MyGo black-hole default and the newer benchmark runs.
- Wire the tile-bucket replay paint into `internal/window`, or drop the build-time tile index.
- Give `internal/perfcheck` a production caller or fold it into the manual bench gate; add a committed reference baseline for `make bench-gate`.
- Unskip `TestWindowBounds` and refresh the Discord scroll comments.
- Reduce the five copied chart engines and remove the unused exported `Range*`/`Tracker*`/`Calendar*` types.
- Reconcile the Phase 07 plan status (`Completed` vs three open items) and the `docs/demos/ui-builder` `.gitignore` allowlist, which points at a directory that does not exist.
- Confirm the `Page.PatchElements` no-op contract and damage bounds match the blinkless fast-path behavior (display-only result, element-box damage).

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] No new root aliases, synonym methods, or example `internal/` imports
- [ ] New exported methods return named `ownframe` types
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
|-----------|------:|-----------:|----------:|
| `.css` | 24 | 370 | 39 |
| `.go` | 367 | 41,203 | 1,121 |
| `.html` | 1 | 108 | 0 |
| `.jpg` | 1 | Binary | Binary |
| `.js` | 3 | 88 | 1 |
| `.json` | 3 | 18 | 0 |
| `.jsx` | 1 | 49 | 14 |
| `.kage` | 6 | 501 | 0 |
| `.md` | 23 | 1,074 | 17 |
| `.mjs` | 1 | 1 | 0 |
| `.mod` | 3 | 19 | 4 |
| `.mp4` | 2 | Binary | Binary |
| `.png` | 18 | Binary | Binary |
| `.py` | 3 | 848 | 0 |
| `.sh` | 2 | 138 | 0 |
| `.sum` | 4 | 60 | 32 |
| `.txt` | 2 | 24 | 0 |
| `.webp` | 3 | Binary | Binary |
| `.yml` | 1 | 1 | 1 |
| No extension | 3 | 19 | 0 |
| **Total** | **471** | **44,521** | **1,229** |
