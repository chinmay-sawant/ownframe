package com.chinmaysawant.depthalert;

import android.Manifest;
import android.app.Activity;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.os.SystemClock;
import android.view.Gravity;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;
import android.opengl.GLSurfaceView;

import com.chinmaysawant.depthalert.mobile.EbitenView;
import com.chinmaysawant.depthalert.mobile.Mobile;
import com.google.ar.core.ArCoreApk;
import com.google.ar.core.Config;
import com.google.ar.core.Session;
import com.google.ar.core.exceptions.CameraNotAvailableException;
import com.google.ar.core.exceptions.UnavailableException;

import java.util.concurrent.atomic.AtomicLong;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;

public final class MainActivity extends Activity implements DepthRenderer.Listener {
    private static final int REQUEST_CAMERA = 1;
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;
    private static final int WRAP = ViewGroup.LayoutParams.WRAP_CONTENT;
    private static final int UNKNOWN = HapticController.UNKNOWN;

    private final Handler main = new Handler(Looper.getMainLooper());
    private final AtomicLong lastDepthAt = new AtomicLong();
    private boolean installRequested;
    private volatile boolean scanning;
    private Session session;
    private boolean sessionResumed;
    private AndroidHost host;
    private TouchCancellation touches;
    private GLSurfaceView depthSurface;
    private EbitenView pageView;
    private DepthRenderer renderer;
    private HapticController haptics;
    private Button startButton;
    private Button stopButton;
    private TextView status;
    private int alertState = -1;

    private final Runnable watchdog = new Runnable() {
        @Override public void run() {
            if (!scanning) return;
            if (SystemClock.elapsedRealtime() - lastDepthAt.get() > 500) {
                applyObservation(Mobile.depthUnavailable(),
                        "No fresh depth frame. Move the phone slowly and keep using your cane or guide dog.");
            }
            main.postDelayed(this, 350);
        }
    };

    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        haptics = new HapticController(this);
        setContentView(buildLayout());
        host = AndroidHost.attach(this, findViewById(android.R.id.content),
                Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING).build(),
                new Callbacks() {
                    @Override public void onResume() {
                        pageView.resumeGame();
                        depthSurface.onResume();
                    }
                    @Override public void onPause() {
                        stopScanning();
                        depthSurface.onPause();
                        touches.cancel();
                        pageView.suspendGame();
                    }
                });
    }

    private View buildLayout() {
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setBackgroundColor(Color.rgb(17, 25, 20));
        int pad = dp(12);

        TextView title = text("Depth alert · experimental", 20, Color.WHITE);
        title.setPadding(pad, dp(8), pad, dp(4));
        root.addView(title, new LinearLayout.LayoutParams(MATCH, WRAP));
        TextView instruction = text("Aim forward and move the phone slowly. Keep using your cane or guide dog.",
                14, Color.rgb(220, 232, 224));
        instruction.setPadding(pad, 0, pad, dp(8));
        root.addView(instruction, new LinearLayout.LayoutParams(MATCH, WRAP));

        depthSurface = new GLSurfaceView(this);
        depthSurface.setEGLContextClientVersion(2);
        depthSurface.setRenderer(renderer = new DepthRenderer(this, this));
        depthSurface.setRenderMode(GLSurfaceView.RENDERMODE_CONTINUOUSLY);
        depthSurface.setContentDescription("Live rear camera view. Aim toward the path ahead.");
        root.addView(depthSurface, new LinearLayout.LayoutParams(MATCH, 0, 5));

        status = text("Not scanning. No haptic signal means the app is not guiding movement.",
                16, Color.WHITE);
        status.setPadding(pad, dp(10), pad, dp(10));
        status.setMinHeight(dp(64));
        status.setGravity(Gravity.CENTER_VERTICAL);
        status.setAccessibilityLiveRegion(View.ACCESSIBILITY_LIVE_REGION_POLITE);
        root.addView(status, new LinearLayout.LayoutParams(MATCH, WRAP));

        LinearLayout controls = new LinearLayout(this);
        controls.setOrientation(LinearLayout.HORIZONTAL);
        startButton = button("Start depth alerts");
        stopButton = button("Stop depth alerts");
        startButton.setOnClickListener(view -> requestStart());
        stopButton.setOnClickListener(view -> stopScanning());
        controls.addView(startButton, new LinearLayout.LayoutParams(0, dp(56), 1));
        controls.addView(stopButton, new LinearLayout.LayoutParams(0, dp(56), 1));
        root.addView(controls, new LinearLayout.LayoutParams(MATCH, WRAP));

        pageView = new EbitenView(this);
        pageView.setImportantForAccessibility(View.IMPORTANT_FOR_ACCESSIBILITY_NO_HIDE_DESCENDANTS);
        touches = TouchCancellation.attach(pageView, Mobile::cancelTouches);
        root.addView(pageView, new LinearLayout.LayoutParams(MATCH, 0, 4));
        return root;
    }

    private void requestStart() {
        if (scanning) return;
        if (checkSelfPermission(Manifest.permission.CAMERA) != PackageManager.PERMISSION_GRANTED) {
            requestPermissions(new String[]{Manifest.permission.CAMERA}, REQUEST_CAMERA);
            return;
        }
        startScanning();
    }

    private void startScanning() {
        try {
            ArCoreApk.InstallStatus install = ArCoreApk.getInstance()
                    .requestInstall(this, !installRequested);
            if (install != ArCoreApk.InstallStatus.INSTALLED) {
                installRequested = true;
                showMessage("ARCore setup started. Tap Start depth alerts again when setup finishes.");
                return;
            }
            Session next = session;
            if (next == null) {
                next = new Session(this);
                if (!next.isDepthModeSupported(Config.DepthMode.AUTOMATIC)) {
                    next.close();
                    showMessage("This device does not support ARCore depth.");
                    return;
                }
                Config config = new Config(next);
                config.setDepthMode(Config.DepthMode.AUTOMATIC);
                config.setFocusMode(Config.FocusMode.AUTO);
                next.configure(config);
                session = next;
            }
            next.resume();
            sessionResumed = true;
            renderer.setSession(next);
            scanning = true;
            Mobile.setScanning(true);
            lastDepthAt.set(SystemClock.elapsedRealtime());
            alertState = -1;
            applyObservation(UNKNOWN,
                    "Depth is warming up. Move the phone slowly across a textured, well-lit view.");
            startButton.setEnabled(false);
            main.removeCallbacks(watchdog);
            main.postDelayed(watchdog, 350);
        } catch (UnavailableException | CameraNotAvailableException e) {
            showMessage("ARCore could not start: " + e.getMessage());
        }
    }

    private void stopScanning() {
        main.removeCallbacks(watchdog);
        if (renderer != null) renderer.setSession(null);
        Session current = session;
        scanning = false;
        if (current != null && sessionResumed) {
            try { current.pause(); } catch (RuntimeException ignored) {}
            sessionResumed = false;
        }
        Mobile.setScanning(false);
        Mobile.depthUnavailable();
        if (haptics != null) haptics.stop();
        alertState = -1;
        if (startButton != null) startButton.setEnabled(true);
        if (status != null) status.setText("Scanning paused. No haptic signal means the app is not guiding movement.");
    }

    @Override public void onDepth(int nearMM, int coveragePercent, boolean tracking) {
        if (!scanning) return;
        if (tracking && coveragePercent >= 15 && nearMM > 0) {
            lastDepthAt.set(SystemClock.elapsedRealtime());
        }
        int next = Mobile.observeDepth(nearMM, coveragePercent, tracking);
        String reason = unknownReason(nearMM, coveragePercent, tracking);
        main.post(() -> {
            if (scanning) applyObservation(next, reason);
        });
    }

    @Override public void onEngineError(String message) {
        main.post(() -> {
            if (scanning) {
                stopScanning();
                showMessage(message);
            }
        });
    }

    private void applyObservation(int next, String reason) {
        if (!scanning) return;
        if (next != alertState) {
            alertState = next;
            haptics.setState(next);
        }
        String message;
        switch (next) {
            case HapticController.CLEAR:
                message = "Walk. No close surface detected in the center view. One pulse each second.";
                break;
            case HapticController.STOP:
                message = "Stop. A surface was detected within 3 meters. Two pulses each second.";
                break;
            case HapticController.STOP_NOW:
                message = "Stop now. A surface was detected within 1.2 meters. Continuous vibration.";
                break;
            default:
                message = reason;
                break;
        }
        if (!message.contentEquals(status.getText())) status.setText(message);
    }

    private String unknownReason(int nearMM, int coveragePercent, boolean tracking) {
        if (!tracking) {
            return "AR tracking is unavailable. Move the phone slowly across a textured view.";
        }
        if (coveragePercent < 15 || nearMM <= 0) {
            return "Depth is warming up or sparse here. Move gently toward a textured, well-lit view.";
        }
        return "Depth has not confirmed a clear view. Move slowly; keep using your cane or guide dog.";
    }

    private void showMessage(String message) {
        if (status != null) status.setText(message);
    }

    @Override public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] results) {
        super.onRequestPermissionsResult(requestCode, permissions, results);
        if (requestCode == REQUEST_CAMERA) {
            if (results.length > 0 && results[0] == PackageManager.PERMISSION_GRANTED) startScanning();
            else showMessage("Camera access is needed for depth alerts. Scanning is off.");
        }
    }

    @Override public void onBackPressed() {
        if (host == null || !host.onBackPressed()) finish();
    }

    @Override protected void onDestroy() {
        stopScanning();
        if (session != null) {
            session.close();
            session = null;
        }
        if (touches != null) touches.close();
        if (host != null) host.close();
        if (haptics != null) haptics.stop();
        super.onDestroy();
    }

    private Button button(String label) {
        Button result = new Button(this);
        result.setText(label);
        result.setAllCaps(false);
        result.setContentDescription(label);
        return result;
    }

    private TextView text(String label, int sizeSp, int color) {
        TextView result = new TextView(this);
        result.setText(label);
        result.setTextSize(sizeSp);
        result.setTextColor(color);
        return result;
    }

    private int dp(int value) {
        return Math.round(value * getResources().getDisplayMetrics().density);
    }
}
