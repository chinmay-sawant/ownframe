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

## Teams

Teams is a native chat and collaboration app built with Go, HTML, and CSS.
[Watch the Teams demo](https://x.com/chinmay_sawant_/status/2107548496194895993/video/1).

- ![Preview of the Teams chat and collaboration app](assets/screenshots/demos/teams.webp)

  **Teams**

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

## Desktop overlays

The desktop cat stays above other windows and lets clicks pass through its
transparent pixels.

- ![Desktop cat overlay preview](assets/screenshots/demos/desktop-cat.webp)

  **Desktop cat overlay**
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106829998409789929)

## Wispr Flow

The Wispr Flow example is an HTML and CSS interface rendered by ownframe.

- ![Wispr Flow example preview](assets/screenshots/demos/wispr-flow.webp)

  **Wispr Flow**
  [Watch the demo](https://x.com/chinmay_sawant_/status/2106788230154871126)

## Dino Run

An endless runner built with HTML, CSS, and Go.

- ![Dino Run example preview](assets/screenshots/demos/dino.webp)

  **Dino Run**
  [View the example source](examples/dino)
