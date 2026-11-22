package com.chinmaysawant.login;

import android.app.Activity;
import android.os.Bundle;
import android.view.ViewGroup;
import android.widget.FrameLayout;

import dev.ownframe.android.AccessibilityBridge;
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
    private AccessibilityBridge a11y;

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
        // TalkBack reads the page through a virtual tree built from the
        // snapshot the Go side publishes.
        a11y = AccessibilityBridge.attach(view, new AccessibilityBridge.Source() {
            @Override public String tree() { return Mobile.accessibility(); }
            @Override public String perform(long document, String node, int action, String value) {
                return Mobile.accessibilityPerform(document, node, action, value);
            }
        });
    }

    @Override protected void onDestroy() {
        if (touches != null) touches.close();
        if (a11y != null) a11y.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
