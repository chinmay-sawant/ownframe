# v0.0.3 - Phase 01: Repository Footprint and Binary Hygiene

> **Parent:** [README.md](README.md)
> **Status:** In progress (8/21 done: `FOOT-01`–`FOOT-04`, `FOOT-14`, `FOOT-15`, `FOOT-17`, `FOOT-GATE-03`)
> **Estimated effort:** 2–4 days (excluding optional git history rewrite coordination)

---

## Overview

At commit `da4f12cc`, the repository carries **3.4 GB in `.git`**, **668 MB tracked in `docs/`** (89 tracked files, dominated by nine 52–66 MB `.wasm`/`.wasm.gz` binaries totaling ~511 MB), and **207 MB of seven tracked Android release APKs** under `examples/*/android/artifacts/`. In addition, local development leaves ~7.0 GB in `temp/` and ~360 MB of compiled `*.test` binaries in the repo root and `examples/`.

Every site rebuild or Android artifact refresh commits hundreds of megabytes of incompressible binaries into `master`, making `git clone`, CI checkouts, and external contributions slow and fragile. This phase stops the ongoing binary growth first, cleans workspace clutter, adds size guardrails, and evaluates a controlled `.git` history shrink.

## Executive Summary

- Remove stray untracked binaries/leftovers from the workspace and tighten `.gitignore`.
- Stop committing generated `.wasm` / `.wasm.gz` demo binaries to `master` by deploying GitHub Pages from a build workflow artifact (or orphan `gh-pages` branch) instead of `docs/` on `master`.
- Move the 7 tracked Android release APKs out of git history onto GitHub Releases / CI artifacts while keeping reproducible build scripts and SHA-256 manifests in git.
- Add an automated tracked-file-size guardrail in CI/lint so >5 MB binaries cannot silently re-enter `master`.
- Audit `.git` packfile bloat and produce a gated decision record before any `git filter-repo` history rewrite.

## Current Source Baseline (`da4f12cc`)

| Path / Metric | Measured Size / Count | Tracked in Git? | Primary Driver |
|---|---|---|---|
| `.git` | **3.4 GB** | N/A (repo history) | Repeated commits of `.wasm`, `.apk`, and `.webp` media |
| `docs/` | **668 MB** (89 files) | Yes | 9 `.wasm`/`.wasm.gz` files (52–66 MB each = ~511 MB) + large `.webp` captures |
| `examples/*/android/artifacts/*-release.apk` | **207 MB** (7 files) | Yes | `lg-remote` (31M), `login` (30M), `platform` (30M), `digital-wellbeing` (30M), `depth-alert` (29M), `eyes` (29M), `dino` (28M) |
| `examples/` (working tree) | **2.5 GB** | Partial | Tracked APKs + MP3 audio + untracked `android/build` & `*.test` binaries |
| `temp/` | **7.0 GB** | No (ignored) | Local scratch captures and test artifacts |
| Root & `examples/` `*.test` + leftovers | **~360 MB** | No (ignored/untracked) | 14 compiled `*.test` binaries (`loglens.test`, `code-view.test`, `page.test`, `window.test`, etc.), `mem.out`, 0-byte `Händen`, `.doc-snippets-h0hluxw9/` |

---

## Phase 1: Repository Footprint and Binary Hygiene

### 1.1 Local workspace cleanup and ignore hygiene

- [x] `FOOT-01` Remove untracked local leftovers from the checkout: `*.test` binaries in root (`code-view.test`, `dirty.test`, `inventory.test`, `loglens.test`, `motion.test`, `ownframe.test`, `page.test`, `perf-benchmarks.test`, `replay.test`, `window.test`) and `examples/` (`code-view.test`, `loglens.test`, `ownframe.test`, `trade.test`), `mem.out`, 0-byte `Händen`, and stale `.doc-snippets-*` temp directories. Evidence: removed all 14 `*.test` binaries, `mem.out`, `Händen`, and `.doc-snippets-h0hluxw9`; `ls -la *.test examples/*.test mem.out Händen .doc-snippets-*` returned `CLEAN`.
- [x] `FOOT-02` Audit `temp/` (7.0 GB) and local Android Gradle `examples/*/android/build/` directories; delete stale scratch outputs and document a `make clean-local` target in `Makefile` for repeatable workspace cleanup. Evidence: audited `temp/` (`temp/frontend-check` 1.3G, `temp/v004-full-bench` 1.1G, `temp/loglens-baseline` 894M, `temp/design-refresh` 894M, `temp/android-runtime` 632M) and `examples/pet-care/android/build` (152K); added `clean-local` target in `Makefile` and ran `make clean-local` (exit 0, preserving `temp/` with size reminder).
- [x] `FOOT-03` Update `.gitignore` to ensure root/nested `*.prof`, `cpu.out`, `mem.out`, `.pytest_cache/`, and temporary doc-snippet directories are consistently ignored without hiding tracked source files. Evidence: updated `.gitignore`; `git status` clean after `make docs-check` and `make tools-check`.
- [x] `FOOT-04` Verify `scripts/check-doc-snippets.py` cleans up its `.doc-snippets-*` temporary directory in a `try/finally` block even on failure or interruption, preventing orphan directories in the repo root. Evidence: verified `with tempfile.TemporaryDirectory(prefix='.doc-snippets-', dir=ROOT)` in `scripts/check-doc-snippets.py` and added `test_snippet_scratch_directory_cleans_up_on_failure` in `scripts/test_doc_claims.py` (`make tools-check` exit 0).

### 1.2 Decouple built `.wasm` demos from `master` git history

- [ ] `FOOT-05` Audit `.github/workflows/` and GitHub Pages settings to determine how `docs/` is currently published (`docs/` on `master` vs `actions/deploy-pages`). Record exact workflow files and triggers. Proof: written audit note referencing exact workflow filenames.
- [ ] `FOOT-06` Add a GitHub Actions Pages workflow (`.github/workflows/pages.yml`) that builds `frontend/` and the starter `.wasm` demos via `frontend/scripts/build-demo.mjs` and publishes them via `actions/upload-pages-artifact` + `actions/deploy-pages` (using CI caching for Go/npm builds). Proof: successful workflow run publishing the full site + live demos.
- [ ] `FOOT-07` Update `scripts/check-generated-site.py`, `Makefile` (`site-check`), and `.gitignore` (removing the `!docs/demos/*.wasm` force-track exceptions) so `docs/demos/*.wasm` and `docs/demos/*.wasm.gz` are no longer committed to `master`. Note: preserve or automate `wispr-flow/ownframe.wasm` via a reproducible build script before removing it from git. Proof: `git ls-files docs/ | grep -E '\.wasm(\.gz)?$'` returns empty and `make site-check` passes against freshly built output.
- [ ] `FOOT-08` Review tracked media assets in `docs/` (`teams.webp` 11 MB, `trade.webp` 7.7 MB, `desktop-cat.webp` 6.6 MB, `wispr-flow.webp` 6.2 MB, `eyes2-60fps-*.webp` 5.7 MB) and compress/re-encode or host heavy demo recordings outside git where lossless quality is unnecessary. Proof: `du -sh docs/` tracked footprint drops from 668 MB to <50 MB (or 0 MB if `docs/` is fully CI-built).
- [ ] `FOOT-09` Update `README.md`, `frontend/README.md`, and `AGENTS.md` to document the CI-built site flow and remove instructions assuming committed `docs/demos/*.wasm` binaries. Proof: `make docs-check` exits 0.

### 1.3 Decouple Android release APKs from git history

- [ ] `FOOT-10` Inventory all tracked binary artifacts under `examples/*/android/artifacts/` (`lg-remote`, `login`, `platform`, `digital-wellbeing`, `depth-alert`, `eyes`, `dino` — 7 files, 207 MB total) and record every doc, script, or CI job referencing them. Proof: `grep -rn "android/artifacts" .` audited and listed.
- [ ] `FOOT-11` Publish reference Android APKs as GitHub Release assets (or workflow build artifacts produced by `scripts/android-release.sh`) and keep only a small text `checksums.txt` / manifest in git if release verification needs pinned hashes. Proof: download URLs and SHA-256 checksums documented in `documentation/android-build.md` / `examples/readme.md`.
- [ ] `FOOT-12` Remove tracked `*.apk` files from `master` with `git rm` and replace the selective `*-unsigned.apk` / `*-debug.apk` ignore lines in `.gitignore` with `*.apk`. Proof: `git ls-files '*.apk'` returns empty.
- [ ] `FOOT-13` Update `scripts/check-release-bundle.py`, `scripts/check-examples.py`, and Android scripts so none assume pre-committed APKs in the checkout. Proof: `make examples-check` and `make tools-check` exit 0.

### 1.4 Git history shrink assessment and size guardrails

- [x] `FOOT-14` Run a read-only object-size audit on `.git` (`git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | awk '/^blob/ {sum[$4]+=$3; cnt[$4]++} ...'`) to attribute the 3.4 GB `.git` packfile by path and extension. Evidence: top cumulative uncompressed blob history drivers at `da4f12cc`:
  - `docs/demos/spotify-player/ownframe.wasm.gz`: **1,502.6 MB** (23x)
  - `docs/demo/demo.wasm`: **1,224.5 MB** (24x)
  - `docs/demos/telegram/ownframe.wasm`: **1,177.5 MB** (23x)
  - `docs/demos/workspace-inspector/demo.wasm`: **1,173.7 MB** (23x)
  - `docs/demos/expense-tracker/demo.wasm`: **1,123.8 MB** (22x)
  - `docs/demos/hello/demo.wasm`: **1,123.1 MB** (22x)
  - `docs/demos/builder/demo.wasm`: **318.1 MB** (6x) & `docs/demos/ui-builder/demo.wasm`: **181.5 MB** (3x)
  - Android APKs (`telegram`, `lg-remote`, `platform`, `login`, `dino`, `eyes`, `digital-wellbeing`, `depth-alert`): **~718 MB** cumulative across 24 commits.
- [x] `FOOT-15` Add a tracked-file size check script (`scripts/check-repo-size.py`, wired into `make lint` / CI) that fails when any newly tracked file exceeds **5 MB** (with an explicit, shrinking allowlist for existing unavoidable assets until resolved). Evidence: added `scripts/check-repo-size.py`, `scripts/repo-size-allowlist.txt` (28 frozen >5 MiB entries), `scripts/test_repo_size.py` (3 unit tests covering allowlist, new >5 MiB rejection, and stale entry rejection), `Makefile` `repo-size-check`, and `scripts/run-lint.py` stage; `make repo-size-check` and `make tools-check` exit 0.
- [ ] `FOOT-16` Decide whether to rewrite `.git` history with `git filter-repo` (stripping historical `.wasm` and `.apk` blobs to shrink `.git` from 3.4 GB toward <200 MB). Document the exact cutover procedure: tag backup mirror, invalidate old PR branches, coordinate force-push, and verify `go.mod` / published tags (`v0.0.1`, `v0.0.2`) remain immutable or clearly note that published Go proxy checksums (`sum.golang.org`) pin existing tags and therefore history rewrites must NOT mutate already-published tag commits. Proof: written decision note (either `[x]` with verified mirror test or `[~]` deferred due to published tag immutability, with shallow-clone guidance in `CONTRIBUTING.md`).
- [x] `FOOT-17` Document `git clone --depth=1` / `--filter=blob:none` instructions in `CONTRIBUTING.md` if full history remains large to preserve published tag commits. Evidence: added partial/shallow clone commands under `## Setup` in `CONTRIBUTING.md`.

---

## Dependencies

- `FOOT-06` (CI Pages deploy) must be live and verified before `FOOT-07` removes `docs/demos/*.wasm` from `master` so the live website never breaks.
- `FOOT-11` (Release asset upload) must complete before `FOOT-12` removes APKs from `master`.
- `FOOT-16` must respect Go module proxy (`sum.golang.org`) immutability for tagged releases `v0.0.1` and `v0.0.2`: never force-rewrite commits reachable from published Go module tags without verifying impact on downstream `go.sum` entries.

## Risks

- **Go proxy checksum breakage:** Rewriting git history across tagged `v0.0.1`/`v0.0.2` commits breaks `GOPROXY=direct` and changes commit SHAs referenced in release notes and ledgers. Prefer stopping future bloat (`FOOT-05`–`FOOT-15`) and using partial clones (`--filter=blob:none`) unless a clean mirror/new-history strategy is explicitly approved.
- **Website build time in CI:** Building 9 wasm demos in `pages.yml` could slow site deploys; mitigate with Go build cache and only rebuilding changed demos or running `pages.yml` on `frontend/`/`examples/`/manual triggers.

## Closure Gates

- [ ] `FOOT-GATE-01` `git ls-files '*.wasm' '*.wasm.gz' '*.apk'` returns empty on `master`.
- [ ] `FOOT-GATE-02` Tracked `docs/` size <50 MB (or CI-deployed only) and tracked `examples/` excludes all APK binaries.
- [x] `FOOT-GATE-03` `scripts/check-repo-size.py` wired into `make lint` and passing. Evidence: `scripts/run-lint.py` stage added; `make repo-size-check` printed `28 allowlisted >5 MiB files; 0 new violations` (exit 0).
- [ ] `FOOT-GATE-04` `make lint`, `make site-check`, `make examples-check`, `make tools-check`, and `make test` exit 0.
