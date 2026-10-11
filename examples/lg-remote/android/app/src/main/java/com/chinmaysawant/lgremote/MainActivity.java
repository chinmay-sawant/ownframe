package com.chinmaysawant.lgremote;

import android.Manifest;
import android.app.Activity;
import android.content.pm.PackageManager;
import android.net.wifi.WifiManager;
import android.view.View;
import android.view.WindowManager;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.ViewGroup;
import android.widget.FrameLayout;

import com.chinmaysawant.lgremote.mobile.EbitenView;
import com.chinmaysawant.lgremote.mobile.Mobile;
import com.chinmaysawant.lgremote.mobile.CommandListener;
import java.util.concurrent.atomic.AtomicBoolean;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;

// MainActivity fills the screen with the remote and polls Bluetooth commands.
public class MainActivity extends Activity {
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;

    private EbitenView view;
    private TouchCancellation touches;
    private AndroidHost host;
    private Hid hid;
    private RemoteAccessibility accessibility;
    private WifiManager.MulticastLock lock;
    private final Handler poller = new Handler(Looper.getMainLooper());
    private final AtomicBoolean dispatchQueued = new AtomicBoolean();
    private final CommandListener listener = new CommandListener() {
        @Override public void ready() {
            if (!dispatchQueued.compareAndSet(false, true)) { return; }
            poller.post(() -> {
                dispatchQueued.set(false);
                String cmd;
                while (hid != null && !(cmd = Mobile.takeCommand()).isEmpty()) { hid.command(cmd); }
            });
        }
    };
    private final Runnable poll = new Runnable() {
        @Override
        public void run() {
            if (accessibility != null) { accessibility.refresh(); }
            poller.postDelayed(this, 40);
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE);
        holdMulticast();

        // Load the Go runtime and set the key directory before the game loop
        // starts. The first frame reads that directory.
        Mobile.setStoreDir(getFilesDir().getAbsolutePath());

        FrameLayout layout = new FrameLayout(this);
        layout.setLayoutParams(new ViewGroup.LayoutParams(MATCH, MATCH));
        view = new EbitenView(this);
        touches = TouchCancellation.attach(view, Mobile::cancelTouches);
        layout.addView(view, new FrameLayout.LayoutParams(MATCH, MATCH));
        view.setImportantForAccessibility(View.IMPORTANT_FOR_ACCESSIBILITY_NO_HIDE_DESCENDANTS);
        accessibility = new RemoteAccessibility(this);
        layout.addView(accessibility, new FrameLayout.LayoutParams(MATCH, MATCH));
        setContentView(layout);
        host = AndroidHost.attach(this, layout,
            Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING)
                .baseTextSizeSp(16f).build(),
            new Callbacks() {
                @Override public void onConfigurationChanged(
                        dev.ownframe.android.DisplayInfo display) {
                    Mobile.setFontSize(Math.round(display.baseTextSizeDp));
                }
                @Override public boolean onBackPressed() { return handleBack(); }
                @Override public void onImeBackPressed() { Mobile.queueAction("blur:"); }
                @Override public void onResume() {
                    if (hid != null) hid.resume();
                    Mobile.setCommandListener(listener);
                    view.resumeGame();
                    poller.removeCallbacks(poll);
                    poller.post(poll);
                }
                @Override public void onPause() {
                    if (accessibility != null) accessibility.cancelTouch();
                    Mobile.setCommandListener(null);
                    if (hid != null) hid.pause();
                    poller.removeCallbacks(poll);
                    touches.cancel();
                    view.suspendGame();
                }
            });
        hid = new Hid(this);
    }

    @Override
    protected void onDestroy() {
        if (lock != null && lock.isHeld()) {
            lock.release();
        }
        Mobile.setCommandListener(null);
        poller.removeCallbacksAndMessages(null);
        if (hid != null) { hid.close(); }
        if (host != null) { host.close(); }
        super.onDestroy();
    }

    @Override
    public void onBackPressed() {
        if (host == null || !host.onBackPressed()) finish();
    }

    private boolean handleBack() {
        try {
            org.json.JSONObject state = new org.json.JSONObject(Mobile.accessibility());
            if (!"remote".equals(state.optString("Panel", "remote"))) {
                Mobile.queueAction("panel:remote");
                return true;
            }
        } catch (org.json.JSONException ex) { }
        return false;
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] results) {
        super.onRequestPermissionsResult(requestCode, permissions, results);
        if (requestCode != 1) { return; }
        boolean allowed = results.length == 2;
        for (int result : results) { allowed &= result == PackageManager.PERMISSION_GRANTED; }
        if (allowed && hid != null) {
            hid.command("start");
        } else {
            Mobile.setBluetooth("Bluetooth permission denied. Wi-Fi controls still work.");
        }
    }

    private void holdMulticast() {
        WifiManager wifi = (WifiManager) getApplicationContext()
            .getSystemService(WIFI_SERVICE);
        if (wifi == null) {
            return;
        }

        lock = wifi.createMulticastLock("lg-remote");
        lock.setReferenceCounted(false);
        lock.acquire();
    }

}
