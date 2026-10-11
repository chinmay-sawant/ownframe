# Telegram demo on Android

`scripts/android.sh` is the debug build and install path. For the signed
release APK, run `scripts/android-release.sh telegram`; it writes
`artifacts/telegram-arm64-release.apk`. This README is the fresh-machine
walkthrough for Linux.

The [runtime guide](../../../documentation/android-runtime.md) describes coordinates,
input cancellation, IME focus, and lifecycle limits. The
[validation record](../../../plans/v0.0.2/android-runtime-validation.md) distinguishes
build verification from device testing.

All Android mobile example APKs are listed in
[the APK index](../../android-host/APPS.md). Keep the local signing key in
`temp/android-release-signing/` backed up. Updates need the same key.

## One-time setup

The toolchain is Java 17, the Android SDK (platform-tools, a platform,
build-tools, an NDK), and `ebitenmobile`. About 2 GB of downloads.

### 1. Java 17

```
java -version
```

Any 17 or newer JDK works. `JAVA_HOME` may stay unset as long as `java` is on
`PATH`.

### 2. Android command-line tools

Download the Linux package from
<https://developer.android.com/studio#command-line-tools-only> (the file is
named `commandlinetools-linux-<build>_latest.zip`):

```
mkdir -p ~/Android/Sdk/cmdline-tools
cd /tmp
curl -O https://dl.google.com/android/repository/commandlinetools-linux-15859902_latest.zip
unzip -q commandlinetools-linux-15859902_latest.zip -d ~/Android/Sdk/cmdline-tools
mv ~/Android/Sdk/cmdline-tools/cmdline-tools ~/Android/Sdk/cmdline-tools/latest
```

### 3. SDK packages

```
export ANDROID_HOME=$HOME/Android/Sdk
export PATH="$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH"

sdkmanager --licenses          # answer y to each
sdkmanager "platform-tools" "platforms;android-34" "build-tools;35.0.0" "ndk;27.2.12479018"
```

Any recent NDK works; `sdkmanager --list | grep ndk` shows the versions.
Put the two exports in `~/.zshrc` so every shell sees them.

### 4. ebitenmobile

```
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.10.4
```

The version tracks the Ebiten in `go.mod`. `$GOPATH/bin` must be on `PATH`.

## Build and install

On the phone, enable Developer options and USB debugging, connect the cable,
and accept the RSA prompt. `adb devices` lists the phone.

```
sh scripts/android.sh install
```

The first Gradle run downloads Gradle and the Android Gradle plugin. To build
without installing:

```
sh scripts/android.sh
adb install -r examples/telegram/android/app/build/outputs/apk/debug/app-debug.apk
```

The debug APK is self-signed with the debug key. Use the release command above
for the signed app artifact. Copy an APK to a phone and open it there if USB is
not an option ("install unknown apps" permission required).

## APK size

The APK is mostly native code: the Go runtime, Ebitengine, and the HTML/CSS
layout engine, compiled into one shared library per Android ABI. The default
bind builds all four, arm64-v8a at 40 MB, armeabi-v7a at 52 MB, x86 at 53 MB,
and x86_64 at 43 MB, so the debug APK lands near 188 MB. A phone runs one of
them.

```
ANDROID_TARGET=android/arm64 sh scripts/android.sh
```

That binds arm64 only and passes `-ldflags "-s -w"`, which drops the Go symbol
tables and DWARF. The same APK comes out at about 29 MB. Use `android/arm`
for a 32-bit ARM phone, or leave the default for an APK that also runs on an
emulator.

Gradle reuses the previous APK file when it repackages, so after switching the
target the file on disk can keep its old size with the removed libraries left
as dead space inside. `rm -rf examples/telegram/android/app/build` before the
rebuild, or `./gradlew clean`, writes it fresh.

## WSL2 and the phone

WSL2 does not see USB devices, so `adb devices` stays empty when the phone is
plugged into the Windows host. Wireless debugging is the short path; Android
11 or newer and the phone on the same Wi-Fi are the requirements.

1. On the phone, enable Developer options and Wireless debugging, then tap
   "Pair device with pairing code". The screen shows an `IP:port` and a
   six-digit code; keep it open.
2. In WSL: `adb pair <ip:port> <code>`.
3. The Wireless debugging screen shows a second `IP:port` at the top. Run
   `adb connect <ip:port>`, and `adb devices` lists the phone.
4. `sh scripts/android.sh install`, or `adb install -r` as above.

The pairing port changes every time the pairing dialog opens. If the network
blocks client-to-client Wi-Fi traffic, forward the cable with
[usbipd-win](https://github.com/dorssel/usbipd-win) instead:

```
winget install usbipd
```

Then in an admin PowerShell: `usbipd list`, `usbipd bind --busid <busid>`
once, and `usbipd attach --wsl --busid <busid>` after each replug. adb in WSL
sees the phone after the attach.

## What runs on the phone

`examples/telegram/mobile` calls `ownframe.BindMobile`, so the phone draws the same
page `go run ./examples/telegram` shows on the desktop. Taps, drag-to-scroll,
the tabs, opening chats, and Send all work.

The demo wires the soft keyboard: tap the Message field and type, and the
composer sits directly above the keyboard. The system back gesture works
too: from a chat it returns to the list, and from the list it closes the
app. The paperclip opens the attachment drawer with Camera and Gallery,
and Anna Petrova's chat has the gift button. Photos arrive as rounded
bubbles. While the messages scroll, the top bar and the composer stay
pinned, drawn above the messages so a scrolled thread never covers them.

The activity draws edge to edge and passes the status bar, cutout, and
gesture area insets to the page, so the list and composer keep clear of
them.

## How the pieces fit

- `examples/telegram/mobile`: the package `ebitenmobile bind` compiles. The
  generated `EbitenView` lives in Java package
  `com.chinmaysawant.telegram.mobile`.
- `telegram/telegram.aar`: the bound Go runtime plus the view classes. The
  bind writes it on every build; it is never committed.
- `app`: one `MainActivity` fills the screen with the `EbitenView` and
  suspends the game on `onPause`.
- `telegram`: a tiny Gradle module that exposes the AAR to `app`.

## When the build fails

- `no usable NDK in ...`: install an NDK under `$ANDROID_HOME/ndk`.
- `SDK location not found`: export `ANDROID_HOME`, or write
  `local.properties` next to `settings.gradle` with `sdk.dir=<path>`.
- `INSTALL_FAILED_UPDATE_INCOMPATIBLE`: a differently signed build is
  installed; run `adb uninstall com.chinmaysawant.telegram` first.
- The app closes when you press the system back button: Android handles that
  key, so use the on-screen back arrow to leave a chat.
