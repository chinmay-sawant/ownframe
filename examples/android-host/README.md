# ownframe Android host

This Android library holds the activity behavior shared by ownframe examples. It targets API 23 and compiles against API 34. The app keeps its own navigation, permissions, camera and Bluetooth flows.

## Add the module

Include `:android-host` in the Android project's `settings.gradle`, point it to this directory, and add `implementation project(':android-host')` to the app module. The project must use Android Gradle Plugin 8.5.1 or newer and Java 17.

Set `android:enableOnBackInvokedCallback="true"` on the application's `<application>` element. This enables the platform Back callback that the host registers on Android 13 and later.

## Attach it

Attach one host to the root content view in `Activity.onCreate`. The default uses edge-to-edge layout and applies safe-area padding to that view.

```java
Options options = Options.builder()
        .edgeToEdge(true)
        .insetMode(InsetMode.HOST_PADDING)
        .baseTextSizeSp(16f)
        .build();

AndroidHost host = AndroidHost.attach(this, rootView, options, new Callbacks() {
    @Override public void onInsetsChanged(Insets value) {
        // Values are dp. Apply them in app layout only with APP_CALLBACK mode.
    }

    @Override public void onConfigurationChanged(DisplayInfo value) {
        // density is the native dp scale; baseTextSizeDp includes Android text scaling.
    }

    @Override public void onViewportChanged(ViewportSnapshot value) {
        // Width and height are the supplied View's measured bounds in dp.
        // Revisions increase when any reported viewport or display value changes.
    }

    @Override public boolean onBackPressed() {
        return closeAppPanelIfOpen();
    }

    @Override public void onImeBackPressed() {
        // Clear field focus before the host hides the keyboard.
    }

    @Override public void onResume() { resumeAppWork(); }
    @Override public void onPause() { pauseAppWork(); }
});
```

For app-owned safe-area layout, select `InsetMode.APP_CALLBACK`. The host still reports left, top, right, and bottom safe insets, IME visibility, and the IME bottom in dp, but it leaves view padding alone. Never apply both policies to the same content.

`onViewportChanged` publishes an immutable snapshot with measured content width and height, all four safe insets, IME visibility and bottom space, density, font scale, Android-scaled base text size, and a revision. Dimensions describe the supplied view's measured bounds. Insets are separate values, so an app can apply its selected inset policy exactly once. The first snapshot can have zero dimensions before Android measures the view. The revision starts at 1 and advances only when a reported value changes.

Call `host.onBackPressed()` from `Activity.onBackPressed()` on API 23 through 32. If it returns false, call `super.onBackPressed()`. API 33 and later use the predictive Back dispatcher; the helper handles the keyboard first, then calls the app hook, then finishes the activity when the app does not consume Back. Call `host.close()` when replacing the root or ending the activity to detach listeners and restore its original padding.

`baseTextSizeSp` is converted with Android's SP conversion and then to dp once. It follows nonlinear scaling where the OS supports it. Use the reported size as a text size; do not multiply it by `fontScale` again. Insets are native pixels converted to dp once. Host padding itself stays in native pixels.

The helper configures edge-to-edge flags but does not choose orientation,
permissions, game viewport, or app-specific gestures. API 23 through 32 use the
Activity's legacy Back override; check keyboard and inset behavior on the
minimum supported OS and on device-specific Android builds.

## Match system bars to the page theme

Call `setSystemBarStyle` on the main thread when the page theme changes. Pass
the background color and whether the bars need light icons:

```java
host.setSystemBarStyle(Color.parseColor(dark ? "#171c19" : "#f4f6f2"), dark);
```

The method updates both system bar colors, icon contrast, and Android's
automatic contrast enforcement.

## Cancel interrupted touches

Attach `TouchCancellation` to the `EbitenView` and queue cancellation through a
small mobile binding:

```java
touches = TouchCancellation.attach(game, Mobile::cancelTouches);
```

Call `touches.cancel()` before suspending the game. Call `touches.close()` when
the Activity is destroyed. In Go, call `Page.Edit().CancelTouches()` from
`Mobile.cancelTouches`.

## Keep Back changes on the game loop

Android calls `Callbacks.onBackPressed` on the UI thread. Return `false` when
the current screen cannot handle Back. Queue accepted requests and apply page
changes from the app's `Page.SetTick` callback. Do not redraw the Go page from
the Android callback.
