# Eyes demo

Cute robot eyes on black that blink, look around, breathe, and cycle
expressions (open, squint, angry, spiral/dizzy). Spiral eyes react to
device motion on Android.

```sh
go run ./examples/eyes          # desktop window, landscape 960x540
go run ./examples/eyes -web     # browser on 127.0.0.1:8140
```

Click to change mood, drag to look, keys 1-4 pick a mood, and S starts
the spiral. Shaking the phone also starts it; it stays until you choose
another mood. The phone auto-cycles open, squint, and angry every 8s with
random saccades every 1.4-3s and blinks every 2.2-4.8s. Mood changes ease
at ~4.5/s while blinks stay snappy. Squint and angry expressions use lids
while keeping the eye width fixed.

The DEV pill at the top right toggles the devtools dock (Elements,
Frame, Ops) without needing F12. The dock answers touch: tabs,
rows, and the picker work from a finger, lists scroll with a drag,
and the X key in the header closes the dock.

## Android

Landscape by default (`sensorLandscape`), portrait works because the
flex layout re-centers and the tick re-bases on generation change.

```sh
sh scripts/android-eyes.sh
ADB=/mnt/c/Users/<you>/platform-tools/adb.exe sh scripts/android-eyes.sh install
```

The app is `com.chinmaysawant.eyes`.
