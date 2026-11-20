# Release checklist

Run this list before tagging. Every item must be checked off; a miss
blocks the tag.

- `make test` and `make lint` pass on both modules.
- The reusable CI workflow passes before release packaging, including library
  and example race checks, tagged debug routes, security/fuzz checks, cross
  compilation, workflow/secrets checks, and archive reproducibility.
- The same-runner parent/candidate benchmark comparison and 60-second page
  soak pass. Retain raw/JSON performance artifacts and record a dated baseline.
- Linux, Windows, and macOS packages pass the release bundle inventory check.
  Android binding compilation is separate from APK and device evidence.
- `govulncheck` is clean; `go.mod` bumps are reviewed.
- Threat-model refresh: re-read documentation/security.md assumptions
  and the data-flow diagram, and update them for what actually ships:
  new screens, inputs, fetch callers, mobile surfaces, IPC channels,
  and debug endpoints. This is the GEN-23 gate.
- New risky options use an `InsecureAllow`/`Unsafe` prefix, zero values
  stay safe, and footguns carry a `// Deprecated:` pointer.
- New complex-input paths (HTML, CSS, images, URLs, click queries)
  have a fuzz target with seed corpus.
- Security regression tests (`go test -run Sec ./... ./examples/...`)
  pass.
- Packaging notes (signing, sandboxing, permissions) still match the
  shipped artifacts. Unsigned `scripts/package.sh` archives are the 1.0
  deliverable; signing and notarization stay in the operator runbook
  ([packaging.md](packaging.md)).

## Two-module release tag sequence (`go.mod` and `examples/go.mod`)

`ownframe` and `examples/` are separate Go modules joined by `go.work` during
development. Tag them in order so published module zips never carry a local
`replace` directive or an untagged engine pseudo-version:

1. **Verify the `blinkless` pin:** Confirm `github.com/chinmay-sawant/blinkless`
   in both `go.mod` and `examples/go.mod` is pinned to the same tagged semver
   release (`vA.B.C`), not a `-0.2026...` commit pseudo-version.
2. **Verify library archive and release bundles:** Run `make module-check` and
   `make release-check` from the repository root.
3. **Tag the root library module:** Commit any final `go.mod` / `go.sum` updates
   and tag `vX.Y.Z` on the release commit.
4. **Update `examples/go.mod` for the published library tag:**
   - Bump `require github.com/chinmay-sawant/ownframe` in `examples/go.mod` to
     `vX.Y.Z`.
   - Remove `replace github.com/chinmay-sawant/ownframe => ../` from
     `examples/go.mod` and run `GOWORK=off go -C examples mod tidy`.
   - Validate the candidate examples module without the workspace:
     `GOWORK=off go -C scripts run ./cmd/module-check -root .. -examples-tag examples/vX.Y.Z`.
5. **Tag the examples module:** Commit `examples/go.mod` and `examples/go.sum`,
   then tag `examples/vX.Y.Z`.
6. **Restore development workspace replacement:** Re-add
   `replace github.com/chinmay-sawant/ownframe => ../` in `examples/go.mod` on
   the development branch after tagging `examples/vX.Y.Z`.
