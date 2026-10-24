# PR: pending website and example updates

Base: `master`. Head: `chore/pending-items`.
Commits on this branch range from `d2b25242` through `e48b91af`.
Scale: 46 files changed vs `origin/master`, 4226 insertions and 339 deletions. Go is 13 of those files.

## Summary

This branch collects pending website and example work. The website docs now focus on examples and comparisons, with one framework comparison page split into per framework pages. It adds hello and grid starter examples, a shared comparison display helper, and a CSS animations knowledge lab. It also aligns the blinkless pin between root and examples modules. The built `docs/` assets are rebuilt twice along the way.

## Changes by area

Website docs. `documentation/compare-go-app-frameworks.md` is gone, replaced by `compare-ebitengine.md`, `compare-fyne.md`, `compare-gio.md`, `compare-wails.md`, and `compare-webview.md`, with `compare-rust-gpui.md` updated. New `documentation/compatibility-matrix.md` adds 1204 lines covering animation properties. `documentation/snippets.json` gains entries. `frontend/src/content.js`, `Documentation.jsx`, `Home.jsx`, and `main.jsx` switch to the examples and comparisons shape, and `frontend/src/docs-overview.js` is deleted. `frontend/README.md` and `frontend/scripts/check-site.mjs` follow.

Comparison display. New `frontend/src/compareDisplay.js` holds the shared header and override logic, with `frontend/src/compareDisplay.test.js` covering it. `content.js` and `main.jsx` use it. The syntax entry it dropped is the display override path the helper now owns.

Starter examples. New `examples/hello` runs a minimal page with desktop and wasm mains. New `examples/grid` runs a small grid page with its own fragment files. `examples/readme.md` and `examples/maintenance.json` list them.

Engine pin. Root `go.mod`, `go.sum`, and `go.work.sum` move the blinkless pin, and `examples/go.mod` plus `examples/go.sum` match it. No replace entries change.

Animations lab. New `examples/animations` runs a quiz over timing, direction, delays, fill modes, WebKit prefixes, and composite properties, with live chips bound through `internal/animation`, `internal/cssanimation`, and `internal/cssmotion`. It registers in `examples/maintenance.json` as experimental and adds three rows to `scripts/example-internal-allowlist.txt`.

Built site. `docs/index.html` and the two hashed `docs/assets` bundles are rebuilt in `d2b25242` and again in `52d563ee`. Final net is about 497 lines in the Documentation bundle and 12 in the index bundle.

## Breaking changes and migration

- No Go API changes. Library callers keep the same `ownframe.Page` set.
- `documentation/compare-go-app-frameworks.md` is deleted. Link to the per framework compare pages instead.
- `frontend/src/docs-overview.js` is deleted. The overview now lives in `content.js`.
- No migration for the engine pin. Both modules resolve the same blinkless commit.

## Test plan

- `go test -p 1 ./examples/animations/...` passes.
- `go vet -p 1 ./... ./examples/...` passes through `make build`.
- Full `make test` passes across both modules.
- `make examples-check` passes with 6 maintained and 52 experimental roots.
- `make internal-import-check` passes with 40 allowlisted pairs.
- `npm test --prefix frontend` passes, 10 tests.
- `npm run check:site --prefix frontend` passes, 120 checks.

## Risks and follow ups

- The built `docs/assets` bundles churn with each website commit, so review reads best with those paths filtered.
- Deleted doc and overview paths may leave stale inbound links in plans or external pages.
- `compatibility-matrix.md` is 1204 lines and may rot as blinkless behavior moves. Nothing checks it against code.
- The animations example adds new internal import rows while the allowlist header says not to add rows for new work. A public export should replace that binding.
- Hello and grid examples duplicate the small page wiring. Later work could fold them into one starter.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
|-----------|-------|------------|-----------|
| `.go` | 13 | 1266 | 0 |
| `.html` | 4 | 808 | 1 |
| `.js` | 6 | 476 | 202 |
| `.json` | 2 | 41 | 6 |
| `.jsx` | 3 | 16 | 25 |
| `.md` | 11 | 1593 | 79 |
| `.mjs` | 1 | 14 | 14 |
| `.mod` | 2 | 2 | 2 |
| `.sum` | 3 | 7 | 10 |
| `.txt` | 1 | 3 | 0 |
| Total files 46, insertions 4226, deletions 339 |

Related: branch `chore/pending-items`, head commit `e48b91af`.
