# Depth Alert

Depth Alert is an experimental Android example that uses ARCore depth to warn
about a close surface in the center of the camera view. It targets phones with
ARCore Depth API support, including Pixel 7.

## Signals

Start scanning with the native Android button and aim the rear camera forward.
Vibration repeats while a state is active:

| State | Signal | Prototype rule |
|---|---|---|
| Walk | One short pulse each second | Twelve consecutive clear depth estimates, with no detected surface closer than 3.5 m |
| Stop | Two short pulses each second | A confident center-view distance at or below 3 m |
| Stop now | Near-continuous vibration | A confident center-view distance at or below 1.2 m |
| Depth unknown | Three short pulses, then a pause | Tracking or depth support is inadequate, or a 3–3.5 m surface contradicts a prior Walk state |

The thresholds are initial prototype values. The app uses ARCore's automatic
depth image, applies a 3×3 median filter to reject isolated noisy pixels, then
uses the 10th percentile of valid depths in the central 40% of the view. The Go
classifier requires twelve consecutive clear frames before reporting “Walk.”
A surface in the 3–3.5 m buffer band resets a prior Walk state to
“Depth unknown.” Stop alerts remain active through brief missing-depth gaps and
until twelve clear frames arrive; after 500 ms without a confident frame, the
state becomes “Depth unknown.” “Walk” only reports that the center camera
region has no detected close surface; it does not mean the path is safe. The
spatial filter can reject thin structures, which is another
reason this prototype cannot guide independent travel. Move the phone slowly
across the scene so ARCore can establish depth.

## Build

Set up the Android SDK, NDK, Java 17, and `ebitenmobile` as described in
[the Telegram Android setup](../telegram/android/README.md). Then run from the
repository root:

```sh
sh scripts/android-depth-alert.sh
```

For local development, `sh scripts/android-depth-alert.sh` builds a debug APK.
Build the signed arm64 release APK with `sh scripts/android-release.sh depth-alert`.
It is written to
`examples/depth-alert/android/artifacts/depth-alert-arm64-release.apk`.
Install it from WSL2 after configuring Windows-host ADB with:

```sh
ADB="$(wslpath -u 'C:\Users\YOUR_WINDOWS_USER\AppData\Local\Android\Sdk\platform-tools\adb.exe')"
"$ADB" install -r "$(wslpath -w 'examples/depth-alert/android/artifacts/depth-alert-arm64-release.apk')"
```

Replace `YOUR_WINDOWS_USER` with the Windows profile directory. The release
build requires the repository's configured Android release-signing key.

## Design

Android owns ARCore, the live camera surface, permission and lifecycle handling,
TalkBack-visible controls, and vibration. It sends only a compact depth summary
to Go: nearest distance, valid-depth coverage, and tracking status. Go classifies
that summary and the OwnFrame page displays the current state. The camera image
never crosses the Go bridge. The camera and OwnFrame views use separate screen
regions because each owns its own rendering surface.

## Limits

This is a research prototype, not a mobility aid. It detects some surfaces in a
small center region. It does not identify objects, moving traffic, stairs,
drop-offs, overhead hazards, safe crossings, or a navigable route. Depth can be
missing or wrong, especially on thin, reflective, dark, distant, or moving
surfaces. Keep using a cane, guide dog, or other established mobility support;
the app reports unknown when it lacks a fresh confident estimate.

This revision was installed and captured on a Pixel 7. That confirms installation
and rendering on this device; it does not validate depth accuracy, haptic
reliability, or safe navigation. The screenshot shows one live observation. It
demonstrates rendering only and does not establish that a walking path is safe.
