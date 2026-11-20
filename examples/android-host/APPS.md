# Android example APKs

Download signed arm64 APKs from the [Android reference release](https://github.com/chinmay-sawant/ownframe/releases/tag/android-reference-2026-11-20). Download [APK-SHA256SUMS](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/APK-SHA256SUMS) alongside the APKs and run `sha256sum -c APK-SHA256SUMS` in that directory. The same hashes are also [recorded in Git](APK-SHA256SUMS).

| Example | Release APK | Local build |
| --- | --- | --- |
| Login | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/login-arm64-release.apk) | `scripts/android-release.sh login` |
| Platform | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/platform-arm64-release.apk) | `scripts/android-release.sh platform` |
| Dino | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/dino-arm64-release.apk) | `scripts/android-release.sh dino` |
| LG Remote | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/lg-remote-arm64-release.apk) | `scripts/android-release.sh lg-remote` |
| Telegram | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/telegram-arm64-release.apk) | `scripts/android-release.sh telegram` |
| Eyes | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/eyes-arm64-release.apk) | `scripts/android-release.sh eyes` |
| Depth Alert | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/depth-alert-arm64-release.apk) | `scripts/android-release.sh depth-alert` |
| Digital Wellbeing | [arm64 APK](https://github.com/chinmay-sawant/ownframe/releases/download/android-reference-2026-11-20/digital-wellbeing-arm64-release.apk) | [Android project](../digital-wellbeing/android/) |

Build one with `scripts/android-release.sh <example>`, for example:

```sh
scripts/android-release.sh lg-remote
```

The script binds the selected Go mobile package, assembles the Gradle release
variant, signs it, and verifies the signature. It defaults to `android/arm64`.
The Android SDK, `ebitenmobile`, Java 17, and a release keystore are required.
Set `ANDROID_BUILD_TOOLS_VERSION` if the installed build-tools version is not
35.0.0.

The current local signing key and password files live under the ignored
`temp/android-release-signing/` directory. Keep a secure backup of that key.
Every future update must use the same key to install over these APKs. The key
and passwords are not part of the repository.

The signed APK, build metadata, and certificate report are written to the
example's local `android/artifacts/` directory. Set `ANDROID_ARTIFACT_DIR`
to write them elsewhere.

Rebuilding an APK can change its bytes. To install over a reference APK, sign
with the same certificate as that APK. Check the certificate with
`apksigner verify --print-certs <apk>`.
