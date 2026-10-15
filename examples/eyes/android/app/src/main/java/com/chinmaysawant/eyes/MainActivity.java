package com.chinmaysawant.eyes;

import android.app.Activity;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.os.Bundle;
import android.os.SystemClock;
import android.view.Surface;
import android.view.ViewGroup;
import android.widget.FrameLayout;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;
import com.chinmaysawant.eyes.mobile.EbitenView;
import com.chinmaysawant.eyes.mobile.Mobile;

// MainActivity fills the screen with the bound ownframe eyes page.
public class MainActivity extends Activity implements SensorEventListener {
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;

    private EbitenView view;
    private TouchCancellation touches;
    private AndroidHost host;
    private SensorManager sensors;
    private Sensor accelerometer;
    private float lastX, lastY, lastZ;
    private float tiltCenterX, tiltCenterY;
    private float shakeEnergy;
    private long lastShakeAt;
    private boolean hasAcceleration;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        FrameLayout layout = new FrameLayout(this);
        layout.setLayoutParams(new ViewGroup.LayoutParams(MATCH, MATCH));

        view = new EbitenView(this);
        touches = TouchCancellation.attach(view, Mobile::cancelTouches);
        layout.addView(view, new FrameLayout.LayoutParams(MATCH, MATCH));
        setContentView(layout);
        host = AndroidHost.attach(this, layout,
            Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING).build(),
            new Callbacks() {
                @Override public void onResume() { view.resumeGame(); }
                @Override public void onPause() { touches.cancel();
                    view.suspendGame(); }
            });
        sensors = (SensorManager) getSystemService(SENSOR_SERVICE);
        if (sensors != null) {
            accelerometer = sensors.getDefaultSensor(Sensor.TYPE_ACCELEROMETER);
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        hasAcceleration = false;
        shakeEnergy = 0f;
        if (sensors != null && accelerometer != null) {
            sensors.registerListener(this, accelerometer, SensorManager.SENSOR_DELAY_GAME);
        }
    }

    @Override
    protected void onPause() {
        if (sensors != null) {
            sensors.unregisterListener(this);
        }
        super.onPause();
    }

    @Override
    public void onSensorChanged(SensorEvent event) {
        if (event == null || event.values == null || event.values.length < 3) return;
        float x = event.values[0];
        float y = event.values[1];
        float z = event.values[2];
        boolean first = !hasAcceleration;
        if (first) {
            lastX = tiltCenterX = x;
            lastY = tiltCenterY = y;
            lastZ = z;
            hasAcceleration = true;
        }
        float tiltX = (x - tiltCenterX) / SensorManager.GRAVITY_EARTH;
        float tiltY = (y - tiltCenterY) / SensorManager.GRAVITY_EARTH;
        // Sensor axes stay in the device's natural orientation. Undo the
        // display rotation before sending tilt in screen coordinates.
        int rotation = getWindowManager().getDefaultDisplay().getRotation();
        switch (rotation) {
            case Surface.ROTATION_90:
                Mobile.setMotion(tiltY, -tiltX);
                break;
            case Surface.ROTATION_180:
                Mobile.setMotion(-tiltX, -tiltY);
                break;
            case Surface.ROTATION_270:
                Mobile.setMotion(-tiltY, tiltX);
                break;
            default:
                Mobile.setMotion(tiltX, tiltY);
                break;
        }
        if (first) {
            return;
        }
        float delta = Math.abs(x - lastX) + Math.abs(y - lastY) + Math.abs(z - lastZ);
        lastX = x;
        lastY = y;
        lastZ = z;
        shakeEnergy = shakeEnergy * 0.45f + delta * 0.55f;
        long now = SystemClock.uptimeMillis();
        if (shakeEnergy > 4.5f && now - lastShakeAt > 500) {
            lastShakeAt = now;
            shakeEnergy = 0f;
            Mobile.shake();
        }
    }

    @Override
    public void onAccuracyChanged(Sensor sensor, int accuracy) {
    }

    @Override
    protected void onDestroy() {
        if (touches != null) touches.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
