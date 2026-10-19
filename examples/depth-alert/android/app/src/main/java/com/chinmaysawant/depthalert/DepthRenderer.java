package com.chinmaysawant.depthalert;

import android.content.Context;
import android.opengl.GLES20;
import android.opengl.GLSurfaceView;
import android.os.SystemClock;
import android.view.Display;
import android.view.Surface;
import android.view.WindowManager;

import com.google.ar.core.Frame;
import com.google.ar.core.Session;
import com.google.ar.core.TrackingState;
import com.google.ar.core.exceptions.CameraNotAvailableException;
import com.google.ar.core.exceptions.NotYetAvailableException;

import android.media.Image;
import javax.microedition.khronos.egl.EGLConfig;
import javax.microedition.khronos.opengles.GL10;

final class DepthRenderer implements GLSurfaceView.Renderer {
    interface Listener {
        void onDepth(int nearMM, int coveragePercent, boolean tracking);
        void onEngineError(String message);
    }

    private final Object sessionLock = new Object();
    private final Display display;
    private final Listener listener;
    private final CameraBackground background = new CameraBackground();
    private volatile Session session;
    private Session textureSession;
    private int cameraTexture;
    private int width;
    private int height;
    private int rotation = -1;
    private long lastDepthAttemptMs;
    private long lastErrorMs;

    DepthRenderer(Context context, Listener listener) {
        WindowManager manager = (WindowManager) context.getSystemService(Context.WINDOW_SERVICE);
        display = manager.getDefaultDisplay();
        this.listener = listener;
    }

    void setSession(Session next) {
        synchronized (sessionLock) {
            session = next;
            textureSession = null;
            lastDepthAttemptMs = 0;
            lastErrorMs = 0;
        }
    }

    @Override public void onSurfaceCreated(GL10 gl, EGLConfig config) {
        synchronized (sessionLock) {
            GLES20.glClearColor(0, 0, 0, 1);
            cameraTexture = background.create();
            textureSession = null;
        }
    }

    @Override public void onSurfaceChanged(GL10 gl, int width, int height) {
        this.width = width;
        this.height = height;
        GLES20.glViewport(0, 0, width, height);
        rotation = -1;
    }

    @Override public void onDrawFrame(GL10 gl) {
        GLES20.glClear(GLES20.GL_COLOR_BUFFER_BIT | GLES20.GL_DEPTH_BUFFER_BIT);
        synchronized (sessionLock) {
            Session current = session;
            if (current == null || width == 0 || height == 0) return;
            try {
                if (textureSession != current) {
                    current.setCameraTextureNames(new int[]{cameraTexture});
                    textureSession = current;
                }
                int displayRotation = display.getRotation();
                if (rotation != displayRotation) {
                    current.setDisplayGeometry(displayRotation, width, height);
                    rotation = displayRotation;
                }
                Frame frame = current.update();
                background.draw(frame);
                long now = SystemClock.elapsedRealtime();
                if (now - lastDepthAttemptMs < 100) return;
                lastDepthAttemptMs = now;
                if (frame.getCamera().getTrackingState() != TrackingState.TRACKING) {
                    listener.onDepth(0, 0, false);
                    return;
                }
                readDepth(frame);
            } catch (CameraNotAvailableException e) {
                reportError("Camera unavailable. Stop and start scanning again.");
            } catch (RuntimeException e) {
                reportError("Depth session stopped. Stop and start scanning again.");
            }
        }
    }

    private void readDepth(Frame frame) {
        Image depth = null;
        try {
            depth = frame.acquireDepthImage16Bits();
            DepthFrameAnalyzer.Summary summary = DepthFrameAnalyzer.center(depth);
            listener.onDepth(summary.nearestMillimeters, summary.coveragePercent, true);
        } catch (NotYetAvailableException e) {
            listener.onDepth(0, 0, true);
        } finally {
            close(depth);
        }
    }

    private void reportError(String message) {
        long now = SystemClock.elapsedRealtime();
        if (now - lastErrorMs > 1500) {
            lastErrorMs = now;
            listener.onEngineError(message);
        }
    }

    private static void close(Image image) {
        if (image != null) image.close();
    }
}
