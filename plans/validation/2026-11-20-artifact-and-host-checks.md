# Artifact and host checks on 2026-11-20

OwnFrame checkout based on `b27c49c2515bf02d610318a0107f29a495457cb0`, with
local v0.0.4 artifact, dependency, and demo changes. Go toolchain
`go1.26.11`. Both modules pin Blinkless `v0.1.2` and Ebitengine `v2.10.6`.

## Artifacts and dependency size

- All eight signed arm64 APKs matched the Git blob contents from
  `origin/master` at `63464cd0f77bed2c0f27f6b052a92dd20f0a3034`.
  All eight GitHub downloads passed `sha256sum -c APK-SHA256SUMS`. The
  manifest is also in [examples/android-host](../../examples/android-host/APK-SHA256SUMS).
  [Android reference release](https://github.com/chinmay-sawant/ownframe/releases/tag/android-reference-2026-11-20).
- `ANDROID_ARTIFACT_DIR=<scratch> sh scripts/android-release.sh login`
  exited 0 after assembling and signing a fresh arm64 APK. Android build-tools
  35.0.0 verified APK signature schemes v1, v2, and v3. The fresh APK's
  SHA-256 was `bd1b5b221f5b5d5500a54ceb93393f612b241dd2252dd6a02f270130f4ce1265`.
  This arm64 APK installed on an Android 14 x86_64 emulator that advertises
  arm64 translation. The emulator translator rejected an ARM instruction
  and the process exited with SIGILL. A native ARM64 launch remains open.
  A second signed build, using `ANDROID_TARGET=android/amd64`, passed
  installation and a cold foreground launch on the same API 34 emulator.
  A captured frame showed the sign-in form. The x86_64 APK SHA-256 was
  `e77e47bda31e79cc5c80864e91bef0b548c84549b99573f0cc38cb8cf4a008d5`.
- The Pages artifact check found nine WASM payloads. GitHub's live Pages
  source used `master:/docs` at the start of the migration. After
  [PR #84](https://github.com/chinmay-sawant/ownframe/pull/84) merged and
  the source switched to GitHub Actions, the
  [first Actions deployment](https://github.com/chinmay-sawant/ownframe/actions/runs/44716541603)
  succeeded at `bad50392`. Chrome opened all nine live demos while the old
  blobs were still tracked on `master`. The local site check separately
  passed 140 layout checks.
- A local upstream Blinkless patch was tested from upstream `05f7df50`.
  `go test -p 2 ./internal/fonts ./internal/svg ./layout` passed with the
  default configuration. Using an alternate workspace against that checkout,
  `go build -p 2 -buildvcs=false -trimpath -ldflags='-s -w' ./examples/hello`
  produced 33,416,808 bytes by default and 26,952,424 bytes with
  `-tags blinkless_minimal` on Linux amd64. The package list for the compact
  build excludes canvas, its PDF and TeX dependencies, Latin Modern, and
  Goldmark. The default build retains SVG and its bundled fonts. Upstream
  review and release of the patch remain open.

## Native desktop

Windows 11 Home Single Language, version `10.0.26200`. The Go executable
was built for `windows/amd64` and launched as a Windows process through
PowerShell from WSL. PowerShell observed an interactive Windows session
and these results.

| Check | Result | Evidence |
|---|---|---|
| `Run` | **Pass** | Win32 reported a 600×414 client area. Sending `WM_CLOSE` returned from `Run`. |
| Custom chrome | **Pass** | The window had no native caption. Clicking the page's custom Close control dispatched its action and returned from `RunWithOptions`. |
| Clipboard read | **Pass** | The executable read nonempty text through the Windows backend. The probe recorded only whether the text was nonempty. |
| Clipboard write | **Open** | The probe left desktop clipboard contents intact. |
| File picker | **Pass** | The native `comdlg32` dialog accepted a fixture path via its filename edit and returned `C:\Windows\win.ini`. |
| Japanese Microsoft IME | **Open** | Windows listed en-US, en-IN, and Japanese Microsoft IME. The native probe selected a Japanese keyboard layout, but the field committed Latin `kana` directly. Native preedit, cancellation, conversion, CJK commit, and mode switching remain unverified. |
| macOS | **Open** | No macOS host was available for this session. |

## Linux accessibility

`dbus-run-session -- xvfb-run -a python3 scripts/check-atspi-actions.py`
verified live focus, activation, Unicode set/insert/delete, and exclusion
of disabled and password fields from `EditableText`. The page tests
covered Unicode offsets and the in-memory clipboard. A follow-up run
used Xvfb in a private mount and network namespace with its own
`/tmp/.X11-unix` directory. The live bus verified copy, cut, and paste
against the X11 clipboard using synthetic Unicode input. Orca speech
remains unverified. Windows UIA, macOS NSAccessibility, and
Android TalkBack remain open.

## Local verification

`xvfb-run -a make test TEST_P=2` exited 0 in a clean copy of the proposed
source tree `b887dd9d511f6546c3652e195a13c641f37d57bb`, before the final
evidence edits. The copy included Git-tracked documentation and no tracked
WASM or APK files. This run used a local verification commit and did not
exercise the independent teammate handoff.

The following also exited 0 on the local changes.

- `xvfb-run -a make test-race`
- `make lint`
- `make docs-check plans-check examples-check tools-check repo-size-check`
- `make internal-import-check module-check release-check`
- `xvfb-run -a make graphics-check`
- `make site-check`
- `make vuln` plus `go tool govulncheck ./...` from `examples/`
- Actionlint for `.github/workflows/pages.yml` and `.github/workflows/ci.yml`

The scans reported zero vulnerabilities reached by application code and two
uncalled findings in imported packages. The size gate lists ten retained
media files over 5 MiB and no violations. Final verification on the merged
tree and the independent teammate handoff remain open.
