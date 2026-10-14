package com.chinmaysawant.digitalwellbeing;

import android.app.Activity;
import android.app.AppOpsManager;
import android.app.usage.UsageStats;
import android.app.usage.UsageStatsManager;
import android.graphics.Bitmap;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.drawable.Drawable;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.content.pm.ResolveInfo;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.provider.Settings;
import android.util.Base64;
import android.view.ViewGroup;
import android.widget.FrameLayout;

import com.chinmaysawant.digitalwellbeing.mobile.EbitenView;
import com.chinmaysawant.digitalwellbeing.mobile.Mobile;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Date;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

import dev.ownframe.android.AndroidHost;
import dev.ownframe.android.Callbacks;
import dev.ownframe.android.InsetMode;
import dev.ownframe.android.Options;
import dev.ownframe.android.TouchCancellation;

// MainActivity hosts the OwnFrame page and reads Android's on-device usage API.
public class MainActivity extends Activity {
    private static final int MATCH = ViewGroup.LayoutParams.MATCH_PARENT;
    private static final String PREFS = "mochi";
    private static final String SELECTED = "selected_package";

    private EbitenView view;
    private TouchCancellation touches;
    private AndroidHost host;
    private boolean restored;
    private boolean previousAccess;
    private String lastSelected = "";
    private long lastAppsAt;
    private long lastUsageAt;
    private int lastWeekOffset = -1;
    private final Handler poller = new Handler(Looper.getMainLooper());
    private final Runnable poll = new Runnable() {
        @Override public void run() {
            if (Mobile.usageAccessRequested()) {
                startActivity(new Intent(Settings.ACTION_USAGE_ACCESS_SETTINGS));
            }
            boolean allowed = hasUsageAccess();
            Mobile.setUsageAccess(allowed);
            if (allowed) {
                if (!previousAccess) {
                    lastAppsAt = 0;
                    lastUsageAt = 0;
                }
                sendRestoreOnce();
                sendCatalogIfDue();
                refreshSelectionIfDue();
            }
            previousAccess = allowed;
            poller.postDelayed(this, 750);
        }
    };

    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        getWindow().setStatusBarColor(Color.rgb(255, 249, 242));
        getWindow().setNavigationBarColor(Color.rgb(255, 249, 242));
        FrameLayout layout = new FrameLayout(this);
        layout.setLayoutParams(new ViewGroup.LayoutParams(MATCH, MATCH));
        view = new EbitenView(this);
        touches = TouchCancellation.attach(view, Mobile::cancelTouches);
        layout.addView(view, new FrameLayout.LayoutParams(MATCH, MATCH));
        setContentView(layout);
        host = AndroidHost.attach(this, layout,
            Options.builder().edgeToEdge(true).insetMode(InsetMode.HOST_PADDING).build(),
            new Callbacks() {
                @Override public void onResume() {
                    view.resumeGame();
                    lastUsageAt = 0;
                    poller.removeCallbacks(poll);
                    poller.post(poll);
                }
                @Override public void onPause() {
                    poller.removeCallbacks(poll);
                    touches.cancel();
                    view.suspendGame();
                }
                @Override public boolean onBackPressed() { return Mobile.back(); }
            });
        host.setSystemBarStyle(Color.rgb(255, 249, 242), false);
    }

    private boolean hasUsageAccess() {
        AppOpsManager ops = (AppOpsManager) getSystemService(APP_OPS_SERVICE);
        if (ops == null) return false;
        try {
            return ops.checkOpNoThrow(AppOpsManager.OPSTR_GET_USAGE_STATS,
                android.os.Process.myUid(), getPackageName()) == AppOpsManager.MODE_ALLOWED;
        } catch (RuntimeException ignored) {
            return false;
        }
    }

    private void sendRestoreOnce() {
        if (restored) return;
        restored = true;
        String saved = getSharedPreferences(PREFS, MODE_PRIVATE).getString(SELECTED, "");
        if (!saved.isEmpty()) Mobile.restoreSelection(saved);
    }

    private void sendCatalogIfDue() {
        long now = System.currentTimeMillis();
        if (now - lastAppsAt < 60_000) return;
        Mobile.setAppsJSON(appCatalog());
        lastAppsAt = now;
    }

    private void refreshSelectionIfDue() {
        String selected = Mobile.selectedPackage();
        int weekOffset = (int) Mobile.weekOffset();
        if (!selected.equals(lastSelected)) {
            lastSelected = selected;
            getSharedPreferences(PREFS, MODE_PRIVATE).edit()
                .putString(SELECTED, selected).apply();
            lastUsageAt = 0;
        }
        if (weekOffset != lastWeekOffset) {
            lastWeekOffset = weekOffset;
            lastUsageAt = 0;
        }
        long now = System.currentTimeMillis();
        if (now - lastUsageAt >= 30_000) {
            Mobile.setUsageJSON(usageSnapshot(selected, weekOffset));
            lastUsageAt = now;
        }
    }

    private UsageStatsManager usageManager() {
        return (UsageStatsManager) getSystemService(USAGE_STATS_SERVICE);
    }

    private Calendar weekStart(int offset) {
        Calendar day = Calendar.getInstance();
        day.set(Calendar.HOUR_OF_DAY, 0);
        day.set(Calendar.MINUTE, 0);
        day.set(Calendar.SECOND, 0);
        day.set(Calendar.MILLISECOND, 0);
        day.add(Calendar.DAY_OF_YEAR, -6 - 7 * offset);
        return day;
    }

    private Map<String, Long> recentTotals(long from, long to) {
        Map<String, Long> totals = new HashMap<>();
        UsageStatsManager manager = usageManager();
        if (manager == null) return totals;
        List<UsageStats> rows = manager.queryUsageStats(
            UsageStatsManager.INTERVAL_DAILY, from, to);
        if (rows == null) return totals;
        for (UsageStats row : rows) {
            totals.merge(row.getPackageName(), row.getTotalTimeInForeground(), Long::sum);
        }
        return totals;
    }

    private String appCatalog() {
        try {
            Calendar start = weekStart(0);
            Map<String, Long> recent = recentTotals(start.getTimeInMillis(),
                System.currentTimeMillis());
            Intent launcher = new Intent(Intent.ACTION_MAIN);
            launcher.addCategory(Intent.CATEGORY_LAUNCHER);
            List<ResolveInfo> resolved = getPackageManager().queryIntentActivities(launcher, 0);
            Map<String, String> labels = new LinkedHashMap<>();
            for (ResolveInfo info : resolved) {
                String pkg = info.activityInfo.packageName;
                if (pkg.equals(getPackageName()) || recent.getOrDefault(pkg, 0L) == 0) continue;
                labels.put(pkg, info.loadLabel(getPackageManager()).toString());
            }
            List<Map.Entry<String, String>> sorted = new ArrayList<>(labels.entrySet());
            sorted.sort((left, right) -> {
                int byUsage = Long.compare(recent.getOrDefault(right.getKey(), 0L),
                    recent.getOrDefault(left.getKey(), 0L));
                if (byUsage != 0) return byUsage;
                return String.CASE_INSENSITIVE_ORDER.compare(left.getValue(), right.getValue());
            });
            JSONArray apps = new JSONArray();
            for (Map.Entry<String, String> entry : sorted) {
                String label = entry.getValue();
                JSONObject item = new JSONObject();
                item.put("package", entry.getKey());
                item.put("label", label);
                item.put("letter", label.isEmpty() ? "•" : label.substring(0, 1).toUpperCase(Locale.getDefault()));
                item.put("icon", iconData(getPackageManager().getApplicationIcon(entry.getKey())));
                apps.put(item);
            }
            return new JSONObject().put("apps", apps).toString();
        } catch (Exception ignored) {
            return "{\"apps\":[]}";
        }
    }

    private String iconData(Drawable icon) {
        Bitmap bitmap = Bitmap.createBitmap(64, 64, Bitmap.Config.ARGB_8888);
        Canvas canvas = new Canvas(bitmap);
        icon.setBounds(0, 0, 64, 64);
        icon.draw(canvas);
        ByteArrayOutputStream output = new ByteArrayOutputStream();
        bitmap.compress(Bitmap.CompressFormat.PNG, 100, output);
        bitmap.recycle();
        return Base64.encodeToString(output.toByteArray(), Base64.NO_WRAP);
    }

    private String usageSnapshot(String pkg, int offset) {
        try {
            String label = pkg.isEmpty() ? "All apps" : appLabel(pkg);
            JSONArray days = new JSONArray();
            Calendar day = weekStart(offset);
            SimpleDateFormat name = new SimpleDateFormat("EEE", Locale.getDefault());
            SimpleDateFormat dateName = new SimpleDateFormat("MMM d", Locale.getDefault());
            SimpleDateFormat rangeName = new SimpleDateFormat("MMM d", Locale.getDefault());
            Map<String, Long> weekly = new HashMap<>();
            Map<String, String> labels = launcherLabels();
            long allMillis = 0;
            long weekdayMillis = 0;
            long weekendMillis = 0;
            for (int i = 0; i < 7; i++) {
                long from = day.getTimeInMillis();
                Calendar next = (Calendar) day.clone();
                next.add(Calendar.DAY_OF_YEAR, 1);
                long to = Math.min(next.getTimeInMillis(), System.currentTimeMillis());
                Map<String, Long> daily = recentTotals(from, to);
                long selectedMillis = 0;
                for (Map.Entry<String, Long> row : daily.entrySet()) {
                    String rowPackage = row.getKey();
                    if (!labels.containsKey(rowPackage)) continue;
                    long millis = row.getValue();
                    allMillis += millis;
                    if (day.get(Calendar.DAY_OF_WEEK) == Calendar.SATURDAY
                        || day.get(Calendar.DAY_OF_WEEK) == Calendar.SUNDAY) {
                        weekendMillis += millis;
                    } else {
                        weekdayMillis += millis;
                    }
                    weekly.merge(rowPackage, millis, Long::sum);
                    if (pkg.isEmpty() || pkg.equals(rowPackage)) selectedMillis += millis;
                }
                Map<String, Long> visibleDaily = new HashMap<>();
                for (Map.Entry<String, Long> row : daily.entrySet()) {
                    if (labels.containsKey(row.getKey())) visibleDaily.put(row.getKey(), row.getValue());
                }
                days.put(new JSONObject().put("label", name.format(new Date(from)))
                    .put("date", dateName.format(new Date(from)))
                    .put("minutes", (int) (selectedMillis / 60_000))
                    .put("apps", usageRows(visibleDaily, labels)));
                day.add(Calendar.DAY_OF_YEAR, 1);
            }
            Calendar end = (Calendar) day.clone();
            end.add(Calendar.DAY_OF_YEAR, -1);
            String weekRange = rangeName.format(weekStart(offset).getTime()) + " – "
                + rangeName.format(end.getTime());
            return new JSONObject().put("package", pkg).put("label", label)
                .put("weekOffset", offset).put("weekRange", weekRange)
                .put("allMinutes", (int) (allMillis / 60_000)).put("days", days)
                .put("weekdayMinutes", (int) (weekdayMillis / 60_000))
                .put("weekendMinutes", (int) (weekendMillis / 60_000))
                .put("topApps", usageRows(weekly, labels)).toString();
        } catch (Exception ignored) {
            return "{}";
        }
    }

    private String appLabel(String pkg) {
        try {
            return getPackageManager().getApplicationLabel(
                getPackageManager().getApplicationInfo(pkg, 0)).toString();
        } catch (Exception ignored) {
            return pkg;
        }
    }

    private Map<String, String> launcherLabels() {
        Intent launcher = new Intent(Intent.ACTION_MAIN);
        launcher.addCategory(Intent.CATEGORY_LAUNCHER);
        List<ResolveInfo> resolved = getPackageManager().queryIntentActivities(launcher, 0);
        Map<String, String> labels = new HashMap<>();
        for (ResolveInfo info : resolved) {
            labels.put(info.activityInfo.packageName,
                info.loadLabel(getPackageManager()).toString());
        }
        labels.remove(getPackageName());
        return labels;
    }

    private JSONArray usageRows(Map<String, Long> totals, Map<String, String> labels)
        throws Exception {
        List<Map.Entry<String, Long>> sorted = new ArrayList<>(totals.entrySet());
        sorted.sort((left, right) -> Long.compare(right.getValue(), left.getValue()));
        JSONArray rows = new JSONArray();
        for (int i = 0; i < sorted.size(); i++) {
            Map.Entry<String, Long> item = sorted.get(i);
            rows.put(new JSONObject().put("package", item.getKey())
                .put("label", labels.getOrDefault(item.getKey(), item.getKey()))
                .put("minutes", (int) (item.getValue() / 60_000)));
        }
        return rows;
    }

    @Override public void onBackPressed() {
        if (host == null || !host.onBackPressed()) finish();
    }

    @Override protected void onDestroy() {
        poller.removeCallbacks(poll);
        if (touches != null) touches.close();
        if (host != null) host.close();
        super.onDestroy();
    }
}
