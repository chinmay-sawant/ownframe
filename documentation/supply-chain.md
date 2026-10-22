# Supply chain, releases, and branch protection

How tags, archives, and the `master` branch stay trustworthy. The audit
items live in `plans/v0.0.3/security/SECURITY_AUDIT_3.md` as GEN-11
through GEN-17; this page is the operator checklist.

## Protected branches and tags

A single compromised session must not be able to push to `master` or
move a tag. In repository Settings > Rules, protect `master` and the
`v*` and `examples/v*` tag patterns with:

- Required status checks: the `gates` job (`make test`, `make lint`),
  the `reproducible` job, CodeQL, Semgrep, and the secret scan.
- Required pull request review before merge, with dismissal of stale
  approvals when new commits land.
- Required signed commits. SSH signing is the simplest setup:
  `git config -S` is not a thing, so set it explicitly:
  ```
  git config commit.gpgsign true
  git config gpg.format ssh
  git config user.signingkey ~/.ssh/id_ed25519.pub
  git config tag.gpgsign true
  ```
  `gitsign` (Sigstore, no long-lived key) is an equivalent choice.
- No force pushes and no deletions, on branches and on tags.
- Required linear history is optional; signed merge commits are fine.

Tags are immutable once published. The Go module proxy caches them
forever, so a bad or vulnerable release is fixed with a `retract`
directive in `go.mod` plus a new tag, never by re-tagging. Verify the
live settings with:

```
gh api repos/{owner}/{repo}/branches/master/protection
gh api repos/{owner}/{repo}/tags/v0.0.2/protection
```

## Signed release artifacts and provenance

`SHA256SUMS` detects a corrupted download. It does not authenticate the publisher. The tag workflow waits for test, lint, graphics, module, and website gates, then prepares a `release-candidate` artifact containing:

- Desktop and wasm login archives plus checksums and a keyless cosign signature over `dist/SHA256SUMS`.
- `sbom.spdx.json`, generated over the checkout, with its own cosign signature and certificate.
- SLSA build provenance covering the SBOM, archive files, and checksum file.

Verify provenance with `gh attestation verify <file> --repo <owner>/<repo>`. Verify cosign blobs against the expected repository workflow identity and OIDC issuer, as well as their certificate and signature. A matching checksum alone is insufficient.

This workflow prepares a candidate; it does not create a GitHub Release or upload published assets. An operator must review the successful tag run and approve publication. Local `make release-check` verifies archive structure and bytes but does not exercise GitHub OIDC signing. APKs use their separate Android signing and device-validation path.

For a local archive the same signature step is opt-in:

```
COSIGN=1 sh scripts/package.sh login
```

That signs `dist/SHA256SUMS` into `SHA256SUMS.sig` plus the Fulcio
certificate `SHA256SUMS.pem`. Keyless signing needs network access to
Fulcio and Rekor; without it the step fails loudly instead of writing
an unsigned file and calling it signed.

## SBOM per release

Downstream users need to know what is inside the desktop, wasm, and
APK artifacts. The workflow scans the checkout, including both Go modules, into SPDX JSON. It uploads the SBOM, signatures, certificates, and desktop/wasm archives as `release-candidate`. A repository SBOM is not a claim that a separately built APK was inspected. Attach only reviewed artifacts to a GitHub Release after the candidate run passes. To reproduce the SBOM
locally without the workflow:

```
syft ./ -o spdx-json=sbom.spdx.json
```

or, for the module graph only:

```
go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest app -output sbom.json
```

## Reproducible builds

`sh scripts/reproducible-check.sh` builds `./examples/login` twice
with `-trimpath -buildvcs=true`, `-ldflags "-s -w"`, and a fixed
`SOURCE_DATE_EPOCH`, then fails when the two binaries differ by even
one byte. The `reproducible` CI job runs it on every push. A mismatch
means the build embeds a timestamp, an absolute path, or randomness;
find the source (often a `time.Now()` in generated code or a missing
`-trimpath`) and fix it before tagging. The login example stands in
for every release binary because they all share the same flags in
`scripts/package.sh`.

## Compiling out debug surfaces

Release binaries must not contain the network-reachable debug code.
The `ownframe_debug` build tag controls it:

- Without the tag, `GET /debug/state` and `GET /debug/pprof/*` are not
  registered even when `ServeOptions` opts in, and `net/http/pprof`
  is not linked. Requesting them logs one startup line pointing at
  the tag.
- `OWNFRAME_FILEPICK_DEBUG` and `OWNFRAME_PACE_DEBUG` output is
  compiled out without the tag; setting the variables does nothing.
- The window devtools overlay and the `Config.DevTools`, `Perf`, and
  `Debug` flags stay runtime-gated (local-only, default off), because
  the overlay has no network path. Its network-observable effects are
  what the tag removes.

Build a debug binary with `go run -tags ownframe_debug ./examples/login`
or `go test -tags ownframe_debug ./internal/web/`. The
`TestDebugRoutesAbsentWithoutTag` test pins the 404s in a default
build and `TestDebugRoutesPresentWithTag` pins the 200s with the tag.

## OS code signing

See `packaging.md` ("OS code signing in CI"): sign and notarize in CI
behind the protected `release-signing` environment with HSM or
KMS-backed keys, never on a dev laptop. `scripts/sign-release.sh`
implements the steps and skips platforms whose tooling or credentials
are absent.
