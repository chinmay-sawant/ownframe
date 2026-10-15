# Android example APKs

Each example with an Android mobile entrypoint builds a signed arm64 release APK locally. Reference builder APKs are committed under `examples/*/android/artifacts/`; the ignore rule below covers new unsigned and debug outputs only.

| Example | Application ID | Release build |
| --- | --- | --- |
| Login | `com.chinmaysawant.login` | `scripts/android-release.sh login` |
| Platform | `com.chinmaysawant.platform` | `scripts/android-release.sh platform` |
| Dino | `com.chinmaysawant.dino` | `scripts/android-release.sh dino` |
| LG Remote | `com.chinmaysawant.lgremote` | `scripts/android-release.sh lg-remote` |
| Telegram | `com.chinmaysawant.telegram` | `scripts/android-release.sh telegram` |

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

Gradle outputs, local SDK paths, generated AARs, and unsigned or debug APKs remain ignored. The signed builder APK committed under the example's `artifacts/` directory is the deliverable. Rebuilding an APK can change its bytes; the signing certificate SHA-256 is
`5779480c16bb8403f6bb58d454f3487623c4f7da5e498403859f85bd09b1816f`.
