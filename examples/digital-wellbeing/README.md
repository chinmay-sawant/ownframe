# Digital Wellbeing with Mochi

Mochi opens to a seven-day overview of foreground time across apps. Use the
week arrows to browse earlier seven-day periods, tap a day bar for its app
breakdown, or tap the weekly total to see the most-used apps and daily average.
The **All apps** screen shows app icons and usage totals, plus separate weekday
and weekend totals. Selecting a day updates its total and the average through
that day. You can also focus the chart on one app. Android supplies the totals
through `UsageStatsManager`; Mochi reads them with Usage Access and does not
keep a separate usage history.

The picker shows launchable apps with activity in the last seven days, sorted
from highest foreground time to lowest. It passes app names to the picker,
then requests and displays daily totals for only the selected package. The
totals stay on the phone. The app has no network permission. Mochi's
illustration was created for this example.

Its Back screens share a centered-title header with a fixed Back control. See
the [Android app creation guide](../../documentation/android-app-creation.md#lay-out-mobile-screen-headers)
for this layout pattern and the narrow-screen checks.

## Build and install

From WSL2, set `ADB` to your Windows `platform-tools/adb.exe` path:

```sh
ADB=/mnt/c/Users/your-windows-user/platform-tools/adb.exe sh examples/digital-wellbeing/android/build.sh install
```

The script binds the Go page, assembles a debug APK, and installs it with
`adb install -r`. Set `ANDROID_HOME` if the SDK is not at `~/Android/Sdk`.
It defaults to `android/arm64`, suitable for the Pixel 7.

Open Digital Wellbeing, tap **Open Usage Access**, and enable access for this
app in Android Settings. Return to Mochi to see the overview. Tap **Choose an
app** to focus the chart on one app. Usage data refreshes every 30 seconds while
the activity is open.

The generated APK is at
`android/app/build/outputs/apk/debug/app-debug.apk`. Generated AARs and Gradle
outputs are local build artifacts.
