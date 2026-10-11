package dev.ownframe.android;

import android.app.Activity;
import android.app.Application;
import android.content.ComponentCallbacks;
import android.content.res.Configuration;
import android.os.Build;
import android.os.Bundle;
import android.os.Looper;
import android.util.DisplayMetrics;
import android.util.TypedValue;
import android.view.View;
import android.view.Window;
import android.view.WindowInsets;
import android.view.inputmethod.InputMethodManager;
import android.window.OnBackInvokedCallback;
import android.window.OnBackInvokedDispatcher;

/** Edge-to-edge, safe-inset, text-scale, Back, and activity-lifecycle helper. */
public final class AndroidHost implements Application.ActivityLifecycleCallbacks, ComponentCallbacks {
    private final Activity activity;
    private final View content;
    private final Options options;
    private final Callbacks callbacks;
    private final int padLeft, padTop, padRight, padBottom;
    private final Application application;
    private final View.OnLayoutChangeListener layoutListener;
    private BackRegistration backRegistration;
    private Insets currentInsets;
    private DisplayInfo currentDisplay;
    private ViewportSnapshot currentViewport;
    private long viewportRevision;
    private boolean closed;

    private AndroidHost(Activity a, View v, Options o, Callbacks c) {
        activity = a;
        content = v;
        options = o;
        callbacks = c;
        padLeft = v.getPaddingLeft(); padTop = v.getPaddingTop();
        padRight = v.getPaddingRight(); padBottom = v.getPaddingBottom();
        application = a.getApplication();
        layoutListener = (view, left, top, right, bottom, oldLeft, oldTop, oldRight, oldBottom) ->
                publishViewport();
        v.addOnLayoutChangeListener(layoutListener);
        if (o.edgeToEdge) enableEdgeToEdge(a.getWindow());
        if (Build.VERSION.SDK_INT >= 33) backRegistration = BackRegistration.install(this, a);
        v.setOnApplyWindowInsetsListener((view, windowInsets) -> {
            applyInsets(windowInsets);
            return view.onApplyWindowInsets(windowInsets);
        });
        application.registerActivityLifecycleCallbacks(this);
        activity.registerComponentCallbacks(this);
        v.requestApplyInsets();
        currentInsets = new Insets(0, 0, 0, 0, 0, false);
        currentDisplay = displayInfo();
        callbacks.onConfigurationChanged(currentDisplay);
        publishViewport();
    }

    public static AndroidHost attach(Activity activity, View content,
                                     Options options, Callbacks callbacks) {
        if (activity == null || content == null || options == null || callbacks == null)
            throw new IllegalArgumentException("activity, content, options, and callbacks are required");
        if (Looper.myLooper() != Looper.getMainLooper())
            throw new IllegalStateException("attach must run on the main thread");
        return new AndroidHost(activity, content, options, callbacks);
    }

    /** Legacy API 28-32 Back entry point. Call super.onBackPressed() when false. */
    public boolean onBackPressed() {
        if (closed) return false;
        if (currentInsets != null && currentInsets.imeVisible) {
            callbacks.onImeBackPressed();
            hideIme();
            return true;
        }
        return callbacks.onBackPressed();
    }

    /** Updates both system bars to match the active page theme. Call on the main thread. */
    public void setSystemBarStyle(int color, boolean lightIcons) {
        if (closed) return;
        if (Looper.myLooper() != Looper.getMainLooper())
            throw new IllegalStateException("system bar style must be set on the main thread");
        Window window = activity.getWindow();
        window.setStatusBarColor(color);
        window.setNavigationBarColor(color);
        if (Build.VERSION.SDK_INT >= 29) {
            window.setStatusBarContrastEnforced(false);
            window.setNavigationBarContrastEnforced(false);
        }
        View decor = window.getDecorView();
        int flags = decor.getSystemUiVisibility();
        if (lightIcons) flags &= ~View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR;
        else flags |= View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR;
        if (Build.VERSION.SDK_INT >= 26) {
            if (lightIcons) flags &= ~View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR;
            else flags |= View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR;
        }
        decor.setSystemUiVisibility(flags);
    }

    public void close() {
        if (closed) return;
        closed = true;
        application.unregisterActivityLifecycleCallbacks(this);
        activity.unregisterComponentCallbacks(this);
        if (backRegistration != null) backRegistration.close();
        content.removeOnLayoutChangeListener(layoutListener);
        content.setOnApplyWindowInsetsListener(null);
        content.setPadding(padLeft, padTop, padRight, padBottom);
    }

    private void applyInsets(WindowInsets insets) {
        float density = density();
        int l, t, r, b, ime;
        boolean visible;
        if (Build.VERSION.SDK_INT >= 30) {
            android.graphics.Insets safe = insets.getInsets(
                    WindowInsets.Type.systemBars() | WindowInsets.Type.displayCutout());
            android.graphics.Insets keyboard = insets.getInsets(WindowInsets.Type.ime());
            l = safe.left; t = safe.top; r = safe.right; b = safe.bottom;
            ime = keyboard.bottom;
            visible = insets.isVisible(WindowInsets.Type.ime());
        } else {
            int cutoutLeft = 0, cutoutTop = 0, cutoutRight = 0, cutoutBottom = 0;
            if (Build.VERSION.SDK_INT >= 28 && insets.getDisplayCutout() != null) {
                cutoutLeft = insets.getDisplayCutout().getSafeInsetLeft();
                cutoutTop = insets.getDisplayCutout().getSafeInsetTop();
                cutoutRight = insets.getDisplayCutout().getSafeInsetRight();
                cutoutBottom = insets.getDisplayCutout().getSafeInsetBottom();
            }
            l = Math.max(insets.getStableInsetLeft(), cutoutLeft);
            t = Math.max(insets.getStableInsetTop(), cutoutTop);
            r = Math.max(insets.getStableInsetRight(), cutoutRight);
            b = Math.max(insets.getStableInsetBottom(), cutoutBottom);
            int systemBottom = insets.getSystemWindowInsetBottom();
            visible = systemBottom > b;
            ime = visible ? systemBottom : 0;
        }
        currentInsets = new Insets(l / density, t / density, r / density,
                b / density, ime / density, visible);
        if (options.insetMode == InsetMode.HOST_PADDING)
            content.setPadding(padLeft + l, padTop + t, padRight + r, padBottom + b);
        callbacks.onInsetsChanged(currentInsets);
        publishViewport();
    }

    private DisplayInfo displayInfo() {
        DisplayMetrics metrics = activity.getResources().getDisplayMetrics();
        float px = TypedValue.applyDimension(TypedValue.COMPLEX_UNIT_SP,
                options.baseTextSizeSp, metrics);
        return new DisplayInfo(metrics.density, activity.getResources().getConfiguration().fontScale,
                px / metrics.density);
    }

    private float density() {
        return activity.getResources().getDisplayMetrics().density;
    }

    private void hideIme() {
        if (Build.VERSION.SDK_INT >= 30) {
            if (activity.getWindow().getInsetsController() != null)
                activity.getWindow().getInsetsController().hide(WindowInsets.Type.ime());
        } else {
            InputMethodManager manager = (InputMethodManager)
                    activity.getSystemService(Activity.INPUT_METHOD_SERVICE);
            manager.hideSoftInputFromWindow(content.getWindowToken(), 0);
        }
    }

    private static void enableEdgeToEdge(Window window) {
        if (Build.VERSION.SDK_INT >= 30) {
            window.setDecorFitsSystemWindows(false);
        } else {
            View decor = window.getDecorView();
            decor.setSystemUiVisibility(decor.getSystemUiVisibility()
                    | View.SYSTEM_UI_FLAG_LAYOUT_STABLE | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION);
        }
    }

    @Override public void onActivityResumed(Activity a) { if (a == activity) callbacks.onResume(); }
    @Override public void onActivityPaused(Activity a) { if (a == activity) callbacks.onPause(); }
    @Override public void onConfigurationChanged(Configuration c) {
        if (!closed) {
            currentDisplay = displayInfo();
            callbacks.onConfigurationChanged(currentDisplay);
            publishViewport();
            content.requestApplyInsets();
        }
    }

    private void publishViewport() {
        if (closed || currentDisplay == null || currentInsets == null) return;
        float density = currentDisplay.density;
        ViewportSnapshot next = new ViewportSnapshot(content.getWidth() / density,
                content.getHeight() / density, currentInsets, currentDisplay, viewportRevision + 1);
        if (currentViewport != null && currentViewport.hasSameValues(next)) return;
        viewportRevision++;
        currentViewport = next;
        callbacks.onViewportChanged(next);
    }

    @Override public void onActivityCreated(Activity a, Bundle b) {}
    @Override public void onActivityDestroyed(Activity a) {}
    @Override public void onActivityStarted(Activity a) {}
    @Override public void onActivitySaveInstanceState(Activity a, Bundle b) {}
    @Override public void onActivityStopped(Activity a) {}
    @Override public void onLowMemory() {}

    private static final class BackRegistration {
        private final OnBackInvokedDispatcher dispatcher;
        private final OnBackInvokedCallback callback;

        private BackRegistration(Activity activity, AndroidHost host) {
            dispatcher = activity.getOnBackInvokedDispatcher();
            callback = () -> {
                if (!host.onBackPressed()) activity.finish();
            };
            dispatcher.registerOnBackInvokedCallback(OnBackInvokedDispatcher.PRIORITY_DEFAULT,
                    callback);
        }

        static BackRegistration install(AndroidHost host, Activity activity) {
            return new BackRegistration(activity, host);
        }

        void close() { dispatcher.unregisterOnBackInvokedCallback(callback); }
    }
}
