package dev.ownframe.android;

import android.graphics.Rect;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.os.Build;
import android.view.View;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityManager;
import android.view.accessibility.AccessibilityNodeInfo;
import android.view.accessibility.AccessibilityNodeProvider;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Answers TalkBack for an ownframe page drawn into one view.
 *
 * <p>The page is a single GPU surface, so TalkBack sees nothing in it without
 * help. This class gives the view a virtual tree built from the snapshot the Go
 * side publishes (package androidaccess), reports focus and content changes,
 * and turns TalkBack's tap, focus, and text actions into requests the page runs
 * on its own loop.
 *
 * <p>Everything that touches views runs on the main thread. Actions run on one
 * worker thread, because the page may take a moment to answer.
 */
public final class AccessibilityBridge {
    /** The Go side. Implement it with the bound package's two functions. */
    public interface Source {
        /** The latest accessibility tree as JSON. Cheap, and safe on any thread. */
        String tree();

        /** Runs one action: 1 focus, 2 activate, 3 set text. "" on success. */
        String perform(long document, String node, int action, String value);
    }

    private static final int ACTION_FOCUS = 1;
    private static final int ACTION_ACTIVATE = 2;
    private static final int ACTION_SET_VALUE = 3;
    private static final long POLL_MS = 250;
    private static final long REFRESH_AFTER_ACTION_MS = 120;

    private final View view;
    private final Source source;
    private final float density;
    private final Handler main = new Handler(Looper.getMainLooper());
    private final ExecutorService actions = Executors.newSingleThreadExecutor();
    private final AccessibilityManager manager;

    private Tree tree = new Tree();
    private String lastJson = "";
    private int accessibilityFocusedVid = View.NO_ID;
    private boolean closed;

    private final Runnable poll = new Runnable() {
        @Override
        public void run() {
            if (closed) return;
            if (manager != null && manager.isEnabled()) refresh();
            main.postDelayed(this, POLL_MS);
        }
    };

    private AccessibilityBridge(View view, Source source) {
        this.view = view;
        this.source = source;
        this.density = view.getResources().getDisplayMetrics().density;
        this.manager = (AccessibilityManager) view.getContext().getSystemService(
                android.content.Context.ACCESSIBILITY_SERVICE);
    }

    /** Attaches the provider to view. Call it on the main thread. */
    public static AccessibilityBridge attach(View view, Source source) {
        if (view == null || source == null) throw new IllegalArgumentException("view and source are required");
        if (Looper.myLooper() != Looper.getMainLooper()) {
            throw new IllegalStateException("attach must run on the main thread");
        }
        AccessibilityBridge bridge = new AccessibilityBridge(view, source);
        view.setImportantForAccessibility(View.IMPORTANT_FOR_ACCESSIBILITY_YES);
        view.setAccessibilityDelegate(bridge.new Delegate());
        bridge.main.post(bridge.poll);
        return bridge;
    }

    /** Stops polling and releases the worker. Call it when the activity goes. */
    public void close() {
        closed = true;
        main.removeCallbacks(poll);
        actions.shutdownNow();
        view.setAccessibilityDelegate(null);
    }

    // ---- the tree ----------------------------------------------------------

    private static final class Node {
        String id, parent, role, name, desc, value;
        boolean hasBounds, disabled, focusable, password, hasChecked, checked;
        double x, y, w, h;
        final List<String> children = new ArrayList<>();
        final List<Integer> actions = new ArrayList<>();
        int vid;
    }

    private static final class Tree {
        long doc, rev;
        String root = "", focus = "";
        final List<Node> nodes = new ArrayList<>();
        final Map<String, Node> byId = new HashMap<>();
        final Map<Integer, Node> byVid = new HashMap<>();
    }

    private static Tree parse(String json) {
        Tree t = new Tree();
        try {
            JSONObject o = new JSONObject(json);
            t.doc = o.optLong("doc");
            t.rev = o.optLong("rev");
            t.root = o.optString("root", "");
            t.focus = o.optString("focus", "");
            JSONArray arr = o.optJSONArray("nodes");
            if (arr == null) return t;
            for (int i = 0; i < arr.length(); i++) {
                JSONObject j = arr.getJSONObject(i);
                Node n = new Node();
                n.vid = i;
                n.id = j.optString("id");
                n.parent = j.optString("parent", "");
                n.role = j.optString("role");
                n.name = j.optString("name", "");
                n.desc = j.optString("desc", "");
                n.value = j.optString("value", "");
                n.x = j.optDouble("x");
                n.y = j.optDouble("y");
                n.w = j.optDouble("w");
                n.h = j.optDouble("h");
                n.hasBounds = j.optBoolean("bounds");
                n.disabled = j.optBoolean("disabled");
                n.focusable = j.optBoolean("focusable");
                n.password = j.optBoolean("password");
                n.hasChecked = j.optBoolean("hasChecked");
                n.checked = j.optBoolean("checked");
                JSONArray kids = j.optJSONArray("children");
                if (kids != null) for (int k = 0; k < kids.length(); k++) n.children.add(kids.getString(k));
                JSONArray acts = j.optJSONArray("actions");
                if (acts != null) for (int k = 0; k < acts.length(); k++) n.actions.add(acts.getInt(k));
                t.nodes.add(n);
                t.byId.put(n.id, n);
                t.byVid.put(n.vid, n);
            }
        } catch (JSONException e) {
            return new Tree();
        }
        return t;
    }

    /**
     * Loads the latest tree if it changed and returns the tree it replaced, or
     * null when nothing changed. Android may ask about a node before anything
     * has polled (a service that just connected, for one), so every query
     * loads first. Main thread.
     */
    private Tree load() {
        String json = source.tree();
        if (json == null || json.equals(lastJson)) return null;
        lastJson = json;
        Tree prev = tree;
        tree = parse(json);
        if (accessibilityFocusedVid != View.NO_ID && !tree.byVid.containsKey(accessibilityFocusedVid)) {
            accessibilityFocusedVid = View.NO_ID;
        }
        return prev;
    }

    /** Loads the latest tree and tells TalkBack what changed. Main thread. */
    private void refresh() {
        Tree prev = load();
        if (prev == null) return;
        sendContentChanged(AccessibilityEvent.CONTENT_CHANGE_TYPE_SUBTREE);
        if (!tree.focus.isEmpty() && !tree.focus.equals(prev.focus)) {
            Node n = tree.byId.get(tree.focus);
            if (n != null) sendVirtualEvent(AccessibilityEvent.TYPE_VIEW_FOCUSED, n.vid);
        }
    }

    private void sendContentChanged(int changeTypes) {
        if (manager == null || !manager.isEnabled()) return;
        AccessibilityEvent e = AccessibilityEvent.obtain(AccessibilityEvent.TYPE_WINDOW_CONTENT_CHANGED);
        e.setContentChangeTypes(changeTypes);
        e.setSource(view);
        e.setPackageName(view.getContext().getPackageName());
        if (view.getParent() != null) view.getParent().requestSendAccessibilityEvent(view, e);
    }

    private void sendVirtualEvent(int type, int vid) {
        if (manager == null || !manager.isEnabled()) return;
        AccessibilityEvent e = AccessibilityEvent.obtain(type);
        e.setSource(view, vid);
        e.setPackageName(view.getContext().getPackageName());
        if (view.getParent() != null) view.getParent().requestSendAccessibilityEvent(view, e);
    }

    // ---- the provider ------------------------------------------------------

    private final class Delegate extends View.AccessibilityDelegate {
        private final AccessibilityNodeProvider provider = new Provider();

        @Override
        public AccessibilityNodeProvider getAccessibilityNodeProvider(View host) {
            return provider;
        }

        @Override
        public void onInitializeAccessibilityNodeInfo(View host, AccessibilityNodeInfo info) {
            super.onInitializeAccessibilityNodeInfo(host, info);
            load();
            addTopLevelChildren(host, info);
        }
    }

    private void addTopLevelChildren(View host, AccessibilityNodeInfo info) {
        for (Node n : tree.nodes) {
            if (n.parent.isEmpty()) {
                info.addChild(host, n.vid);
            }
        }
    }

    private final class Provider extends AccessibilityNodeProvider {
        @Override
        public AccessibilityNodeInfo createAccessibilityNodeInfo(int virtualViewId) {
            load();
            if (virtualViewId == View.NO_ID) {
                AccessibilityNodeInfo info = AccessibilityNodeInfo.obtain(view);
                view.onInitializeAccessibilityNodeInfo(info);
                return info;
            }
            Node n = tree.byVid.get(virtualViewId);
            if (n == null) return null;
            return describe(n);
        }

        @Override
        public boolean performAction(int virtualViewId, int action, Bundle arguments) {
            if (virtualViewId == View.NO_ID) return view.performAccessibilityAction(action, arguments);
            Node n = tree.byVid.get(virtualViewId);
            if (n == null) return false;
            switch (action) {
                case AccessibilityNodeInfo.ACTION_ACCESSIBILITY_FOCUS:
                    if (accessibilityFocusedVid == n.vid) return false;
                    accessibilityFocusedVid = n.vid;
                    sendVirtualEvent(AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUSED, n.vid);
                    return true;
                case AccessibilityNodeInfo.ACTION_CLEAR_ACCESSIBILITY_FOCUS:
                    if (accessibilityFocusedVid != n.vid) return false;
                    accessibilityFocusedVid = View.NO_ID;
                    sendVirtualEvent(AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUS_CLEARED, n.vid);
                    return true;
                case AccessibilityNodeInfo.ACTION_CLICK:
                    if (!n.actions.contains(ACTION_ACTIVATE) || n.disabled) return false;
                    run(n, ACTION_ACTIVATE, "");
                    return true;
                case AccessibilityNodeInfo.ACTION_FOCUS:
                    if (!n.focusable) return false;
                    run(n, ACTION_FOCUS, "");
                    return true;
                case AccessibilityNodeInfo.ACTION_SET_TEXT: {
                    if (!n.actions.contains(ACTION_SET_VALUE) || n.disabled || n.password) return false;
                    CharSequence text = arguments == null ? null
                            : arguments.getCharSequence(AccessibilityNodeInfo.ACTION_ARGUMENT_SET_TEXT_CHARSEQUENCE);
                    run(n, ACTION_SET_VALUE, text == null ? "" : text.toString());
                    return true;
                }
                default:
                    return false;
            }
        }

        @Override
        public AccessibilityNodeInfo findFocus(int focus) {
            if (focus == AccessibilityNodeInfo.FOCUS_ACCESSIBILITY) {
                Node n = tree.byVid.get(accessibilityFocusedVid);
                return n == null ? null : describe(n);
            }
            Node n = tree.byId.get(tree.focus);
            return n == null ? null : describe(n);
        }
    }

    /** Sends one action to the page and refreshes shortly after. */
    private void run(final Node n, final int action, final String value) {
        final long doc = tree.doc;
        final String id = n.id;
        actions.execute(new Runnable() {
            @Override
            public void run() {
                source.perform(doc, id, action, value);
                main.postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        if (!closed) refresh();
                    }
                }, REFRESH_AFTER_ACTION_MS);
            }
        });
    }

    private AccessibilityNodeInfo describe(Node n) {
        AccessibilityNodeInfo info = AccessibilityNodeInfo.obtain(view, n.vid);
        info.setPackageName(view.getContext().getPackageName());
        info.setClassName(className(n));
        info.setParent(view, parentVid(n));
        for (String child : n.children) {
            Node c = tree.byId.get(child);
            if (c != null) info.addChild(view, c.vid);
        }

        Rect inParent = resolvedBounds(n);
        info.setBoundsInParent(inParent);
        int[] origin = new int[2];
        view.getLocationOnScreen(origin);
        Rect onScreen = new Rect(inParent);
        onScreen.offset(origin[0], origin[1]);
        info.setBoundsInScreen(onScreen);
        info.setVisibleToUser(!inParent.isEmpty() && Rect.intersects(inParent,
                new Rect(0, 0, view.getWidth(), view.getHeight())));

        // A node with no geometry of its own (the document, a wrapper, a label
        // whose control already carries its name) would cover its parent and
        // read twice, so it stays out of swipe navigation. Its children remain.
        if (Build.VERSION.SDK_INT >= 24) {
            info.setImportantForAccessibility(n.hasBounds || n.focusable);
        }
        info.setEnabled(!n.disabled);
        info.setFocusable(n.focusable);
        info.setFocused(n.id.equals(tree.focus));
        info.setAccessibilityFocused(n.vid == accessibilityFocusedVid);
        info.setPassword(n.password);
        if (n.hasChecked) {
            info.setCheckable(true);
            info.setChecked(n.checked);
        }

        // The accessible name: a control reads its name and, for a text field,
        // its value; plain text reads itself.
        boolean field = "textbox".equals(n.role) || "combobox".equals(n.role);
        if (field) {
            info.setContentDescription(n.name.isEmpty() ? null : n.name);
            if (!n.password) info.setText(n.value);
            info.setEditable(!n.disabled && n.actions.contains(ACTION_SET_VALUE));
            if (Build.VERSION.SDK_INT >= 26 && !n.name.isEmpty()) info.setHintText(n.name);
        } else {
            info.setText(n.name.isEmpty() ? n.value : n.name);
            if (!n.desc.isEmpty()) info.setContentDescription(n.name.isEmpty() ? n.desc : n.name + ", " + n.desc);
        }
        if ("heading".equals(n.role) && Build.VERSION.SDK_INT >= 28) info.setHeading(true);

        if (n.actions.contains(ACTION_ACTIVATE) && !n.disabled) {
            info.setClickable(true);
            info.addAction(AccessibilityNodeInfo.AccessibilityAction.ACTION_CLICK);
        }
        if (n.focusable && !n.disabled) info.addAction(AccessibilityNodeInfo.AccessibilityAction.ACTION_FOCUS);
        if (field && !n.disabled && !n.password && n.actions.contains(ACTION_SET_VALUE)) {
            info.addAction(AccessibilityNodeInfo.AccessibilityAction.ACTION_SET_TEXT);
        }
        info.addAction(n.vid == accessibilityFocusedVid
                ? AccessibilityNodeInfo.AccessibilityAction.ACTION_CLEAR_ACCESSIBILITY_FOCUS
                : AccessibilityNodeInfo.AccessibilityAction.ACTION_ACCESSIBILITY_FOCUS);
        return info;
    }

    private int parentVid(Node n) {
        if (n.parent.isEmpty()) return View.NO_ID;
        Node p = tree.byId.get(n.parent);
        return p == null ? View.NO_ID : p.vid;
    }

    /**
     * The node's rectangle in view pixels. A node with no geometry of its own,
     * such as the document or a wrapper, takes its parent's, and the top of
     * the tree takes the whole view, so Android never treats it as invisible
     * and drops its subtree.
     */
    private Rect resolvedBounds(Node n) {
        if (n.hasBounds) return boundsInParent(n);
        Node p = n.parent.isEmpty() ? null : tree.byId.get(n.parent);
        if (p == null) return new Rect(0, 0, view.getWidth(), view.getHeight());
        return resolvedBounds(p);
    }

    /** Page CSS pixels are density-independent pixels; the view works in pixels. */
    private Rect boundsInParent(Node n) {
        if (!n.hasBounds) return new Rect();
        int left = Math.round((float) (n.x * density));
        int top = Math.round((float) (n.y * density));
        int right = Math.round((float) ((n.x + n.w) * density));
        int bottom = Math.round((float) ((n.y + n.h) * density));
        return new Rect(left, top, right, bottom);
    }

    private static String className(Node n) {
        switch (n.role) {
            case "button": return "android.widget.Button";
            case "link": return "android.widget.Button";
            case "textbox": return "android.widget.EditText";
            case "combobox": return "android.widget.Spinner";
            case "checkbox": return "android.widget.CheckBox";
            case "radio": return "android.widget.RadioButton";
            case "image": return "android.widget.ImageView";
            case "list": return "android.widget.ListView";
            default: return "android.widget.TextView";
        }
    }
}
