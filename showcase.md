# Showcase

Choose an example in the left navigation to view its screenshots or demo.

## Telegram

`examples/telegram` is the phone-first example. These Pixel 7 screens use the
same HTML template that `go run ./examples/telegram` opens on desktop, bound to
Android with `ebitenmobile`. `sh scripts/android.sh install` binds it, builds a
debug APK with Gradle, and installs it over `adb`. See the
[Android setup guide](examples/telegram/android/README.md).

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-chats.webp" alt="Chat list with search, unread badges, pinned and muted chats" width="250" height="600" /><br /><b>Chats</b></td>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-thread.webp" alt="Conversation with bubbles, timestamps, read ticks, and a composer" width="250" height="600" /><br /><b>Conversation</b></td>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-contacts.webp" alt="Contacts tab with last-seen times" width="250" height="600" /><br /><b>Contacts</b></td>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-settings.webp" alt="Settings tab with a dark mode toggle" width="250" height="600" /><br /><b>Settings</b></td>
  </tr>
  <tr>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-settings-dark.webp" alt="Settings in the dark theme" width="250" height="600" /><br /><b>Settings, dark</b></td>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-chats-dark.webp" alt="Chat list in the dark theme" width="250" height="600" /><br /><b>Chats, dark</b></td>
    <td align="center"><img src="assets/screenshots/android/telegram/telegram-thread-dark.webp" alt="Conversation in the dark theme" width="250" height="600" /><br /><b>Conversation, dark</b></td>
    <td></td>
  </tr>
</table>

Taps, tabs, opening a chat, drag-to-scroll, and the soft keyboard work on the
phone. The system Back gesture returns from a chat. The top bar and composer
stay pinned above the scrolling messages. The arm64 debug APK is 29 MB; binding
all four ABIs puts one APK at 188 MB.

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/telegram) ·
Build the arm64 release APK with `scripts/android-release.sh telegram`.

## Teams

Teams is a native chat and collaboration app built with Go, HTML, and CSS.
[Watch the Teams demo](https://x.com/chinmay_sawant_/status/2107548496194895993/video/1).

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/demos/teams.webp" alt="Animated preview of the Teams chat and collaboration app" width="600" height="250" /><br /><b>Teams</b></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/teams) ·
[Watch the local demo recording](assets/screenshots/demos/teams.webp)

## Spotify Player

The Spotify Player is a Go desktop app. ownframe renders its HTML and CSS with
Blinkless, and Ebitengine handles the native window and bundled audio. It uses
no WebView. Phone apps use ownframe's Ebitengine Mobile binding; this example
does not have a mobile wrapper. Its `-web` mode serves a picture of the page
over HTTP.

<table>
  <tr>
    <td align="center"><img src="frontend/public/demos/videos/spotify.webp" alt="Animated Spotify Player demo showing its Home, Browse, Profile, and Library screens" width="600" height="250" /><br /><b>Spotify Player</b></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/spotify-player)

## Pet Care

Pet Care is a cross-platform example. The Android captures are from a Pixel 7,
and the desktop capture shows the same app in its native Windows window.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/android/pet-care/pet-care-login.webp" alt="Pet Care login screen in the light theme" width="250" height="600" /><br /><b>Login, light</b></td>
    <td align="center"><img src="assets/screenshots/android/pet-care/pet-care-home.webp" alt="Pet Care home screen after sign-in" width="250" height="600" /><br /><b>Today</b></td>
    <td align="center"><img src="assets/screenshots/android/pet-care/pet-care-settings-dark.webp" alt="Pet Care settings screen in the dark theme" width="250" height="600" /><br /><b>Settings, dark</b></td>
  </tr>
  <tr>
    <td align="center" colspan="3"><img src="assets/screenshots/windows/pet-care/pet-care.webp" alt="Pet Care desktop application showing the Today screen" width="600" height="250" /><br /><b>Pet Care Windows</b></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/pet-care)

## Digital Wellbeing

Mochi is a Pixel 7 Android example that turns Android Digital Wellbeing usage
stats into a gentle dashboard. Browse a week, tap a day for its breakdown, see
weekly insights and weekday/weekend totals, or choose an app to focus the chart.
Usage data comes from Android's Usage Access permission and stays on the phone.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/digital-wellbeing/dashboard.webp" alt="Mochi dashboard with a seven-day usage chart and daily summary" width="250" height="600" /><br /><b>Dashboard</b></td>
    <td align="center"><img src="assets/screenshots/digital-wellbeing/day-detail.webp" alt="Mochi dashboard focused on a selected day" width="250" height="600" /><br /><b>Selected day</b></td>
    <td align="center"><img src="assets/screenshots/digital-wellbeing/weekly-insights.webp" alt="Weekly Insights with most-used apps and daily average" width="250" height="600" /><br /><b>Weekly Insights</b></td>
  </tr>
  <tr>
    <td align="center"><img src="assets/screenshots/digital-wellbeing/all-apps.webp" alt="All Apps list with application icons and usage totals" width="250" height="600" /><br /><b>All apps</b></td>
    <td align="center"><img src="assets/screenshots/digital-wellbeing/choose-app.webp" alt="Choose an app list sorted by usage time" width="250" height="600" /><br /><b>Choose an app</b></td>
    <td></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/digital-wellbeing) ·
Build the APK with `examples/digital-wellbeing/android/build.sh`.

## Desktop overlays

The desktop cat stays above other windows and lets clicks pass through its
transparent pixels.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/demos/desktop-cat.webp" alt="Desktop cat overlay preview" width="600" height="250" /><br /><b>Desktop cat overlay</b></td>
  </tr>
</table>
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106829998409789929)

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/desktop-cat)

## Wispr Flow

The Wispr Flow example is an HTML and CSS interface rendered by ownframe.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/demos/wispr-flow.webp" alt="Wispr Flow example preview" width="600" height="250" /><br /><b>Wispr Flow</b></td>
  </tr>
</table>
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106788230154871126)

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/wispr-flow-dashboard)

## Dino Run

An endless runner built with HTML, CSS, and Go.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/demos/dino.webp" alt="Dino Run example preview" width="600" height="250" /><br /><b>Dino Run</b></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/dino) ·
Build the arm64 release APK with `scripts/android-release.sh dino`.

## Eyes

`examples/eyes` is a cute robot-eyes demo: two big boxy rounded eyes on
black that blink, look around, breathe, and cycle moods. The same page runs
on desktop with `go run ./examples/eyes` and on Android through
`ebitenmobile`. `sh scripts/android-eyes.sh install` binds it, builds a debug
APK with Gradle, and installs it over `adb`.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/android/eyes/eyes-idle.webp" alt="Idle robot eyes on a black phone screen" width="600" height="250" /><br /><b>Idle</b></td>
    <td align="center"><img src="assets/screenshots/android/eyes/eyes-curious.webp" alt="Curious robot eyes with one eye stretched tall" width="600" height="250" /><br /><b>Curious</b></td>
    <td align="center"><img src="assets/screenshots/android/eyes/eyes-sleepy.webp" alt="Sleepy robot eyes half closed into cups" width="600" height="250" /><br /><b>Sleepy</b></td>
    <td align="center"><img src="assets/screenshots/android/eyes/eyes-surprised.webp" alt="Surprised robot eyes wide open" width="600" height="250" /><br /><b>Surprised</b></td>
  </tr>
  <tr>
    <td align="center" colspan="4"><img src="assets/screenshots/android/eyes/eyes2-60fps.webp" alt="Second animated eyes demo blinking and cycling moods" width="600" height="250" /><br /><b>Animated demo 2</b></td>
  </tr>
  <tr>
    <td align="center" colspan="4"><img src="assets/screenshots/android/eyes/eyes-60fps.webp" alt="First animated eyes demo blinking and cycling moods" width="600" height="250" /><br /><b>Animated demo 1</b></td>
  </tr>
</table>

Taps change the mood, swipes look around, keys 1-5 pick a mood, and the
phone build auto-cycles a mood every 8 seconds with random glances and
blinks.

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/eyes)

[Download arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/eyes/android/artifacts/eyes-arm64-release.apk)

## Depth Alert

Depth Alert is an experimental Android app that reads ARCore depth from the
center of the camera view and maps it to haptic signals. WALK requires twelve
consecutive clear observations. A nearby surface reports STOP, and a surface
within 1.2 m reports STOP NOW. It is a prototype, not a mobility aid.

<table>
  <tr>
    <td align="center"><img src="assets/screenshots/android/depth-alert/depth-alert-pixel7.webp" alt="Depth Alert running on a Pixel 7 with its live camera preview, current signal, distance, and depth coverage" width="250" height="600" /><br /><b>Depth Alert on Pixel 7</b></td>
  </tr>
</table>

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/depth-alert)

[Download signed arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/depth-alert/android/artifacts/depth-alert-arm64-release.apk) ·
[Build and safety notes](examples/depth-alert/README.md)
