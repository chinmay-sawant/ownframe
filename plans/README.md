# Current maintenance ledger

This is the current status record. Historical release plans, `PHASES.md`, and `RENDERING-FINDINGS.md` preserve rationale, not present support. Public behavior belongs in [features](../documentation/features.md) and [platforms](../documentation/platforms.md). Earlier pass-by-pass results remain in Git history through `b32ea940`.

Use one state per row: implemented `[x]` with evidence, open `[ ]` awaiting work or validation, deferred `[-]` with a reconsideration condition, or superseded `[s]` with a replacement. A passing compile is not device evidence. Keep evidence beside its claim instead of appending duplicate closure sections.

## Implemented responsibilities

- [x] `COHESION` retire the character cap and consolidate related code and tests. Evidence: `AGENTS.md`, `CONTRIBUTING.md`, and the page/window/replay consolidations in `c59f1a51`; responsibility review replaces size bookkeeping.
- [x] `STATE-OWNER` put subscription cancellation, binding generations, and metadata synchronization in `internal/state.Binding`. Evidence: binding and page integration race tests; `b32ea940`. Page paint invalidates encoded PNGs, and each page owns its changed-property lock.
- [x] `PACING-OWNER` keep scheduling policy in `internal/window/pacing` and Ebiten modes, timers, and synchronization in its host adapter. Evidence: policy tests and `TestWindowIdleDeadline`, which checks the actual idle mode, timer wake, and resulting pixel.
- [x] `GPU-FIXTURES` move replay fixture ownership into the library and run alpha pixel assertions inside the GPU game loop. Evidence: `internal/replay/testdata`, `gpu_alpha_pixels_test.go`, and the graphics gate; no library test reads downstream example fixtures.
- [x] `GRAPHICS` require five isolated graphics sessions, including the scrollbar and actual window deadline. Evidence: `scripts/check-graphics.sh` rejects absent or skipped tests; `TestWindowRepaintPixels` also checks a surface cleared between clean frames.
- [x] `DOCS` compile public guide examples and check operational claims. Evidence: `make docs-check` compiles 31 snippets, explicitly reports eight external illustrations, and rejects stale Go/engine pins, retired tick semantics, or invalid performance flags.
- [x] `LINT` execute the exact declared linter with a curated policy. Evidence: version-selection regression, both Go module lint passes, and `.golangci.yml`; a mismatched PATH binary is preserved and replaced only for this invocation.
- [x] `EXAMPLES` keep six teaching examples and classify experiments explicitly. Evidence: `examples/maintenance.json`, inventory rejection tests, and compatibility tests for every example package.
- [x] `MODULE-CONSUMER` validate archive contents and an isolated public consumer. Evidence: `scripts/cmd/module-check`; nested modules are excluded, library fixtures remain, and the consumer resolves the archive without a workspace or replacement. An examples-tag check rejects filesystem replacements and vets the declared published dependency.
- [x] `WEBSITE-GATE` check source tests, generated output, and browser layouts in CI. Evidence: `make site-check`, generated-output rejection tests, and `.github/workflows/ci.yml`.
- [x] `RELEASE-CANDIDATE` verify local archive inventory, checksums, paths, modes, and wasm bytes; prepare signed and attested candidates after tag gates. Evidence: `make release-check`, archive corruption tests, and `release.yml`. The workflow definition is implemented; a successful OIDC tag run and publication are separate open checks below.
- [x] `PERF-WORKLOAD` drive a repeatable page soak and reject missing web metrics. Evidence: `page_soak_test.go`, strict sampler regressions, and the 600-second result in the validation record. Web perf now returns actual runtime metrics without a callback that production never installed.
- [x] `MOBILE-METRICS` use a collapsible bottom panel with a separate 48 dp X target on mobile. Evidence: geometry/touch/close regressions and cleared-surface GPU regression; desktop retains its right dock. Device results are recorded separately below.
- [s] `OLD-CLOSURE` retire blanket production-readiness completion. Replacement: this ledger; archived rows under `v0.0.3/prod-ready-task` are explicitly superseded.

## Validation and remaining gaps

- [x] `FINAL-GATES` run full tests, library race detection, lint, graphics, module, local archive, and website checks on the isolated source snapshot. Evidence: [validation record](validation/2026-10-23-session.md); all passed after incorporating `69f083de`. Concurrent unfinished shared-checkout edits are excluded.
- [x] `ANDROID-SESSION` complete guarded Pixel 7 keyboard, rotation, and same-process resume captures and visual review. Evidence: [session record](validation/2026-10-23-session.md), APK digest, reviewed captures, and the keyboard-gap regression. Complex IME composition remains outside this session.
- [ ] `NATIVE-IME` validate native Windows preedit, cancellation, conversion, and commit. The native window was confirmed and Japanese Microsoft IME was added, but English typing and layout switching are insufficient.
- [-] `MACOS-IME` defer native macOS validation at the author's request because no host is available. Reconsider when a macOS host can run the composing-input-method session.
- [ ] `INDEPENDENT-HANDOFF` have a teammate follow the clean-checkout exercise without author assistance. The reusable [review prompt](../documentation/maintenance-review.md) is available; the exercise has not happened.
- [x] `TAG-SIGNING` verify a successful candidate workflow, artifact checksums, and provenance at `a1d66b2d`. Evidence: [verification record](validation/2026-10-23-session.md). The [validation branch run](https://github.com/chinmay-sawant/ownframe/actions/runs/39836584963) prepares evidence without publishing a version tag or GitHub Release; a production tag remains a separate operator action.
- [-] `INTERNAL-IMPORTS` preserve the current frozen exception inventory. Reconsider removing each pair when a coherent public seam replaces it or its experimental caller is retired. The maintained desktop-cat custom reloadable preview still needs its internal web adapter; adding public forwarding APIs solely to erase a count would make ownership worse.
- [-] `PACKAGE-SPLIT` extract more page/window behavior only when it can own a responsibility with explicit inputs and outputs. Reconsider when a concrete change exposes duplicated state or crosses ownership boundaries.
- [-] `NATIVE-A11Y` live screen-reader adapters remain unshipped. Reconsider with an adapter owner and real assistive-technology validation; see [native accessibility](../documentation/accessibility-native.md).
- [-] `EXPERIMENTS` promote another teaching example only when it teaches a missing public capability and has a maintainer. Reconsider on that evidence; experiments remain in the compatibility suite.

This ledger does not award a perfect rating. Dependency pins are checked against the modules, not copied as release promises. Prepared artifacts, published releases, and platform observations are different claims.
