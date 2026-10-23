# Local review fixes

The reviewed snapshot is `78306469`. Implementation starts from `24d4267a`
and preserves its later runtime fixes. Implementation and validation were
completed locally before the user authorized a commit and push.

## Workflow

1. Read the repository guidance and working-tree state. The starting tree is clean.
2. Fix measurement names, frame work accounting, and dirty-region counters.
3. Fix benchmark, sampler, and soak tooling with falsifiable regression checks.
4. Fix archive/signing tooling and add the missing CI validation paths.
5. Fix accessibility state publication and focused test gaps.
6. Correct documentation and enforce the corrected claims with scripts.
7. Run targeted checks, both-module tests, lint, graphics, and distribution checks.

Performance tooling, distribution tooling, and accessibility use isolated local
worktrees. Their uncommitted patches were reviewed and applied here. During implementation,
commits, pushes, tags, releases, repository settings, and device installs were
not authorized. The temporary worktrees were removed after integration.

## Limits

Native Windows/macOS interaction, Android device validation, real signing,
GitHub protection settings, and a completed hosted release cannot be proven by
local Linux changes. No timing supplied by the earlier review is treated as
reproduced until a corresponding run is recorded. API counts and unverified
unreachable-code claims do not justify deleting public APIs or packages.

Validation results and unresolved findings will be recorded after integration.


## Finding disposition

| Findings | Local change | Remaining limit |
|---|---|---|
| A1 | StyleTime, LayoutTime and OtherTime match measured work; all example callers migrated | Layout includes list emission; cache-hit styling is not guaranteed constant |
| A2 | Cold fixture changes executed HTML and requires one parse per iteration | Timing ordering depends on workload and machine |
| A3/A4 | Repeated micro/app benchmarks, raw JSON provenance and compatible median comparison | Sample p95 from five runs is not a tail-confidence estimate |
| A5 | Same-runner parent/candidate CI comparison, release reuse, dated local report | Hosted execution is unverified; 1.5 ratio is a coarse regression guard |
| A6 | Frame ring records Update/Draw work; 60 Hz long-work threshold | Presentation gaps remain separate; no native GPU completion timing or web ring export |
| A7 | Distinct diff-region count/area and deduplicated intersecting-op count | Host padding/coalescing/tile work can differ |
| A8 | Three soak scenarios and separate configurable redraw/PNG budgets | Software workload, not GPU replay or native input |
| A9/A10 | Linux RSS, warmup slope limit, failure metadata, explicit unknown provenance | SIGKILL leaves running metadata; no native Windows/macOS RSS proof |
| A11/A12 | Explicit debug opt-ins and stale SVG timing removed | Historical reports retain their original measurement scope |
| B1 | Normalized archives and archive-level reproducibility | Real repeat build tested on Linux/wasm only |
| B2 | Extract/sign/notarize/staple/repack/rechecksum with invocation fixtures | Real certificates/platform tools not exercised |
| B3/B4 | Release reuses all CI gates and collects native desktop archives | Hosted native runs and production publication remain open |
| B5 | Correct protection claims and retain operator settings procedure | No repository settings changed |
| C1 | Native cgo compilation/package jobs; clipboard runtime subset uses memory only | Native click/type/resize/clipboard sessions still needed |
| C2 | Android bind job and trimpath/provenance/certificate metadata in release helper | Existing APKs were not rebuilt or installed; device and key required |
| C3 | IME test build tags and wasm vet | Browser interaction is not established |
| D1 | Independent committed fill golden; whole damage assertions; half-pixel mutation fails | Other GPU parity tests retain a shared implementation oracle |
| D2 | History branching regression rejects append mutation | No history implementation change needed |
| D3/D4 | Example race shards, debug-tag route test and bounded soak CI | Hosted CI remains unverified |
| E1 | Probe mutation contract corrected; DevTools getter available through Window | No speculative API deletion or count target enforced |
| E2 | Six renderer-owned caches, independent eviction/close, Run cancellation, public ownership contract | Package compatibility helpers retain a default renderer |
| E3 | Removed reproducibly redundant toolchain directives; standalone readonly vet | No historical retraction or tag changes; module archive checker scope unchanged |
| E4 | No deletion based on unverified importer counts | Unverified finding remains open |
| F1 | Correct active docs/API claims and scan example readmes | Historical plans remain historical records |
| F2 | Static anonymous-control semantics, content publication, password rejection and semantic AT-SPI paths | Anonymous live editing and stable identity across anonymous DOM moves still need form identity work |
| G | Closed FakeBus rejection, safe accessibility logging example, Node minimum docs and stronger evidence syntax | Plan consolidation is not justified by file count; evidence syntax is not proof of execution |

The wider public API redesign, anonymous live form identity, and unreachable-code
inventory remain separate work. They were not established as safe edits by the
source review. External platform evidence and protection settings cannot be
closed by these local changes.


## Completed local validation

All commands below exited zero on the combined local changes:

- `make test`, both library and examples modules.
- `make build`, both-module compile/vet check.
- `make test-race`, library module.
- `bash scripts/test-examples-race.sh 0 1`, every examples package with race detection.
- `make lint`, pinned linter, maintenance tooling, docs, plans, imports and website lint.
- `make graphics-check`, all five required software OpenGL/Xvfb sessions.
- `make docs-check plans-check examples-check`.
- `go test -p 1 -tags ownframe_debug ./internal/web`.
- `GOOS=js GOARCH=wasm go vet -p 1 ./internal/... .`.
- `GOWORK=off go vet -mod=readonly -p 1 ./internal/page/draw`.
- `go mod tidy -diff` at root, and `GOWORK=off go mod tidy -diff` in examples.
- `make module-check` and `make release-check`.
- `bash scripts/reproducible-check.sh`, complete Linux and wasm archive sums matched.
- `make site-check`, generated output matched and 120 browser layout checks passed.
- `make soak-check SOAK_SECONDS=60`, all three scenarios passed their separate budgets.
- Five-sample 100 ms benchmark report, including one parse/op for cold redraw.
- `go test -p 1 -run '^TestPerfProbeStages$' -v .`, row scaling and cache/stage accounting.
- `git diff --check`.

[Stage probe](stages-2026-10-23.txt), [soak log](soak-2026-10-23.txt),
[benchmark samples](performance-2026-10-23.txt), and
[benchmark provenance](performance-2026-10-23.json) retain the measured evidence.
Detailed gate output is retained under `temp/review-validation/` locally.

The independently specified golden rejected a half-pixel fixture displacement
with `golden differs at 12,12`; the original fixture was restored before the
passing graphics run. The history branching test also rejected a temporary
append-only implementation, which was restored. Sampler, signing and baseline
rejection fixtures run in the maintenance tooling gate. Real signing and
hosted platform runs remain unverified.

Validation ran from `24d4267a` with the review changes uncommitted. No tags,
releases, device installs, or remote repository settings were changed. The user
subsequently authorized a detailed commit and push to `feat/prod-readiness`.
