package com.chinmaysawant.login;

import android.app.Activity;
import android.os.Bundle;
import android.view.ViewGroup;
import android.widget.FrameLayout;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;
import com.chinmaysawant.login.mobile.EbitenView;
import com.chinmaysawant.login.mobile.Mobile;

public class MainActivity extends Activity {
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;
    private EbitenView view;
    private TouchCancellation touches;
    private AndroidHost host;

    @Override public void onCreate(Bundle state) {
        super.onCreate(state);
        FrameLayout layout = new FrameLayout(this);
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
    }

    @Override protected void onDestroy() {
        if (touches != null) touches.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
