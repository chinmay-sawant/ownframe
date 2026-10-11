package com.chinmaysawant.telegram;

import android.Manifest;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.graphics.Color;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.provider.MediaStore;
import android.view.View;
import android.view.ViewGroup;
import android.view.WindowInsetsController;
import android.widget.FrameLayout;

import com.chinmaysawant.telegram.mobile.EbitenView;
import com.chinmaysawant.telegram.mobile.Mobile;

import java.io.ByteArrayOutputStream;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;

// MainActivity fills the screen with the bound ownframe page and forwards
// Back, camera capture, and gallery selection to the page.
public class MainActivity extends Activity {
    private static final int REQ_PHOTO = 1;
    private static final int REQ_PERMISSION = 2;
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;

    private EbitenView view;
    private TouchCancellation touches;
    private FrameLayout layout;
    private AndroidHost host;
    private boolean waitingPermission;

    private final Handler poller = new Handler(Looper.getMainLooper());
    private final Runnable poll = new Runnable() {
        @Override
        public void run() {
            String attach = Mobile.attachRequest();
            if ("camera".equals(attach)) {
                openCamera();
            } else if ("gallery".equals(attach)) {
                openGallery();
            }
            setBarIcons(Mobile.darkTheme());
            poller.postDelayed(this, 250);
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().setStatusBarColor(Color.TRANSPARENT);
        getWindow().setNavigationBarColor(Color.TRANSPARENT);
        setContentView(buildLayout());
        host = AndroidHost.attach(this, findViewById(android.R.id.content),
            Options.builder().edgeToEdge(true).insetMode(InsetMode.APP_CALLBACK).build(),
            new Callbacks() {
                @Override public void onInsetsChanged(dev.ownframe.android.Insets insets) {
                    float density = getResources().getDisplayMetrics().density;
                    layout.setPadding(Math.round(insets.leftDp * density), 0,
                        Math.round(insets.rightDp * density), 0);
                    Mobile.setInsets(Math.round(insets.topDp),
                        Math.round(Math.max(insets.bottomDp, insets.imeBottomDp)));
                }
                @Override public void onImeBackPressed() { Mobile.blur(); }
                @Override public boolean onBackPressed() { return Mobile.back(); }
                @Override public void onResume() {
                    view.resumeGame();
                    poller.removeCallbacks(poll);
                    poller.post(poll);
                }
                @Override public void onPause() {
                    poller.removeCallbacks(poll);
                    touches.cancel();
                    view.suspendGame();
                }
            });
    }

    // buildLayout holds the Ebiten view below the shared host container.
    private FrameLayout buildLayout() {
        layout = new FrameLayout(this);
        layout.setLayoutParams(new ViewGroup.LayoutParams(MATCH, MATCH));

        view = new EbitenView(this);
        touches = TouchCancellation.attach(view, Mobile::cancelTouches);
        layout.addView(view, new FrameLayout.LayoutParams(MATCH, MATCH));

        return layout;
    }

    // setBarIcons keeps the status bar icons readable in both themes.
    private void setBarIcons(boolean dark) {
        if (Build.VERSION.SDK_INT < 30) {
            return;
        }
        WindowInsetsController c = getWindow().getInsetsController();
        if (c == null) {
            return;
        }
        int mask = WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS
            | WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS;
        c.setSystemBarsAppearance(dark ? 0 : mask, mask);
    }

    // onBackPressed shares IME-first and page navigation handling with the host.
    @Override
    public void onBackPressed() {
        if (host == null || !host.onBackPressed()) finish();
    }

    // openCamera asks for the camera permission once, then captures.
    private void openCamera() {
        if (Build.VERSION.SDK_INT >= 23
            && checkSelfPermission(Manifest.permission.CAMERA) != PackageManager.PERMISSION_GRANTED) {
            waitingPermission = true;
            requestPermissions(new String[]{Manifest.permission.CAMERA}, REQ_PERMISSION);
            return;
        }
        Intent intent = new Intent(MediaStore.ACTION_IMAGE_CAPTURE);
        if (intent.resolveActivity(getPackageManager()) != null) {
            startActivityForResult(intent, REQ_PHOTO);
        }
    }

    // openGallery opens the system photo picker on Android 13 and newer, and
    // the document picker on older releases.
    private void openGallery() {
        Intent intent;
        if (Build.VERSION.SDK_INT >= 33) {
            intent = new Intent(MediaStore.ACTION_PICK_IMAGES);
        } else {
            intent = new Intent(Intent.ACTION_GET_CONTENT);
            intent.addCategory(Intent.CATEGORY_OPENABLE);
        }
        intent.setType("image/*");
        if (intent.resolveActivity(getPackageManager()) == null) {
            intent = new Intent(Intent.ACTION_GET_CONTENT);
            intent.addCategory(Intent.CATEGORY_OPENABLE);
            intent.setType("image/*");
        }
        startActivityForResult(intent, REQ_PHOTO);
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] results) {
        super.onRequestPermissionsResult(requestCode, permissions, results);
        if (requestCode != REQ_PERMISSION) {
            return;
        }
        boolean granted = results.length > 0 && results[0] == PackageManager.PERMISSION_GRANTED;
        if (granted && waitingPermission) {
            openCamera();
        }
        waitingPermission = false;
    }

    // onActivityResult hands the captured thumbnail or the picked image to
    // the page.
    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode != REQ_PHOTO || resultCode != RESULT_OK || data == null) {
            return;
        }
        Bitmap bitmap = photo(data);
        if (bitmap == null) {
            return;
        }
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        bitmap.compress(Bitmap.CompressFormat.PNG, 100, out);
        Mobile.setPhoto(out.toByteArray());
    }

    // photo returns the image the gallery picker chose, or the thumbnail a
    // camera app attached to its result.
    private Bitmap photo(Intent data) {
        Uri uri = data.getData();
        if (uri != null) {
            try {
                try (java.io.InputStream stream = getContentResolver().openInputStream(uri)) {
                    return BitmapFactory.decodeStream(stream);
                }
            } catch (Exception e) {
                return null;
            }
        }
        if (data.getExtras() == null) {
            return null;
        }
        Object extra = data.getExtras().get("data");
        return extra instanceof Bitmap ? (Bitmap) extra : null;
    }

    @Override
    protected void onDestroy() {
        poller.removeCallbacks(poll);
        if (touches != null) touches.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
