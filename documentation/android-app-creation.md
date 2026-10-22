# Create an Android app from an ownframe example

Use this guide when you add an Android app for an ownframe page. The shared
Java helpers live in `examples/android-host`. App navigation and data stay in
the Go example.

## 1.0 native Android

Android `BindMobile` is a 1.0 native host: one compiled Go process inside an
Ebiten Activity, no Chromium, no WebView. That is not OS widgets and not
TalkBack.

- `ownframe.BindMobile` registers the page. Call it from the `mobile/`
  package that `ebitenmobile bind` compiles.
- `examples/android-host` owns insets, Back, lifecycle (`onResume` /
  `onPause`), and `TouchCancellation`.
- Unset `GOWORK` for the bind. Gomobile copies packages into a temporary
  module.
- `scripts/android-release.sh` binds, assembles, and signs an arm64 APK.
- The APK index is [examples/android-host/APPS.md](../examples/android-host/APPS.md):
  login, platform, dino, lg-remote, telegram, eyes.

1.0 limits on this host:

- Clipboard is process memory ([clipboard.md](clipboard.md)).
- There is no file picker dialog. Telegram camera and gallery are app Java,
  not `internal/filepick`.
- Idle park is off (`internal/window/idle_supported_other.go`).
- There is no TalkBack node. Process death does not restore page state.
- IME is `exp/textinput`.
- The APK is a large `.so` inside the package.

Play Store restore and TalkBack stay out of 1.0. Telegram and eyes are the
exhibits ([examples/readme.md](../examples/readme.md)). The remaining Play
gaps (TalkBack, process death, IME contract, `GOWORK` unset for bind,
signing) are listed in [play-store.md](play-store.md).

## Set up the mobile Go package

Create a `mobile` package that constructs the app and binds its page:

```go
func start(page *ownframe.Page) error {
    return ownframe.BindMobile(context.Background(), page)
}
```

Export small functions from the mobile package for Android callbacks. Keep
Android callbacks short. They run on Android's UI thread, so queue page or
application changes and apply them from `Page.SetTick`.

Run `ebitenmobile bind` from the repository root with `GOWORK` unset. The bind
tool copies packages into a temporary module and does not use the workspace.
Write the generated AAR under the example's Android project. Do not commit it.

## Connect the shared Android host

Add `:android-host` to the Android project's Gradle settings and depend on it
from the application module. Use `examples/pet-care/android` as a working
project layout. Add
`android:enableOnBackInvokedCallback="true"` to the application manifest.

Create one `EbitenView` and attach both shared helpers:

```java
game = new EbitenView(this);
touches = TouchCancellation.attach(game, Mobile::cancelTouches);
root.addView(game, new FrameLayout.LayoutParams(
    ViewGroup.LayoutParams.MATCH_PARENT,
    ViewGroup.LayoutParams.MATCH_PARENT));
setContentView(root);
host = AndroidHost.attach(this, root,
    Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING).build(),
    new Callbacks() {
        @Override public boolean onBackPressed() { return Mobile.back(); }
        @Override public void onResume() { game.resumeGame(); }
        @Override public void onPause() {
            touches.cancel();
            game.suspendGame();
        }
    });
```

Call `host.onBackPressed()` from the Activity's legacy Back override. On Android
13 and later, `AndroidHost` registers predictive Back. Return `false` when the
app has no screen to pop. The host then closes the Activity.

Queue Back in Go instead of changing page state from the Android callback.
Return whether the current screen can handle Back, store the request safely,
then update and redraw the page on its game-loop tick. Handle the keyboard
before screen navigation; `AndroidHost` hides the IME after
`onImeBackPressed` runs.

`TouchCancellation` handles Android `ACTION_CANCEL`, which Ebitengine v2.10.4
does not pass through as a pointer release. Call `touches.cancel()` before
suspending the game. Call `touches.close()` and `host.close()` when the Activity
is destroyed.

## Choose an inset policy

Choose one owner for safe-area padding:

- Use `InsetMode.HOST_PADDING` when `AndroidHost` should pad the root view.
- Use `InsetMode.APP_CALLBACK` when the app applies `Callbacks.onInsetsChanged`
  to its own layout.

Do not apply both policies. `Insets` reports status bars, navigation bars,
display cutouts, and the IME in dp. `ViewportSnapshot` reports the measured
view size and the current display configuration. Read
[Android runtime](android-runtime.md) for coordinate conversion, rotation, and
IME behavior.

## Lay out mobile screen headers

For an HTML/CSS screen with Back navigation, keep Back in a fixed leading slot
and center a short title across the content width. Reserve a matching trailing
slot so the Back control does not push the title off center. Use one shared
header class for screens with the same layout. If a long or contextual title
belongs beside Back, make that a deliberate variant instead of relying on
`justify-content: space-between` with only two items.

```html
<div class="title-row back-title-row centered-title-row">
  <button class="back-link" data-action="back">
    <span data-action="back">Back</span>
  </button>
  <h1>All apps</h1>
  <span class="title-spacer" aria-hidden="true"></span>
</div>
```

Give the Back control a touch target of at least 48 dp. Keep the title and
control vertically aligned, preserve the Back action on nested text elements,
and check the rendered header on a device after changing its CSS. Long titles
need room at narrow phone widths, so test the centered and contextual variants
at the smallest supported width.

## Keep the system bars in the active theme

Call `host.setSystemBarStyle(color, lightIcons)` on Android's UI thread when the
page theme changes. Pass the page background color and `true` for light icons
on a dark background. The host sets both bar colors, icon contrast, and Android
contrast enforcement.

## Use touch scrolling and pinch zoom

OwnFrame scrolls an overflowing page when the user drags it. Desktop users can
also use the mouse wheel. `ownframe.ScrollbarAuto` hides thumbs on Android and iOS. Set
`Scrollbar: ownframe.ScrollbarVisible` in the page config to show them, or use
`ownframe.ScrollbarHidden` to hide them on every platform. Hidden thumbs do not
disable touch scrolling. Use `Page.ScrollTo` or `Page.ScrollBy` when an app
needs to move to a specific point.

Pinch zoom is enabled by default. Two fingers change the page zoom from 0.25 to
4. Set `Config.DisableTouchZoom` or call `Page.Testing().SetTouchZoomAllowed(false)` when
the app needs two-finger gestures for another action. Disabling pinch zoom does
not disable touch scrolling. Read [scrolling](scrolling.md) and
[interaction](interaction.md) for gesture thresholds and scroll behavior.

## Build and install the APK

Use the app's build script to bind Go, assemble the debug APK, and copy it to
`temp`. Follow `scripts/pet-care-android.sh` when you add a new script. Pass
`install` only when the script also has access to ADB:

```sh
scripts/<app>-android.sh
adb install -r temp/<app>/<app>-debug.apk
```

Set `ANDROID_TARGET=android/arm64` to bind only for a 64-bit ARM device. Keep
`GOWORK` unset for the bind step. The
[Android build guide](android-build.md) covers toolchain setup and device
checks.
