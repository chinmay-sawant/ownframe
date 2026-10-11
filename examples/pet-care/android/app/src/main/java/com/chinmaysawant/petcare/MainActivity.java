package com.chinmaysawant.petcare;

import android.app.Activity;
import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.graphics.Color;
import android.view.ViewGroup;
import android.widget.FrameLayout;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;

import com.chinmaysawant.petcare.mobile.EbitenView;
import com.chinmaysawant.petcare.mobile.Mobile;
import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;

public final class MainActivity extends Activity {
    private static final int EXPORT = 1;
    private static final int RESTORE = 2;
    private EbitenView game;
    private AndroidHost host;
    private TouchCancellation touches;
    private Boolean darkBars;
    private final android.os.Handler handler = new android.os.Handler();
    private final Runnable poll = new Runnable() {
        @Override public void run() {
            String request = Mobile.pickerRequest();
            if ("export".equals(request)) exportBackup();
            if ("restore".equals(request)) chooseBackup();
            updateSystemBars();
            handler.postDelayed(this, 250);
        }
    };

    private void updateSystemBars() {
        boolean dark = Mobile.darkTheme();
        if (darkBars != null && darkBars == dark) return;
        darkBars = dark;
        int color = Color.parseColor(dark ? "#171c19" : "#f4f6f2");
        if (host != null) host.setSystemBarStyle(color, dark);
    }

    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        String startupError = Mobile.start(getFilesDir().getAbsolutePath());
        if (!startupError.isEmpty()) throw new IllegalStateException(startupError);
        FrameLayout root = new FrameLayout(this);
        game = new EbitenView(this);
        touches = TouchCancellation.attach(game, Mobile::cancelTouches);
        root.addView(game, new FrameLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
        setContentView(root);
        host = AndroidHost.attach(this, root,
            Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING).build(),
            new Callbacks() {
                @Override public boolean onBackPressed() { return Mobile.back(); }
                @Override public void onImeBackPressed() { }
                @Override public void onResume() { game.resumeGame(); handler.removeCallbacks(poll); handler.post(poll); }
                @Override public void onPause() {
                    handler.removeCallbacks(poll);
                    touches.cancel();
                    game.suspendGame();
                }
            });
    }

    private void exportBackup() {
        Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.setType("application/json");
        intent.putExtra(Intent.EXTRA_TITLE, "pet-care-backup.json");
        startActivityForResult(intent, EXPORT);
    }

    private void chooseBackup() {
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.setType("application/json");
        startActivityForResult(intent, RESTORE);
    }

    @Override protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (resultCode != RESULT_OK || data == null || data.getData() == null) return;
        if (requestCode == EXPORT) writeBackup(data.getData());
        if (requestCode == RESTORE) readBackup(data.getData());
    }

    private void writeBackup(Uri uri) {
        byte[] bytes = Mobile.backupData();
        if (bytes == null) return;
        try (OutputStream out = getContentResolver().openOutputStream(uri)) {
            if (out != null) out.write(bytes);
        } catch (Exception ignored) { }
    }

    private void readBackup(Uri uri) {
        try (InputStream in = getContentResolver().openInputStream(uri);
             ByteArrayOutputStream out = new ByteArrayOutputStream()) {
            if (in == null) return;
            byte[] chunk = new byte[8192];
            int count;
            while ((count = in.read(chunk)) != -1) {
                if (out.size() + count > 1 << 20) return;
                out.write(chunk, 0, count);
            }
            Mobile.setBackup(out.toByteArray());
        } catch (Exception ignored) { }
    }

    @Override public void onBackPressed() { if (host == null || !host.onBackPressed()) finish(); }

    @Override protected void onDestroy() {
        handler.removeCallbacks(poll);
        if (touches != null) touches.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
