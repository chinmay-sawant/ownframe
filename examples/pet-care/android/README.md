# Android build

Install Go, Java 17, Android SDK platform 34, build tools, NDK, and
`ebitenmobile` v2.10.4. Set `ANDROID_HOME` or `ANDROID_SDK_ROOT`, then run
`scripts/pet-care-android.sh` from the repository root. The debug APK is
copied to `temp/pet-care/pet-care-debug.apk` after Gradle assembles it.

Pass `install` to build and run `adb install -r`:

```sh
scripts/pet-care-android.sh install
```

To target only an arm64 phone, set `ANDROID_TARGET=android/arm64`. Do not
export `GOWORK` for the bind step. The generated AAR is a build artifact.

The Activity uses `examples/android-host` for system insets, lifecycle, and
Back dispatch. Backup import and export use Android's system document picker.
APK compilation does not establish picker or other physical-device behavior.
