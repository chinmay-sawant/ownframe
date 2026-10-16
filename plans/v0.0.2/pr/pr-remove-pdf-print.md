# chore(review): land architecture wave and remove PDF print path

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md), saved before opening the PR. Local only: nothing is published from this file.

- Base: `origin/master` (`8bc0dde`)
- Head: `chore/review` (`65be197` plus this record)
- Commits: 9 on the branch plus this PR record
- Diff: 389 files, +5978 / -4466
- Review: 10 parallel subagents covered the branch diff slice by slice; findings are folded in below

---

## Summary

This branch lands the improve-codebase architecture wave, closes the Phase F road-to-10 items, and removes the PDF export and print path end to end. The engine still stops at the display list, so the `ErrNoPDF` stubs carried no behavior worth keeping: `Page.PDF`, `WritePDF`, `SavePDF`, `Print`, `SupportsPDF`, `PDFOptions`, `ErrNoPDF`, `ErrNoPrinter`, `internal/print`, `GET /pdf`, and `examples/print` are all gone, with docs, routes, and the packaging test moved to the `login` example.

## Motivation / context

- The PDF surface was stub-only. Blinkless returns a drawing list and has no document writer, so every PDF call refused and the only working fallback was already `Page.PNG`.
- The architecture wave (roles, guards, probes, splits) and the Phase F close-out were reviewed in-branch under `plans/v0.0.2/code-review/` and needed one PR body against `master`.
- Plans: `plans/v0.0.2/code-review/`, `plans/v0.0.2/security/`
- Issues: none. This repo has no tracker entries for this work.

---

## Changes

### PDF and print removal (tip commit `65be197`)

- Deletes `print.go` (public `PDFOptions`, `SavePDF`, `ErrNoPDF`, `ErrNoPrinter`), all of `internal/print/` (OS handoff plus `ValidPath`), all of `internal/page/print*.go` (11 files: `PDF`, `WritePDF`, `SavePDF`, `Print`, `SupportsPDF`, `pdfSource`, theme-inject helpers), `internal/host/pdf.go` (`PDFExporter`), `internal/web/pdf_http.go` plus its test, and `examples/print/` (6 files).
- `internal/web/serve_build.go` no longer registers `GET /pdf`. The remaining mux (`/`, `/frame.png`, click, type, backspace, debug, pprof) is intact.
- `internal/host/screen.go` comment drops PDF from the optional-interfaces list. The `Screen` method set is unchanged.
- Docs updated to match: `documentation/printing.md` deleted, printing sections removed from `features.md`, `GET /pdf` rows removed from `web.md`, capability matrix and prose fixed in `platforms.md`, `screen.md`, `frames.md`, `security.md`, `features-examples.md`, `rebrand.md`, `examples/readme.md`, plus `AGENTS.md` and `PHASES.md`.
- Follow-up fix in this record: `scripts/package_test.sh` and `documentation/packaging.md` still invoked the deleted `print` example, so `package_test.sh` failed. Both now use the `login` example. `sh scripts/package_test.sh` passes.

### Architecture wave (earlier branch commits)

- `host.Screen` now composes `Boxer` plus the `Pointer`, `Keyboard`, and `Clipboard` roles with the same method set. Web handlers narrow to one role at each call site (`roles.go`, `areas`, `click`, `input`, `reload`). `screen_shim_test.go` and `screen_roles_test.go` pin the pattern.
- New `ServeWithOptionsContext` and `ServeWithContext` with nil guards. Debug and pprof registration stay gated behind `ServeOptions`.
- New `internal/masking` package holds fill radii, blend, color, grid, and image-decode helpers. `replay` no longer imports `render`, fixing the one-way import direction.
- `internal/replay/grid.go` segments now inherit owner opacity, transform, and fixed placement through `masking.GridLine`, matching the bitmap painter. Faded or transformed tables render differently than before, toward parity.
- `page` internals regrouped into role-sized files to stay under the 2000-character file limit. No violations in new or changed `.go` files.
- `examples/print` is the only deleted demo. Four new files extend `spotify-player/player`. `Makefile` example discovery is glob-based, so nothing else changed.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No measured change. Grid replay does less per-segment allocation. Full `make test TEST_P=4` ran green. |
| **Memory** | No measured change. Text and image caches split their locks; one duplicate parse is possible with one winner kept, benign. |
| **Behavior / correctness** | Grid segments on faded or transformed tables now match the bitmap painter. `GET /pdf` is gone; `GET /frame.png` still serves the picture. |
| **API / CLI** | Breaking, see table below. `Serve` signature unchanged. `scripts/package.sh` example renamed `print` to `login`. |
| **Dependencies** | None added or removed. `codeberg.org/go-pdf/fpdf` stays as an indirect require through blinkless via `examples/departure-board/dump`; `go mod tidy` must not drop it. |
| **Binary size / build time** | Not measured in this PR. `make build` (`go vet -p 1 ./... ./examples/...`) passes. |

---

## Breaking changes / migration

All breaks are intentional removals from `65be197`. A 10-agent review found no remaining callers in `examples/` or root (only stdlib `log.Print` hits) and no surviving test references the removed symbols.

| Item | Migration |
|------|-----------|
| `ownframe.PDFOptions` | Delete the type. There is no PDF output. |
| `ownframe.SavePDF`, `(*Page).SavePDF` | Snapshot with `p.PNG()` and write the bytes yourself. |
| `(*Page).PDF`, `(*Page).WritePDF` | Use `p.PNG()` or `PNGWithContext` for a bitmap snapshot. |
| `(*Page).Print`, `ownframe.ErrNoPrinter` | Hand a PNG or HTML file to the OS print path or the browser. |
| `ownframe.ErrNoPDF`, `(*Page).SupportsPDF` | Delete `errors.Is` branches for both errors. |
| `GET /pdf` on `Serve` | Serve `GET /frame.png`, which still exists. |
| `examples/print` demo | Gone. No caller action. |

---

## Test plan

- [x] `make build` (`go vet -p 1 ./... ./examples/...`) passes.
- [x] `go test -count=1 -p 1 ./internal/page/... ./internal/web/... ./internal/host/...` passes.
- [x] `go test -count=1 -p 1 ./internal/clipboard/... ./internal/fetch/... ./internal/crash/... ./internal/ipc/...` passes.
- [x] `go vet -p 1 ./examples/...` and `./examples/spotify-player/...` pass.
- [x] `sh scripts/package_test.sh` passes (`package layout ok`).
- [x] Full `make test TEST_P=4` ran across both modules; results below.
- [ ] `make run` wall time vs baseline (no runtime behavior in scope beyond grid parity).
- [ ] Reference metrics (no detector surface change claimed).

### Commands

```sh
make build
go test -count=1 -p 1 ./internal/page/... ./internal/web/... ./internal/host/...
sh scripts/package_test.sh
make test TEST_P=4
bash scripts/pr-diff-stat.sh master
```

---

## Screenshots / sample output

```
package layout ok
```

```
ok  github.com/chinmay-sawant/ownframe/internal/page
ok  github.com/chinmay-sawant/ownframe/internal/web
ok  github.com/chinmay-sawant/ownframe/internal/host
```

---

## Related issues

- None. No tracker entries exist for this work.

---

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied (`enhancement`, plus `documentation` for the plans copy)
- [ ] Related issues filled with real ticket IDs (none exist)
- [x] Filled body committed under `plans/v0.0.2/pr/pr-remove-pdf-print.md`
- [ ] PR has assignee and labels (on `gh pr create`)

Suggested command once the author decides to open the PR. Do not run this until the author asks:

```sh
gh pr create \
  --base master \
  --head "$(git branch --show-current)" \
  --title "chore(review): land architecture wave and remove PDF print path" \
  --body-file plans/v0.0.2/pr/pr-remove-pdf-print.md \
  --assignee "@me" \
  --label enhancement \
  --label documentation
```

---

## Follow-ups (out of scope)

- No mux route smoke test: nothing hits `POST /click`, `/type`, `/backspace`, or `GET /frame.png`, so a dropped route would pass CI.
- `internal/masking/mask_geometry.go` sits at 1942 of 2000 characters. Pre-split it on the next touch.
- `frame.Anim.Bind` stale-guard is opt-in with zero non-test callers. Consider binding at construction.
- The `</style>`-breakout escaping test left with the deleted theme-inject helper. The live theme path goes through parsed CSS, so the gap is moot unless HTML injection returns.
- Confirm faded and transformed table parity is pinned by the grid parity tests after the `masking.GridLine` change.
- `frontend/dist/` bundle still mentions `printing.md` and `GET /pdf`. It is gitignored build output and regenerates.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 361 | 4863 | 4292 |
| `.html` | 4 | 475 | 34 |
| `.md` | 23 | 637 | 140 |
| No extension | 1 | 3 | 0 |
| **Total** | **389** | **5978** | **4466** |
