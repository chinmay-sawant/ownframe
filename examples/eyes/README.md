# Eyes demo

Robot eyes on black with 12 selectable expressions: open, squint, angry,
spiral, relaxed, sleepy, surprised, curious, skeptical, left wink, right
wink, and bored. The spiral also reacts to device motion on Android.

```sh
go run ./examples/eyes          # desktop window, landscape 960x540
go run ./examples/eyes -web     # browser on 127.0.0.1:8140
```

The AUTO button cycles through all 12 expressions. Each holds for 8 seconds;
spiral holds for 4 seconds before it fades into the next expression. Tap an
eye to skip to the next expression, drag to look, keys 1-4 pick the original
four, and S starts the spiral. Shaking the phone starts a short spiral too.
Both modes keep the eyes animated: they blink, breathe, move their gaze, and
finish each expression animation. Angry brows sway, and wink expressions close
and reopen the viewer's named eye. STATIC stops automatic expression cycling;
tap an eye to select another expression. The spiral keeps rotating until you
select another expression. Mood changes ease at about 4.5/s in both modes.

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
