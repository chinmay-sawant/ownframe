# Improve codebase architecture: path from 7.3 to 10

> Parent: `plans/v0.0.2/code-review/critical-golang-architecture-review.md` and `plans/v0.0.2/code-review/improve-codebase-checklist.md`
> Frozen at: `e2886cb` on `chore/review`, clean tree, 2026-10-16
> Scope: root `*.go` plus `internal/...` plus `examples/...` weight and `documentation/...`
> Status: plan only, no code changed in this file

## Score as of right now

Overall: 7.3 out of 10.

This is a straight average of five fresh sub-scores: architecture 8, correctness 8.5, maintainability 6, tests 6, docs 8. Sum is 36.5 across 5, so 7.3.

The tree is the same commit as yesterday, `e2886cb`, with a clean `git status`. Nothing in the code moved. The number moved from 6.9 because todays probes verified that the prior fixes actually landed. Take drains, the shared masking policy, the canonical sentinel, `boxAt`, the link classifier, and the front selector all held up under re-reads. The drag is unchanged: examples weight, committed binaries, and the two large types.

## Why 7.3

What earned it:

- Root plus `internal/...` has zero Go files over 2000 chars. The audit command `git ls-files '*.go' | grep -v '^examples/' | grep -v '^temp/' | xargs wc -m | awk '$1>2000'` returns only the total line. Yesterday I measured `internal/replay/draw_rect.go` at 1995 and `internal/window/update.go` at 1979, so the cap holds but files are packed tight.
- Paint policy is now shared. `internal/masking/paint_color.go:23`, `internal/masking/grid_line.go:10`, `internal/masking/blend.go:17`, and `internal/masking/image_decode.go:19` give both sinks one alpha, grid, mode, and decode contract.
- Contracts that lied now tell the truth or are fixed. `internal/page/page_patch_props.go:28` drains on Take, `internal/page/form_input.go:5` guards disabled controls, `internal/page/page_boxat.go:13` is the one scan, `internal/page/page_classify_link.go:18` is the one parser, `internal/page/page_front.go` is the one front selector, `internal/state/path_split.go:5` rejects empty segments.
- Error identity is canonical. `fetch.ErrNilContext` is aliased in `internal/filepick/filepick.go:11`, `internal/web/serve_errors.go:10`, `internal/ipc/ipc.go:19`, and `internal/page/page.go:29`.
- Docs state the hard limit honestly. `internal/page/print.go:27`, `documentation/printing.md:4`, `documentation/features.md:101`, and `documentation/screen.md:13` agree that no PDF is ever written.

What costs the remaining 2.7:

- `internal/page/page_types.go:19` still holds around 45 fields plus `pageGeom`, `pageContent`, `frameState`, and `accessibilityState` embeds. Form, viewport, scroll, and accessibility state live in one struct.
- `internal/host/screen.go:22` still carries about 25 methods after `Boxer` was split out. Every fake implements pointer, keyboard, and clipboard even when the sink uses two of them.
- Scale without map. `internal/window` has 189 non-test files and `internal/page` has 170. `internal/replay` has 63 test files. Devtools, idle, touch, and tile code form subsystems with no package line.
- Documented but unenforced. `internal/frame/anim.go:42` says an animation must not outlive a Redraw, but nothing invalidates it. The replay to render rule lives in a comment in `render/paint.go` with no gate test.
- Stale reads. `internal/clipboard/clipboard.go:34` returns in-memory text when the OS read fails, so callers cannot tell stale from live. `internal/crash/validate.go:67` accepts relative paths, so report location depends on working directory. `internal/print/command_args.go:30` rejects any dot segment, which blocks benign relative paths while doing no absolute or canonical check.
- Dead end without probe. `internal/page/print.go:27` always returns `ErrNoPDF`, and bake stays under the chunk lock by design in `internal/replay/chunk_cache.go:11`, so callers and contributors hit runtime surprises instead of compile or probe signals.
- Repo weight. Four example files break the stated cap: `examples/spotify-player/player/playback.go` at 2357, `logger.go` at 2238, `actions.go` at 2097, `display_patch_helper_test.go` at 2028. Six release APKs at 28 to 30 MB each sit in git, about 170 MB total, plus `teams.mp4` at 4.4 MB committed twice under `docs/demos/` and `frontend/public/demos/`.
- Pin drift. `go.mod:6` and `examples/go.mod:6` pin blinkless `c085b3423260` while `AGENTS.md:43` still names `34d93862f9ed`.

## What 10 means here

A 10 is not perfect code. It is a tree with no known structural blocker, no committed weight that slows every clone, no doc that disagrees with source, and gates that fail the build when a rule breaks. Every row below has a named proof so a validator can re-run it without asking the author.

## Phase A: shrink the two large types

### ARCH-01 form and focus state leaves Page

Location: `internal/page/page_types.go:19`, form fields, `hover`, `active`, `pressUsed`, accessibility state.

Change: add one owned struct for form and focus state with narrow methods for set, clear, and hit routing. `Page` delegates. No behavior change.

Proof: grep shows zero direct writes to the moved fields outside the owner, page suite green, `page_types.go` stays under cap.

### ARCH-02 geometry, content, and frame ownership is written down

Location: `internal/page/page_types.go`, `pageGeom`, `pageContent`, `frameState`, `windowingState` embeds.

Change: either split content loading from geometry or add a short ownership comment per embed naming the single writer and the lock. Prefer the comment first, split only if two writers appear.

Proof: each embed has a named owner in code, grep confirms one writer per field group.

### ARCH-03 Screen becomes roles plus compat shim

Location: `internal/host/screen.go:22`, `internal/host/box.go`, `internal/host/window_only.go`.

Change: keep `Screen` for compat. Add `Pointer`, `Keyboard`, and `Clipboard` role interfaces. Move window and web call sites to the narrowest role they need. Fakes implement roles, not the full surface.

Proof: window and web build against roles, old `Screen` callers still compile, host suite green.

### ARCH-04 window subsystems get a map

Location: `internal/window`, 189 non-test files.

Change: no package split yet. Add one index doc that maps file groups to owners: devtools, idle, touch, pointer, replay tile, reload. Each group names its entry file and its state owner.

Proof: new contributor can find the idle owner and the tile owner from the index without grep across 189 files.

### ARCH-05 replay to render rule gets a gate

Location: `internal/replay/`, `internal/render/paint.go`.

Change: add a test that lists production imports of replay and fails on any import of render outside an allow list. The current zero stays zero by build, not by comment.

Proof: gate test green, temporary render import fails the gate.

## Phase B: enforce instead of document

### CORR-01 anims carry a generation

Location: `internal/frame/anim.go:42`.

Change: record the display generation at anim creation. `Step` checks it and goes inactive when a Redraw has replaced the list. Alternative accepted: copy target values at creation so no pointer into the old list survives.

Proof: test creates an anim, forces a Redraw, steps, and asserts no patch lands on freed ops.

### CORR-02 print gets a probe

Location: `internal/page/print.go:27`.

Change: add a `SupportsPDF` or `Available` probe next to `PDF`, `WritePDF`, `SavePDF`, and `Print` so callers branch before calling. Keep `ErrNoPDF` as the error for direct calls.

Proof: unit asserts the probe is false on every sink, docs show the branch pattern.

### CORR-03 chunk bake policy is measured in the comment

Location: `internal/replay/chunk_cache.go:11`.

Change: keep bake under the lock, but record the bench name and the race reason in the comment so the next reader does not retry the narrowing blindly.

Proof: comment cites bench and invariant, suite green under `-race`.

## Phase C: stop stale reads

### CORR-04 clipboard read stops returning stale text

Location: `internal/clipboard/clipboard.go:34`.

Change: return an error or `ok=false` when the OS read fails. Keep the in-memory copy for `UseMemory` tests only.

Proof: failing OS read returns an error and no text, memory mode still works in tests.

### CORR-05 crash dir is absolute at set time

Location: `internal/crash/validate.go:67`.

Change: reject relative paths or resolve them to absolute once and store the absolute form.

Proof: relative input is rejected or stored absolute, test asserts the stored path is absolute.

### CORR-06 print path is canonicalized, not substring filtered

Location: `internal/print/command_args.go:30`.

Change: clean and absolutize the PDF path, then reject traversal after cleaning. Allow benign relative paths.

Proof: table over benign relative, traversal, and absolute paths passes.

## Phase D: repo weight and example discipline

### REPO-01 four example files meet the cap or get an exemption

Locations: `examples/spotify-player/player/playback.go`, `logger.go`, `actions.go`, `display_patch_helper_test.go`.

Change: split each to at or under 2000 chars, or write the exemption into `AGENTS.md` if examples are out of scope. Either way the rule stops being half true.

Proof: `git ls-files 'examples/**/*.go' | xargs wc -m | awk '$1>2000'` returns only the total, or the exemption names these paths.

### REPO-02 binaries leave git

Locations: six `examples/*/android/artifacts/*-release.apk`, `docs/demos/teams.mp4`, `frontend/public/demos/teams.mp4`.

Change: remove APKs from git history or move them to release artifacts or LFS. Keep one `teams.mp4` and link it from the other spot.

Proof: `git ls-files | grep -E 'apk$|teams.mp4$'` shows zero APKs and one mp4, clone size drops by about 170 MB.

### REPO-03 pin and stale notes are synced

Locations: `AGENTS.md:43`, `go.mod:6`, `examples/go.mod:6`, `README.md:31`.

Change: update `AGENTS.md` to blinkless `c085b3423260`, fix README rename and frontend notes to match the tree.

Proof: grep shows one pin across all three files, README paths resolve.

### TEST-01 replay tests merge by family

Location: `internal/replay/`, 63 test files including 3 `chunk_parity_*`, 4 `chunk_bench_*`, about 10 `gpu_parity_*`.

Change: merge each family into one file per family with table subtests. No coverage loss.

Proof: file count drops, `go test ./internal/replay/` green, parity and bench cases still present by name.

## Phase E: gates that keep 10

Run after every phase, record exits:

- `make build` exit 0, which runs `go vet -p 1 ./... ./examples/...`
- `make test` exit 0, 97 packages ok, zero FAIL, plus `-race` on `internal/state`, `internal/ipc`, `internal/replay`, and `internal/page`
- `gofmt -l` over changed trees returns empty
- `git ls-files '*.go' | xargs wc -m | awk '$1>2000'` returns only totals, or the written exemption covers the hits
- `git ls-files | grep -E 'apk$|mp4$'` shows no release binaries
- grep shows one blinkless pin and zero replay to render production imports

Estimated effort: Phase A about 4 days, Phase B about 2 days, Phase C about 2 days, Phase D about 2 days, Phase E rides with each phase. Total near 2 weeks for one engineer, less with two validators re-running proofs as the prior waves did.

## Dependencies

- Blinkless engine for parse, cascade, and layout, pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, and text.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules proposed.

## Risks

- Screen role split touches every host fake. Keep the old interface as a shim so examples compile untouched.
- Clipboard signature change may break callers that ignore the ok flag. Prefer the error return and fix call sites in the same row.
- APK removal rewrites history or breaks links. Prefer moving to release artifacts with redirects from docs.

## Parked

Revisit only with fresh evidence: scroll offset triple ownership, sink capability matrix comment on `host/screen.go`, dual threading contract, web string-only errors, reload-note dedup, scroll and zoom unit math, text shaping face cache sharing.
