# Current maintenance ledger

This is the current status record. Release folders, PR records, `PHASES.md`, and `RENDERING-FINDINGS.md` preserve historical evidence and rationale. They do not define present support. User-facing behavior belongs in [documentation/features.md](../documentation/features.md) and [documentation/platforms.md](../documentation/platforms.md).

Use one status per row:

- `[x]` implemented, with a code or validation reference.
- `[ ]` open, awaiting work or validation.
- `[-]` deferred, with a reason and a condition for reconsideration.
- `[s]` superseded, with a replacement reference.

A documented non-goal is deferred, not implemented. Passing a compiler or a unit test does not prove device behavior. Do not copy a phase closure paragraph into other plans. Add evidence beside the row it proves.

## Repository maintenance

- [x] `COHESION` replace the character cap with responsibility and cohesion review; consolidate repaint, window-frame, Wayland-fixture, and GPU comparison tests. Evidence: `AGENTS.md`, `CONTRIBUTING.md`, and the consolidated test families; `make test` exit 0 on 2026-10-22.
- [x] `DOCS` correct API references and compile guide snippets. Evidence: `make docs-check` compiled 31 snippets on 2026-10-22; eight third-party illustrations are explicitly excluded; public Page method references are checked.
- [x] `GRAPHICS` run retained window pixels and replay parity in isolated sessions. Evidence: `make graphics-check` exit 0 on 2026-10-22, all three required tests passed under Xvfb software OpenGL; missing or skipped tests fail the gate.
- [x] `LINT` execute the exact declared golangci-lint version. Evidence: `make lint-version` exit 0 on 2026-10-22; the fixture proves a wrong PATH runner is replaced and the exact cached version is reused.
- [x] `EXAMPLES` maintain a focused teaching set and inventory experiments separately. Evidence: `make examples-check` exit 0 on 2026-10-22, six maintained teaching roots and 49 experimental or support roots; every example remains in `make test`.
- [x] `GATES` run the final repository checks. Evidence: `make test`, `make lint`, `make graphics-check`, frontend `npm run build`, and frontend `npm run check:site` exited 0 on 2026-10-22. The browser check covered 120 layouts including the testing guide. `zizmor` reported no new findings on the edited workflows in offline mode.
- [s] `OLD-CLOSURE` replace the blanket production-readiness closure. Replacement: this ledger; `v0.0.3/prod-ready-task/README.md` explains the archived rows.
- [-] `PACKAGE-SPLIT` further extract page/window responsibilities. Reconsider when a concrete change crosses ownership boundaries; mechanical merging of all files is not this maintenance pass.
- [-] `NATIVE-A11Y` live screen-reader adapters remain unshipped. Reconsider with an adapter owner and real assistive-technology validation; see `../documentation/accessibility-native.md`.
- [-] `EXPERIMENTS` Reconsider promoting additional demos only when they teach a missing public capability and have a maintenance owner. Existing experimental examples remain in the compatibility suite.

Validation used the local checkout with Go 1.26.10. CI workflow changes have been inspected locally, not executed on GitHub. The first maintenance pass was committed as `f8a2100a` and pushed to `origin/feat/prod-readiness` at the author's request. No release or version bump was made. Logs are in the ignored `temp/maintenance-*.log` files.

## Follow-up maintenance

- [x] `PNG-CACHE` invalidate encoded PNGs through page paint damage. Evidence: the new regression failed for operation patches, display patches, ticks, and explicit invalidation before the fix; all four now decode to the updated color without a redraw. The no-op cache test passes.
- [x] `PAGE-LOCK` give each page its own changed-property mutex and remove the global-lock lint exemption. Evidence: `make test-race` exit 0 on 2026-10-22, including the concurrent changed-property regression.
- [x] `COHESION-FILES` consolidate related production methods and update the subsystem maps. Evidence: page production files 185 to 148, window 200 to 179, replay 73 to 71, and the public Testing facade 18 files to one. The public API and platform file constraints are preserved; `make test` exit 0 on 2026-10-22.
- [x] `LINT-OUTPUT` summarize each lint gate and preserve full failure diagnostics. Evidence: `make lint` exit 0 on 2026-10-22 prints eight gate results; `make tools-check` verifies compact success output, full failure output, verbose mode, missing-tool failure, and safe Go file consolidation.
- [x] `FOLLOWUP-GATES` validate the final tooling and generated website. Evidence: `make test`, `make test-race`, `make lint`, `make graphics-check`, frontend build, and 120 browser layout checks exited 0 on 2026-10-22. Logs are in `temp/maintainability-*.log`.

The follow-up maintenance was committed as `c59f1a51` and pushed to `origin/feat/prod-readiness` at the author's request. This ledger does not assign a perfect rating or claim untested platform support. Native input methods, hardware graphics drivers, Android lifecycle behavior, and live screen-reader adapters need platform-specific evidence; see the published support limits.

## State ownership and gate validation

- [x] `STATE-LIFECYCLE` cancel former state registrations on rebind and reject already-dispatched callbacks. Evidence: three regressions failed before the fix; full tests and the library race detector passed on 2026-10-22. Cancellation, reentry, concurrent rebinding, and independent page subscriptions have regressions. The obsolete internal SetState and RemoveListener callers were migrated to BindState and cancellation handles; the root public API is unchanged.
- [x] `VALIDATION-GATES` reject malformed maintenance records even when Python assertions are disabled. Evidence: `PYTHONOPTIMIZE=1 make tools-check plans-check examples-check lint-version` and final `make lint` exited 0 on 2026-10-22. Negative tests reject bad states, missing evidence, duplicate identifiers, conflicting classifications, and stale example inventories.
- [x] `STATE-GATES` run final full tests, race detection, lint, and generated-site checks. Evidence: `make test`, `make test-race`, `make lint`, `make graphics-check`, frontend build, and 120 browser layout checks exited 0 on 2026-10-22. Logs are in `temp/state-lifecycle-*.log`. Validation covered this pass after the pushed `c59f1a51` snapshot.

A local display-free window unit test passed, so no additional Xvfb wrappers were added to non-graphics CI jobs on assumption alone. The graphics gate continues to require a real Xvfb game session. Native host and device claims remain subject to the platform limits in the guides.
