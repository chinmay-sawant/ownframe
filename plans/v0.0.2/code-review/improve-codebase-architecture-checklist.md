# Improve codebase architecture - phase-wise checklist

> Parent: `plans/v0.0.2/code-review/improve-codebase-architecture.md` and `improve-codebase-architecture.html` beside this file
> Frozen at: `69fcd32` on `chore/review`, 2026-10-16
> Status: open, 15 rows plus gates
> Scope: root `*.go` plus `internal/...` plus `examples/...` weight plus `documentation/...`

## Overview

This checklist turns the architecture plan into checkable rows. Each row names one absorbing module, one proof, and one owner file set so 10 workers can run in parallel without touching the same files. Rule: touch only assigned files, keep every Go file at or under 2000 chars counted with `wc -m`, run `gofmt` on edits, and prove with the named targeted test plus `go vet` on the touched package. Full `make build` and `make test` run once at integration, not inside every worker.

Agent to row map: A1 takes ARCH-01 and ARCH-02, A2 takes ARCH-03, A3 takes ARCH-04, A4 takes ARCH-05, A5 takes CORR-01, A6 takes CORR-02, A7 takes CORR-03 plus TEST-01 gate coordination, A8 takes CORR-04, A9 takes CORR-05 and CORR-06, A10 takes REPO-01, REPO-02, REPO-03. REPO rows are grouped because they are docs plus example splits plus git rm, all outside `internal/...` hot paths. If REPO-02 history rewrite is refused, do plain `git rm` plus release links, never rewrite pushed history without explicit approval.

## Phase A: shrink the two large types

- [ ] `ARCH-01` `internal/page/page_types.go:19` - form and focus state leaves Page; add one owned struct for form plus `hover`, `active`, `pressUsed`, accessibility focus with narrow set, clear, hit-routing methods; Page delegates; no behavior change; proof: grep shows zero direct writes outside owner, `go test ./internal/page/` green, `wc -m internal/page/page_types.go` under 2000. Owner files: new `internal/page/page_form_focus*.go`, edits to `page_types.go` plus call-site delegation only. Agent A1.
- [ ] `ARCH-02` `internal/page/page_types.go` embeds - geometry, content, frame ownership written down; add one short comment per embed (`pageGeom`, `pageContent`, `frameState`, `windowingState`) naming single writer and lock, split only if two writers found; proof: each embed has owner comment, grep confirms one writer per group. Owner files: `internal/page/page_types.go` comment block plus any embed def file owned by A1 only. Agent A1.
- [ ] `ARCH-03` `internal/host/screen.go:22` - Screen becomes roles plus compat shim; keep `Screen`, add `Pointer`, `Keyboard`, `Clipboard` role interfaces; move window and web call sites to narrowest role; fakes implement roles; proof: window plus web plus host suites green, old Screen callers compile. Owner files: `internal/host/*.go`, window and web call-site edits, host fakes. Agent A2.
- [ ] `ARCH-04` `internal/window`, 189 non-test files - subsystem map; add one index doc mapping devtools, idle, touch, pointer, replay tile, reload groups to entry file plus state owner; no code change; proof: idle owner and tile owner found from index alone. Owner files: one new doc under `internal/window/` or `documentation/`, no Go edits. Agent A3.
- [ ] `ARCH-05` `internal/replay/`, `internal/render/paint.go` - replay to render rule gets a gate; add test listing production imports of replay, fail on render outside allow list; proof: gate green, trial render import fails gate. Owner files: one new `internal/replay/*_gate_test.go` only. Agent A4.

## Phase B: enforce instead of document

- [ ] `CORR-01` `internal/frame/anim.go:42` - anims carry a generation; record display generation at creation, Step goes inactive after Redraw, or copy target values so no old-list pointer survives; proof: test creates anim, forces Redraw, steps, asserts no patch on freed ops. Owner files: `internal/frame/anim*.go` plus one new test. Agent A5.
- [ ] `CORR-02` `internal/page/print.go:27` - print gets a probe; add `SupportsPDF` or `Available` probe next to PDF, WritePDF, SavePDF, Print, keep ErrNoPDF on direct calls; update printing docs branch pattern; proof: unit asserts probe false on every sink. Owner files: `internal/page/print*.go` plus docs `documentation/printing.md`. Agent A6.
- [ ] `CORR-03` `internal/replay/chunk_cache.go:11` - chunk bake policy measured in comment; keep bake under lock, cite bench name plus race reason so narrowing is not retried blindly; proof: comment cites bench plus invariant, replay suite green under `-race`. Owner files: `internal/replay/chunk_cache.go` comment only. Agent A7.

## Phase C: stop stale reads

- [ ] `CORR-04` `internal/clipboard/clipboard.go:34` - clipboard read stops stale text; return error or ok false on OS failure, in-memory copy for UseMemory tests only; fix call sites in same row; proof: failing OS read returns error and no text, memory mode green. Owner files: `internal/clipboard/*.go` plus callers touched by A8 only. Agent A8.
- [ ] `CORR-05` `internal/crash/validate.go:67` - crash dir absolute at set time; reject relative or resolve to absolute once and store absolute; proof: relative input rejected or stored absolute, test asserts absolute. Owner files: `internal/crash/*.go` owned by A9 only. Agent A9.
- [ ] `CORR-06` `internal/print/command_args.go:30` - print path canonicalized; clean plus absolutize, reject traversal after cleaning, allow benign relative; proof: table over benign relative, traversal, absolute passes. Owner files: `internal/print/*.go` owned by A9 only. Agent A9.

## Phase D: repo weight and example discipline

- [ ] `REPO-01` `examples/spotify-player/player/playback.go:2357`, `logger.go:2238`, `actions.go:2097`, `display_patch_helper_test.go:2028` - four files meet the cap; split each to at or under 2000 chars, no behavior change; proof: `git ls-files 'examples/**/*.go' | xargs wc -m | awk '$1>2000'` returns only total. Owner files: `examples/spotify-player/player/` only. Agent A10.
- [ ] `REPO-02` six `examples/*/android/artifacts/*-release.apk` plus two `teams.mp4` - binaries leave git; plain `git rm` plus release artifact links, keep one mp4, no history rewrite without approval; proof: `git ls-files | grep -E 'apk$|teams.mp4$'` shows zero APKs and one mp4. Owner files: git index plus docs links only. Agent A10.
- [ ] `REPO-03` `AGENTS.md:43`, `go.mod:6`, `examples/go.mod:6`, `README.md:31` - pin plus stale notes synced; AGENTS.md moves to `c085b3423260`, README paths fixed; proof: grep shows one pin across all three files. Owner files: `AGENTS.md`, `README.md` only. Agent A10.
- [ ] `TEST-01` `internal/replay/`, 63 test files - replay tests merge by family; merge chunk parity, chunk bench, gpu parity families into one file each with subtests, no coverage loss; proof: file count drops, `go test ./internal/replay/` green, case names kept. Owner files: `internal/replay/*_test.go` merges only, coordinated with A4 gate file name and A7 comment to avoid same-file edits. Agent A7.

## Phase E: gates that keep 10

Run once at integration after all workers land, record exits in this file:

- [ ] `make build` exit 0, which runs `go vet -p 1 ./... ./examples/...`
- [ ] `make test` exit 0, 97 packages ok, zero FAIL, plus `-race` on `internal/state`, `internal/ipc`, `internal/replay`, `internal/page`
- [ ] `gofmt -l` over changed trees returns empty
- [ ] `git ls-files '*.go' | xargs wc -m | awk '$1>2000'` returns only totals
- [ ] `git ls-files | grep -E 'apk$|mp4$'` shows no release binaries except the one kept mp4
- [ ] grep shows one blinkless pin and zero replay to render production imports

## Dependencies

- Blinkless engine for parse, cascade, layout, pinned in `go.mod`.
- Ebiten v2.10.4 for window, canvas, text.
- `golang.org/x/image` and `golang.org/x/text` for raster and shaping.
- No new modules proposed.

## Risks

- Screen role split touches every host fake, keep old interface as shim so examples compile untouched.
- Clipboard signature change may break callers that ignore ok flag, fix call sites in CORR-04 row.
- APK removal breaks links, move to release artifacts with redirects from docs, no history rewrite without approval.
