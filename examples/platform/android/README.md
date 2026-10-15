# Ownframe Platform Android app

This project wraps `examples/platform/mobile` as an Android app and uses the shared `examples/android-host` module for insets and lifecycle handling.

From the repository root, run `scripts/android-release.sh platform` to bind the Go package and build the signed arm64 release APK. The generated AAR and Gradle build outputs are ignored. Release signing uses the local Android keystore under the ignored `temp/android-release-signing/` directory. Keep a backup of that key because future updates need the same signing identity.

The release build writes `artifacts/platform-arm64-release.apk`, which is ignored and not committed. All Android example APKs are listed in [APPS.md](../../android-host/APPS.md).
