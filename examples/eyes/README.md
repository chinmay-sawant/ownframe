# Eyes demo

Robot eyes on black with 21 selectable expressions: open, squint, angry,
spiral, relaxed, sleepy, surprised, curious, left wink, right wink, bored,
happy, worried, focused, excited, shy, nervous, relieved, sad, scared, and
frustrated. The spiral also reacts to device motion on Android.

```sh
go run ./examples/eyes          # desktop window, landscape 960x540
go run ./examples/eyes -web     # browser on 127.0.0.1:8140
```

The AUTO button cycles through all 21 expressions. Each holds for 8 seconds;
spiral holds for 4 seconds before it fades into the next expression. Tap the
left half to go back and the right half to go forward. A double tap on either
half selects the same direction once. Drag to look, keys 1-4 pick the original
four, and S starts the spiral. Shaking the phone starts a short spiral too.
Both modes keep the eyes animated: they blink, breathe, move their gaze, and
finish each expression animation. The eyes keep their rounded shape as the lids
close and reopen. Wink expressions close the viewer's named eye. STATIC stops
automatic expression cycling;
tap either half to select another expression. The spiral keeps rotating until
you select another expression. Mood changes ease at about 4.5/s in both modes.

The mood label is centered. The DEV and AUTO/STATIC controls fade after
5 seconds without input, then return when you interact.
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
