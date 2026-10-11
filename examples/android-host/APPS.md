# Android example APKs

Each example with an Android mobile entrypoint has a signed arm64 release APK:

| Example | Application ID | APK |
| --- | --- | --- |
| Login | `com.chinmaysawant.login` | `examples/login/android/artifacts/login-arm64-release.apk` |
| Platform | `com.chinmaysawant.platform` | `examples/platform/android/artifacts/platform-arm64-release.apk` |
| Dino | `com.chinmaysawant.dino` | `examples/dino/android/artifacts/dino-arm64-release.apk` |
| LG Remote | `com.chinmaysawant.lgremote` | `examples/lg-remote/android/artifacts/lg-remote-arm64-release.apk` |
| Telegram | `com.chinmaysawant.telegram` | `examples/telegram/android/artifacts/telegram-arm64-release.apk` |

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

Gradle outputs, local SDK paths, and generated AARs remain ignored. The APKs
listed above are the deliverables.

## Verify the committed APKs

From the repository root, check the committed arm64 APK snapshot with:

```sh
sha256sum -c examples/android-host/APK-SHA256SUMS
```

All five APKs use the release variant and the same existing signing certificate.
They contain only `arm64-v8a` native libraries and are not debuggable. The
certificate SHA-256 is
`5779480c16bb8403f6bb58d454f3487623c4f7da5e498403859f85bd09b1816f`.
Rebuilding an APK can change its bytes; update the checksum snapshot when
committing rebuilt artifacts.
