# Release checklist

Run this list before tagging. Every item must be checked off; a miss
blocks the tag.

- `make test` and `make lint` pass on both modules.
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
