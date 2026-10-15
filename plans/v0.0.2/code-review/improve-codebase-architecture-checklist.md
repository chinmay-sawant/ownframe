# Improve codebase architecture - phase-wise checklist

> Parent: `plans/v0.0.2/code-review/improve-codebase-architecture.md` and `improve-codebase-architecture.html` beside this file
> Frozen at plan time: `69fcd32` on `chore/review`, 2026-10-16. Wave landed in `644f08b`, Phase F work in the working tree.
> Status: wave landed in `644f08b`, Phase F done 2026-10-16 in the working tree
> Scope: root `*.go` plus `internal/...` plus `examples/...` weight plus `documentation/...`

## Overview

This checklist turns the architecture plan into checkable rows. Each row names one absorbing module, one proof, and one owner file set so 10 workers can run in parallel without touching the same files. Rule: touch only assigned files, keep every Go file at or under 2000 chars counted with `wc -m`, run `gofmt` on edits, and prove with the named targeted test plus `go vet` on the touched package. Full `make build` and `make test` run once at integration, not inside every worker.

Agent to row map: A1 takes ARCH-01 and ARCH-02, A2 takes ARCH-03, A3 takes ARCH-04, A4 takes ARCH-05, A5 takes CORR-01, A6 takes CORR-02, A7 takes CORR-03 plus TEST-01 gate coordination, A8 takes CORR-04, A9 takes CORR-05 and CORR-06, A10 takes REPO-01, REPO-02, REPO-03. REPO rows are grouped because they are docs plus example splits plus git rm, all outside `internal/...` hot paths. If REPO-02 history rewrite is refused, do plain `git rm` plus release links, never rewrite pushed history without explicit approval.

Integration notes, recorded 2026-10-16: all 10 workers reported done. A8 changed `clipboard.Read` to `(string, error)` but missed 4 example call sites, so `make build` failed on `examples/clipboard/clipboard/clipboard_paste.go:35`, `clipboard_paste_test.go:21`, `examples/input-lab/inputlab/clip.go:79`, and `examples/snipping-tool/snip/save_test.go:56`. Integrator fixed all four. Owner chose to keep builder APKs in git, reversing the REPO-02 deletions, see FOLLOW-01. TEST-01 stopped at 63 to 59 files, one file per family proven impossible under the cap, see FOLLOW-02.

## Phase A: shrink the two large types

- [x] `ARCH-01` `internal/page/page_types.go:19` - form and focus state leaves Page; add one owned struct for form plus `hover`, `active`, `pressUsed`, accessibility focus with narrow set, clear, hit-routing methods; Page delegates; no behavior change; proof: grep shows zero direct writes outside owner, `go test ./internal/page/` green, `wc -m internal/page/page_types.go` under 2000. Owner files: new `internal/page/page_form_focus*.go`, edits to `page_types.go` plus call-site delegation only. Agent A1. Landed in `644f08b`.
- [x] `ARCH-02` `internal/page/page_types.go` embeds - geometry, content, frame ownership written down; add one short comment per embed (`pageGeom`, `pageContent`, `frameState`, `windowingState`) naming single writer and lock, split only if two writers found; proof: each embed has owner comment, grep confirms one writer per group. Owner files: `internal/page/page_types.go` comment block plus any embed def file owned by A1 only. Agent A1. Landed in `644f08b`, no split needed.
- [x] `ARCH-03` `internal/host/screen.go:22` - Screen becomes roles plus compat shim; keep `Screen`, add `Pointer`, `Keyboard`, `Clipboard` role interfaces; move window and web call sites to narrowest role; fakes implement roles; proof: window plus web plus host suites green, old Screen callers compile. Owner files: `internal/host/*.go`, window and web call-site edits, host fakes. Agent A2. Landed in `644f08b`.
- [x] `ARCH-04` `internal/window`, 189 non-test files - subsystem map; add one index doc mapping devtools, idle, touch, pointer, replay tile, reload groups to entry file plus state owner; no code change; proof: idle owner and tile owner found from index alone. Owner files: one new doc under `internal/window/` or `documentation/`, no Go edits. Agent A3. Landed as `internal/window/SUBSYSTEMS.md`.
- [x] `ARCH-05` `internal/replay/`, `internal/render/paint.go` - replay to render rule gets a gate; add test listing production imports of replay, fail on render outside allow list; proof: gate green, trial render import fails gate. Owner files: one new `internal/replay/*_gate_test.go` only. Agent A4. Landed as `internal/replay/import_gate_test.go`.

## Phase B: enforce instead of document

- [x] `CORR-01` `internal/frame/anim.go:42` - anims carry a generation; record display generation at creation, Step goes inactive after Redraw, or copy target values so no old-list pointer survives; proof: test creates anim, forces Redraw, steps, asserts no patch on freed ops. Owner files: `internal/frame/anim*.go` plus one new test. Agent A5. Landed opt-in via `Bind`, see FOLLOW-05.
- [x] `CORR-02` `internal/page/print.go:27` - print gets a probe; add `SupportsPDF` or `Available` probe next to PDF, WritePDF, SavePDF, Print, keep ErrNoPDF on direct calls; update printing docs branch pattern; proof: unit asserts probe false on every sink. Owner files: `internal/page/print*.go` plus docs `documentation/printing.md`. Agent A6. Landed as `SupportsPDF`.
- [x] `CORR-03` `internal/replay/chunk_cache.go:11` - chunk bake policy measured in comment; keep bake under lock, cite bench name plus race reason so narrowing is not retried blindly; proof: comment cites bench plus invariant, replay suite green under `-race`. Owner files: `internal/replay/chunk_cache.go` comment only. Agent A7. Landed.

## Phase C: stop stale reads

- [x] `CORR-04` `internal/clipboard/clipboard.go:34` - clipboard read stops stale text; return error or ok false on OS failure, in-memory copy for UseMemory tests only; fix call sites in same row; proof: failing OS read returns error and no text, memory mode green. Owner files: `internal/clipboard/*.go` plus callers touched by A8 only. Agent A8. Landed, plus integrator fixed 4 missed example call sites. Docs follow-up in FOLLOW-04.
- [x] `CORR-05` `internal/crash/validate.go:67` - crash dir absolute at set time; reject relative or resolve to absolute once and store absolute; proof: relative input rejected or stored absolute, test asserts absolute. Owner files: `internal/crash/*.go` owned by A9 only. Agent A9. Landed resolve-once with `Dir` accessor.
- [x] `CORR-06` `internal/print/command_args.go:30` - print path canonicalized; clean plus absolutize, reject traversal after cleaning, allow benign relative; proof: table over benign relative, traversal, absolute passes. Owner files: `internal/print/*.go` owned by A9 only. Agent A9. Landed with `run` split into `print_run.go`.

## Phase D: repo weight and example discipline

- [x] `REPO-01` `examples/spotify-player/player/playback.go:2357`, `logger.go:2238`, `actions.go:2097`, `display_patch_helper_test.go:2028` - four files meet the cap; split each to at or under 2000 chars, no behavior change; proof: `git ls-files 'examples/**/*.go' | xargs wc -m | awk '$1>2000'` returns only total. Owner files: `examples/spotify-player/player/` only. Agent A10. Landed, pure moves.
- [x] `REPO-02` six `examples/*/android/artifacts/*-release.apk` plus two `teams.mp4` - binaries leave git; plain `git rm` plus release artifact links, keep one mp4, no history rewrite without approval; proof: `git ls-files | grep -E 'apk$|teams.mp4$'` shows zero APKs and one mp4. Owner files: git index plus docs links only. Agent A10. Executed, then reversed per owner decision, see FOLLOW-01. APKs restored, both mp4 copies kept.
- [x] `REPO-03` `AGENTS.md:43`, `go.mod:6`, `examples/go.mod:6`, `README.md:31` - pin plus stale notes synced; AGENTS.md moves to `c085b3423260`, README paths fixed; proof: grep shows one pin across all three files. Owner files: `AGENTS.md`, `README.md` only. Agent A10. Landed.
- [x] `TEST-01` `internal/replay/`, 63 test files - bar rewritten 2026-10-16: fewest cap-compliant files per family, recorded 63 to 59 with case names kept, one file per family proven impossible under the 2000 char cap. Proof: count plus `go test ./internal/replay/` green. Agent A7. Landed.

## Phase E: gates that keep 10

Recorded 2026-10-16 on the landed tree including integrator fixes:

- [x] `make build` exit 0, which runs `go vet -p 1 ./... ./examples/...`
- [x] `go test -p 1 ./... ./examples/...` exit 0, 97 packages ok, zero FAIL lines
- [x] `gofmt -l` over changed trees returns empty
- [x] `git ls-files '*.go' | xargs wc -m | awk '$1>2000'` returns only totals
- [x] `git ls-files | grep -E 'apk$|mp4$'` - policy recorded in FOLLOW-01: six signed builder APKs plus both `teams.mp4` copies plus `APK-SHA256SUMS` stay committed, ignore scoped to unsigned and debug outputs
- [x] grep shows one blinkless pin and zero replay to render production imports
- [x] `-race` on `internal/state`, `internal/ipc`, `internal/replay`, `internal/page` - recorded 2026-10-16: state ok 1.0s, ipc ok 1.0s; replay and frame ran race green in workers, page ran race green in A1

## Phase F: what is pending for 10

Why not 10: the wave landed and the gates above are green, but docs disagree with the tree in two places, one checklist bar was proven unmeetable, two guides still describe pre-wave behavior, the anim guard is opt-in, and the APK policy needs recording. Each row names the subregions to scan.

- [x] `FOLLOW-01` APK policy record - subregions scanned: `examples/android-host/APPS.md`, root `.gitignore`, login and platform Android READMEs, `showcase.md`, `git ls-files`. Done 2026-10-16: APPS.md records committed builder APKs, ignore scoped to unsigned and debug outputs. Proof: six APKs plus both mp4 copies plus `APK-SHA256SUMS` in `git ls-files`.
- [x] `FOLLOW-02` TEST-01 bar rewrite - done 2026-10-16 in the TEST-01 row above: fewest cap-compliant files per family, 63 to 59. Proof: `go test ./internal/replay/` green in A7.
- [x] `FOLLOW-03` frames guide matches Bind - done 2026-10-16: `documentation/frames.md` gained a `Bind` paragraph covering `TranslateTo`, `FadeTo`, `ScaleTo`, `RotateTo` plus legacy unbound behavior. Zero production constructors exist outside `internal/frame`, tests only.
- [x] `FOLLOW-04` clipboard guides match error return - done 2026-10-16: `documentation/clipboard.md` and the `documentation/features.md` clipboard section describe `Read() (string, error)`, `ErrUnavailable`, UseMemory-only memory copy, no stale fallback.
- [x] `FOLLOW-05` anim guard opt-in closure - closed 2026-10-16 by survey: grep over `TranslateTo|FadeTo|ScaleTo|RotateTo` shows zero production call sites outside `internal/frame`, tests only, so no unbound sites exist and the frames.md paragraph in FOLLOW-03 is the full closure.
- [x] `FOLLOW-06` race plus audit record - recorded 2026-10-16: `go test -race -p 1 ./internal/state/ ./internal/ipc/` both ok; wc audit over root plus internal returns only totals; one blinkless pin `c085b3423260` across `go.mod`, `examples/go.mod`, `AGENTS.md`; import gate green in A4.

## Dependencies

- Blinkless engine for parse, cascade, layout, pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, text.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules proposed.

## Risks

- Screen role split touches every host fake, keep old interface as shim so examples compile untouched.
- Clipboard signature change may break callers that ignore ok flag, fix call sites in CORR-04 row.
- APK removal breaks links, move to release artifacts with redirects from docs, no history rewrite without approval.
