# Packaging

`scripts/package.sh <example>` builds one example for this system and writes
a release archive to `dist/`. The build uses `-trimpath` and
`-ldflags "-s -w"`. The script disables module downloads and toolchain downloads. Populate and verify
the dependency cache before packaging. Missing cached inputs fail the build.
Optional `COSIGN=1` signing uses the network, as described below.

```
sh scripts/package.sh login          # archive the login example
sh scripts/package.sh -n login       # show the layout, build nothing
```

The dry run prints the archive names and every entry, so you can check the
layout without a toolchain for the target. It still asks Go for `GOOS` and
`GOARCH`, so `GOOS=darwin sh scripts/package.sh -n login` prints the macOS
layout on any system.

## Archive layouts

Linux, a `tar.gz`:

```
ownframe-login-linux-amd64.tar.gz
  ownframe-login/login
  ownframe-login/login.desktop
  ownframe-login/README.md
```

macOS, a zip with an app bundle:

```
ownframe-login-macos-arm64.zip
  ownframe-login.app/Contents/Info.plist
  ownframe-login.app/Contents/MacOS/login
```

Windows, a zip with the executable:

```
ownframe-login-windows-amd64.zip
  ownframe-login.exe
```

wasm, on every system:

```
ownframe-login-wasm.zip
  ownframe.wasm
  wasm_exec.js
  index.html
```

`index.html` is the repository's `browser/index.html`. It loads
`ownframe.wasm` through `wasm_exec.js` from the same directory, so serve the
unpacked directory over HTTP.

`SHA256SUMS` sits next to the archives and covers the archives from the same
run. `scripts/package_test.sh` checks the dry-run layout against this file
and needs no build.

## Signing and notarization

Unsigned archives from `scripts/package.sh` are the 1.0 deliverable. Signing,
notarization, and Authenticode are operator runbook steps, not library code.
The script ships unsigned archives. Each system treats one differently.

- Linux has no signing step. A `tar.gz` unpacks and runs.
- macOS Gatekeeper blocks an unsigned `.app` from an unknown developer on
  first launch. The person can right-click and choose Open, or you can sign
  and notarize outside the script with an Apple Developer certificate,
  `codesign`, `notarytool`, and `stapler`. That work needs an Apple account
  and belongs in your release pipeline.
- Windows SmartScreen warns on an unsigned `.exe` from an unknown publisher.
  An Authenticode certificate and `signtool` remove the warning, again
  outside this script.
- The wasm zip has no signing step; the browser trusts the server it came
  from.

Icons are the platform default. Custom icons need per-platform resource
tools for `icns` and `.ico` files, which this script does not include.

## Auto-update

The application or its distribution channel owns updates. OS stores and
package managers update the apps they install. An application distributed
outside those channels can ship its own updater process and update policy.

The package script creates archives and SHA-256 checksums. A checksum detects
changes to an archive; it does not authenticate the publisher. The script
does not contact an update service, sign releases, install an update, or
restart an application.

A custom updater belongs to the application. It must verify a signed release
manifest and downloaded artifact, replace the installed files safely, roll
back on failure, and restart only after the application saves its state and
closes its window.

See the [browser API and updater proposal](../plans/v0.0.2/browser-apis-and-updater-proposal.md)
for the library boundary and future acceptance criteria.

## Signed updates

A naive updater is remote code execution. An updater that downloads and
runs files must authenticate every byte before executing or installing it:

- Sign a release manifest (artifact names plus SHA-256 hashes) with an
  offline key, and verify the signature in the updater before fetching
  anything else. Reference: TUF (`github.com/theupdateframework/go-tuf`)
  or a minimal `minisign`/`ed25519` signature over the manifest.
- Pin the manifest's version and refuse rollbacks: never install an
  artifact older than the running build.
- Fetch manifests and artifacts over HTTPS only, with the platform TLS
  roots; do not add opt-out flags.
- Stage the download next to the install, verify the hash, swap it in,
  and keep the previous copy until the new build starts cleanly.

## OS sandboxing

Sandboxing limits the blast radius of a renderer bug. Apply the platform
mechanism in your release pipeline; the package script stays unsigned and
unsandboxed by default.

- macOS: sign with the hardened runtime and a minimal entitlements plist.
  Ship only the entitlements the app needs; no `com.apple.security.cs.allow-jit`
  or `allow-unsigned-executable-memory` unless the app justifies it:
  ```xml
  <dict>
    <key>com.apple.security.app-sandbox</key><true/>
    <key>com.apple.security.network.client</key><true/>
  </dict>
  ```
- Linux: ship a Flatpak manifest with narrow permissions. Start from no
  filesystem access and add only what the app reads or writes:
  ```json
  "finish-args": ["--share=ipc", "--socket=x11", "--socket=wayland",
                  "--share=network", "--filesystem=xdg-download"]
  ```
- Windows: embed an `asInvoker` application manifest so the app never
  gains elevation it did not ask for:
  ```xml
  <requestedExecutionLevel level="asInvoker" uiAccess="false" />
  ```

## Secret scan

Before tagging, run `sh scripts/scan-secrets.sh`. API keys stay server
side (see [secrets.md](secrets.md)): the scan greps tracked source for
token patterns and runs `strings` over the `dist/` archives. Install
`gitleaks` for the full rule set; `.gitleaks.toml` allowlists only the
demo credentials by path.

## Binary hardening

`scripts/package.sh` links position-independent binaries on Linux and
macOS (`-buildmode=pie`; Windows gets ASLR from the OS loader) with the
C stack protector and full RELRO:

```
CGO_CFLAGS="-O2 -D_FORTIFY_SOURCE=2 -fstack-protector-strong"
CGO_LDFLAGS="-Wl,-z,relro,-z,now"
```

Verify the result before tagging. Install `checksec` and confirm PIE
and RELRO on the unpacked binary:

```
checksec --file=dist/ownframe-login/login
```

Look for `PIE enabled` and `Full RELRO`. A `No PIE` or `Partial RELRO`
line means the flags did not reach the link; rebuild with
`sh scripts/package.sh` and check that no environment override cleared
`CGO_CFLAGS` or `CGO_LDFLAGS`.

## Signed artifacts and provenance

Checksums detect corruption, not tampering. Tags additionally ship:

- `sbom.spdx.json`: the release workflow generates it with
  `anchore/sbom-action` and uploads it with the build provenance.
- A SLSA build attestation over the SBOM
  (`actions/attest-build-provenance`), verifiable on the release page
  and with `gh attestation verify`.
- A keyless cosign signature (`cosign sign-blob --yes`) over the same
  file, verifiable with `cosign verify-blob`.

For a local archive, opt into the same signature step:

```
COSIGN=1 sh scripts/package.sh login
```

That signs `dist/SHA256SUMS` into `SHA256SUMS.sig` plus the Fulcio
certificate `SHA256SUMS.pem`. Verify with
`cosign verify-blob --cert SHA256SUMS.pem --signature SHA256SUMS.sig SHA256SUMS`.

Reproducibility backs all of this: `sh scripts/reproducible-check.sh`
builds the login example twice with fixed inputs and fails on any byte
difference, and the `reproducible` CI job runs it on every push.
`documentation/supply-chain.md` has the full release checklist.

## OS code signing in CI

Unsigned apps train users to bypass Gatekeeper and SmartScreen. Sign in
CI, never on a dev laptop:

- macOS: `codesign --deep --sign "$APPLE_DEVELOPER_ID"` with the
  hardened runtime, then `xcrun notarytool submit --wait` and
  `xcrun stapler staple`. Notarization needs an App Store Connect API
  key, not a developer password.
- Windows: `AzureSignTool sign` against a certificate in Azure Key Vault
  (or `signtool` with a cloud-backed key). The private key never leaves
  the HSM; CI holds only short-lived credentials.

`scripts/sign-release.sh` runs exactly those steps. It reads every
secret from the environment and skips a platform with a message when
its tooling or credentials are absent. Wire it to a protected
`release-signing` environment (required reviewers, short-lived secrets,
no secret ever echoed) once a publish job lands in
`.github/workflows/release.yml`, and keep the signing keys out of
developer machines, chat logs, and the repository.


## Archive reproducibility and signing

`package.sh` writes sorted tar/gzip and ZIP entries with normalized timestamps,
owners and modes through `scripts/release_archive.py`. `SOURCE_DATE_EPOCH`
defaults to the source commit timestamp. `scripts/reproducible-check.sh` builds
Linux and wasm candidates twice on this host and compares complete archive
checksums. This does not promise identical bytes across different toolchains
or operating systems.

`sign-release.sh --platform windows` or `--platform macos` extracts matching
archives, signs executable members, notarizes and staples macOS app bundles,
repacks and refreshes `SHA256SUMS`. Missing tools, credentials, or signable
inputs fail. Old checksum signatures are removed; re-sign checksums after the
archive contents change. Fixture tests prove invocation and repacking; real
platform signatures require an operator run with credentials.
