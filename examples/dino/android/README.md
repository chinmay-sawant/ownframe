# Dino Run on Android

`scripts/android-dino.sh` is the debug build and install path.
`scripts/android-release.sh dino` creates the signed release APK at
`artifacts/dino-arm64-release.apk`. The toolchain setup and the WSL2 USB story are in
[examples/telegram/android/README.md](../../telegram/android/README.md).

Release APKs for every Android mobile example are listed in
[the APK index](../../android-host/APPS.md). Keep the local signing key in
`temp/android-release-signing/` backed up. Updates need the same key.

## Build and install

    export ANDROID_HOME=$HOME/Android/Sdk
    sh scripts/android-dino.sh
    sh scripts/android-dino.sh install

The default bind is arm64 only, which is what a Pixel-class phone runs. The
script builds every Android ABI with `ANDROID_TARGET=android`.

From WSL2 the phone is visible to the Windows adb, not the Linux one:

    ADB=/mnt/c/Users/<you>/platform-tools/adb.exe sh scripts/android-dino.sh install

When `ADB` ends in `.exe` the script hands the APK path over with
`wslpath -w`.

## What runs on the phone

`examples/dino/mobile` calls `ownframe.BindMobile`, so the phone draws the same
page `go run ./examples/dino` shows on the desktop. The activity is locked
to landscape, where the 900x300 scene fills the view. A tap anywhere jumps:
it starts the first run, leaps while running, and restarts after a crash.
The key controls stay on the desktop build.

The app is `com.chinmaysawant.dino`. `adb uninstall com.chinmaysawant.dino`
clears a build that was signed differently.
