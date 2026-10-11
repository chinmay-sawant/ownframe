package dev.ownframe.android;

import android.view.MotionEvent;
import android.view.View;

/** Clears native touches when Android cancels a gesture or suspends the Activity. */
public final class TouchCancellation {
    private final View view;
    private final Runnable notifyGame;
    private MotionEvent last;

    private TouchCancellation(View view, Runnable notifyGame) {
        this.view = view;
        this.notifyGame = notifyGame;
        view.setOnTouchListener((v, event) -> {
            if (event.getActionMasked() == MotionEvent.ACTION_CANCEL) {
                remember(event);
                cancel();
                return true;
            }
            remember(event);
            if (event.getActionMasked() == MotionEvent.ACTION_UP) clear();
            return false;
        });
    }

    public static TouchCancellation attach(View view, Runnable notifyGame) {
        return new TouchCancellation(view, notifyGame);
    }

    public void cancel() {
        // Notify first. The runtime suppresses input until native touches clear.
        notifyGame.run();
        if (last == null) return;
        // Ebiten v2.10.4 ignores ACTION_CANCEL. Send one indexed UP per pointer.
        for (int i = 0; i < last.getPointerCount(); i++) {
            MotionEvent.PointerProperties properties = new MotionEvent.PointerProperties();
            MotionEvent.PointerCoords coordinates = new MotionEvent.PointerCoords();
            last.getPointerProperties(i, properties);
            last.getPointerCoords(i, coordinates);
            MotionEvent up = MotionEvent.obtain(last.getDownTime(), last.getEventTime(),
                MotionEvent.ACTION_UP, 1, new MotionEvent.PointerProperties[]{properties},
                new MotionEvent.PointerCoords[]{coordinates}, last.getMetaState(),
                last.getButtonState(), last.getXPrecision(), last.getYPrecision(),
                last.getDeviceId(), last.getEdgeFlags(), last.getSource(), last.getFlags());
            view.onTouchEvent(up);
            up.recycle();
        }
        clear();
    }

    public void close() {
        cancel();
        view.setOnTouchListener(null);
    }

    private void remember(MotionEvent event) {
        clear();
        last = MotionEvent.obtain(event);
    }

    private void clear() {
        if (last != null) last.recycle();
        last = null;
    }
}
