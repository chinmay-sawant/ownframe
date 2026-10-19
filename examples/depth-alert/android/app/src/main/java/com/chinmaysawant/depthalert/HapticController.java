package com.chinmaysawant.depthalert;

import android.content.Context;
import android.os.Build;
import android.os.VibrationEffect;
import android.os.Vibrator;

final class HapticController {
    static final int UNKNOWN = 0;
    static final int CLEAR = 1;
    static final int STOP = 2;
    static final int STOP_NOW = 3;

    private final Vibrator vibrator;
    private int state = -1;

    HapticController(Context context) {
        vibrator = (Vibrator) context.getSystemService(Context.VIBRATOR_SERVICE);
    }

    void setState(int next) {
        if (next == state) return;
        state = next;
        if (vibrator == null || !vibrator.hasVibrator()) return;
        vibrator.cancel();
        long[] pattern;
        switch (next) {
            case CLEAR:
                pattern = new long[]{0, 65, 935};
                break;
            case STOP:
                pattern = new long[]{0, 80, 120, 80, 720};
                break;
            case STOP_NOW:
                pattern = new long[]{0, 600, 50};
                break;
            default:
                pattern = new long[]{0, 65, 110, 65, 110, 65, 1385};
                break;
        }
        if (Build.VERSION.SDK_INT >= 26) {
            vibrator.vibrate(VibrationEffect.createWaveform(pattern, 0));
        } else {
            vibrator.vibrate(pattern, 0);
        }
    }

    void stop() {
        state = -1;
        if (vibrator != null) vibrator.cancel();
    }
}
