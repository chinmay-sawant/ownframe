# Showcase

Choose an example in the left navigation to view its screenshots or demo.

## Telegram

`examples/telegram` is the phone-first example. These Pixel 7 screens use the
same HTML template that `go run ./examples/telegram` opens on desktop, bound to
Android with `ebitenmobile`. `sh scripts/android.sh install` binds it, builds a
debug APK with Gradle, and installs it over `adb`. See the
[Android setup guide](examples/telegram/android/README.md).

- ![Chat list with search, unread badges, pinned and muted chats](assets/screenshots/android/telegram/telegram-chats.webp)

  **Chats**
- ![Conversation with bubbles, timestamps, read ticks, and a composer](assets/screenshots/android/telegram/telegram-thread.webp)

  **Conversation**
- ![Contacts tab with last-seen times](assets/screenshots/android/telegram/telegram-contacts.webp)

  **Contacts**
- ![Settings tab with a dark mode toggle](assets/screenshots/android/telegram/telegram-settings.webp)

  **Settings**
- ![Settings in the dark theme](assets/screenshots/android/telegram/telegram-settings-dark.webp)

  **Settings, dark**
- ![Chat list in the dark theme](assets/screenshots/android/telegram/telegram-chats-dark.webp)

  **Chats, dark**
- ![Conversation in the dark theme](assets/screenshots/android/telegram/telegram-thread-dark.webp)

  **Conversation, dark**

Taps, tabs, opening a chat, drag-to-scroll, and the soft keyboard work on the
phone. The system Back gesture returns from a chat. The top bar and composer
stay pinned above the scrolling messages. The arm64 debug APK is 29 MB; binding
all four ABIs puts one APK at 188 MB.

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/telegram) ·
[Download arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/telegram/android/artifacts/telegram-arm64-release.apk)

## Teams

Teams is a native chat and collaboration app built with Go, HTML, and CSS.
[Watch the Teams demo](https://x.com/chinmay_sawant_/status/2107548496194895993/video/1).

- ![Animated preview of the Teams chat and collaboration app](assets/screenshots/demos/teams.webp)

  **Teams**

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/teams)

## Spotify Player

The Spotify Player is a Go desktop app. ownframe renders its HTML and CSS with
Blinkless, and Ebitengine handles the native window and bundled audio. It uses
no WebView. Phone apps use ownframe's Ebitengine Mobile binding; this example
does not have a mobile wrapper. Its `-web` mode serves a picture of the page
over HTTP.

- ![Animated Spotify Player demo showing its Home, Browse, Profile, and Library screens](frontend/public/demos/videos/spotify.webp)

  **Spotify Player**

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/spotify-player)

## Pet Care

Pet Care is a cross-platform example. The Android captures are from a Pixel 7,
and the desktop capture shows the same app in its native Windows window.

- ![Pet Care login screen in the light theme](assets/screenshots/android/pet-care/pet-care-login.webp)

  **Login, light**
- ![Pet Care home screen after sign-in](assets/screenshots/android/pet-care/pet-care-home.webp)

  **Today**
- ![Pet Care settings screen in the dark theme](assets/screenshots/android/pet-care/pet-care-settings-dark.webp)

  **Settings, dark**
- ![Pet Care desktop application showing the Today screen](assets/screenshots/windows/pet-care/pet-care.webp)

  **Pet Care Windows**

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/pet-care)

## Digital Wellbeing

Mochi is a Pixel 7 Android example that turns Android Digital Wellbeing usage
stats into a gentle dashboard. Browse a week, tap a day for its breakdown, see
weekly insights and weekday/weekend totals, or choose an app to focus the chart.
Usage data comes from Android's Usage Access permission and stays on the phone.

- ![Mochi dashboard with a seven-day usage chart and daily summary](assets/screenshots/digital-wellbeing/dashboard.webp)

  **Dashboard**
- ![Mochi dashboard focused on a selected day](assets/screenshots/digital-wellbeing/day-detail.webp)

  **Selected day**
- ![Weekly Insights with most-used apps and daily average](assets/screenshots/digital-wellbeing/weekly-insights.webp)

  **Weekly Insights**
- ![All Apps list with application icons and usage totals](assets/screenshots/digital-wellbeing/all-apps.webp)

  **All apps**
- ![Choose an app list sorted by usage time](assets/screenshots/digital-wellbeing/choose-app.webp)

  **Choose an app**

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/digital-wellbeing)
[Download arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/digital-wellbeing/android/artifacts/digital-wellbeing-arm64-release.apk)

## Desktop overlays

The desktop cat stays above other windows and lets clicks pass through its
transparent pixels.

- ![Desktop cat overlay preview](assets/screenshots/demos/desktop-cat.webp)

  **Desktop cat overlay**
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106829998409789929)

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/desktop-cat)

## Wispr Flow

The Wispr Flow example is an HTML and CSS interface rendered by ownframe.

- ![Wispr Flow example preview](assets/screenshots/demos/wispr-flow.webp)

  **Wispr Flow**
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106788230154871126)

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/wispr-flow-dashboard)

## Dino Run

An endless runner built with HTML, CSS, and Go.

- ![Dino Run example preview](assets/screenshots/demos/dino.webp)

  **Dino Run**

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/dino) ·
[Download arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/dino/android/artifacts/dino-arm64-release.apk)

## Eyes

`examples/eyes` is a cute robot-eyes demo: two big boxy rounded eyes on
black that blink, look around, breathe, and cycle moods. The same page runs
on desktop with `go run ./examples/eyes` and on Android through
`ebitenmobile`. `sh scripts/android-eyes.sh install` binds it, builds a debug
APK with Gradle, and installs it over `adb`.

- ![Idle robot eyes on a black phone screen](assets/screenshots/android/eyes/eyes-idle.webp)

  **Idle**
- ![Curious robot eyes with one eye stretched tall](assets/screenshots/android/eyes/eyes-curious.webp)

  **Curious**
- ![Sleepy robot eyes half closed into cups](assets/screenshots/android/eyes/eyes-sleepy.webp)

  **Sleepy**
- ![Surprised robot eyes wide open](assets/screenshots/android/eyes/eyes-surprised.webp)

  **Surprised**
- ![Animated eyes demo blinking and cycling moods](assets/screenshots/android/eyes/eyes-60fps.webp)

  **Animated demo**

Taps change the mood, swipes look around, keys 1-5 pick a mood, and the
phone build auto-cycles a mood every 8 seconds with random glances and
blinks.

[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/eyes)
