# Improve codebase - review ledger

> **Parent:** `plans/v0.0.2/code-review/critical-golang-architecture-review.md` - prior ledger, all rows closed
> **Status:** review complete, implementation open
> **Estimated effort:** 2 days for P0, 1 week for P1, 1 week for P2 and P3 combined

---

## Overview

This is a review-only wave under the improve-codebase pack: architecture deepening, extension seams, and Go practices. Ten read-only agents scanned the library at frozen HEAD `66ed433` on branch `chore/review` with a clean tree. Scope is root `*.go` plus `internal/...`. Scope excludes `examples/...`. Nothing was implemented, no test or build ran, and nothing is committed by this wave.

Findings in: 71 blocks from 10 agents. Rows out: 25 active. Refused: 3 agent blocks plus settled-ledger re-files. Parked: 24 lookalikes in the overview below, not fake phases. Dedup key was path plus absorbing module. The SetImage ownership write-up appeared three times and the nil-context split appeared three times; each collapsed to one row with the strongest evidence kept.

Placement note: the pack default is `plans/reviews/improve-codebase/`, but this review was ordered into `plans/v0.0.2/code-review/`, so both artifacts live there. No `plans/reviews` tree was created.

## Executive Summary

The prior wave closed its ledger by fixing guards, locks, aliases, and seams. This wave asked a different question: where does the next honest feature go, and will the code force a fork. The answer is mostly healthy. The pipeline stages, the host optionals, and the sentinel discipline held up under three lenses.

What fails is paint parity between the two sinks. Translucent alpha, GridRun expansion, image prepare, text shaping, and viewport math each live twice with visible drift. Four of the six P0 rows sit in that fork. The other two are contract breaks: Take methods that never drain, and exec seams that panic on nil context. P1 is dominated by one theme as well: the nil-context sentinel exists in four identities, and the state store accepts dotted paths it cannot resolve.

## Phase 1: P0 Integrity

Security, output, and broken public contracts. All six rows are defects with current-source proof.

### 1.1 Take and routing contracts

- [ ] `EXT-01` `internal/page/page_patch_props.go:25` - Take aliases return a snapshot without draining; expected behavior: Take drains changedProps or is renamed to snapshot; proof: unit that sets one prop, takes twice, and asserts the second take is empty.
- [ ] `EXT-02` `internal/page/form_input.go:5` - disabled guard covers button only, so a disabled checkbox still invokes the app Click handler; expected behavior: any disabled control returns before the handler; proof: unit with disabled checkbox, text, and file asserting zero handler calls, unchanged value, and unchanged generation.

### 1.2 Shared paint policy

- [ ] `EXT-03` `internal/bitmap/color.go:14` and `internal/replay/color.go:15` - bitmap composites unmultiplied RGB with Porter-Duff while replay premultiplies alpha into RGB; expected behavior: one Opacity to color helper shared by both painters, write-out stays separate; proof: unit painting one translucent FillRect through each sink and comparing pixels.
- [ ] `EXT-04` `internal/bitmap/op.go:13` and `internal/replay/grid.go:12` - bitmap retains owner transform, fixed, and alpha on GridRun segments while replay rebuilds a bare struct; expected behavior: one GridRun to Line expansion both painters call; proof: unit with a transformed, fixed, or translucent GridRun asserting both sinks agree.

### 1.3 Entry guards

- [ ] `PRAC-01` `internal/print/print.go:50` and `internal/filepick/pick_linux.go:32` - exec-backed entries forward context to `exec.CommandContext` with no nil check, so nil panics instead of returning an error; expected behavior: reject nil with the canonical ErrNilContext; proof: nil-context tests asserting `errors.Is` and no panic.
- [ ] `ARC-01` `internal/web/serve_build.go:21` - serve substitutes Background for nil context and never checks a nil app, so a nil screen becomes an HTTP 500 instead of a sentinel; expected behavior: mirror the window nil guards or document the divergence; proof: nil inputs return matchable sentinels.

## Phase 2: P1 Seams

Job type, DAG leaks, and half-wired extension points. Each row is one change in the named absorbing module.

### 2.1 Sentinel and error identity

- [ ] `ARC-02` `internal/fetch/doc.go:12` is canonical ErrNilContext; `internal/ipc/ipc.go:14`, `internal/page/page.go:28`, and `internal/window/run.go:33` re-construct it - alias all three to the canonical value; expected behavior: one `errors.Is` branch works across Fetch, Request, Prepare, and Run; proof: cross-package identity test plus converting the substring asserts in `internal/fetch/context_test.go:35` and `internal/ipc/request_test.go:53` and the `==` checks in fetch, ipc, and page tests to `errors.Is`.
- [ ] `ARC-03` `run.go:46` and `report.go:13` - panic path stringifies with `%v` while the error path wraps with `%w`; expected behavior: one report helper both call that keeps the chain; proof: recovered panic report matches `errors.Is`.

### 2.2 State store contracts

- [ ] `EXT-05` `internal/state/store_patch.go:6` - PatchChange appends to changes and undo and notifies listeners even when setPath fails; expected behavior: return failure and skip all three side effects; proof: invalid-path unit asserting no event and no undo entry.
- [ ] `EXT-06` `internal/state/path_read.go:8` and `internal/state/path_write.go:11` - read splits on dots while write drops trailing empties, so `a.` writes one field and reads another; expected behavior: one shared splitter that rejects empty segments; proof: round-trip table over `a.`, `a..b`, `.a`, and empty.
- [ ] `EXT-07` `internal/state/patch_projected.go:26` - Patch skips missing values silently, writes origin outside the lock at line 49, and notifies per path in map order; expected behavior: report skipped paths, hold one lock for the origin write, batch one notify; proof: deny unit plus `-race` origin test.

### 2.3 Boundary ownership

- [ ] `ARC-04` `internal/page/page_set_image.go:24` stores caller bytes without cloning, and `internal/page/page_image.go:40` plus `:81` return the live PNG and box slices; expected behavior: clone on intake, copy on return; proof: mutate-after-set tests for images, PNG bytes, and boxes.
- [ ] `ARC-05` `internal/page/page_nav_route.go:16`, `page_new.go:20`, and `theme.go:30` return bare ErrSourceTooLarge while `page_source.go:74` wraps both sentinels; expected behavior: one size guard helper builds the identical wrapped error at all six sites; proof: table test over New, Load, Route, SetTheme, and poll asserting both sentinels.

### 2.4 Import direction and sink seams

- [ ] `ARC-06` `internal/replay/fill.go:18` and four sibling call sites resolve radii through `render.FillRadii`, a one-line shim over `masking.RadiiXY`; expected behavior: call masking directly and delete the render wrappers; proof: grep shows zero replay to render production imports plus green suites.
- [ ] `EXT-08` `internal/window/bounds.go:5` and `internal/window/bitmap_viewport.go:11` assert private `maxSizer` and `fallbackDisplay` that host never declares and web never honors; expected behavior: promote both to host one-method optionals or mark them window-only; proof: host declares them or the comment states the local contract.
- [ ] `EXT-09` `internal/window/sync.go:19`, `internal/web/page_http.go:20`, and `internal/web/debug_http.go:37` each reimplement display-else-image size and fallback preference; expected behavior: one front-half selector behind the page that all three call; proof: three-state unit (display present, image only, neither) exercised through both sinks.
- [ ] `EXT-10` `internal/page/page_source.go:48`, `internal/page/page_reload.go:64`, and `internal/page/page_reload_retry.go:11` call `os.Stat` and `os.ReadFile` directly, so every reload test stages real temp files; expected behavior: a two-method sourceReader seam with an os adapter and a memory fake; proof: PollReload against the fake with zero temp dirs.

## Phase 3: P2 Shared forks

Prepare, paint, and settings helpers that already drifted. Fix the shared table, keep both adapters.

### 3.1 Adapter pairs

- [ ] `EXT-11` `internal/bitmap/image.go:20` and `internal/replay/image_cache.go:33` - bitmap decodes per paint and fails loudly while replay caches and skips silently, with different alpha compositing; expected behavior: one decode, error, and alpha contract, forking only at blit versus upload; proof: bad-payload and translucent image units per sink.
- [ ] `EXT-12` `internal/page/page_deeplink.go:26` and `internal/page/page_link.go:25` - two parsers enforce overlapping scheme policy in different shapes with no shared deny proof; expected behavior: one classifyLink helper plus a shared allow and deny corpus; proof: trust unit covering allow, javascript deny, empty-host http, and relative fallthrough through both entries.
- [ ] `EXT-13` `internal/page/page_click.go:67` and `internal/page/pointer_state.go:81` - two linear scans with different winner rules papered over by a fallback second scan; expected behavior: one boxAt method on the box index that both delegate to; proof: overlapping-box fixture where Click, Press, and Hover agree.
- [ ] `ARC-07` `internal/bitmap/blend.go:8`, `internal/render/replayable_group.go:37`, and `internal/replay/group_composite.go:55` - the same modes enumerated three times with opposite polarity and a silent zero default; expected behavior: one canonical table with a contract test pinning gate-accepted modes to nonzero shader ids; proof: the contract test plus a diff of the three lists.
- [ ] `ARC-08` `internal/bitmap/bitmap.go:14` swallows err to nil while `Paint` at `:20` returns it, and `internal/render/paint.go:3` plus `:30` and `document.go:32` reference a `replay.Picture` that does not exist; expected behavior: one fallible write-out with explicit handling at call sites, and comments naming the real fillers; proof: grep shows zero stale references plus green suites.

### 3.2 Scoping risk

- [~] `ARC-09` `internal/ipc/ipc.go:26` - registry is process-global keyed only by channel, so two pages share one namespace with no cap; intentionally deferred as risk because scoping changes the public surface; owner boundary: repo owner; next gate: a two-page same-channel test proving collision or proving isolation, then scope per owner or document the single-owner constraint with a cap.

## Phase 4: P3 Locality

God-adjacent helpers, stale docs, and oracles that lie. Small blast radius, no behavior change except where stated.

### 4.1 Extraction and honesty

- [ ] `ARC-10` `internal/page/page_dirty_take.go:10`, `page_dirty.go:68`, `page_invalidate.go:30`, and `page_new_parts.go:40` - dirty pending, last, rect, and full bookkeeping lives in four places and needs a full page fixture to test; expected behavior: a dirty tracker struct owns mark, take, and coalesce while Page delegates; proof: tracker-only test driving rects past the collapse threshold.
- [ ] `PRAC-02` `internal/page/page_source_test.go:62` and `repaint_equality_test.go:47` assert byte-identical PNGs instead of pictures, and `documentation/printing.md:3` plus `documentation/features.md:100` promise PDF rendering the source refuses with ErrNoPDF; expected behavior: compare decoded pixels or box geometry, and rewrite the printing paragraphs to the ErrNoPDF contract; proof: pixel diff on both cases plus doc grep against the source refusal.

## Phase 5: P4 Closure

- [ ] Run `make build` after the implementation wave and record the exit.
- [ ] Run `make test` after the implementation wave and record pass counts.

## Dependencies

- Blinkless engine for parse, cascade, and layout. Pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, and text.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules are proposed in this review.

## Parked P3 list

Not rows. Revisit only with fresh evidence. Risk entries name their validation.

- Scroll offset has three owners (`window/game_state.go:43`, `scroll_window.go:8`, `sync.go:66`); declare shell the owner or prove the mirror never drifts.
- Sink capability matrix on `host.Screen` (web drives 4 methods plus Selector, window drives all plus 35 asserts); comment-only.
- Dual threading contract (window single-thread run ctx versus web mutex plus request ctx); comment on `host/screen.go`.
- Web string-only errors (`web/page_http.go:50`, `http_security.go:39`); bless with comments or export sentinels.
- Shared reload-note dedup (`window/reload.go:48`, `web/reload.go:30`); one helper, throttle difference documented.
- Scroll, zoom, and fixed math in three unit systems (`bitmap/viewport.go:12`, `replay/image_key.go:33`, `window/bitmap_viewport.go:29`).
- Text shaping stack per sink (`bitmap/font.go:17`, `replay/text_cache.go:32`); share policy, keep face caches.
- Partial repaint drops fixed ops unless the caller repaints them (`replay/draw_rect.go:37`); compose one entry or pin the contract. Validation: dirty-rect-over-fixed unit.
- Pixel budgets in six constants at two magnitudes (`bitmap/validate.go:10`, `masking/limits.go:3`, `render/replayable_group.go:13`); one table with reasons.
- Op box geometry in replay versus frame (`replay/op_bounds.go:18`, `frame/anim_bounds.go:13`); share helper or pin superset contract.
- Global replay image cache ownership (`replay/image_cache.go:29`); scope per display or document affinity. Validation: cross-page eviction test.
- Shared path policy across print, crash, and filepick (`print/command_args.go:16`, `crash/validate.go:31`, `filepick/pick_wsl_path_safe.go:12`); one helper or documented trust split.
- ScrollGate and ViewportPinner doubles (`window/scroll_gate.go:9`, `page/page_pinz.go:7`); add exercised fakes.
- OnFieldChange cancel (`state/listeners.go:51`); return cancel like `ipc.Listen` and settle empty path.
- Nil Option panic (`options.go:10`); skip nil plus round-trip test.
- Print and filepick deny tables (`print/command_args.go:9`, `filepick/pick_wsl_path_safe.go:12`); allow-plus-deny units.
- Fetch and focus alias markers (`fetch/scheme.go:30`, `state/focus_store_aliases.go:1`); Deprecated shims matching the IPC convention.
- Print forwarder (`print/command_args.go:41`); call ValidPath directly.
- Nil-form accessor (`page/form_apply.go:4`); sweep test under `-race` plus guard-count canary.
- Bind file-as-string matrix (`page/bind.go:27`); document and add round-trip row.
- Cursor disabled shape (`page/page_cursor.go:8`); return default for disabled.
- Caret up, down, and unknown no-op lock (`page/form_caret_keys.go:23`); generation-unchanged test.
- PatchOp perf and last-rect coverage (`page/page_patch_display.go:47`); route through mark owners.
- Control tag pipeline order (`page/form_scan.go:30`); one allowlist comment plus round-trip test.
- ClipRect receiver mix (`runtimebehavior/transition.go:59`); unify receivers.

## Refused

- Collapse ten source-size aliases (`page/page_source_limit.go:10`); deletion-only churn, no behavior changes.
- Window shell context storage (`window/game_state.go:23`); the single documented Ebiten-forced exception.
- Serve cancel and PNG context re-files; closed rows with no current-source regression.
- Closed critical-review rows re-filed without regression proof; returned with pointers.
