export const showcaseIntro = 'Choose an example in the left navigation to view its screenshots or demo.';

const leadingShowcaseIds = ['wispr-flow', 'code-view', 'spotify-player', 'teams', 'telegram', 'desktop-overlays', 'eyes'];

export const showcaseSections = [
  {
    id: 'code-view',
    title: 'Code Review',
    layout: 'single',
    before: '`examples/code-view` is the flagship desktop application of ownframe. It is a native multi-pane code viewer with syntax highlighting in pure Go, a visual git commit graph, and word-level diff pairing with 26 ns idle ticks and zero webview overhead.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/code-view)',
    media: [
      { src: 'frontend/public/demos/code-view.png', alt: 'Code Review dark multi-pane editor with pure Go syntax highlighting, git graph, and diffs', caption: 'Code Review', width: 1500, height: 900, nativeDisplay: true },
    ],
  },
  {
    id: 'telegram',
    title: 'Telegram',
    layout: 'grid-4',
    before: '`examples/telegram` is the phone-first example. These Pixel 7 screens use the same HTML template that `go run ./examples/telegram` opens on desktop, bound to Android with `ebitenmobile`. `sh scripts/android.sh install` binds it, builds a debug APK with Gradle, and installs it over `adb`. See the [Android setup guide](examples/telegram/android/README.md).',
    after: 'Taps, tabs, opening a chat, drag-to-scroll, and the soft keyboard work on the phone. The system Back gesture returns from a chat. The top bar and composer stay pinned above the scrolling messages. The arm64 debug APK is 29 MB; binding all four ABIs puts one APK at 188 MB.\n\n[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/telegram)\u00a0·\u00a0Build the arm64 release APK with `scripts/android-release.sh telegram`.',
    media: [
      { src: 'assets/screenshots/android/telegram/telegram-chats.webp', alt: 'Chat list with search, unread badges, pinned and muted chats', caption: 'Chats', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-thread.webp', alt: 'Conversation with bubbles, timestamps, read ticks, and a composer', caption: 'Conversation', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-contacts.webp', alt: 'Contacts tab with last-seen times', caption: 'Contacts', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-settings.webp', alt: 'Settings tab with a dark mode toggle', caption: 'Settings', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-settings-dark.webp', alt: 'Settings in the dark theme', caption: 'Settings, dark', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-chats-dark.webp', alt: 'Chat list in the dark theme', caption: 'Chats, dark', width: 540, height: 1095 },
      { src: 'assets/screenshots/android/telegram/telegram-thread-dark.webp', alt: 'Conversation in the dark theme', caption: 'Conversation, dark', width: 540, height: 1095 },
    ],
  },
  {
    id: 'teams',
    title: 'Teams',
    layout: 'single',
    before: 'Teams is a native chat and collaboration app built with Go, HTML, and CSS. [Watch the Teams demo](https://x.com/chinmay_sawant_/status/2107548496194895993/video/1).',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/teams)\u00a0·\u00a0[Watch the local demo recording](assets/screenshots/demos/teams.webp)',
    media: [
      { src: 'assets/screenshots/demos/teams.webp', video: 'demos/teams.mp4', alt: 'Animated preview of the Teams chat and collaboration app', caption: 'Teams', width: 720, height: 402, nativeDisplay: true },
    ],
  },
  {
    id: 'spotify-player',
    title: 'Spotify Player',
    layout: 'single',
    before: 'The Spotify Player is a Go desktop app. ownframe renders its HTML and CSS with Blinkless, and Ebitengine handles the native window and bundled audio. It uses no WebView. Phone apps use ownframe\'s Ebitengine Mobile binding; this example does not have a mobile wrapper. Its `-web` mode serves a picture of the page over HTTP.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/spotify-player)',
    media: [
      { src: 'frontend/public/demos/videos/spotify.webp', alt: 'Animated Spotify Player demo showing its Home, Browse, Profile, and Library screens', caption: 'Spotify Player', width: 780, height: 520, nativeDisplay: true },
    ],
  },
  {
    id: 'pet-care',
    title: 'Pet Care',
    layout: 'grid-3',
    before: 'Pet Care is a cross-platform example. The Android captures are from a Pixel 7, and the desktop capture shows the same app in its native Windows window.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/pet-care)',
    media: [
      { src: 'assets/screenshots/android/pet-care/pet-care-login.webp', alt: 'Pet Care login screen in the light theme', caption: 'Login, light', width: 1080, height: 2400 },
      { src: 'assets/screenshots/android/pet-care/pet-care-home.webp', alt: 'Pet Care home screen after sign-in', caption: 'Today', width: 1080, height: 2400 },
      { src: 'assets/screenshots/android/pet-care/pet-care-settings-dark.webp', alt: 'Pet Care settings screen in the dark theme', caption: 'Settings, dark', width: 1080, height: 2400 },
      { src: 'assets/screenshots/windows/pet-care/pet-care.webp', alt: 'Pet Care desktop application showing the Today screen', caption: 'Pet Care Windows', width: 1882, height: 982, wide: true },
    ],
  },
  {
    id: 'digital-wellbeing',
    title: 'Digital Wellbeing',
    layout: 'grid-3',
    before: 'Mochi is a Pixel 7 Android example that turns Android Digital Wellbeing usage stats into a gentle dashboard. Browse a week, tap a day for its breakdown, see weekly insights and weekday/weekend totals, or choose an app to focus the chart. Usage data comes from Android\'s Usage Access permission and stays on the phone.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/digital-wellbeing)\u00a0·\u00a0Build the APK with `examples/digital-wellbeing/android/build.sh`.',
    media: [
      { src: 'assets/screenshots/digital-wellbeing/dashboard.webp', alt: 'Mochi dashboard with a seven-day usage chart and daily summary', caption: 'Dashboard', width: 1080, height: 2400 },
      { src: 'assets/screenshots/digital-wellbeing/day-detail.webp', alt: 'Mochi dashboard focused on a selected day', caption: 'Selected day', width: 1080, height: 2400 },
      { src: 'assets/screenshots/digital-wellbeing/weekly-insights.webp', alt: 'Weekly Insights with most-used apps and daily average', caption: 'Weekly Insights', width: 1080, height: 2400 },
      { src: 'assets/screenshots/digital-wellbeing/all-apps.webp', alt: 'All Apps list with application icons and usage totals', caption: 'All apps', width: 1080, height: 2400 },
      { src: 'assets/screenshots/digital-wellbeing/choose-app.webp', alt: 'Choose an app list sorted by usage time', caption: 'Choose an app', width: 1080, height: 2400 },
    ],
  },
  {
    id: 'desktop-overlays',
    title: 'Desktop overlays',
    layout: 'single',
    before: 'The desktop cat stays above other windows and lets clicks pass through its transparent pixels.',
    after: '[Watch the demo](https://x.com/chinmay_sawant_/status/2106829998409789929)\n\n[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/desktop-cat)',
    media: [
      { src: 'assets/screenshots/demos/desktop-cat.webp', video: 'demos/desktop-cat.mp4', alt: 'Desktop cat overlay preview', caption: 'Desktop cat overlay', width: 880, height: 476, nativeDisplay: true },
    ],
  },
  {
    id: 'wispr-flow',
    title: 'Wispr Flow',
    layout: 'single',
    before: 'The Wispr Flow example is an HTML and CSS interface rendered by ownframe.',
    after: '[Try the live WebAssembly demo](frontend/public/demos/wispr-flow/index.html) (52 MB download)\n\n[Watch the demo](https://x.com/chinmay_sawant_/status/2106788230154871126)\n\n[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/wispr-flow-dashboard)',
    media: [
      { src: 'assets/screenshots/demos/wispr-flow.webp', video: 'demos/preview.mp4', autoplayPreview: true, alt: 'Wispr Flow example preview', caption: 'Wispr Flow', width: 1906, height: 1030 },
    ],
  },
  {
    id: 'dino-run',
    title: 'Dino Run',
    layout: 'single',
    before: 'An endless runner built with HTML, CSS, and Go.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/dino)\u00a0·\u00a0Build the arm64 release APK with `scripts/android-release.sh dino`.',
    media: [
      { src: 'assets/screenshots/demos/dino.webp', alt: 'Dino Run example preview', caption: 'Dino Run', width: 1888, height: 1074 },
    ],
  },
  {
    id: 'eyes',
    title: 'Eyes',
    layout: 'grid-4',
    before: '`examples/eyes` is a cute robot-eyes demo: two big boxy rounded eyes on black that blink, look around, breathe, and cycle moods. The same page runs on desktop with `go run ./examples/eyes` and on Android through `ebitenmobile`. `sh scripts/android-eyes.sh install` binds it, builds a debug APK with Gradle, and installs it over `adb`.',
    after: 'Taps change the mood, swipes look around, keys 1-5 pick a mood, and the phone build auto-cycles a mood every 8 seconds with random glances and blinks.\n\n[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/eyes)\n\n[Download arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/eyes/android/artifacts/eyes-arm64-release.apk)',
    media: [
      { src: 'assets/screenshots/android/eyes/eyes-idle.webp', alt: 'Idle robot eyes on a black phone screen', caption: 'Idle', width: 2400, height: 1080 },
      { src: 'assets/screenshots/android/eyes/eyes-curious.webp', alt: 'Curious robot eyes with one eye stretched tall', caption: 'Curious', width: 2400, height: 1080 },
      { src: 'assets/screenshots/android/eyes/eyes-sleepy.webp', alt: 'Sleepy robot eyes half closed into cups', caption: 'Sleepy', width: 2400, height: 1080 },
      { src: 'assets/screenshots/android/eyes/eyes-surprised.webp', alt: 'Surprised robot eyes wide open', caption: 'Surprised', width: 2400, height: 1080 },
      { src: 'assets/screenshots/android/eyes/eyes2-60fps.webp', alt: 'Second animated eyes demo blinking and cycling moods', caption: 'Animated demo 2', width: 1066, height: 480, wide: true },
      { src: 'assets/screenshots/android/eyes/eyes-60fps.webp', alt: 'First animated eyes demo blinking and cycling moods', caption: 'Animated demo 1', width: 1066, height: 480, wide: true },
    ],
  },
  {
    id: 'depth-alert',
    title: 'Depth Alert',
    layout: 'portrait',
    before: 'Depth Alert is an experimental Android app that reads ARCore depth from the center of the camera view and maps it to haptic signals. WALK requires twelve consecutive clear observations. A nearby surface reports STOP, and a surface within 1.2 m reports STOP NOW. It is a prototype, not a mobility aid.',
    after: '[Source code](https://github.com/chinmay-sawant/ownframe/tree/master/examples/depth-alert)\n\n[Download signed arm64 APK](https://github.com/chinmay-sawant/ownframe/raw/refs/heads/master/examples/depth-alert/android/artifacts/depth-alert-arm64-release.apk)\u00a0·\u00a0[Build and safety notes](examples/depth-alert/README.md)',
    media: [
      { src: 'assets/screenshots/android/depth-alert/depth-alert-pixel7.webp', alt: 'Depth Alert running on a Pixel 7 with its live camera preview, current signal, distance, and depth coverage', caption: 'Depth Alert on Pixel 7', width: 1080, height: 2400 },
    ],
  },
].sort((left, right) => {
  const leftOrder = leadingShowcaseIds.indexOf(left.id);
  const rightOrder = leadingShowcaseIds.indexOf(right.id);
  if (leftOrder === -1) return rightOrder === -1 ? 0 : 1;
  if (rightOrder === -1) return -1;
  return leftOrder - rightOrder;
});
