# Eyes demo

Cute robot eyes like the Grok bot and the OpenAI dot: two big boxy
rounded eyes on black that blink, look around, breathe, and cycle
moods (idle, happy, sleepy, surprised, curious).

```sh
go run ./examples/eyes          # desktop window, landscape 960x540
go run ./examples/eyes -web     # browser on 127.0.0.1:8140
```

Click to change mood, drag to look, keys 1-5 pick a mood, space
blinks, arrows look. The phone build auto-cycles a mood every 8s
with random saccades every 1.4-3s and blinks every 2.2-4.8s.
Mood squints ease at ~4.5/s while blinks stay snappy, so both look
smooth at 60 Hz and 90 Hz. Squints narrow onto the bottom lid instead of
widening, and happy shapes a dome while sleepy shapes a cup, so the eyes
never pinch into a fish. Corners always clamp inside the box.

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
