# Packaging

`scripts/package.sh <example>` builds one example for this system and writes
a release archive to `dist/`. The build uses `-trimpath` and
`-ldflags "-s -w"`. The script makes no network calls and adds no modules.

```
sh scripts/package.sh print          # archive the print example
sh scripts/package.sh -n print       # print the layout, build nothing
```

The dry run prints the archive names and every entry, so you can check the
layout without a toolchain for the target. It still asks Go for `GOOS` and
`GOARCH`, so `GOOS=darwin sh scripts/package.sh -n print` prints the macOS
layout on any system.

## Archive layouts

Linux, a `tar.gz`:

```
ownframe-print-linux-amd64.tar.gz
  ownframe-print/print
  ownframe-print/print.desktop
  ownframe-print/README.md
```

macOS, a zip with an app bundle:

```
ownframe-print-macos-arm64.zip
  ownframe-print.app/Contents/Info.plist
  ownframe-print.app/Contents/MacOS/print
```

Windows, a zip with the executable:

```
ownframe-print-windows-amd64.zip
  ownframe-print.exe
```

wasm, on every system:

```
ownframe-print-wasm.zip
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
